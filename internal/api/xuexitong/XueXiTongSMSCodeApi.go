package xuexitong

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/yatori-dev/yatori-go-core/v3/internal/ocr"
)

const loginSMSCaptchaAttempts = 8

// sendLoginSMSCode 触发 checkmessage 页面的“获取验证码”。
//
// 页面脚本实际流程：
//
//	POST /num/booleanCode?key=<phone>&type=1      // 风控预检
//	GET  /num/code?<timestamp>                    // 需要验证码时的数字图片
//	GET  /num/phonecode?phone=...&code=<ocr>&type=1&countrycode=...&validate=
//
// booleanCode 返回 "alert" 表示当天发送次数超限，返回 "true" 表示需要
// 图片验证码。图片验证码由内置 ddddocr 识别后自动重试，不会打扰调用方；
// 调用方只需要在用户确认后触发 SendSMS，再输入手机收到的短信验证码。
func sendLoginSMSCode(ctx context.Context, client *http.Client, phone, countryCode, userAgent, referer string) error {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return errors.New("发送短信验证码失败：手机号为空")
	}
	countryCode = strings.TrimSpace(countryCode)
	if countryCode == "" {
		countryCode = "86"
	}

	needCaptcha, err := precheckLoginSMSCode(ctx, client, phone, userAgent, referer)
	if err != nil {
		return err
	}
	if !needCaptcha {
		return requestLoginSMSCode(ctx, client, phone, countryCode, "", userAgent, referer)
	}

	var lastErr error
	for attempt := 0; attempt < loginSMSCaptchaAttempts; attempt++ {
		captchaCode, captchaErr := fetchLoginCaptchaCode(ctx, client, userAgent, referer)
		if captchaErr != nil {
			lastErr = captchaErr
			continue
		}
		if captchaCode == "" {
			lastErr = errors.New("图片验证码识别为空")
			continue
		}
		sendErr := requestLoginSMSCode(ctx, client, phone, countryCode, captchaCode, userAgent, referer)
		if sendErr == nil {
			return nil
		}
		lastErr = sendErr
		if !strings.Contains(sendErr.Error(), "验证码") {
			return sendErr
		}
	}
	if lastErr == nil {
		lastErr = errors.New("发送短信验证码失败")
	}
	return fmt.Errorf("图片验证码重试 %d 次仍未通过: %w", loginSMSCaptchaAttempts, lastErr)
}

// precheckLoginSMSCode 复刻页面的 /num/booleanCode 风控预检。
func precheckLoginSMSCode(ctx context.Context, client *http.Client, phone, userAgent, referer string) (bool, error) {
	query := url.Values{}
	query.Set("key", phone)
	query.Set("type", "1")
	endpoint := "https://passport2.chaoxing.com/num/booleanCode?" + query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	applySMSCodeHeaders(req, userAgent, referer)
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("短信发送预检失败: %w", err)
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	_ = resp.Body.Close()
	switch strings.TrimSpace(string(body)) {
	case "alert":
		return false, errors.New("发送短信验证码失败：超出今日短信发送上限")
	case "true":
		return true, nil
	}
	return false, nil
}

// fetchLoginCaptchaCode 下载 /num/code 的数字图片验证码并用 ddddocr 识别。
func fetchLoginCaptchaCode(ctx context.Context, client *http.Client, userAgent, referer string) (string, error) {
	endpoint := "https://passport2.chaoxing.com/num/code?" + strconv.FormatInt(time.Now().UnixMilli(), 10)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	applySMSCodeHeaders(req, userAgent, referer)
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("获取图片验证码失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("获取图片验证码失败: HTTP %d", resp.StatusCode)
	}
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return "", fmt.Errorf("解析图片验证码失败: %w", err)
	}
	return strings.TrimSpace(ocr.Solve(img)), nil
}

// requestLoginSMSCode 调用 /num/phonecode 发送短信。captchaCode 是图片
// 验证码；预检不需要验证码时传空字符串。
func requestLoginSMSCode(ctx context.Context, client *http.Client, phone, countryCode, captchaCode, userAgent, referer string) error {
	query := url.Values{}
	query.Set("phone", phone)
	query.Set("code", captchaCode)
	query.Set("type", "1")
	query.Set("countrycode", countryCode)
	query.Set("validate", "")
	endpoint := "https://passport2.chaoxing.com/num/phonecode?" + query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	applySMSCodeHeaders(req, userAgent, referer)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送短信验证码失败: %w", err)
	}
	body, _ := readAll(resp)
	_ = resp.Body.Close()

	var result struct {
		Result bool   `json:"result"`
		Msg    string `json:"msg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("发送短信验证码响应异常: %s", appBodySummary(body))
	}
	if !result.Result {
		message := strings.TrimSpace(result.Msg)
		if message == "" {
			message = appBodySummary(body)
		}
		return fmt.Errorf("发送短信验证码失败: %s", message)
	}
	return nil
}

func applySMSCodeHeaders(req *http.Request, userAgent, referer string) {
	if strings.TrimSpace(userAgent) == "" {
		userAgent = appLoginUserAgent
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Origin", "https://passport2.chaoxing.com")
	if strings.TrimSpace(referer) != "" {
		req.Header.Set("Referer", referer)
	}
}
