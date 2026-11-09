package xuexitong

import (
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"

	"io/ioutil"

	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/yatori-dev/yatori-go-core/v3/utils"
)

var randChar []string = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f", "A", "B", "C", "D", "E", "F"}

// 获取学习通验证码
// XueXiTProcessVerifyImage 只拉取反爬验证码（processVerifyPng.ac）。
//
// 注意：这个接口只有会话被重定向到 antispiderShowVerify.ac 时才返回图片；
// 正常会话会返回 404 HTML，因此不能和章节图片验证码混在一起重试。
func (cache *XueXiTUserCache) XueXiTProcessVerifyImage(retry int, lastErr error) (image.Image, error) {
	if retry < 0 {
		return nil, lastErr
	}

	urlStr := "https://mooc1.chaoxing.com/processVerifyPng.ac?t=" + fmt.Sprintf("%d", time.Now().UnixMilli())
	method := "GET"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{},
	}
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse("http://" + cache.ProxyIP)
		}
	}
	client := newRequestClient(tr)
	req, err := http.NewRequest(method, urlStr, nil)
	if err != nil {
		return nil, err
	}
	cache.applyCourseSession(client, req, GetUA("mobile"))
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "mooc1.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")

	res, err := client.Do(req)
	if err != nil {
		return cache.XueXiTProcessVerifyImage(retry-1, err)
	}
	defer res.Body.Close()
	utils.CookiesAddNoRepetition(&cache.cookies, res.Cookies())

	img, _, err := image.Decode(res.Body)
	if err != nil {
		return cache.XueXiTProcessVerifyImage(retry-1, err)
	}
	return img, nil
}

// XueXiTVerificationCodeApi 兼容旧调用：先取反爬验证码，失败再取章节图片验证码。
func (cache *XueXiTUserCache) XueXiTVerificationCodeApi(retry int, lastErr error) (image.Image, error) {
	img, err := cache.XueXiTProcessVerifyImage(retry, lastErr)
	if err == nil {
		return img, nil
	}
	return cache.XueXiTChapterVerificationCodeApi(retry, err)
}

// 获取学习通验证码
// 获取学习通验证码（直接返回 image.Image）
func (cache *XueXiTUserCache) XueXiTChapterVerificationCodeApi(retry int, lastErr error) (image.Image, error) {
	if retry < 0 {
		return nil, lastErr
	}

	urlStr := "https://mooc1.chaoxing.com/mooc-ans/kaptcha-img/code?" +
		fmt.Sprintf("%d", time.Now().UnixMilli())

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{},
	}

	// 代理
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse("http://" + cache.ProxyIP)
		}
	}

	client := &http.Client{Transport: tr}

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}

	// cookie
	cache.applyCourseSession(client, req, GetUA("mobile"))
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "mooc1-api.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")

	res, err := client.Do(req)
	if err != nil {
		return cache.XueXiTChapterVerificationCodeApi(retry-1, err)
	}
	defer res.Body.Close()
	utils.CookiesAddNoRepetition(&cache.cookies, res.Cookies())

	// 直接 decode 图片
	img, _, err := image.Decode(res.Body)
	if err != nil {
		// 可能是坏图，尝试重试
		return cache.XueXiTChapterVerificationCodeApi(retry-1, err)
	}

	return img, nil
}

// 提交学习通反爬验证码（processVerify）。
//
// 对应 antispiderShowVerify.ac 页面表单：GET /html/processVerify.ac
// 参数 ucode/app，成功返回 302 到原请求地址；POST 会被 405 拒绝。
func (cache *XueXiTUserCache) XueXiTPassVerificationCode(code string, retry int, lastErr error) (bool, error) {
	if retry < 0 {
		return false, lastErr
	}
	urlStr := "https://mooc1.chaoxing.com/html/processVerify.ac?ucode=" + url.QueryEscape(code) + "&app=0"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{},
	}
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse("http://" + cache.ProxyIP)
		}
	}
	client := &http.Client{
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return false, err
	}
	cache.applyCourseSession(client, req, GetUA("mobile"))
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "mooc1.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Referer", "https://mooc1.chaoxing.com/antispiderShowVerify.ac")

	res, err := client.Do(req)
	if err != nil {
		return cache.XueXiTPassVerificationCode(code, retry-1, err)
	}
	defer res.Body.Close()
	utils.CookiesAddNoRepetition(&cache.cookies, res.Cookies())
	if res.StatusCode == http.StatusFound || res.StatusCode == http.StatusMovedPermanently {
		return true, nil
	}
	return false, nil
}

