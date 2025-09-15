package ketangx

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/yatori-dev/yatori-go-core/utils"
)

type KetangxUserCache struct {
	Account  string         //账号
	Password string         //用户密码
	Cookies  []*http.Cookie //验证码用的session
	UserName string         //用户名称
	UserId   string         //用户ID，目前不知道能有啥用
	Id       string         //不知道啥玩意的ID，不过学时提交要用
	UserUnit string         //单位
}

func (cache *KetangxUserCache) LoginApi() (string, error) {

	url := "https://www.ketangx.cn/Login/AccLogin"
	method := "POST"

	payload := strings.NewReader("userAccount=" + base64.StdEncoding.EncodeToString([]byte(cache.Account)) + "&password=" + base64.StdEncoding.EncodeToString([]byte(cache.Password)) + "&returnUrl=")

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "www.ketangx.cn")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

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
	cache.Cookies = res.Cookies()

	//fmt.Println(string(body))
	return string(body), nil
}

// 获取个人信息接口
func (cache *KetangxUserCache) PullPersonInfoApi() (string, error) {

	//urlStr := "https://www.ketangx.cn/Comment/MyInfo?topicId=fc235f564f7f464a8f9bb34e00e861c0&topicType=2"
	urlStr := "https://www.ketangx.cn/Comment/MyInfo?topicType=2"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, urlStr, nil)

	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	for _, cookie := range cache.Cookies {
		req.AddCookie(cookie)
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "www.ketangx.cn")
	req.Header.Add("Connection", "keep-alive")

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
