package haiqikeji

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/api/haiqikeji"
)

// 登录
func HqkjLoginAction(cache *haiqikeji.HqkjUserCache) error {
	hostname, _ := extractDomain(cache.PreUrl)
	schoolInfoStr := cache.PullSchoolInfoApi(hostname, 3, nil)
	cache.SchoolId = strconv.Itoa(int(gojsonq.New().JSONString(schoolInfoStr).Find("data.id").(float64)))
	loginResult, err := cache.LoginApi(3, nil)
	if err != nil {
		return err
	}

	code := int(gojsonq.New().JSONString(loginResult).Find("code").(float64))
	//登录失败处理
	if code != 200 {
		return fmt.Errorf("登录失败：%s", loginResult)
	}
	cache.Token = gojsonq.New().JSONString(loginResult).Find("data").(string) //登陆成功赋值Token
	//fmt.Println(loginResult)
	userInfo, err := cache.PullUserInfoApi(3, nil)
	cache.UserId = strconv.Itoa(int(gojsonq.New().JSONString(userInfo).Find("data.id").(float64))) //赋值userId
	return nil
}

func extractDomain(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return u.Hostname(), nil
}
