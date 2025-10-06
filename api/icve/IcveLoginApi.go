package icve

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/yatori-dev/yatori-go-core/utils"
)

type IcveUserCache struct {
	PreUrl         string         //前置url
	Account        string         //账号
	Password       string         //用户密码
	VerCodeRandStr string         //验证码参数
	VerCodeTicket  string         //验证码参数
	Cookies        []*http.Cookie //验证码用的session
	Token          string
	AccessToken    string //智慧职教AccessToken的Token
	ZYKAccessToken string //资源库AccessToken
	sign           string //签名
}

// IcveLoginApi 智慧职教登录后拉取cookie接口
func (cache *IcveUserCache) IcveLoginApi() (string, error) {

	url := "https://sso.icve.com.cn/prod-api/data/userLoginV2"
	method := "POST"
	payload := strings.NewReader(`{` + `"userName":"` + cache.Account + `","randstr":"` + cache.VerCodeRandStr + `","ticket":"` + cache.VerCodeTicket + `","password": "` + cache.Password + `",` + `"type": 1,` + `"webPageSource": 1` + `}`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "sso.icve.com.cn")
	req.Header.Add("Connection", "keep-alive")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	//修改cookie
	utils.CookiesAddNoRepetition(&cache.Cookies, res.Cookies())
	return string(body), err
}

// 拉取UserEncrypt
func (cache *IcveUserCache) IcveUserEncryptApi() (string, error) {
	url := "https://sso.icve.com.cn/prod-api/user/userEncrypt"
	method := "POST"

	payload := strings.NewReader(`{` + `"source":"25","token": "` + cache.Token + `"}`)
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "sso.icve.com.cn")
	req.Header.Add("Connection", "keep-alive")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	utils.CookiesAddNoRepetition(&cache.Cookies, res.Cookies())
	return string(body), err
}

// 拉取Authentication
func (cache *IcveUserCache) IcveAccessTokenApi() (string, error) {

	url := "https://www.icve.com.cn/prod-api/uc/passLogin?token=" + cache.Token
	method := "POST"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "www.icve.com.cn")
	req.Header.Add("Connection", "keep-alive")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	utils.CookiesAddNoRepetition(&cache.Cookies, res.Cookies())
	return string(body), err
}

// 拉取Authentication的另一个接口，一般用这个
func (cache *IcveUserCache) IcveZYKAccessTokenApi() (string, error) {

	url := "https://zyk.icve.com.cn/prod-api/auth/passLogin?token=" + cache.Token
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "www.icve.com.cn")
	req.Header.Add("Connection", "keep-alive")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	utils.CookiesAddNoRepetition(&cache.Cookies, res.Cookies())
	return string(body), err
}
