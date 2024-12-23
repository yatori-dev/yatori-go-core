package ttcdw

import (
	"errors"
	"fmt"
	"github.com/thedevsaddam/gojsonq"
	"io/ioutil"
	"net/http"
	"strings"
)

type TtcdwUserCache struct {
	PreUrl   string         //前置url
	Account  string         //账号
	Password string         //用户密码
	verCode  string         //验证码
	cookies  []*http.Cookie //验证码用的session
	token    string         //保持会话的Token
	sign     string         //签名
}

func (cache *TtcdwUserCache) TtcdwLoginApi() error {

	url := "https://www.ttcdw.cn/p/uc/userLogin?type=0&pageType=login&service=https%253A%252F%252Fwww.ttcdw.cn"
	method := "POST"

	payload := strings.NewReader("username=" + cache.Account + "&password=" + cache.Password + "&platformId=13145854983311&key=1f0280d0-a000-43b4-8968-aca3a7180b65&service=https%3A%2F%2Fwww.ttcdw.cn")

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return err
	}
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "www.ttcdw.cn")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	//req.Header.Add("Cookie", "HWWAFSESID=c92a3799bef8ba22d2; HWWAFSESTIME=1734968345848")

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
	if gojsonq.New().JSONString(string(body)).Find("success").(bool) != true {
		return errors.New(string(body))
	}
	return nil
}
