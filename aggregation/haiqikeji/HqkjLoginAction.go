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
	schoolInfoStr := cache.PullSchoolInfoApi(hostname, 5, nil)

	if schoolId, ok := gojsonq.New().JSONString(schoolInfoStr).Find("data.id").(float64); ok {
		cache.SchoolId = strconv.Itoa(int(schoolId))
	} else {
		return fmt.Errorf("未找到url对应学校id，请检查url是否填写正确(注：有些学校可能会登录后才会显示正真的域名)：%s", schoolInfoStr)
	}
	loginResult, err := cache.LoginApi(5, nil)
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
	userInfo, err := cache.PullUserInfoApi(5, nil)
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
