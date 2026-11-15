package xuexitong

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"math/big"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/yatori-dev/yatori-go-core/v3/core"
	log2 "github.com/yatori-dev/yatori-go-core/v3/utils/log"
)

const (
	appLoginAESKey    = "z4ok6lu^oWp4_AES"
	appLoginUserAgent = "Mozilla/5.0 (Linux; Android 11; M2102J2SC Build/RKQ1.200826.002; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/87.0.4280.141 Mobile Safari/537.36 (schild:0386d381aec6f8797cca82e4b2dff099) (device:M2102J2SC) Language/zh_CN com.chaoxing.mobile/ChaoXingStudy_3_7.0.4_android_phone_10995_342 (@Kalimdor)_4f38c61e5ec74b159f07437c0f426f6b"
	appFacePublicKey  = "MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC79d8Ot0hCbxxSISC6x8SCwTBspFSzlLKHJUYqoFNu1TSRaw4hEYkOnvEaL1VyoxV6HXcDrzwYvaFZaZaPQPFnfCHZy5dQwxcmifgSHqS+oKXw40Ys4cVIqnU5d90S7EWSRdBglX489jlqVaNcQSkDx2TYmC+DbAq9FV/BU09ISQIDAQAB"
)

// AppLoginResult 是学习通 App 登录后拿到的设备信息。
type AppLoginResult struct {
	ClientID string
	Cxcid    string
	Sc       string
	PUID     string
	UID      string
	Name     string
	Cookies  []*http.Cookie
}

var (
	facePublicKeyOnce sync.Once
	facePublicKey     *rsa.PublicKey
	facePublicKeyErr  error
)

