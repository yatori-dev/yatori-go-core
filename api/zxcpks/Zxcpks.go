package zxcpks

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

type ZxcpksUserCache struct {
	PreUrl    string //前置url
	Account   string //账号
	Password  string //用户密码
	IpProxySW bool   //是否开启IP代理
	ProxyIP   string //代理IP
	verCode   string //验证码
	cookie    string //验证码用的session
	cookies   []*http.Cookie
	token     string //保持会话的Token
	sign      string //签名
}

func (cache *ZxcpksUserCache) LoginApi() {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://swxy.haiqikeji.com/api/user/login?number="+cache.Account+"&password="+cache.Password+"&schoolId=15", nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("accept-language", "zh-CN,zh;q=0.9,zh-TW;q=0.8,en;q=0.7")
	req.Header.Set("cache-control", "no-cache")
	req.Header.Set("pragma", "no-cache")
	req.Header.Set("priority", "u=1, i")
	req.Header.Set("referer", "https://swxy.haiqikeji.com/student/login")
	req.Header.Set("sec-ch-ua", `"Not:A-Brand";v="99", "Google Chrome";v="145", "Chromium";v="145"`)
	req.Header.Set("sec-ch-ua-mobile", "?1")
	req.Header.Set("sec-ch-ua-platform", `"Android"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-origin")
	req.Header.Set("user-agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Mobile Safari/537.36")
	req.Header.Set("cookie", "__root_domain_v=.haiqikeji.com; _qddaz=QD.247873038960618; _qdda=3-1.1; _qddab=3-vn3xs9.mmitm9t9")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", bodyText)
}
