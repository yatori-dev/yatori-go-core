package qingshuxuetang

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type QsxtUserCache struct {
	Account        string         //账号
	Password       string         //用户密码
	VerCodeSession string         //验证码需要的参数
	VerCode        string         //验证码答案
	Cookies        []*http.Cookie //验证码用的session
	Token          string         //保持会话的Token
	sign           string         //签名
	IpProxySW      bool           // 是否开启代理
	ProxyIP        string         //代理IP
}

// 手机端登录接口
func (cache *QsxtUserCache) QsxtPhoneLoginApi() (string, error) {

	url := "https://api.qingshuxuetang.com/v25_10/account/login"
	method := "POST"

	payload := strings.NewReader(`{"name":"` + cache.Account + `","password":"` + cache.Password + `","type":1,"validation":{"sessionId":"` + cache.VerCodeSession + `","type":3,"userInput":"` + cache.VerCode + `"}}`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	req.Header.Add("User-Agent", "okhttp/4.2.2")
	req.Header.Add("Connection", "Keep-Alive")
	req.Header.Add("Authorization-QS", "")
	req.Header.Add("Device-Trace-Id-QS", "b0afcf7e-a8ae-48f2-b438-66982a13dc16")
	req.Header.Add("Device-Info-QS", "{\"appType\":1,\"appVersion\":\"25.10.0\",\"clientType\":2,\"deviceName\":\"xiaomi MI 5X\",\"netType\":1,\"osVersion\":\"8.1.0\"}")
	req.Header.Add("User-Agent-QS", "QSXT")
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "api.qingshuxuetang.com")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	//fmt.Println(string(body))
	return string(body), nil
}

// 拉取验证码
func (cache *QsxtUserCache) QsxtPhoneValidationCodeApi() (string, error) {

	url := "https://api.qingshuxuetang.com/v25_10/account/getValidationCode"
	method := "POST"

	payload := strings.NewReader(`{"recv":"` + cache.Account + `","validationType":3}`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	req.Header.Add("User-Agent", "okhttp/4.2.2")
	req.Header.Add("Connection", "Keep-Alive")
	req.Header.Add("Authorization-QS", "")
	req.Header.Add("Device-Trace-Id-QS", "b0afcf7e-a8ae-48f2-b438-66982a13dc16")
	req.Header.Add("Device-Info-QS", "{\"appType\":1,\"appVersion\":\"25.10.0\",\"clientType\":2,\"deviceName\":\"xiaomi MI 5X\",\"netType\":1,\"osVersion\":\"8.1.0\"}")
	req.Header.Add("User-Agent-QS", "QSXT")
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "api.qingshuxuetang.com")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	//fmt.Println(string(body))
	return string(body), nil
}
