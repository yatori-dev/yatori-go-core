package utils

import "net/http"

// 添加Cookies，并且是无重复添加，意思就是添加到目标Cookies里面时会检测是否有重复Key的Cookie，如果有则直接替换Cookie值
func CookiesAddNoRepetition(addTarget []*http.Cookie, oldTarget []*http.Cookie) {
	//替换cookie
	for i := range oldTarget {
		for i2 := range addTarget {
			if oldTarget[i].Name == addTarget[i2].Name {
				addTarget[i2] = oldTarget[i]
				continue
			}
			// 如果没有对应Key的Cookie则直接添加
			addTarget = append(addTarget, oldTarget[i])
		}
	}
}
