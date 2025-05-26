package xuexitong

import (
	"errors"
	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/api/xuexitong"
	log2 "github.com/yatori-dev/yatori-go-core/utils/log"
)

// XueXiTLoginAction 学习通登录Action
func XueXiTLoginAction(cache *xuexitong.XueXiTUserCache) error {
	jsonStr, err := cache.LoginApi()
	log2.Print(log2.DEBUG, "["+cache.Name+"] "+" 登录成功", jsonStr, err)
	if err != nil {
		if gojsonq.New().JSONString(jsonStr).Find("msg2") != nil {
			return errors.New(gojsonq.New().JSONString(jsonStr).Find("msg2").(string))
		} else {
			return err
		}

	}
	return nil
}
