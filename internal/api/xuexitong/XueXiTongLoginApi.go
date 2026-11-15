package xuexitong

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yatori-dev/yatori-go-core/v3/utils"
)

// 注意Api类文件主需要写最原始的接口请求和最后的json的string形式返回，不需要用结构体序列化。
// 序列化和具体的功能实现请移步到Action代码文件中
const (
	KEY             = "u2oh6Vu^HWe4_AES" // 注意 Go 语言中字符串默认就是 UTF-8 编码
	APP_VERSION     = "6.7.5"
	DEVICE_VENDOR   = "MI10"
	BUILD           = "10941_314" //这个对应6.7.5版本
	ANDROID_VERSION = "Android 11"
)

// xxtDeviceConfigPath 设备指纹持久化路径: 优先用户配置目录，避免污染工作目录
func xxtDeviceConfigPath() string {
	if dir, err := os.UserConfigDir(); err == nil && dir != "" {
		return filepath.Join(dir, "yatori", "xxt_device.json")
	}
	return filepath.Join(".", "config", "xxt_device.json")
}

// loadOrCreateIMEI 固定设备指纹(32位hex): 首次运行生成后持久化。
// 每次启动随机生成会导致同一账号多会话"设备漂移"，容易被风控判定为设备异常。
func loadOrCreateIMEI() string {
	filePath := xxtDeviceConfigPath()
	if data, err := ioutil.ReadFile(filePath); err == nil {
		var payload struct {
			IMEI string `json:"imei"`
		}
		if json.Unmarshal(data, &payload) == nil {
			value := strings.ToLower(strings.TrimSpace(payload.IMEI))
			if len(value) == 32 {
				valid := true
				for _, c := range value {
					if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
						valid = false
						break
					}
				}
				if valid {
					return value
				}
			}
		}
	}
	value := utils.TokenHex(16)
	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err == nil {
		_ = ioutil.WriteFile(filePath, []byte(fmt.Sprintf(`{"imei":%q}`, value)), 0644)
	}
	return value
}

var IMEI = loadOrCreateIMEI()

type XueXiTUserCache struct {
	Name     string //用户使用Phone
	Password string //用户密码

	UserID      string // 用户ID
	JsonContent map[string]interface{}
	cookies     []*http.Cookie
	cookie      string //验证码用的session
	IpProxySW   bool   // 是否开启代理
	ProxyIP     string //代理IP

	// FaceSign 是可选的课程人脸签名（新版 faceNeedSignature 课程需要）。
	// 由学习通 App 的 CLIENT_FORM_SIGN 生成，可通过 cmd/facesign 采集。
	FaceSign *FaceSign
	// FaceSignChapter 是章节级人脸签名（knowledge/startface 页面，type=4）。
	// 章节被人脸锁住时（卡页返回 knowledge/startface）必须提供。
	FaceSignChapter *FaceSign
	// FaceSc / FaceKeyPEM 用于纯本地生成签名（不再依赖 App 抓包）。
	FaceSc     string
	FaceKeyPEM []byte
	// ClientMode is "web" or "app". Web login/2FA use WebUserAgent; the
	// app-only study APIs always use the account's Android identity.
	ClientMode string
	// UserAgent is the Android/API UA bound to this account. Study APIs
	// (gas/clazz, knowledge/cards?isPhone=1) reject web UAs with 403.
	UserAgent string
	// WebUserAgent is the browser UA used by web login and 2FA endpoints.
	WebUserAgent string
	// LoginCodeProvider 是登录过程中图片验证码/短信验证码的回调。
	// 返回用户输入的验证码；nil 时由 SDK 内置 OCR 处理图片验证码。
	LoginCodeProvider LoginCodeProvider
}

// LoginCodeChallenge 描述一次登录验证码挑战。
//
// 图片验证码由 SDK 内置 ddddocr 持续识别重试，不会回调到这里；目前会回调
// LoginCodeKindSMS（短信验证码）和 LoginCodeKindPassword（二次密码）。
// 短信挑战里 SendSMS 由调用方在用户确认后触发，SDK 不会自动发送。
type LoginCodeChallenge struct {
	Kind      string
	Prompt    string
	Phone     string
	Image     []byte
	SendSMS   func(ctx context.Context) error
	Attempt   int
	LastError string
}

// LoginCodeProvider 返回用户输入/识别出的验证码。
type LoginCodeProvider func(ctx context.Context, challenge LoginCodeChallenge) (string, error)

func (cache *XueXiTUserCache) GetCookie() string {
	return cache.cookie
}
func (cache *XueXiTUserCache) GetCookies() []*http.Cookie        { return cache.cookies }
func (cache *XueXiTUserCache) SetCookies(cookies []*http.Cookie) { cache.cookies = cookies }

