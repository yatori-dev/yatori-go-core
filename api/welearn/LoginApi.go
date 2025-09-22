package welearn

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/yatori-dev/yatori-go-core/utils"
)

type WeLearnCache struct {
	Account  string
	Password string
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

// 登录接口
func (cache *WeLearnCache) WeLearnLoginApi() {

	url := "https://sso.sflep.com/idsvr/account/login"
	method := "POST"
	gePass, ts := GenerateCipherText(cache.Password)
	payload := strings.NewReader("rturl=%2Fconnect%2Fauthorize%2Fcallback%3Fclient_id%3Dwelearn_web%26redirect_uri%3Dhttps%253A%252F%252Fwelearn.sflep.com%252Fsignin-sflep%26response_type%3Dcode%26scope%3Dopenid%2520profile%2520email%2520phone%2520address%26code_challenge%3Dp18_2UckWpdGfknVKQp6Ang64zAYH6__0Z8eQu2uuZE%26code_challenge_method%3DS256%26state%3DOpenIdConnect.AuthenticationProperties%253DBhc1Qn6lYFZrxO_KhC7UzXZTYACtsAnIVT0PgzDlhtuxIXeSFLwXaNbthEeuwSCbzvhrw2wECCxFTq8tbd7k2OFPfH0_TCnMkuh8oBFmlhEsZ3ZXUYecidfT2h2YpAyAoaBaXfpuQj2SGCIEW3KVRYpnljmx-mso97xCbjz72URywiBJRMqDS9TqY-0vaviUIH1X72u_phfuiBdbR1s-WOyUj21KAPdNPJXi1nQtUd-hRoeI53WBTrv2EC0U4SNFvhivPgE6YseB2fdYbPv4u0NiFeHPD3EBQyqE_iUVI1QrGPG3VvhD5xs8odx21WncybewKIuTQpH3MAfJkTmDeQ%26x-client-SKU%3DID_NET472%26x-client-ver%3D6.32.1.0&account=" + cache.Account + "&pwd=" + gePass + "&ts=" + strconv.FormatInt(ts, 10))

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "sso.sflep.com")
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
	fmt.Println(string(body))
}
