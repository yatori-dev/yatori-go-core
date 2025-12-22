package icve

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/yatori-dev/yatori-go-core/utils"
)

type IcveUserCache struct {
	PreUrl         string //前置url
	Account        string //账号
	Password       string //用户密码
	IpProxySW      bool
	ProxyIP        string
	VerCodeRandStr string         //验证码参数
	VerCodeTicket  string         //验证码参数
	Cookies        []*http.Cookie //验证码用的session
	Token          string
	AccessToken    string //智慧职教AccessToken的Token
	ZYKAccessToken string //资源库AccessToken
	UserId         string //用户ID
	NickName       string //用户名称
	PhoneNumber    string //电话号码
	Sex            string //性别
}

// IcveLoginApi 智慧职教登录后拉取cookie接口
func (cache *IcveUserCache) IcveLoginApi() (string, error) {

	urlStr := "https://sso.icve.com.cn/prod-api/data/userLoginV2"
	method := "POST"
	payload := strings.NewReader(`{` + `"userName":"` + cache.Account + `","randstr":"` + cache.VerCodeRandStr + `","ticket":"` + cache.VerCodeTicket + `","password": "` + cache.Password + `",` + `"type": 1,` + `"webPageSource": 1` + `}`)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}

	//如果开启了IP代理，那么就直接添加代理
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse(cache.ProxyIP) // 设置代理
		}
	}
	client := &http.Client{
		Transport: tr,
	}
	req, err := http.NewRequest(method, urlStr, payload)

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
	urlStr := "https://sso.icve.com.cn/prod-api/user/userEncrypt"
	method := "POST"

	payload := strings.NewReader(`{` + `"source":"25","token": "` + cache.Token + `"}`)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}

	//如果开启了IP代理，那么就直接添加代理
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse(cache.ProxyIP) // 设置代理
		}
	}
	client := &http.Client{
		Transport: tr,
	}
	req, err := http.NewRequest(method, urlStr, payload)

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

	urlStr := "https://www.icve.com.cn/prod-api/uc/passLogin?token=" + cache.Token
	method := "POST"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}

	//如果开启了IP代理，那么就直接添加代理
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse(cache.ProxyIP) // 设置代理
		}
	}
	client := &http.Client{
		Transport: tr,
	}
	req, err := http.NewRequest(method, urlStr, nil)

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

// 拉取资源课Authentication的接口
func (cache *IcveUserCache) IcveZYKAccessTokenApi() (string, error) {

	urlStr := "https://zyk.icve.com.cn/prod-api/auth/passLogin?token=" + cache.Token
	method := "GET"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}

	//如果开启了IP代理，那么就直接添加代理
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse(cache.ProxyIP) // 设置代理
		}
	}
	client := &http.Client{
		Transport: tr,
	}
	req, err := http.NewRequest(method, urlStr, nil)

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

// 拉取用户信息
func (cache *IcveUserCache) IcveZYKPullUserInfoApi() (string, error) {

	urlStr := "https://zyk.icve.com.cn/prod-api/system/user/getInfo"
	method := "GET"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}

	//如果开启了IP代理，那么就直接添加代理
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse(cache.ProxyIP) // 设置代理
		}
	}
	client := &http.Client{
		Transport: tr,
	}
	req, err := http.NewRequest(method, urlStr, nil)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "www.icve.com.cn")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Authorization", "Bearer "+cache.ZYKAccessToken)
	for _, cookie := range cache.Cookies {
		req.AddCookie(cookie)
	}

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
