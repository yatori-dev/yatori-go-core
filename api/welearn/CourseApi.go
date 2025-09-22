package welearn

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"

	"github.com/yatori-dev/yatori-go-core/utils"
)

// 拉取课程列表json
func (cache *WeLearnUserCache) PullCourseListApi(retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	urlStr := "https://welearn.sflep.com/ajax/authCourse.aspx?action=gmc&nocache=" + fmt.Sprintf("%.16f", rand.Float32())
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, urlStr, nil)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Add("Referer", "https://welearn.sflep.com/student/index.aspx")
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Host", "welearn.sflep.com")
	req.Header.Add("Connection", "keep-alive")
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
	return string(body), nil
}

// 拉取课程必要的信息，用于后续请求
// 必要信息有，uid,classid
func (cache *WeLearnUserCache) PullCourseInfoApi(cid string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	url := "https://welearn.sflep.com/student/course_info.aspx?cid=" + cid
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("Referer", "https://welearn.sflep.com/student/index.aspx")
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "welearn.sflep.com")
	req.Header.Add("Connection", "keep-alive")
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
	//fmt.Println(string(body))
	return string(body), nil
}

// 拉取大章节
func (cache *WeLearnUserCache) PullCourseChapterApi(cid, stuid, classid string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	params := url.Values{}
	params.Set("action", "courseunits")
	params.Set("cid", cid)
	params.Set("stuid", stuid)
	params.Set("classid", classid)
	params.Set("nocache", fmt.Sprintf("%f", rand.Float64()))

	urlStr := "https://welearn.sflep.com/ajax/StudyStat.aspx?" + params.Encode()
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, urlStr, nil)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("Referer", "https://welearn.sflep.com/student/index.aspx")
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "welearn.sflep.com")
	req.Header.Add("Connection", "keep-alive")
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
	//fmt.Println(string(body))
	return string(body), nil
}

// 拉取大章节点对应的任务点
func (cache *WeLearnUserCache) PullCoursePointApi(cid, stuid, classid, unitidx string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	params := url.Values{}
	params.Set("action", "scoLeaves")
	params.Set("cid", cid)
	params.Set("stuid", stuid)
	params.Set("unitidx", unitidx)
	params.Set("classid", classid)
	params.Set("nocache", fmt.Sprintf("%f", rand.Float64()))

	urlStr := "https://welearn.sflep.com/ajax/StudyStat.aspx?" + params.Encode()
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, urlStr, nil)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("Referer", "https://welearn.sflep.com/student/index.aspx")
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "welearn.sflep.com")
	req.Header.Add("Connection", "keep-alive")
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
	//fmt.Println(string(body))
	return string(body), nil
}
