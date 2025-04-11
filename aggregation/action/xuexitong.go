package action

import log2 "github.com/yatori-dev/yatori-go-core/utils/log"

func (y YatoriCache) LoginAction() interface{} {
	_, err := y.XueXiTUserCache.LoginApi()
	if err == nil {
		log2.Print(log2.INFO, "["+y.XueXiTUserCache.Name+"] "+" 登录成功")
	}
	return "登录成功"
}
