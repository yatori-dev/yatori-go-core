package welearn

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/yatori-dev/yatori-go-core/utils"
)

type WeLearnUserCache struct {
	Account  string
	Password string
	Cookies  []*http.Cookie
}

// 生成加密后的密码和时间戳
func GenerateCipherText(password string) (string, int64) {
	// 当前时间戳（毫秒）
	t0 := time.Now().UnixMilli()

	// 初始 v
	v := byte((t0 >> 16) & 0xFF)

	// 遍历密码字节，逐个异或
	pBytes := []byte(password)
	for _, b := range pBytes {
		v ^= b
	}

	remainder := int64(v % 100)
	t1 := (t0/100)*100 + remainder

	// 转换为十六进制字符串
	p1 := hex.EncodeToString(pBytes)

	// 拼接 s
	s := fmt.Sprintf("%d*%s", t1, p1)

	// Base64 编码
	encrypted := base64.StdEncoding.EncodeToString([]byte(s))

	return encrypted, t1
}

//func (cache *WeLearnUserCache) WeLearnSSOLoginApi(cid string) (string, int64) {
//
//}

// 登录接口
func (cache *WeLearnUserCache) WeLearnLoginApi(retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	urlStr := "https://sso.sflep.com/idsvr/account/login"
	method := "POST"
	gePass, ts := GenerateCipherText(cache.Password)
	payload := strings.NewReader("rturl=%2Fconnect%2Fauthorize%2Fcallback%3Fclient_id%3Dwelearn_web%26redirect_uri%3Dhttps%253A%252F%252Fwelearn.sflep.com%252Fsignin-sflep%26response_type%3Dcode%26scope%3Dopenid%2520profile%2520email%2520phone%2520address%26code_challenge%3Dp18_2UckWpdGfknVKQp6Ang64zAYH6__0Z8eQu2uuZE%26code_challenge_method%3DS256%26state%3DOpenIdConnect.AuthenticationProperties%253DBhc1Qn6lYFZrxO_KhC7UzXZTYACtsAnIVT0PgzDlhtuxIXeSFLwXaNbthEeuwSCbzvhrw2wECCxFTq8tbd7k2OFPfH0_TCnMkuh8oBFmlhEsZ3ZXUYecidfT2h2YpAyAoaBaXfpuQj2SGCIEW3KVRYpnljmx-mso97xCbjz72URywiBJRMqDS9TqY-0vaviUIH1X72u_phfuiBdbR1s-WOyUj21KAPdNPJXi1nQtUd-hRoeI53WBTrv2EC0U4SNFvhivPgE6YseB2fdYbPv4u0NiFeHPD3EBQyqE_iUVI1QrGPG3VvhD5xs8odx21WncybewKIuTQpH3MAfJkTmDeQ%26x-client-SKU%3DID_NET472%26x-client-ver%3D6.32.1.0&account=" + cache.Account + "&pwd=" + gePass + "&ts=" + strconv.FormatInt(ts, 10))

	client := &http.Client{}
	req, err := http.NewRequest(method, urlStr, payload)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "sso.sflep.com")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

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
	//fmt.Println(string(body))
	return string(body), nil
}

// 处理登录SSO回调
func (cache *WeLearnUserCache) WeLearnLoginSsoCallApi(retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}

	callbackParams := url.Values{}
	callbackParams.Set("client_id", "welearn_web")
	callbackParams.Set("redirect_uri", "https://welearn.sflep.com/signin-sflep")
	callbackParams.Set("response_type", "code")
	callbackParams.Set("scope", "openid profile email phone address")
	callbackParams.Set("code_challenge", "p18_2UckWpdGfknVKQp6Ang64zAYH6__0Z8eQu2uuZE")
	callbackParams.Set("code_challenge_method", "S256")
	callbackParams.Set("state", "OpenIdConnect.AuthenticationProperties=Bhc1Qn6lYFZrxO_KhC7UzXZTYACtsAnIVT0PgzDlhtuxIXeSFLwXaNbthEeuwSCbzvhrw2wECCxFTq8tbd7k2OFPfH0_TCnMkuh8oBFmlhEsZ3ZXUYecidfT2h2YpAyAoaBaXfpuQj2SGCIEW3KVRYpnljmx-mso97xCbjz72URywiBJRMqDS9TqY-0vaviUIH1X72u_phfuiBdbR1s-WOyUj21KAPdNPJXi1nQtUd-hRoeI53WBTrv2EC0U4SNFvhivPgE6YseB2fdYbPv4u0NiFeHPD3EBQyqE_iUVI1QrGPG3VvhD5xs8odx21WncybewKIuTQpH3MAfJkTmDeQ")
	callbackParams.Set("x-client-SKU", "ID_NET472")
	callbackParams.Set("x-client-ver", "6.32.1.0")
	urlStr := "https://sso.sflep.com/idsvr/connect/authorize/callback?" + callbackParams.Encode()
	method := "GET"

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequest(method, urlStr, nil)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	//req.Header.Add("Referer", "https://welearn.sflep.com/student/index.aspx")
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "sso.sflep.com")
	req.Header.Add("sec-ch-ua-platform", "Windows")
	req.Header.Add("Referer", "https://sso.sflep.com/idsvr/login.html")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Origin", "https://sso.sflep.com")
	for _, cookie := range cache.Cookies {
		req.AddCookie(cookie)
	}

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer res.Body.Close()

	utils.CookiesAddNoRepetition(&cache.Cookies, res.Cookies())

	client1 := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req1, err1 := http.NewRequest(method, res.Header.Get("Location"), nil)

	if err1 != nil {
		fmt.Println(err1)
		return "", err1
	}
	req1.Header.Add("User-Agent", utils.DefaultUserAgent)
	req1.Header.Add("Accept", "*/*")
	req1.Header.Add("Host", "sso.sflep.com")
	req1.Header.Add("sec-ch-ua-platform", "Windows")
	req1.Header.Add("Referer", "https://sso.sflep.com/idsvr/login.html")
	req1.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req1.Header.Add("Origin", "https://sso.sflep.com")
	for _, cookie := range cache.Cookies {
		req1.AddCookie(cookie)
	}

	res1, err1 := client1.Do(req1)
	if err1 != nil {
		fmt.Println(err1)
		return "", err
	}
	defer res1.Body.Close()

	utils.CookiesAddNoRepetition(&cache.Cookies, res1.Cookies())

	client2 := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req2, err2 := http.NewRequest(method, "https://welearn.sflep.com/user/loginredirect.aspx", nil)

	if err2 != nil {
		fmt.Println(err2)
		return "", err2
	}

	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "sso.sflep.com")
	req.Header.Add("sec-ch-ua-platform", "Windows")
	req.Header.Add("Referer", "https://sso.sflep.com/idsvr/login.html")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Origin", "https://sso.sflep.com")
	for _, cookie := range cache.Cookies {
		req2.AddCookie(cookie)
	}

	res2, err2 := client2.Do(req2)
	if err2 != nil {
		fmt.Println(err2)
		return "", err2
	}
	defer res2.Body.Close()

	body2, err2 := ioutil.ReadAll(res2.Body)
	if err2 != nil {
		fmt.Println(err2)
		return "", err2
	}
	utils.CookiesAddNoRepetition(&cache.Cookies, res2.Cookies())
	return string(body2), nil
}
