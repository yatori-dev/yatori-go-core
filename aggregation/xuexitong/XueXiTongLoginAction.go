package xuexitong

import (
	"errors"

	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/api/xuexitong"
	log2 "github.com/yatori-dev/yatori-go-core/utils/log"
)

// XueXiTLoginAction
func XueXiTLoginAction(cache *xuexitong.XueXiTUserCache) error {
	jsonStr, err := cache.LoginApi(5)
	log2.Print(log2.DEBUG, "["+cache.Name+"] "+" 登录成功", jsonStr, err)
	if err != nil {
		return err
	}
	//如果登录成功
	if gojsonq.New().JSONString(jsonStr).Find("status") != nil && gojsonq.New().JSONString(jsonStr).Find("status").(bool) == true {
		return nil
	}
	//如果失败
	if gojsonq.New().JSONString(jsonStr).Find("msg2") != nil {
		return errors.New(gojsonq.New().JSONString(jsonStr).Find("msg2").(string))
	} else {
		return errors.New(jsonStr)
	}
}
