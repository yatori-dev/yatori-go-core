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
	token    string         //保持会话的Token
}

func (cache *KetangxUserCache) LoginApi() {

	url := "https://www.ketangx.cn/Login/AccLogin"
	method := "POST"

	payload := strings.NewReader("userAccount=" + base64.StdEncoding.EncodeToString([]byte(cache.Account)) + "&password=" + base64.StdEncoding.EncodeToString([]byte(cache.Password)) + "&returnUrl=")

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "www.ketangx.cn")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	cache.Cookies = res.Cookies()

	fmt.Println(string(body))
}
