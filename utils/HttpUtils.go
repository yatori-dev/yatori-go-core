package utils

import "net/http"

// 添加Cookies，并且是无重复添加，意思就是添加到目标Cookies里面时会检测是否有重复Key的Cookie，如果有则直接替换Cookie值
func CookiesAddNoRepetition(addTarget *[]*http.Cookie, oldTarget []*http.Cookie) {
	for i := range oldTarget {
		flag := false
		for i2 := range *addTarget {
			if oldTarget[i].Name == (*addTarget)[i2].Name {
				(*addTarget)[i2] = oldTarget[i]
				flag = true
				break
			}
		}
		if !flag {
			*addTarget = append(*addTarget, oldTarget[i])
		}
	}
}

// 过滤并获取指定Cookies并返回
func CookiesFiltration(keys []string, cookies []*http.Cookie) []*http.Cookie {
	res := make([]*http.Cookie, 0)
	for i := range cookies {
		for i2 := range keys {
			if cookies[i].Name == keys[i2] {
				res = append(res, cookies[i])
			}
		}
	}
	return res
}