// SetClientMode 固定本次会话的身份。
//
// 学习通的网页登录/二次验证使用浏览器 UA；但课程、章节卡片、视频文档提交
// 这些接口只认 Android App UA（用桌面或移动浏览器 UA 会 403），所以这些
// 接口统一使用该账号持久化的 Android 设备身份，避免每次请求随机切换 UA。
func (cache *XueXiTUserCache) SetClientMode(mode string) {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "app":
		cache.ClientMode = "app"
	default:
		cache.ClientMode = "web"
	}
	cache.WebUserAgent = webLoginUserAgent
	cache.ensureAPIUserAgent()
}

// IsAppMode reports whether this session uses the Android/App identity.
func (cache *XueXiTUserCache) IsAppMode() bool {
	return cache.ClientMode == "app"
}

// UA returns the Android/API UA bound to this account. It is stable for the
// whole session and shared by web and App login modes because the study APIs
// require an App UA.
func (cache *XueXiTUserCache) UA() string {
	return cache.ensureAPIUserAgent()
}

// WebUA returns the browser UA used by web login and 2FA endpoints.
func (cache *XueXiTUserCache) WebUA() string {
	if ua := strings.TrimSpace(cache.WebUserAgent); ua != "" {
		return ua
	}
	return webLoginUserAgent
}

func (cache *XueXiTUserCache) ensureAPIUserAgent() string {
	if ua := strings.TrimSpace(cache.UserAgent); ua != "" {
		return ua
	}
	if name := strings.TrimSpace(cache.Name); name != "" {
		profile := loadOrCreateAppDeviceProfile(name)
		cache.UserAgent = profile.UserAgent
		return cache.UserAgent
	}
	cache.UserAgent = GetUA("mobile")
	return cache.UserAgent
}

// GetUA 构建并获取 UA
func GetUA(uaType string) string {
	switch uaType {
	case "mobile":
		//return strings.Join([]string{
		//	fmt.Sprintf("Dalvik/2.1.0 (Linux; U; %s; %s Build/SKQ1.210216.001)", ANDROID_VERSION, DEVICE_VENDOR),
		//	fmt.Sprintf("(schild:%s)", MobileUASign(DEVICE_VENDOR, "zh_CN", APP_VERSION, BUILD, IMEI)),
		//	fmt.Sprintf("(device:%s)", DEVICE_VENDOR),
		//	"Language/zh_CN",
		//	fmt.Sprintf("com.chaoxing.mobile/ChaoXingStudy_3_%s_android_phone_%s", APP_VERSION, BUILD),
		//	//APP_VERSION,
		//	fmt.Sprintf("(@Kalimdor)_%s", IMEI),
		//}, " ")

		return strings.Join([]string{
			fmt.Sprintf("Mozilla/5.0 (Linux; %s; %s Build/QKQ1.190929.002; wv)", ANDROID_VERSION, DEVICE_VENDOR),
			fmt.Sprintf("AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/79.0.3945.116 Mobile Safari/537.36"),
			fmt.Sprintf("(schild:%s)", MobileUASign(DEVICE_VENDOR, "zh_CN", APP_VERSION, BUILD, IMEI)),
			fmt.Sprintf("(device:%s)", DEVICE_VENDOR),
			"Language/zh_CN",
			fmt.Sprintf("com.chaoxing.mobile/ChaoXingStudy_3_%s_android_phone_%s", APP_VERSION, BUILD),
			fmt.Sprintf("(@Kalimdor)_%s", IMEI),
		}, " ")
	case "iphone":

		return "Mozilla/5.0 (iPhone; CPU iPhone OS 18_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.5 Mobile/15E148 Safari/604.1"

	case "web":
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36 Edg/107.0.1418.35"
	default:
		return ""
	}
}

func MobileUASign(model, locale, version, build, imei string) string {
	// 拼接字符串，和 Python 的 " ".join 相同
	str := strings.Join([]string{
		"(schild:ipL$TkeiEmfy1gTXb2XHrdLN0a@7c^vu)",
		fmt.Sprintf("(device:%s)", model),
		fmt.Sprintf("Language/%s", locale),
		fmt.Sprintf("com.chaoxing.mobile/ChaoXingStudy_3_%s_android_phone_%s", version, build),
		fmt.Sprintf("(@Kalimdor)_%s", imei),
	}, " ")

	// 计算 md5
	hash := md5.Sum([]byte(str))

	// 转换为小写 hex 字符串
	return fmt.Sprintf("%x", hash)
}

// pad 确保数据长度是块大小的整数倍，以便符合块加密算法的要求
func pad(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padText...)
}

