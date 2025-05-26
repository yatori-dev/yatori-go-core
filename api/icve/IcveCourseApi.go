package icve

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// CourseListApi 拉取课程
func (cache *IcveUserCache) CourseListApi() {

	url := "http://www.icve.com.cn/studycenter/MyCourse/studingCourse"
	method := "GET"

	payload := strings.NewReader(`{` + `"isFinished": 0,` + `"page":1,` + `"pageSize": 8` + `}`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "mooc.icve.com.cn")
	req.Header.Add("Connection", "keep-alive")
	for _, v := range cache.cookies {
		req.AddCookie(v)
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
}
