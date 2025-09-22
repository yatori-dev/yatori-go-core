package welearn

import (
	"errors"

	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/api/welearn"
	"github.com/yatori-dev/yatori-go-core/utils"
)

// 登录
func WeLearnLoginAction(cache *welearn.WeLearnUserCache) error {
	loginJson, err := cache.WeLearnLoginApi(3, nil)
	if err != nil {
		return err
	}
	msg := gojsonq.New().JSONString(loginJson).Find("msg")
	if msg == nil {
		return errors.New(loginJson)
	}
	if msg.(string) != "OK" {
		return errors.New(loginJson)
	}
	return nil
}

// Cookie登录方式，方便测试
func WeLearnCookieLoginAction(cache *welearn.WeLearnUserCache, cookies string) {
	cache.Cookies = utils.TurnCookiesFromString(cookies)
}