// 提交学习通章节图片验证码（kaptcha-img）。
//
// 图片来自 /mooc-ans/kaptcha-img/code，提交到同目录的 ajaxValidate2，
// 返回 {"status":true/false}。
func (cache *XueXiTUserCache) XueXiTPassCahpterVerificationCode(code string, retry int, lastErr error) (bool, error) {
	if retry < 0 {
		return false, lastErr
	}
	urlStr := "https://mooc1.chaoxing.com/mooc-ans/kaptcha-img/ajaxValidate2"
	form := url.Values{}
	form.Set("code", code)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{},
	}
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse("http://" + cache.ProxyIP)
		}
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("POST", urlStr, strings.NewReader(form.Encode()))
	if err != nil {
		return false, err
	}
	cache.applyCourseSession(client, req, GetUA("mobile"))
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "mooc1.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Referer", "https://mooc1.chaoxing.com/mooc-ans/mycourse/studentstudy")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")

	res, err := client.Do(req)
	if err != nil {
		return cache.XueXiTPassCahpterVerificationCode(code, retry-1, err)
	}
	defer res.Body.Close()
	utils.CookiesAddNoRepetition(&cache.cookies, res.Cookies())

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return false, err
	}
	var payload struct {
		Status bool `json:"status"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return false, nil
	}
	return payload.Status, nil
}

// XueXiTSliderVerificationCodeApi 获取学习通滑块验证码相关信息，返回信息cx_captcha_function({"t":1764584640340,"captchaId":"Ew0z9skxsLzVKQjmeObQiRVLxkxbPkRF"})
func (cache *XueXiTUserCache) XueXiTSliderVerificationCodeApi(captchaId string, retry int, lastErr error) (string, error) {

	urlStr := "https://captcha.chaoxing.com/captcha/get/conf?callback=cx_captcha_function&captchaId=" + captchaId + "&_=" + fmt.Sprintf("%d", time.Now().UnixMilli())
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
	client := newRequestClient(tr)
	req, err := http.NewRequest(method, urlStr, nil)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("User-Agent", GetUA("mobile"))
	req.Header.Add("Accept-Language", "zh-CN,en-US;q=0.9")
	req.Header.Add("X-Requested-With", "com.chaoxing.mobile")
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "captcha.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return cache.XueXiTSliderVerificationCodeApi(captchaId, retry-1, err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	//fmt.Println(string(body))
	return string(body), nil
}

// 生成滑块动态iv - 对齐官方前端JS: iv=md5(captchaId+"slide"+当前毫秒+uuid)
// image接口上报后服务器记录，check时必须传同一iv，否则报 verification error
func NewSliderCaptchaIv(captchaId string) string {
	ivHash := md5.Sum([]byte(fmt.Sprintf("%s%s%d%s", captchaId, "slide", time.Now().UnixMilli(), uuid.New().String())))
	return hex.EncodeToString(ivHash[:])
}

// 拉取验证码图片等信息
// iv: 动态iv(NewSliderCaptchaIv生成)，会随URL上报并在check时必须复用；为空时自动生成
func (cache *XueXiTUserCache) XueXiTSliderVerificationImgApi(captchaId, serverTime, referer, iv string, retry int, lastErr error) (string, error) {
	// 计算 MD5
	captchaKeyHash := md5.Sum([]byte(serverTime + uuid.New().String()))
	captchaKey := hex.EncodeToString(captchaKeyHash[:])

	if iv == "" {
		iv = NewSliderCaptchaIv(captchaId)
	}

	// 计算token
	sum := md5.Sum([]byte(fmt.Sprintf("%s%s%s%s", serverTime, captchaId, "slide", captchaKey)))
	md5hex := hex.EncodeToString(sum[:])
	serverTimeInt, _ := strconv.ParseInt(serverTime, 10, 64)
	token := fmt.Sprintf("%s:%d", md5hex, serverTimeInt+300000)
	version := "1.1.20"
	urlStr := "https://captcha.chaoxing.com/captcha/get/verification/image?callback=cx_captcha_function&captchaId=" + captchaId + "&type=slide&version=" + version + "&captchaKey=" + captchaKey + "&token=" + token + "&referer=" + url.QueryEscape(referer) + "&iv=" + iv + "&_=" + fmt.Sprintf("%d", time.Now().UnixMilli())
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
		return "", err
	}
	req.Header.Add("User-Agent", GetUA("mobile"))
	req.Header.Add("Accept-Language", "zh-CN,en-US;q=0.9")
	req.Header.Add("X-Requested-With", "com.chaoxing.mobile")
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}
	if referer != "" {
		// 浏览器请求image接口时自带来源页Referer头
		req.Header.Add("Referer", referer)
	}
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "captcha.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return cache.XueXiTSliderVerificationImgApi(captchaId, serverTime, referer, iv, retry-1, err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return cache.XueXiTSliderVerificationImgApi(captchaId, serverTime, referer, iv, retry-1, err)
	}
	//fmt.Println(string(body))
	return string(body), nil
}

// 请求并获取图片
func (cache *XueXiTUserCache) PullSliderImgApi(imgUrl string, retry int, lastErr error) (image.Image, error) {
	if retry < 0 {
		return nil, lastErr
	}
	method := "GET"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{},
	}
	client := &http.Client{
		Transport: tr,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return fmt.Errorf("滑块图片请求不允许重定向")
		},
	}
	req, err := http.NewRequest(method, imgUrl, nil)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	req.Header.Add("User-Agent", GetUA("mobile"))
	req.Header.Add("Accept-Language", "zh-CN,en-US;q=0.9")
	req.Header.Add("X-Requested-With", "com.chaoxing.mobile")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "captcha.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")

	resp, err := client.Do(req)
	if err != nil {
		return cache.PullSliderImgApi(imgUrl, retry-1, fmt.Errorf("请求失败: %v", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP 状态码异常: %d", resp.StatusCode)
	}

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("图片解码失败: %v", err)
	}

	return img, nil
}

// 过滑块接口
// referer: 触发滑块的页面地址(浏览器发起check时会自动携带来源页Referer)
// iv: 与image接口一致的动态iv(NewSliderCaptchaIv生成)
// 注意: 该请求不携带任何cookie(仅代理+UA)。check接口启TLS指纹风控，
// 必须用Chrome指纹客户端(完整ClientHello+HTTP/1.1，服务器协商h2时降级HTTP/2)
func (cache *XueXiTUserCache) PassSliderApi(captchaId, token, xPoint, runEnv, referer, iv string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	if iv == "" {
		iv = "cdd9bfb9e7805d0d2d5f1ad4498f70e1"
	}
	urlStr := "https://captcha.chaoxing.com/captcha/check/verification/result?callback=cx_captcha_function&captchaId=" + captchaId + "&type=slide&token=" + token + "&textClickArr=" + url.QueryEscape(`[{"x":`+xPoint+`}]`) + "&coordinate=" + url.QueryEscape(`[]`) + "&runEnv=" + runEnv + "&version=1.1.20&t=a&iv=" + iv + "&_=" + fmt.Sprintf("%d", time.Now().UnixMilli())

	headers := map[string]string{
		"User-Agent":       "Mozilla/5.0 (Linux; Android 8.1.0; MI 5X Build/OPM1.171019.019; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/71.0.3578.99 Mobile Safari/537.36 (schild:ce5175d20950c8ee955fb03246f762da) (device:MI 5X) Language/zh_CN com.chaoxing.mobile/ChaoXingStudy_3_6.7.2_android_phone_10936_311 (@Kalimdor)_76c82452584d47e39ab79aa54ea86554",
		"Accept-Language":  "zh-CN,en-US;q=0.9",
		"X-Requested-With": "com.chaoxing.mobile",
		"Accept":           "*/*",
		"Connection":       "keep-alive",
	}
	if referer != "" {
		headers["Referer"] = referer
	}
	proxyIP := ""
	if cache.IpProxySW {
		proxyIP = cache.ProxyIP
	}

	// 优先Chrome指纹(完整ClientHello, HTTP/1.1)；服务器协商h2时降级HTTP/2客户端
	body, resp, autoErr := captchaFingerprintGet(urlStr, headers, proxyIP, fingerprintAuto)
	if autoErr != nil {
		body, resp, err := captchaFingerprintGet(urlStr, headers, proxyIP, fingerprintHTTP2)
		if err != nil {
			fmt.Println(err)
			return cache.PassSliderApi(captchaId, token, xPoint, runEnv, referer, iv, retry-1, fmt.Errorf("HTTP/1.1: %v; HTTP/2: %v", autoErr, err))
		}
		if err := validateResponse("提交滑块验证码", resp, body); err != nil {
			return "", err
		}
		return string(body), nil
	}
	if err := validateResponse("提交滑块验证码", resp, body); err != nil {
		return "", err
	}
	return string(body), nil
}
