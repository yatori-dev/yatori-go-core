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
	type loginReply struct {
		Status bool   `json:"status"`
		URL    string `json:"url"`
		WebURL string `json:"webUrl"`
		Mes    string `json:"mes"`
	}
	var first loginReply
	if err := json.Unmarshal(step1Body, &first); err != nil {
		return nil, fmt.Errorf("App 登录第一步响应异常（%s）", appLoginReplyShape(step1.StatusCode, step1Body))
	}
	loginAddress := first.URL
	if first.URL == "" && first.WebURL != "" {
		// type=2：服务端要求短信/消息验证。拉起 checkmessage 页面，
		// 解析隐藏参数，通过 LoginCodeProvider 让调用方输入短信验证码，
		// /v11/tocheckmessage 验证成功后通过 CLIENT_THIRD_LOGIN 继续 SSO。
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
		if countryCode == "" {
			countryCode = "86"
		}
		checkType := hiddenInputValue(string(checkPage), "type")
		if checkType == "" {
			checkType = "0"
		}
		// 发送短信是显式动作：SDK 只把 SendSMS 交给调用方，由调用方
		// 在用户确认后触发，避免一登录就给手机发验证码。
		sendSMS := func(sendCtx context.Context) error {
			return sendLoginSMSCode(sendCtx, client, phone, countryCode, userAgent, first.WebURL)
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
		checkForm.Set("countrycode", countryCode)
		checkForm.Set("phone", phone)
		checkForm.Set("vcode", code)
		checkForm.Set("fid", fid)
		checkForm.Set("loginTimeout", loginTimeout)
		verifyReq, verifyErr := http.NewRequest("POST", checkURL, strings.NewReader(checkForm.Encode()))
		if verifyErr != nil {
			return nil, verifyErr
		}
		verifyReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		verifyReq.Header.Set("User-Agent", userAgent)
		verifyReq.Header.Set("Accept", "*/*")
		verifyResp, verifyErr := client.Do(verifyReq)
		if verifyErr != nil {
			return nil, verifyErr
		}
		verifyBody, _ := readAll(verifyResp)
		verifyResp.Body.Close()
		loginAddress, err = appSMSLoginContinuation(verifyResp.StatusCode, verifyBody)
		if err != nil {
			return nil, err
		}
	} else if !first.Status || loginAddress == "" {
		return nil, fmt.Errorf("App 登录未完成（%s）", appLoginReplyShape(step1.StatusCode, step1Body))
	}

	deviceInfo, err := buildAppDeviceInfo(profile)
	if err != nil {
		return nil, err
	}
	loginURL, err := url.Parse(loginAddress)
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
		return nil, fmt.Errorf("App 单点登录响应异常（%s）", appLoginReplyShape(step2.StatusCode, step2Body))
	}
	clientID := stringValue(second.Msg["clientId"])
	if clientID == "" {
		return nil, fmt.Errorf("App 单点登录未返回 clientId（%s）", appLoginReplyShape(step2.StatusCode, step2Body))
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
		k8sReq.Header.Set("User-Agent", userAgent)
		k8sReq.Header.Set("Accept", "*/*")
		if k8sResp, k8sErr := client.Do(k8sReq); k8sErr == nil {
			k8sResp.Body.Close()
		}
	}
	cookies := collectAppCookies(jar)
	cache.SetCookies(cookies)
	cache.appSession = client
	cache.appUserAgent = userAgent
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

// appSMSLoginContinuation follows the success handler in /v11/tocheckmessage.
func appSMSLoginContinuation(statusCode int, body []byte) (string, error) {
	var response struct {
		Status bool            `json:"status"`
		Type   json.RawMessage `json:"type"`
	}
	if statusCode < 200 || statusCode >= 300 || json.Unmarshal(body, &response) != nil || !response.Status {
		return "", fmt.Errorf("短信验证失败（%s）", appLoginReplyShape(statusCode, body))
	}
	switch string(bytes.TrimSpace(response.Type)) {
	case "1", `"1"`:
		return "", errors.New("短信验证已通过，学习通要求先在官方页面修改密码")
	}
	return "https://sso.chaoxing.com/apis/login/userLogin4Uname.do?ft=true", nil
}

// appLoginReplyShape records protocol structure without exposing response values.
func appLoginReplyShape(httpStatus int, body []byte) string {
	var fields map[string]json.RawMessage
	if json.Unmarshal(body, &fields) != nil || fields == nil {
		return fmt.Sprintf("HTTP %d，非 JSON 对象", httpStatus)
	}

	status := "缺失"
	if raw, ok := fields["status"]; ok {
		switch string(bytes.TrimSpace(raw)) {
		case "true":
			status = "true"
		case "false":
			status = "false"
		default:
			status = "非布尔"
		}
	}

	present := make([]string, 0, 8)
	for _, key := range []string{"url", "webUrl", "redirectUrl", "type", "result", "data", "msg", "mes"} {
		switch string(bytes.TrimSpace(fields[key])) {
		case "", "null", `""`, "{}", "[]":
		default:
			present = append(present, key)
		}
	}
	if len(present) == 0 {
		present = append(present, "无")
	}
	return fmt.Sprintf("HTTP %d，status=%s，非空字段=%s", httpStatus, status, strings.Join(present, "/"))
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
