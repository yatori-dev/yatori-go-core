package qingshuxuetang

import "net/http"

type QsxtUserCache struct {
	Account   string         //账号
	Password  string         //用户密码
	verCode   string         //验证码
	Cookies   []*http.Cookie //验证码用的session
	Token     string         //保持会话的Token
	sign      string         //签名
	IpProxySW bool           // 是否开启代理
	ProxyIP   string         //代理IP
}

// 登录接口
func (cache *QsxtUserCache) QsxtLoginApi() {

}