// AppLoginApi 走学习通 App 的登录链路：
//
//	logininfo(AES/ECB) -> v11/loginregister -> 本地 deviceInfo（按账号持久化）
//	-> userLogin4Uname.do -> clientId -> 本地 RSA 解密出 {cid, sc}
//
// 这样 cxcid/sc 不再依赖 App 抓包或手动配置。
func (cache *XueXiTUserCache) AppLoginApi(ctx context.Context) (*AppLoginResult, error) {
	username := strings.TrimSpace(cache.Name)
	password := cache.Password
	if username == "" || password == "" {
		return nil, errors.New("App 登录需要账号和密码")
	}
	// 同一账号复用同一套设备指纹；不同账号的型号、标识和 UA 令牌都不同。
	profile := loadOrCreateAppDeviceProfile(username)
	cache.ClientMode = "app"
	cache.UserAgent = profile.UserAgent
	userAgent := strings.TrimSpace(profile.UserAgent)
	if userAgent == "" {
		userAgent = appLoginUserAgent
	}
	loginInfo, err := buildAppLoginInfo(username, password)
	if err != nil {
		return nil, err
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	transport := &http.Transport{ForceAttemptHTTP2: false}
	if cache.IpProxySW {
		if proxyURL, parseErr := url.Parse("http://" + cache.ProxyIP); parseErr == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}
	client := &http.Client{Jar: jar, Transport: transport, Timeout: 30 * time.Second}

	form := url.Values{}
	form.Set("logininfo", loginInfo)
	form.Set("cfid", "-1")
	form.Set("countrycode", "86")
	form.Set("loginType", "1")
	form.Set("roleSelect", "true")
	form.Set("entype", "1")
	req1, reqErr := http.NewRequest("POST", "https://passport2-api.chaoxing.com/v11/loginregister", strings.NewReader(form.Encode()))
	if reqErr != nil {
		return nil, reqErr
	}
	req1.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req1.Header.Set("User-Agent", userAgent)
	req1.Header.Set("Accept", "*/*")
	req1.Header.Set("X-Requested-With", "XMLHttpRequest")
	step1, err := client.Do(req1)
	if err != nil {
		return nil, fmt.Errorf("App 登录第一步失败: %w", err)
	}
	step1Body, _ := readAll(step1)
	step1.Body.Close()
	var first struct {
		Status bool   `json:"status"`
		URL    string `json:"url"`
		WebURL string `json:"webUrl"`
		Mes    string `json:"mes"`
	}
	if err := json.Unmarshal(step1Body, &first); err != nil {
		return nil, fmt.Errorf("App 登录第一步响应异常: %s", appBodySummary(step1Body))
	}
	var retryBody []byte
	if first.URL == "" && first.WebURL != "" {
		// type=2：服务端要求短信/消息验证。拉起 checkmessage 页面，
		// 解析隐藏参数，通过 LoginCodeProvider 让调用方输入短信验证码，
		// 再提交 /v11/checkmessage 并重试登录。
		checkReq, checkErr := http.NewRequest("GET", first.WebURL, nil)
		if checkErr != nil {
			return nil, checkErr
		}
		checkReq.Header.Set("User-Agent", userAgent)
		checkReq.Header.Set("Accept", "*/*")
		checkResp, checkErr := client.Do(checkReq)
		if checkErr != nil {
			return nil, checkErr
		}
		checkPage, _ := readAll(checkResp)
		checkResp.Body.Close()
		if cache.LoginCodeProvider == nil {
			return nil, fmt.Errorf("App 登录触发短信验证，请配置 LoginCodeProvider 输入验证码: %s", appBodySummary(step1Body))
		}
		enc2 := hiddenInputValue(string(checkPage), "enc2")
		checkTime := hiddenInputValue(string(checkPage), "time")
		phone := hiddenInputValue(string(checkPage), "msg")
		if phone == "" || strings.Contains(phone, "*") {
			// 页面上可能有脱敏手机号输入框，发送短信必须用完整手机号。
			phone = username
		}
		fid := hiddenInputValue(string(checkPage), "fid")
		loginTimeout := hiddenInputValue(string(checkPage), "loginTimeout")
		countryCode := hiddenInputValue(string(checkPage), "countrycode")
		checkType := hiddenInputValue(string(checkPage), "type")
		refer := hiddenInputValue(string(checkPage), "refer")
		if checkType == "" {
			checkType = "0"
		}

		kind := core.LoginCodeKindSMS
		prompt := "新设备登录需要短信验证码"
		checkPath := "/v11/checkmessage"
		if checkType == "2" || checkType == "4" {
			kind = core.LoginCodeKindPassword
			prompt = "新设备登录需要密码验证"
			checkPath = "/v11/checkPassword"
		}

		// 发送短信是显式动作：SDK 只把 SendSMS 交给调用方，由调用方
		// 在用户确认后触发，避免一登录就给手机发验证码。
		var sendSMS func(context.Context) error
		if kind == core.LoginCodeKindSMS {
			sendSMS = func(sendCtx context.Context) error {
				return sendLoginSMSCode(sendCtx, client, phone, countryCode, userAgent, first.WebURL)
			}
		}

		verify := func(value string) error {
			// 用户输入验证码可能较慢，提交前重新打开验证页刷新 enc2/time，
			// 避免服务端返回“认证地址已经失效”。
			if refreshReq, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, first.WebURL, nil); reqErr == nil {
				refreshReq.Header.Set("User-Agent", userAgent)
				refreshReq.Header.Set("Accept", "*/*")
				if refreshResp, doErr := client.Do(refreshReq); doErr == nil {
					if freshPage, readErr := readAll(refreshResp); readErr == nil {
						if v := hiddenInputValue(string(freshPage), "enc2"); v != "" {
							enc2 = v
						}
						if v := hiddenInputValue(string(freshPage), "time"); v != "" {
							checkTime = v
						}
						if v := hiddenInputValue(string(freshPage), "fid"); v != "" {
							fid = v
						}
						if v := hiddenInputValue(string(freshPage), "loginTimeout"); v != "" {
							loginTimeout = v
						}
						if v := hiddenInputValue(string(freshPage), "countrycode"); v != "" {
							countryCode = v
						}
						if v := hiddenInputValue(string(freshPage), "type"); v != "" {
							checkType = v
						}
						if v := hiddenInputValue(string(freshPage), "refer"); v != "" {
							refer = v
						}
					}
					refreshResp.Body.Close()
				}
			}
			checkURL := "https://passport2.chaoxing.com" + checkPath + "?enc2=" + url.QueryEscape(enc2) +
				"&time=" + url.QueryEscape(checkTime) + "&type=" + url.QueryEscape(checkType)
			checkForm := url.Values{}
			checkForm.Set("fid", fid)
			if kind == core.LoginCodeKindPassword {
				checkForm.Set("phone", phone)
				checkForm.Set("password", value)
			} else {
				checkForm.Set("countrycode", countryCode)
				checkForm.Set("phone", phone)
				checkForm.Set("vcode", value)
				checkForm.Set("loginTimeout", loginTimeout)
			}
			verifyReq, verifyErr := http.NewRequestWithContext(ctx, "POST", checkURL, strings.NewReader(checkForm.Encode()))
			if verifyErr != nil {
				return verifyErr
			}
			verifyReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			verifyReq.Header.Set("User-Agent", userAgent)
			verifyReq.Header.Set("Accept", "*/*")
			verifyResp, verifyErr := client.Do(verifyReq)
			if verifyErr != nil {
				return verifyErr
			}
			verifyBody, _ := readAll(verifyResp)
			verifyResp.Body.Close()
			log2.Print(log2.INFO, "xuexitong: App 登录验证码提交返回:", appBodySummary(verifyBody))
			var verifyResult struct {
				Status bool   `json:"status"`
				Mes    string `json:"mes"`
			}
			if err := json.Unmarshal(verifyBody, &verifyResult); err != nil {
				return fmt.Errorf("二次验证响应异常: %s", appBodySummary(verifyBody))
			}
			if !verifyResult.Status {
				message := strings.TrimSpace(verifyResult.Mes)
				if message == "" {
					message = appBodySummary(verifyBody)
				}
				return &loginCodeRejectedError{message: message}
			}
			// 二次验证成功后页面会跳转 refer 补齐会话，App 登录同样要跟一次。
			followRaw := strings.TrimSpace(refer)
			if followRaw == "" {
				followRaw = first.WebURL
			}
			if followURL, followErr := resolveChaoxingURL(followRaw); followErr == nil {
				followReq, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, followURL, nil)
				if reqErr == nil {
					followReq.Header.Set("User-Agent", userAgent)
					followReq.Header.Set("Accept", "*/*")
					if followResp, doErr := client.Do(followReq); doErr == nil {
						followResp.Body.Close()
						log2.Print(log2.INFO, "xuexitong: App 登录二次验证后已跟随会话跳转")
					}
				}
			}
			return nil
		}

		if err := cache.requestLoginCode(ctx, func(attempt int, lastErr string) LoginCodeChallenge {
			return LoginCodeChallenge{
				Kind:      kind,
				Prompt:    prompt,
				Phone:     phone,
				SendSMS:   sendSMS,
				Attempt:   attempt,
				LastError: lastErr,
			}
		}, verify); err != nil {
			return nil, err
		}
		// 验证通过后重试第一步，服务端应返回正常登录 url。
		retryReq, retryErr := http.NewRequest("POST", "https://passport2-api.chaoxing.com/v11/loginregister", strings.NewReader(form.Encode()))
		if retryErr != nil {
			return nil, retryErr
		}
		retryReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		retryReq.Header.Set("User-Agent", userAgent)
		retryReq.Header.Set("Accept", "*/*")
		retryReq.Header.Set("X-Requested-With", "XMLHttpRequest")
		retryResp, retryErr := client.Do(retryReq)
		if retryErr != nil {
			return nil, retryErr
		}
		retryBody, _ = readAll(retryResp)
		retryResp.Body.Close()
		log2.Print(log2.INFO, "xuexitong: App 登录重试返回:", appBodySummary(retryBody))
		_ = json.Unmarshal(retryBody, &first)
	}
	if !first.Status || first.URL == "" {
		return nil, fmt.Errorf("App 登录未完成，仍需要短信/消息验证: %s", appBodySummary(retryBody))
	}

	deviceInfo, err := buildAppDeviceInfo(profile)
	if err != nil {
		return nil, err
	}
	log2.Print(log2.INFO, "xuexitong: App 登录第一步完成，第二步地址:", appURLSummary(first.URL))
	loginURL, err := url.Parse(first.URL)
	if err != nil {
		return nil, err
	}
	query := loginURL.Query()
	query.Set("puid", "")
	query.Set("hddInfo", deviceInfo)
	query.Set("hddInfoLen", strconv.Itoa(len(deviceInfo)))
	loginURL.RawQuery = query.Encode()
	step2Form := url.Values{"data": {deviceInfo}}
	req2, reqErr := http.NewRequest("POST", loginURL.String(), strings.NewReader(step2Form.Encode()))
	if reqErr != nil {
		return nil, reqErr
	}
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req2.Header.Set("User-Agent", userAgent)
	req2.Header.Set("Accept", "*/*")
	req2.Header.Set("X-Requested-With", "XMLHttpRequest")
	step2, err := client.Do(req2)
	if err != nil {
		return nil, fmt.Errorf("App 登录第二步失败: %w", err)
	}
	step2Body, _ := readAll(step2)
	step2.Body.Close()
	var second struct {
		Msg    map[string]interface{} `json:"msg"`
		Result int                    `json:"result"`
	}
	if err := json.Unmarshal(step2Body, &second); err != nil {
		return nil, fmt.Errorf("App 登录第二步响应异常: %s", appBodySummary(step2Body))
	}
	log2.Print(log2.INFO, "xuexitong: App 登录第二步返回 result=", second.Result, " msgKeys=", appMapKeys(second.Msg))
	clientID := stringValue(second.Msg["clientId"])
	if clientID == "" {
		return nil, fmt.Errorf("App 登录未返回 clientId: %s", appBodySummary(step2Body))
	}
	device, err := decryptAppClientID(clientID)
	if err != nil {
		return nil, err
	}
	if device.CID == "" || device.Sc == "" {
		return nil, errors.New("clientId 解密结果缺少 cid/sc")
	}
	log2.Print(log2.INFO, "xuexitong: App 登录设备校验完成，当前 cookie:", appCookieNames(jar))

	puid := stringValue(second.Msg["puid"])
	if puid == "" {
		puid = stringValue(second.Msg["uid"])
	}
	// App 登录只拿到 cid/sc；学习通的课程/章节/作业/考试仍然走 Web 接口，
	// 还需要用同一份设备信息调用 SSO，换取 _uid/_d/uf/vc3 等 Web 会话 cookie。
	// 缺少这一步时 backclazzdata 会返回登录页 HTML。
	if ssoErr := completeAppWebSSOLogin(ctx, client, deviceInfo, puid, userAgent); ssoErr != nil {
		log2.Print(log2.INFO, "xuexitong: App 登录换取 Web 会话失败:", ssoErr)
	} else {
		log2.Print(log2.INFO, "xuexitong: App 登录 SSO 完成，当前 cookie:", appCookieNames(jar))
	}
	// 触发 mooc1-api 下发 k8s，后续章节/人脸接口需要。
	k8sReq, _ := http.NewRequest("GET", "https://mooc1-api.chaoxing.com/", nil)
	if k8sReq != nil {
		k8sReq.Header.Set("User-Agent", userAgent)
		k8sReq.Header.Set("Accept", "*/*")
		if k8sResp, k8sErr := client.Do(k8sReq); k8sErr == nil {
			k8sResp.Body.Close()
		}
	}
	cookies := collectAppCookies(jar)
	cache.SetCookies(cookies)
	cache.UserID = puid
	cache.FaceSc = device.Sc
	cache.FaceSign = &FaceSign{Cxcid: device.CID}
	cache.FaceSignChapter = &FaceSign{Cxcid: device.CID}

	return &AppLoginResult{
		ClientID: clientID,
		Cxcid:    device.CID,
		Sc:       device.Sc,
		PUID:     puid,
		UID:      stringValue(second.Msg["uid"]),
		Name:     stringValue(second.Msg["name"]),
		Cookies:  cookies,
	}, nil
}

