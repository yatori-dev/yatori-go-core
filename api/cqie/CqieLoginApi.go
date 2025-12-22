package cqie

import (
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/yatori-dev/yatori-go-core/utils"
)

type CqieUserCache struct {
	studentId    string //信息id
	Account      string //账号
	Password     string //用户密码
	verCode      string //验证码
	cookie       string //验证码用的session
	uuid         string //验证码uuid
	access_token string // token
	token        string //token
	userId       string //userId
	userName     string //用户名称
	appId        string //appId不知道啥玩意
	ipaddr       string //这玩意还会记录IP？
	deptId       string //不知道啥玩意
	mobile       string //手机号
	orgId        string //不知道啥玩意
	orgMajorId   string //专业Id
	IpProxySW    bool   // 是否开启代理
	ProxyIP      string //代理IP
	Version      string //平台版本
}

var randChar = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f", "A", "B", "C", "D", "E", "F"}

// VerificationCodeApi 获取验证码
func (cache *CqieUserCache) VerificationCodeApi() (string, string) {
	uuid := uuid.New() //生成验证码UUID
	cache.uuid = uuid.String()
	urlStr := "https://study.cqie.edu.cn/gateway/auth/createCaptcha?uuid=" + uuid.String()
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
		return "", ""
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", ""
	}
	defer res.Body.Close()

	codeFileName := "code" + randChar[rand.Intn(len(randChar))] //生成验证码文件名称
	for i := 0; i < 10; i++ {
		codeFileName += randChar[rand.Intn(len(randChar))]
	}
	codeFileName += ".png"
	utils.PathExistForCreate("./assets/code/") //检测是否存在路径，如果不存在则创建
	filepath := fmt.Sprintf("./assets/code/%s", codeFileName)
	file, err := os.Create(filepath)
	if err != nil {
		log.Println(err)
		return "", ""
	}
	defer file.Close()

	_, err = io.Copy(file, res.Body)
	if err != nil {
		log.Println(err)
		return "", ""
	}
	return filepath, res.Header.Get("Set-Cookie")
}

// LoginApi 登录API
func (cache *CqieUserCache) LoginApi() (string, error) {

	urlStr := "https://study.cqie.edu.cn/gateway/auth/login"
	method := "POST"
	acc := fmt.Sprintf("%x", utils.CqieEncrypt(cache.Account))
	pass := fmt.Sprintf("%x", utils.CqieEncrypt(cache.Password))
	payload := strings.NewReader(`{ "account": "` + acc + `", "password": "` + pass + `","code": "` + cache.verCode + `","status": 0,` + `"uuid": "` + cache.uuid + `","student": true}`)
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
	return string(body), nil
}
