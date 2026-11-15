package xuexitong

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"github.com/yatori-dev/yatori-go-core/v3/core"
	"github.com/yatori-dev/yatori-go-core/v3/utils"
)

const webLoginUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36"

// CompleteWebTwoFactorLogin 处理网页登录返回的 containTwoFactorLogin。
//
// fanyalogin 在账号需要短信二次验证时会返回：
//
//	{"status":true,"containTwoFactorLogin":true,"twoFactorLoginUrl":"/v11/tocheckmessage?..."}
//
// 这里用同一会话打开验证页，通过 LoginCodeProvider 让调用方触发发送短信
// 并输入验证码，再提交 /v11/checkmessage 完成登录。
func (cache *XueXiTUserCache) CompleteWebTwoFactorLogin(ctx context.Context, twoFactorURL string) error {
	if cache.LoginCodeProvider == nil {
		return errors.New("网页登录需要短信二次验证，请配置 LoginCodeProvider 输入验证码")
	}
	pageURL, err := resolveChaoxingURL(twoFactorURL)
	if err != nil {
		return err
	}
	client, jar, err := cache.webSessionClient()
	if err != nil {
		return err
	}

	pageBody, err := cache.getWebPage(ctx, client, pageURL)
	if err != nil {
		return fmt.Errorf("打开二次验证页面失败: %w", err)
	}
	cache.syncJarCookies(jar)

	enc2 := hiddenInputValue(string(pageBody), "enc2")
	checkTime := hiddenInputValue(string(pageBody), "time")
	phone := hiddenInputValue(string(pageBody), "msg")
	if phone == "" || strings.Contains(phone, "*") {
		phone = cache.Name
	}
	fid := hiddenInputValue(string(pageBody), "fid")
	loginTimeout := hiddenInputValue(string(pageBody), "loginTimeout")
	countryCode := hiddenInputValue(string(pageBody), "countrycode")
	checkType := hiddenInputValue(string(pageBody), "type")
	if uid := hiddenInputValue(string(pageBody), "uid"); uid != "" {
		cache.UserID = uid
	}
	if checkType == "" {
		checkType = "3"
	}
	if enc2 == "" || checkTime == "" {
		return errors.New("二次验证页面缺少 enc2/time 参数")
	}

	kind := core.LoginCodeKindSMS
	prompt := "网页登录需要短信二次验证"
	checkPath := "/v11/checkmessage"
	if checkType == "2" || checkType == "4" {
		kind = core.LoginCodeKindPassword
		prompt = "网页登录需要密码二次验证"
		checkPath = "/v11/checkPassword"
	}

	var sendSMS func(context.Context) error
	if kind == core.LoginCodeKindSMS {
		sendSMS = func(sendCtx context.Context) error {
			return sendLoginSMSCode(sendCtx, client, phone, countryCode, cache.WebUA(), pageURL)
		}
	}

	verify := func(value string) error {
		form := url.Values{}
		form.Set("fid", fid)
		if kind == core.LoginCodeKindPassword {
			form.Set("phone", phone)
			form.Set("password", value)
		} else {
			form.Set("countrycode", countryCode)
			form.Set("phone", phone)
			form.Set("vcode", value)
			form.Set("loginTimeout", loginTimeout)
		}
		checkURL := "https://passport2.chaoxing.com" + checkPath +
			"?enc2=" + url.QueryEscape(enc2) + "&time=" + url.QueryEscape(checkTime) + "&type=" + url.QueryEscape(checkType)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, checkURL, strings.NewReader(form.Encode()))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("User-Agent", cache.WebUA())
		req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
		req.Header.Set("X-Requested-With", "XMLHttpRequest")

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("提交二次验证失败: %w", err)
		}
		body, _ := readAll(resp)
		_ = resp.Body.Close()
		cache.captureResponseCookies(resp)
		cache.syncJarCookies(jar)

		var result struct {
			Status bool   `json:"status"`
			Mes    string `json:"mes"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			return fmt.Errorf("二次验证响应异常: %s", appBodySummary(body))
		}
		if !result.Status {
			message := strings.TrimSpace(result.Mes)
			if message == "" {
				message = appBodySummary(body)
			}
			return &loginCodeRejectedError{message: message}
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
		return err
	}

	// 页面成功后 endSuccessFunc 会跳到 refer；跟随一次补齐会话 Cookie。
	refer := hiddenInputValue(string(pageBody), "refer")
	if refer == "" {
		refer = "http://i.mooc.chaoxing.com"
	}
	if followURL, followErr := resolveChaoxingURL(refer); followErr == nil {
		followReq, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, followURL, nil)
		if reqErr == nil {
			followReq.Header.Set("User-Agent", cache.WebUA())
			if followResp, doErr := client.Do(followReq); doErr == nil {
				cache.captureResponseCookies(followResp)
				_ = followResp.Body.Close()
			}
		}
	}
	cache.syncJarCookies(jar)
	return nil
}

func resolveChaoxingURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errors.New("二次验证地址为空")
	}
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return raw, nil
	}
	if strings.HasPrefix(raw, "/") {
		return "https://passport2.chaoxing.com" + raw, nil
	}
	return "", fmt.Errorf("无法识别的二次验证地址: %s", raw)
}

func (cache *XueXiTUserCache) webSessionClient() (*http.Client, *cookiejar.Jar, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, nil, err
	}
	host, err := url.Parse("https://passport2.chaoxing.com")
	if err != nil {
		return nil, nil, err
	}
	cookies := make([]*http.Cookie, 0, len(cache.cookies))
	for _, cookie := range cache.cookies {
		if cookie == nil {
			continue
		}
		if cookie.Domain == "" || strings.Contains(cookie.Domain, "chaoxing.com") {
			cookies = append(cookies, cookie)
		}
	}
	jar.SetCookies(host, cookies)
	return &http.Client{Jar: jar}, jar, nil
}

func (cache *XueXiTUserCache) getWebPage(ctx context.Context, client *http.Client, urlStr string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", cache.WebUA())
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := readAll(resp)
	if err != nil {
		return nil, err
	}
	cache.captureResponseCookies(resp)
	return body, nil
}

func (cache *XueXiTUserCache) captureResponseCookies(resp *http.Response) {
	if resp == nil {
		return
	}
	utils.CookiesAddNoRepetition(&cache.cookies, resp.Cookies())
}

func (cache *XueXiTUserCache) syncJarCookies(jar *cookiejar.Jar) {
	if jar == nil {
		return
	}
	for _, raw := range []string{
		"https://passport2.chaoxing.com",
		"https://chaoxing.com",
		"https://i.mooc.chaoxing.com",
		"https://mooc1.chaoxing.com",
		"https://mooc1-api.chaoxing.com",
	} {
		u, err := url.Parse(raw)
		if err != nil {
			continue
		}
		cache.captureCookies(jar.Cookies(u))
	}
	cache.refreshCookieHeader()
}

func (cache *XueXiTUserCache) captureCookies(cookies []*http.Cookie) {
	utils.CookiesAddNoRepetition(&cache.cookies, cookies)
}

func (cache *XueXiTUserCache) refreshCookieHeader() {
	parts := make([]string, 0, len(cache.cookies))
	for _, cookie := range cache.cookies {
		if cookie == nil || cookie.Name == "" {
			continue
		}
		parts = append(parts, cookie.Name+"="+cookie.Value)
	}
	cache.cookie = strings.Join(parts, "; ")
}
