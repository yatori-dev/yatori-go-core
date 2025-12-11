package utils

import (
	"net/http"
	"strings"
)

// 常用的User-Agent
const (
	DefaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36 Edg/143.0.0.0"
)

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

// TurnCookiesFromString 将一整串 cookie 字符串解析成*http.Cookie数组并返回
func TurnCookiesFromString(cookieStr string) []*http.Cookie {
	var cookies = []*http.Cookie{}
	parts := strings.Split(cookieStr, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part) // 去掉前后空格
		if part == "" {
			continue
		}

		kv := strings.SplitN(part, "=", 2) // 拆成 key=value
		if len(kv) != 2 {
			continue
		}

		cookie := &http.Cookie{
			Name:  kv[0],
			Value: kv[1],
		}
		cookies = append(cookies, cookie)
	}
	return cookies
}
