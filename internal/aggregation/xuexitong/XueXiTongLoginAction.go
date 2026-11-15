package xuexitong

import (
	"context"
	"errors"

	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/v3/internal/api/xuexitong"
	"github.com/yatori-dev/yatori-go-core/v3/utils"
	log2 "github.com/yatori-dev/yatori-go-core/v3/utils/log"
)

// XueXiTLoginAction 兼容旧调用，内部使用 Background context。
func XueXiTLoginAction(cache *xuexitong.XueXiTUserCache) error {
	return XueXiTLoginActionContext(context.Background(), cache)
}

// XueXiTLoginActionContext 执行网页端登录。
//
// fanyalogin 在账号需要短信二次验证时会返回 status:true 且
// containTwoFactorLogin:true，此时不能直接当作登录成功，必须继续完成
// /v11/checkmessage 的短信验证。
func XueXiTLoginActionContext(ctx context.Context, cache *xuexitong.XueXiTUserCache) error {
	jsonStr, err := cache.LoginApi(5)
	log2.Print(log2.DEBUG, "["+cache.Name+"] 登录返回", jsonStr, err)
	if err != nil {
		return err
	}

	if twoFactor, ok := gojsonq.New().JSONString(jsonStr).Find("containTwoFactorLogin").(bool); ok && twoFactor {
		urlStr, _ := gojsonq.New().JSONString(jsonStr).Find("twoFactorLoginUrl").(string)
		if urlStr == "" {
			urlStr, _ = gojsonq.New().JSONString(jsonStr).Find("twoFactorLoginPCUrl").(string)
		}
		if urlStr == "" {
			return errors.New("网页登录需要短信二次验证，但响应中没有 twoFactorLoginUrl")
		}
		return cache.CompleteWebTwoFactorLogin(ctx, urlStr)
	}

	if status, ok := gojsonq.New().JSONString(jsonStr).Find("status").(bool); ok && status {
		return nil
	}
	if msg2 := gojsonq.New().JSONString(jsonStr).Find("msg2"); msg2 != nil {
		if message, ok := msg2.(string); ok {
			return errors.New(message)
		}
	}
	return errors.New(jsonStr)
}

// Cookie 登录取 cache.Password 里的 Cookie 字符串。
func XueXiTCookieLoginAction(cache *xuexitong.XueXiTUserCache) error {
	cookies := utils.TurnCookiesFromString(cache.Password)
	cache.SetCookies(cookies)
	return nil
}
