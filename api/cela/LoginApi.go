package cela

import (
	"bytes"
	"crypto/aes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/utils"
	log2 "github.com/yatori-dev/yatori-go-core/utils/log"
)

type CelaUserCache struct {
	Account     string //账号
	Password    string //密码
	Cookies     []*http.Cookie
	asuss       string //token
	Code        string //验证码
	LoginParams string //登录用的params
}

// 初始化登录数据接口
func (cache *CelaUserCache) InitLoginDataApi() {

	url := "https://www.cela.gov.cn/home/default?"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "www.cela.gov.cn")
	req.Header.Add("Connection", "keep-alive")

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
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		fmt.Println(err)
	}
	doc.Find(`#accountFrom input[name="params"]`).Each(func(i int, s *goquery.Selection) {
		val, exists := s.Attr("value")
		if exists {
			cache.LoginParams = val
			log2.Print(log2.DEBUG, val)
		}
	})
	cache.Cookies = res.Cookies()
}

var randChar []string = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f", "A", "B", "C", "D", "E", "F"}

// 获取验证码的API
func (cache *CelaUserCache) GetCaptchaApi() (string, error) {

	url := "https://www.cela.gov.cn/home/captcha"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Add("Host", "www.cela.gov.cn")
	req.Header.Add("Connection", "keep-alive")
	//req.Header.Add("Cookie", "SESSION=4545f911-60e8-472f-9a84-0b3012a9daf2")
	for _, cookie := range cache.Cookies {
		req.AddCookie(cookie)
	}

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer res.Body.Close()

	//body, err := ioutil.ReadAll(res.Body)
	//if err != nil {
	//	fmt.Println(err)
	//	return "", err
	//}
	//fmt.Println(string(body))
	utils.CookiesAddNoRepetition(&cache.Cookies, res.Cookies()) //重新设置Cookies

	codeFileName := "code" + randChar[rand.Intn(len(randChar))] //生成验证码文件名称
	for i := 0; i < 10; i++ {
		codeFileName += randChar[rand.Intn(len(randChar))]
	}
	codeFileName += ".png"
	utils.PathExistForCreate("./assets/code/") //检测是否存在路径，如果不存在则创建
	filepath := fmt.Sprintf("./assets/code/%s", codeFileName)
	file, err := os.Create(filepath)
	if err != nil {
		res.Body.Close() //立即释放
		log.Println(err)
		return "", err
	}

	_, err = io.Copy(file, res.Body)
	if err != nil {
		res.Body.Close() //立即释放
		log.Println(err)
		return "", err
	}

	file.Close()
	if utils.IsBadImg(filepath) {
		res.Body.Close()           //立即释放
		utils.DeleteFile(filepath) //删除坏的文件
		return cache.GetCaptchaApi()
	}
	defer res.Body.Close()
	return filepath, nil
}

// 检测验证码是否正确接口
func (cache *CelaUserCache) CheckCaptchaApi() (string, error) {

	url := "https://www.cela.gov.cn/home/captcha?v=" + fmt.Sprintf("%d", time.Now().Unix())
	method := "POST"

	payload := strings.NewReader("captcha=" + cache.Code + "&account=" + cache.Account)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Add("Host", "www.cela.gov.cn")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	//req.Header.Add("Cookie", "SESSION=4545f911-60e8-472f-9a84-0b3012a9daf2")
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
	fmt.Println(string(body))
	utils.CookiesAddNoRepetition(&cache.Cookies, res.Cookies()) //重新设置Cookies
	return string(body), nil
}

// 登录
func (cache *CelaUserCache) LoginApi() string {
	// 1. 获取时间戳（假设是秒级时间戳，跟 JS 保持一致）
	timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
	//timestamp := "1757440611387"
	// 2. 拼接 timestamp + code[:3]
	// 注意 Go 要检查 code 长度，避免越界
	var prefix string
	if len(cache.Code) >= 3 {
		prefix = cache.Code[:3]
	} else {
		prefix = cache.Code
	}
	keyStr := []byte(timestamp + prefix)
	fmt.Println(keyStr)
	formData := map[string]string{
		"account":          cache.Account,
		"password":         cache.Password,
		"verificationCode": cache.Code,
	}
	jsonBytes, _ := json.Marshal(formData)
	//utf8Bytes := []byte(string((jsonBytes)))
	fmt.Printf("%s\n", string(jsonBytes))
	encrypt, err := aesEncryptECB(jsonBytes, keyStr)
	if err != nil {
		fmt.Println(err)
	}
	//
	fmt.Println(base64.StdEncoding.EncodeToString(encrypt))

	urlPath := "https://www.cela.gov.cn/cas/account/check?v=" + timestamp
	method := "POST"

	payload := strings.NewReader("params=" + url.QueryEscape(cache.LoginParams) + "&code=" + cache.Code + "&content=" + url.QueryEscape(base64.StdEncoding.EncodeToString(encrypt)))

	client := &http.Client{}
	req, err := http.NewRequest(method, urlPath, payload)

	if err != nil {
		fmt.Println(err)
		return ""
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Add("Host", "www.cela.gov.cn")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	//req.Header.Add("Cookie", "SESSION=4545f911-60e8-472f-9a84-0b3012a9daf2")
	for _, cookie := range cache.Cookies {
		req.AddCookie(cookie)
	}

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	fmt.Println(string(body))
	find := gojsonq.New().JSONString(string(body)).Find("data")
	if find != nil {
		cache.Cookies = append(cache.Cookies, &http.Cookie{Name: "cela#sso#logged", Value: find.(string)})
	}
	utils.CookiesAddNoRepetition(&cache.Cookies, res.Cookies()) //重新设置Cookies
	return string(body)
}

// 登录后获取数据
func (cache *CelaUserCache) GetLoginAfterData(data string) {

	url1 := "https://www.cela.gov.cn/cas/account/login/process/" + data
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url1, nil)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "www.cela.gov.cn")
	req.Header.Add("Connection", "keep-alive")
	//req.Header.Add("Referer", "https://www.cela.gov.cn/home/redirect/url?ticket=145940318dac11f0a078eb3fe0b5e919")
	for _, cookie := range cache.Cookies {
		req.AddCookie(cookie)
	}

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
	utils.CookiesAddNoRepetition(&cache.Cookies, res.Cookies()) //重新设置Cookies
}

// PKCS7 填充
func pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

// AES-ECB 加密
func aesEncryptECB(src, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	bs := block.BlockSize()
	src = pkcs7Padding(src, bs)
	encrypted := make([]byte, len(src))
	for start := 0; start < len(src); start += bs {
		block.Encrypt(encrypted[start:start+bs], src[start:start+bs])
	}
	return encrypted, nil
}
