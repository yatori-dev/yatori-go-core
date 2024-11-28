package cqie

import (
	"fmt"
	"github.com/Yatori-Dev/yatori-go-core/utils"
	"github.com/google/uuid"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
)

type CqieUserCache struct {
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
}

var randChar []string = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f", "A", "B", "C", "D", "E", "F"}

func (cache *CqieUserCache) GetCookie() string {
	return cache.cookie
}
func (cache *CqieUserCache) SetCookie(cookie string) {
	cache.cookie = cookie
}

func (cache *CqieUserCache) GetVerCode() string {
	return cache.verCode
}
func (cache *CqieUserCache) SetVerCode(verCode string) {
	cache.verCode = verCode
}

func (cache *CqieUserCache) GetAccess_Token() string {
	return cache.access_token
}
func (cache *CqieUserCache) SetAccess_Token(access_token string) {
	cache.access_token = access_token
}

func (cache *CqieUserCache) GetToken() string {
	return cache.token
}
func (cache *CqieUserCache) SetToken(token string) {
	cache.token = token
}

func (cache *CqieUserCache) GetUserId() string {
	return cache.userId
}
func (cache *CqieUserCache) SetUserId(userId string) {
	cache.userId = userId
}

func (cache *CqieUserCache) GetAppId() string {
	return cache.appId
}
func (cache *CqieUserCache) SetAppId(appId string) {
	cache.appId = appId
}

func (cache *CqieUserCache) GetIpaddr() string {
	return cache.ipaddr
}
func (cache *CqieUserCache) SetIpaddr(ipaddr string) {
	cache.ipaddr = ipaddr
}

func (cache *CqieUserCache) GetDeptId() string {
	return cache.deptId
}
func (cache *CqieUserCache) SetDeptId(deptId string) {
	cache.deptId = deptId
}

// VerificationCodeApi 获取验证码
func (cache *CqieUserCache) VerificationCodeApi() (string, string) {
	uuid := uuid.New() //生成验证码UUID
	cache.uuid = uuid.String()
	url := "https://study.cqie.edu.cn/gateway/auth/createCaptcha?uuid=" + uuid.String()
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return "", ""
	}
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")

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

	url := "https://study.cqie.edu.cn/gateway/auth/login"
	method := "POST"
	acc := fmt.Sprintf("%x", utils.CqieEncrypt(cache.Account))
	pass := fmt.Sprintf("%x", utils.CqieEncrypt(cache.Password))
	payload := strings.NewReader(`{ "account": "` + acc + `", "password": "` + pass + `","code": "` + cache.verCode + `","status": 0,` + `"uuid": "` + cache.uuid + `","student": true}`)
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")
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
