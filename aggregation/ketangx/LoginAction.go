package ketangx

import (
	"fmt"

	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/api/ketangx"
)

// 登录
func LoginAction(cache *ketangx.KetangxUserCache) error {
	api, err := cache.LoginApi()
	if err != nil {
		return err
	}
	infoApi, err := cache.PullPersonInfoApi()
	if err != nil {
		return err
	}
	userId := gojsonq.New().JSONString(infoApi).Find("UserId")
	if userId != nil {
		cache.UserId = userId.(string)
	}
	userName := gojsonq.New().JSONString(infoApi).Find("UserName")
	if userName != nil {
		cache.UserName = userName.(string)
	}
	userUnit := gojsonq.New().JSONString(infoApi).Find("UserUnit")
	if userUnit != nil {
		cache.UserUnit = userUnit.(string)
	}
	id := gojsonq.New().JSONString(infoApi).Find("Id")
	if id != nil {
		cache.Id = id.(string)
	}
	fmt.Println(api)
	return nil
}
