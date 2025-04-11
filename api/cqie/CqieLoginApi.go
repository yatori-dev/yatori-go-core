package cqie

import (
	"fmt"
	"github.com/yatori-dev/yatori-go-core/global"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
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
}

func NewCache(input int) CqieUserCache {
	return CqieUserCache{
		Account:  global.Config.Users[input].Account,
		Password: global.Config.Users[input].Password,
	}
}

func NewCacheList() []CqieUserCache {
	var cacheList []CqieUserCache
	for i, user := range global.Config.Users {
		if user.AccountType != "CQIE" {
			continue
		}
		cacheList = append(cacheList, NewCache(i))
	}
	return cacheList
}

var randChar = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f", "A", "B", "C", "D", "E", "F"}

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
