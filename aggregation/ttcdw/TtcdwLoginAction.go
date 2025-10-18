package ttcdw

import (
	"errors"

	"github.com/thedevsaddam/gojsonq"
	ttcdwApi "github.com/yatori-dev/yatori-go-core/api/ttcdw"
	"github.com/yatori-dev/yatori-go-core/utils"
)

// 登录
func TTCDWLoginAction(cache *ttcdwApi.TtcdwUserCache) error {
	loginResult, err := cache.TtcdwLoginApi() //登录账号
	if err != nil {
		return err
	}
	if gojsonq.New().JSONString(loginResult).Find("success").(bool) != true {
		return errors.New(loginResult)
	}
	return nil
}

// Cookie登录方式
func TTCDWCookieLoginAction(cache *ttcdwApi.TtcdwUserCache) error {
	cache.Cookies = utils.TurnCookiesFromString(cache.Password)
	return nil
}
