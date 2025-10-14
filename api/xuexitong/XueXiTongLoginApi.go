package xuexitong

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/yatori-dev/yatori-go-core/utils"
)

// 注意Api类文件主需要写最原始的接口请求和最后的json的string形式返回，不需要用结构体序列化。
// 序列化和具体的功能实现请移步到Action代码文件中
const (
	ApiLoginWeb = "https://passport2.chaoxing.com/fanyalogin"

	ApiPullCourses = "https://mooc1-api.chaoxing.com/mycourse/backclazzdata"

	// ApiChapterPoint 接口-课程章节任务点状态
	ApiChapterPoint = "https://mooc1-api.chaoxing.com/job/myjobsnodesmap"
	ApiChapterCards = "https://mooc1-api.chaoxing.com/gas/knowledge"
	ApiPullChapter  = "https://mooc1-api.chaoxing.com/gas/clazz"

	// PageMobileChapterCard SSR页面-客户端章节任务卡片
	PageMobileChapterCard = "https://mooc1-api.chaoxing.com/knowledge/cards"

	// APIChapterCardResource 接口-课程章节卡片资源
	APIChapterCardResource = "https://mooc1-api.chaoxing.com/ananas/status"
	// APIVideoPlayReport 接口-视频播放上报
	APIVideoPlayReport  = "https://mooc1.chaoxing.com/mooc-ans/multimedia/log/a"
	APIVideoPlayReport2 = "https://mooc1-api.chaoxing.com/multimedia/log/a" // cxkitty的

	// ApiWorkCommit 接口-单元作业答题提交
	ApiWorkCommit = "https://mooc1-api.chaoxing.com/work/addStudentWorkNew"
	// ApiWorkCommitNew 接口-新的作业提交答案接口
	ApiWorkCommitNew = "https://mooc1.chaoxing.com/mooc-ans/work/addStudentWorkNew"

	// 接口-课程文档阅读上报
	ApiDocumentReadingReport = "https://mooc1.chaoxing.com/ananas/job/document"

	// PageMobileWork SSR页面-客户端单元测验答题页
	PageMobileWork  = "https://mooc1-api.chaoxing.com/android/mworkspecial"           // 这是个cxkitty中的
	PageMobileWorkY = "https://mooc1-api.chaoxing.com/mooc-ans/work/phone/doHomeWork" // 这个是自己爬的

	KEY             = "u2oh6Vu^HWe4_AES" // 注意 Go 语言中字符串默认就是 UTF-8 编码
	APP_VERSION     = "6.4.5"
	DEVICE_VENDOR   = "MI10"
	BUILD           = "10831_263"
	ANDROID_VERSION = "Android 9"
)

var IMEI = utils.TokenHex(16)

type XueXiTUserCache struct {
	Name     string //用户使用Phone
	Password string //用户密码

	UserID      string // 用户ID
	JsonContent map[string]interface{}
	cookies     []*http.Cookie
	cookie      string //验证码用的session
	IpProxySW   bool   // 是否开启代理
	ProxyIP     string //代理IP
}

func (cache *XueXiTUserCache) GetCookie() string {
	return cache.cookie
}
func (cache *XueXiTUserCache) GetCookies() []*http.Cookie        { return cache.cookies }
func (cache *XueXiTUserCache) SetCookies(cookies []*http.Cookie) { cache.cookies = cookies }

// GetUA 构建并获取 UA
func GetUA(uaType string) string {
	switch uaType {
	case "mobile":
		return strings.Join([]string{
			fmt.Sprintf("Dalvik/2.1.0 (Linux; U; %s; %s Build/SKQ1.210216.001)", ANDROID_VERSION, DEVICE_VENDOR),
			fmt.Sprintf("(schild:%s)", MobileUASign(DEVICE_VENDOR, "zh_CN", APP_VERSION, BUILD, IMEI)),
			fmt.Sprintf("(device:%s)", DEVICE_VENDOR),
			"Language/zh_CN",
			fmt.Sprintf("com.chaoxing.mobile/ChaoXingStudy_3_%s_android_phone_%s", APP_VERSION, BUILD),
			//APP_VERSION,
			fmt.Sprintf("(@Kalimdor)_%s", IMEI),
		}, " ")
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
func (cache *XueXiTUserCache) LoginApi() (string, error) {
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
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	//如果开启了IP代理，那么就直接添加代理
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse(cache.ProxyIP) // 设置代理
		}
	}
	client := &http.Client{
		Transport: tr,
	}
	// 发送请求
	resp, err := client.PostForm(ApiLoginWeb, url.Values{
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
	})
	if err != nil {
		fmt.Println("Error sending request:", err)
		return "", err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var jsonContent map[string]interface{}
	err = json.Unmarshal(body, &jsonContent)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return "", err
	}

	if status, ok := jsonContent["status"].(bool); !ok || !status {
		return "", errors.New(string(body))
	}
	values := resp.Header.Values("Set-Cookie")
	for _, v := range values {
		cache.cookie += strings.ReplaceAll(strings.ReplaceAll(v, "HttpOnly", ""), "Path=/", "")
		//if strings.Contains(v, "UUID=") {
		//
		//}
	}
	utils.CookiesAddNoRepetition(&cache.cookies, resp.Cookies()) //赋值cookie

	cache.JsonContent = jsonContent
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
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}

	//如果开启了IP代理，那么就直接添加代理
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse(cache.ProxyIP) // 设置代理
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
	req.Header.Add("User-Agent", GetUA("mobile"))
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
