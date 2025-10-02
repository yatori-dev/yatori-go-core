package ttcdw

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/utils"
)

type TtcdwUserCache struct {
	PreUrl    string         //前置url
	Account   string         //账号
	Password  string         //用户密码
	verCode   string         //验证码
	cookies   []*http.Cookie //验证码用的session
	token     string         //保持会话的Token
	sign      string         //签名
	IpProxySW bool           // 是否开启代理
	ProxyIP   string         //代理IP
}

// TtcdwLoginApi TTCDW学习公社登录
func (cache *TtcdwUserCache) TtcdwLoginApi() error {

	url := "https://www.ttcdw.cn/p/uc/userLogin?type=0&pageType=login&service=https%253A%252F%252Fwww.ttcdw.cn"
	method := "POST"

	payload := strings.NewReader("username=" + cache.Account + "&password=" + fmt.Sprintf("%x", md5.Sum([]byte(cache.Password))) + "&platformId=13145854983311&key=1f0280d0-a000-43b4-8968-aca3a7180b65&service=https%3A%2F%2Fwww.ttcdw.cn")

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return err
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "www.ttcdw.cn")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

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
