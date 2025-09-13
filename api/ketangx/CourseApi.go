package ketangx

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/yatori-dev/yatori-go-core/utils"
)

func (cache *KetangxUserCache) PullCourse() (string, error) {
	url := "https://www.ketangx.cn/Activity/Query"
	method := "POST"

	payload := strings.NewReader("actType=2&actStart=&actClose=&formId=&classId=&actKey=&actState=&timeId=" + fmt.Sprintf("%d", time.Now().UnixMilli()))

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "www.ketangx.cn")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	for _, cookie := range cache.Cookies {
		req.AddCookie(cookie)
	}

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	cache.Cookies = res.Cookies()

	//fmt.Println(string(body))
	return string(body), nil
}
