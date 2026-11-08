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
//	logininfo(AES/ECB) -> v11/loginregister -> 本地生成 deviceInfo
//	-> userLogin4Uname.do -> clientId -> 本地 RSA 解密出 {cid, sc}
//
// 这样 cxcid/sc 不再依赖 App 抓包或手动配置。
func (cache *XueXiTUserCache) AppLoginApi(ctx context.Context) (*AppLoginResult, error) {
	username := strings.TrimSpace(cache.Name)
	password := cache.Password
	if username == "" || password == "" {
		return nil, errors.New("App 登录需要账号和密码")
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
	req1.Header.Set("User-Agent", appLoginUserAgent)
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
	if first.URL == "" && first.WebURL != "" {
		// type=2：服务端要求短信/消息验证。拉起 checkmessage 页面，
		// 解析隐藏参数，通过 LoginCodeProvider 让调用方输入短信验证码，
		// 再提交 /v11/checkmessage 并重试登录。
		checkReq, checkErr := http.NewRequest("GET", first.WebURL, nil)
		if checkErr != nil {
			return nil, checkErr
		}
		checkReq.Header.Set("User-Agent", appLoginUserAgent)
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
		if checkType == "" {
			checkType = "0"
		}
		// 发送短信是显式动作：SDK 只把 SendSMS 交给调用方，由调用方
		// 在用户确认后触发，避免一登录就给手机发验证码。
		sendSMS := func(sendCtx context.Context) error {
			return sendLoginSMSCode(sendCtx, client, phone, countryCode, first.WebURL)
		}
		code, codeErr := cache.LoginCodeProvider(ctx, LoginCodeChallenge{
			Kind:    "sms",
			Prompt:  "新设备登录需要短信验证码",
			Phone:   phone,
			SendSMS: sendSMS,
		})
		if codeErr != nil {
			return nil, codeErr
		}
		code = strings.TrimSpace(code)
		if code == "" {
			return nil, errors.New("未提供短信验证码")
		}
		checkURL := "https://passport2.chaoxing.com/v11/checkmessage?enc2=" + url.QueryEscape(enc2) + "&time=" + url.QueryEscape(checkTime) + "&type=" + url.QueryEscape(checkType)
		checkForm := url.Values{}
		checkForm.Set("countrycode", hiddenInputValue(string(checkPage), "countrycode"))
		checkForm.Set("phone", phone)
		checkForm.Set("vcode", code)
		checkForm.Set("fid", fid)
		checkForm.Set("loginTimeout", loginTimeout)
		verifyReq, verifyErr := http.NewRequest("POST", checkURL, strings.NewReader(checkForm.Encode()))
		if verifyErr != nil {
			return nil, verifyErr
		}
		verifyReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		verifyReq.Header.Set("User-Agent", appLoginUserAgent)
		verifyReq.Header.Set("Accept", "*/*")
		verifyResp, verifyErr := client.Do(verifyReq)
		if verifyErr != nil {
			return nil, verifyErr
		}
		verifyBody, _ := readAll(verifyResp)
		verifyResp.Body.Close()
		var verifyResult struct {
			Status bool   `json:"status"`
			Mes    string `json:"mes"`
		}
		if err := json.Unmarshal(verifyBody, &verifyResult); err != nil || !verifyResult.Status {
			return nil, fmt.Errorf("短信验证失败: %s", appBodySummary(verifyBody))
		}
		// 验证通过后重试第一步，服务端应返回正常登录 url。
		retryReq, retryErr := http.NewRequest("POST", "https://passport2-api.chaoxing.com/v11/loginregister", strings.NewReader(form.Encode()))
		if retryErr != nil {
			return nil, retryErr
		}
		retryReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		retryReq.Header.Set("User-Agent", appLoginUserAgent)
		retryReq.Header.Set("Accept", "*/*")
		retryReq.Header.Set("X-Requested-With", "XMLHttpRequest")
		retryResp, retryErr := client.Do(retryReq)
		if retryErr != nil {
			return nil, retryErr
		}
		retryBody, _ := readAll(retryResp)
		retryResp.Body.Close()
		_ = json.Unmarshal(retryBody, &first)
	}
	if !first.Status || first.URL == "" {
		return nil, fmt.Errorf("App 登录未完成，仍需要短信/消息验证: %s", appBodySummary(step1Body))
	}

	deviceInfo, err := buildAppDeviceInfo()
	if err != nil {
		return nil, err
	}
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
	req2.Header.Set("User-Agent", appLoginUserAgent)
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

	// 触发 mooc1-api 下发 k8s，后续章节/人脸接口需要。
	k8sReq, _ := http.NewRequest("GET", "https://mooc1-api.chaoxing.com/", nil)
	if k8sReq != nil {
		k8sReq.Header.Set("User-Agent", appLoginUserAgent)
		k8sReq.Header.Set("Accept", "*/*")
		if k8sResp, k8sErr := client.Do(k8sReq); k8sErr == nil {
			k8sResp.Body.Close()
		}
	}
	cookies := collectAppCookies(jar)
	cache.SetCookies(cookies)
	puid := stringValue(second.Msg["puid"])
	if puid == "" {
		puid = stringValue(second.Msg["uid"])
	}
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

func buildAppDeviceInfo() (string, error) {
	pub, err := appFacePublicKeyValue()
	if err != nil {
		return "", err
	}
	info := map[string]string{
		"platform":       "android",
		"brand":          "Xiaomi",
		"app_ver":        "3.7.0.4",
		"os_name":        "android",
		"cdid":           "",
		"deviceUniqueId": randomHex(16),
		"mediaDrmId":     randomHex(16),
		"os_ver":         "11",
		"time_stamp":     strconv.FormatInt(time.Now().UnixMilli(), 10),
		"os_lang":        "zh_CN",
		"app_name":       "com.chaoxing.mobile",
		"cpu_ar":         "arm64-v8a",
		"hardware":       "qcom",
		"resolution":     "1080x2206",
		"android_id":     randomHex(8),
		"signatures":     "",
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
