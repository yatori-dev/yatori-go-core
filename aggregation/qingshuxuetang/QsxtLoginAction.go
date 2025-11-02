package qingshuxuetang

import "github.com/yatori-dev/yatori-go-core/api/qingshuxuetang"

func QsxtLoginAction(cache *qingshuxuetang.QsxtUserCache) {

}

func QsxtCookieLoginAction(cache *qingshuxuetang.QsxtUserCache) {
	cache.Token = cache.Password
}