// LoginApi 登录Api
func (cache *XueXiTUserCache) LoginApi(retry int) (string, error) {
	key := []byte(KEY)
	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println("Error creating cipher:", err)
		return "", err
	}

	// 加密电话号码
	phonePadded := pad([]byte(cache.Name), block.BlockSize())
	phoneCipherText := make([]byte, len(phonePadded))
	mode := cipher.NewCBCEncrypter(block, key)
	mode.CryptBlocks(phoneCipherText, phonePadded)
	phoneEncrypted := base64.StdEncoding.EncodeToString(phoneCipherText)

	// 加密密码
	passwdPadded := pad([]byte(cache.Password), block.BlockSize())
	passwdCipherText := make([]byte, len(passwdPadded))
	mode = cipher.NewCBCEncrypter(block, key)
	mode.CryptBlocks(passwdCipherText, passwdPadded)
	passwdEncrypted := base64.StdEncoding.EncodeToString(passwdCipherText)
	//设置代理
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{},
	}
	//如果开启了IP代理，那么就直接添加代理
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse("http://" + cache.ProxyIP) // 设置代理
		}
	}
	client := &http.Client{
		Transport: tr,
	}
	// 发送请求
	form := url.Values{
		"fid":               {"-1"},
		"uname":             {phoneEncrypted},
		"password":          {passwdEncrypted},
		"refer":             {"http%3A%2F%2Fi.mooc.chaoxing.com"},
		"t":                 {"true"},
		"forbidotherlogin":  {"0"},
		"validate":          {""},
		"doubleFactorLogin": {"0"},
		"independentId":     {"0"},
		"independentNameId": {"0"},
	}
	req, reqErr := http.NewRequest(http.MethodPost, "https://passport2.chaoxing.com/fanyalogin", strings.NewReader(form.Encode()))
	if reqErr != nil {
		return "", reqErr
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", cache.WebUA())
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Origin", "https://passport2.chaoxing.com")
	req.Header.Set("Referer", "https://passport2.chaoxing.com/")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return "", err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	//触发风控直接重新登录
	if strings.Contains(string(body), "很抱歉，您所浏览的页面暂时不能访问！") {
		time.Sleep(time.Duration(retry*5) * time.Second)
		return cache.LoginApi(retry - 1)
	}

	//var jsonContent map[string]interface{}
	//err = json.Unmarshal(body, &jsonContent)
	//if err != nil {
	//	fmt.Println("Error parsing JSON:", err)
	//	return "", err
	//}

	//if status, ok := jsonContent["status"].(bool); !ok || !status {
	//	return "", errors.New(string(body))
	//}
	values := resp.Header.Values("Set-Cookie")
	for _, v := range values {
		cache.cookie += strings.ReplaceAll(strings.ReplaceAll(v, "HttpOnly", ""), "Path=/", "")
		//if strings.Contains(v, "UUID=") {
		//
		//}
	}
	utils.CookiesAddNoRepetition(&cache.cookies, resp.Cookies()) //赋值cookie

	//cache.JsonContent = jsonContent
	return string(body), nil
}

// VerificationCodeApi 获取验证码和SESSION验证码,并返回文件路径和SESSION字符串
func (cache *XueXiTUserCache) VerificationCodeApi() (string, string) {
	//TODO 待完成
	return "", ""
}

// 学习通不知道用来干嘛的接口，但是每隔一段时间都会发送一次
func (cache *XueXiTUserCache) MonitorApi() (string, error) {
	fid := ""
	for _, cookie := range cache.cookies {
		if cookie.Name == "fid" {
			fid = cookie.Value
		}
	}
	urlStr := fmt.Sprintf("https://detect.chaoxing.com/api/monitor?version=%s&refer=%s&from=&fid=%s&jsoncallback=jsonp%s&t=%d", "1748956725820", "http%%253A%252F%252Fi.mooc.chaoxing.com", fid, generate17DigitNumber(), time.Now().UnixMilli())
	method := "GET"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{},
	}

	//如果开启了IP代理，那么就直接添加代理
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse("http://" + cache.ProxyIP) // 设置代理
		}
	}
	client := &http.Client{
		Transport: tr,
	}
	req, err := http.NewRequest(method, urlStr, nil)

	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	//req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36 Edg/136.0.0.0")
	req.Header.Add("User-Agent", cache.UA())
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "detect.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	fmt.Println(string(body))
	utils.CookiesAddNoRepetition(&cache.cookies, res.Cookies()) //赋值cookie
	return string(body), nil
}

// 随机生成17位置数字，带前导0
func generate17DigitNumber() string {
	rand.Seed(time.Now().UnixNano())

	result := ""
	for i := 0; i < 17; i++ {
		digit := rand.Intn(10) // 生成0-9的随机数字
		result += fmt.Sprintf("%d", digit)
	}
	return result
}
