package zhihuishu

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/thedevsaddam/gojsonq"
)

func GetValidate() string {
	return ""
}

// 通过用户账号和密码获取SecretStr值
func GenerateSecretStr(username, password string) (string, error) {
	validate := GetValidate()
	if validate == "" {
		return "", fmt.Errorf("validate is empty")
	}
	// 创建参数
	params := map[string]string{
		"account":  username,
		"password": password,
		"validate": validate,
	}

	// 将参数转换为JSON
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return "", err
	}
	// URL编码
	encodedParams := url.QueryEscape(string(paramsJSON))

	// 替换特定字符
	encodedParams = strings.ReplaceAll(encodedParams, "%3A", ":")
	encodedParams = strings.ReplaceAll(encodedParams, "%2C", ",")

	// Base64编码
	encodedStr := base64.StdEncoding.EncodeToString([]byte(encodedParams))
	return encodedStr, nil
}

// 知到扫码登录拉取二维码
func ZhidaoQrCode() (string, string, error) {

	url := "https://passport.zhihuishu.com/qrCodeLogin/getLoginQrImg"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return "", "", err
	}
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "passport.zhihuishu.com")
	req.Header.Add("Connection", "keep-alive")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", "", nil
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", "", nil
	}
	imgBase64 := gojsonq.New().JSONString(string(body)).Find("img").(string)
	qrToken := gojsonq.New().JSONString(string(body)).Find("qrToken").(string)
	return imgBase64, qrToken, nil
}

// 知到登录扫码成功扫描函数
func ZhidaoQrCheck(qrToken string) (string, error) {

	url := "https://passport.zhihuishu.com/qrCodeLogin/getLoginQrInfo?qrToken=" + qrToken
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "passport.zhihuishu.com")
	req.Header.Add("Connection", "keep-alive")
	//req.Header.Add("Cookie", "acw_tc=ac11000117561157012271282ee399905d1a14eaef7d65da0950cf4402dfbc; INGRESSCOOKIE=1756115702.239.53.818564; SERVERID=472b148b148a839eba1c5c1a8657e3a7|1756115701|1756115701")

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
	//fmt.Println(string(body))
	return string(body), err
}