type appDeviceInfo struct {
	Sc    string `json:"sc"`
	CID   string `json:"cid"`
	Stamp string `json:"_t"`
}

func buildAppLoginInfo(username, password string) (string, error) {
	payload, err := json.Marshal(struct {
		Uname string `json:"uname"`
		Code  string `json:"code"`
	}{Uname: username, Code: password})
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher([]byte(appLoginAESKey))
	if err != nil {
		return "", err
	}
	padding := aes.BlockSize - len(payload)%aes.BlockSize
	padded := append(payload, bytes.Repeat([]byte{byte(padding)}, padding)...)
	encrypted := make([]byte, len(padded))
	for offset := 0; offset < len(padded); offset += aes.BlockSize {
		block.Encrypt(encrypted[offset:offset+aes.BlockSize], padded[offset:offset+aes.BlockSize])
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func buildAppDeviceInfo(profile *appDeviceProfile) (string, error) {
	pub, err := appFacePublicKeyValue()
	if err != nil {
		return "", err
	}
	info := map[string]string{
		"platform":       profile.Platform,
		"brand":          profile.Brand,
		"app_ver":        profile.AppVer,
		"os_name":        profile.OSName,
		"cdid":           profile.CDID,
		"deviceUniqueId": profile.DeviceUniqueID,
		"mediaDrmId":     profile.MediaDrmID,
		"os_ver":         profile.OSVer,
		"time_stamp":     strconv.FormatInt(time.Now().UnixMilli(), 10),
		"os_lang":        profile.OSLang,
		"app_name":       profile.AppName,
		"cpu_ar":         profile.CPUArch,
		"hardware":       profile.Hardware,
		"resolution":     profile.Resolution,
		"android_id":     profile.AndroidID,
		"signatures":     profile.Signatures,
	}
	raw, err := json.Marshal(info)
	if err != nil {
		return "", err
	}
	chunk := pub.Size() - 11
	encrypted := make([]byte, 0, (len(raw)/chunk+1)*pub.Size())
	for offset := 0; offset < len(raw); offset += chunk {
		end := offset + chunk
		if end > len(raw) {
			end = len(raw)
		}
		part, encryptErr := rsa.EncryptPKCS1v15(rand.Reader, pub, raw[offset:end])
		if encryptErr != nil {
			return "", encryptErr
		}
		encrypted = append(encrypted, part...)
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func decryptAppClientID(clientID string) (*appDeviceInfo, error) {
	pub, err := appFacePublicKeyValue()
	if err != nil {
		return nil, err
	}
	ciphertext, err := base64.StdEncoding.DecodeString(clientID)
	if err != nil {
		return nil, err
	}
	size := pub.Size()
	plaintext := make([]byte, 0, len(ciphertext))
	for offset := 0; offset < len(ciphertext); offset += size {
		end := offset + size
		if end > len(ciphertext) {
			end = len(ciphertext)
		}
		value := new(big.Int).SetBytes(ciphertext[offset:end])
		value.Exp(value, big.NewInt(int64(pub.E)), pub.N)
		block := value.FillBytes(make([]byte, size))
		if len(block) < 11 || block[0] != 0 || block[1] != 1 {
			return nil, errors.New("clientId 解密填充格式错误")
		}
		index := 2
		for index < len(block) && block[index] == 0xff {
			index++
		}
		if index >= len(block) || block[index] != 0 {
			return nil, errors.New("clientId 解密填充结束符缺失")
		}
		plaintext = append(plaintext, block[index+1:]...)
	}
	var device appDeviceInfo
	if err := json.Unmarshal(plaintext, &device); err != nil {
		return nil, fmt.Errorf("clientId 解密结果解析失败: %w", err)
	}
	return &device, nil
}

func appFacePublicKeyValue() (*rsa.PublicKey, error) {
	facePublicKeyOnce.Do(func() {
		der, err := base64.StdEncoding.DecodeString(appFacePublicKey)
		if err != nil {
			facePublicKeyErr = err
			return
		}
		key, err := x509.ParsePKIXPublicKey(der)
		if err != nil {
			facePublicKeyErr = err
			return
		}
		rsaKey, ok := key.(*rsa.PublicKey)
		if !ok {
			facePublicKeyErr = errors.New("人脸公钥不是 RSA 公钥")
			return
		}
		facePublicKey = rsaKey
	})
	return facePublicKey, facePublicKeyErr
}

// completeAppWebSSOLogin 用 App 登录的设备信息换取 Web 会话。
//
// 学习通 App 登录成功后仍会请求 sso.chaoxing.com/apis/login/userLogin.do，
// 由服务端写入 _uid/_d/uf/vc3/fid 等 .chaoxing.com cookie。课程、章节、作业、
// 考试等接口都是 Web 接口，缺这组 cookie 会返回登录页 HTML。
func completeAppWebSSOLogin(ctx context.Context, client *http.Client, deviceInfo, puid, userAgent string) error {
	if strings.TrimSpace(puid) == "" {
		return errors.New("App 登录响应缺少 puid")
	}
	// 先访问一次 SSO 首页，确保拿到 sso 域名下的 JSESSIONID/route，
	// 否则 userLogin.do 会返回 cookieStatus=lost。
	if warmReq, warmErr := http.NewRequestWithContext(ctx, http.MethodGet, "https://sso.chaoxing.com/", nil); warmErr == nil {
		warmReq.Header.Set("User-Agent", userAgent)
		warmReq.Header.Set("Accept", "*/*")
		if warmResp, doErr := client.Do(warmReq); doErr == nil {
			warmResp.Body.Close()
		}
	}
	ssoURL := "https://sso.chaoxing.com/apis/login/userLogin.do?puid=" + url.QueryEscape(puid) +
		"&hddInfo=" + url.QueryEscape(deviceInfo) + "&len=" + strconv.Itoa(len(deviceInfo))
	form := url.Values{}
	form.Set("data", deviceInfo)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ssoURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "zh_CN")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	body, _ := readAll(resp)
	resp.Body.Close()
	var result struct {
		Msg    map[string]interface{} `json:"msg"`
		Result int                    `json:"result"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("返回内容不是 JSON: %s", appBodySummary(body))
	}
	if len(result.Msg) == 0 {
		return fmt.Errorf("SSO 登录失败: %s", appBodySummary(body))
	}
	return nil
}

func appURLSummary(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Host == "" {
		return "(invalid)"
	}
	return parsed.Scheme + "://" + parsed.Host + parsed.Path
}

func appMapKeys(values map[string]interface{}) string {
	if len(values) == 0 {
		return ""
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	return strings.Join(keys, ",")
}

// appCookieNames 汇总当前会话在关键域名下的 cookie 名称，仅用于日志排查。
func appCookieNames(jar http.CookieJar) string {
	if jar == nil {
		return ""
	}
	seen := map[string]bool{}
	var names []string
	for _, raw := range []string{
		"https://sso.chaoxing.com",
		"https://mooc1-api.chaoxing.com",
		"https://chaoxing.com",
	} {
		parsed, err := url.Parse(raw)
		if err != nil {
			continue
		}
		for _, cookie := range jar.Cookies(parsed) {
			if seen[cookie.Name] {
				continue
			}
			seen[cookie.Name] = true
			names = append(names, cookie.Name)
		}
	}
	return strings.Join(names, ",")
}

func collectAppCookies(jar http.CookieJar) []*http.Cookie {
	if jar == nil {
		return nil
	}
	urls := []string{
		"https://chaoxing.com",
		"https://passport2-api.chaoxing.com",
		"https://sso.chaoxing.com",
		"https://mooc1-api.chaoxing.com",
		"https://mooc1.chaoxing.com",
		"https://pan-yz.chaoxing.com",
	}
	seen := map[string]bool{}
	var cookies []*http.Cookie
	for _, raw := range urls {
		parsed, err := url.Parse(raw)
		if err != nil {
			continue
		}
		for _, cookie := range jar.Cookies(parsed) {
			key := cookie.Name + "|" + cookie.Domain + "|" + cookie.Path
			if seen[key] {
				continue
			}
			seen[key] = true
			cookies = append(cookies, cookie)
		}
	}
	return cookies
}

func randomHex(size int) string {
	buffer := make([]byte, size)
	if _, err := rand.Read(buffer); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 16)
	}
	return fmt.Sprintf("%x", buffer)
}

func readAll(response *http.Response) ([]byte, error) {
	if response == nil {
		return nil, errors.New("响应为空")
	}
	buffer := new(bytes.Buffer)
	_, err := buffer.ReadFrom(response.Body)
	return buffer.Bytes(), err
}

func appBodySummary(body []byte) string {
	text := strings.Join(strings.Fields(string(body)), " ")
	if len(text) > 200 {
		return text[:200]
	}
	return text
}

func hiddenInputValue(htmlText, id string) string {
	pattern := regexp.MustCompile(`(?s)value="([^"]*)"[^>]*id="` + regexp.QuoteMeta(id) + `"`)
	if matched := pattern.FindStringSubmatch(htmlText); len(matched) > 1 {
		return html.UnescapeString(matched[1])
	}
	pattern = regexp.MustCompile(`(?s)id="` + regexp.QuoteMeta(id) + `"[^>]*value="([^"]*)"`)
	if matched := pattern.FindStringSubmatch(htmlText); len(matched) > 1 {
		return html.UnescapeString(matched[1])
	}
	return ""
}
