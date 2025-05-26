package icve

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type IcveUserCache struct {
	PreUrl   string         //前置url
	Account  string         //账号
	Password string         //用户密码
	verCode  string         //验证码
	cookies  []*http.Cookie //验证码用的session
	token    string         //保持会话的Token
	sign     string         //签名
}

// IcveLoginApi 智慧职教登录接口
func (cache *IcveUserCache) IcveLoginApi() error {

	url := "https://sso.icve.com.cn/prod-api/data/userLoginV2"
	method := "POST"

	payload := strings.NewReader(`{` + `"userName":"` + cache.Account + `",` + `"password": "` + cache.Password + `",` + `"type": 1,` + `"webPageSource": 1` + `}`)
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return err
	}
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "sso.icve.com.cn")
	req.Header.Add("Connection", "keep-alive")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return err
	}
	cache.cookies = res.Cookies()

	fmt.Println(string(body))
	return nil
}
