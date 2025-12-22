package external

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/que-core/qentity"
	"github.com/yatori-dev/yatori-go-core/utils"
)

// 全能免费题库
func QuanNengQueRequestApi(problem qentity.Question, retry int, err error) (*qentity.ResultQuestion, error) {
	urlStr := "https://api.wkexam.com/api?q=" + url.QueryEscape(problem.Content) + "&token=qqqqq"
	method := "GET"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}

	client := &http.Client{
		Transport: tr,
	}
	req, err := http.NewRequest(method, urlStr, nil)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "api.wkexam.com")
	req.Header.Add("Connection", "keep-alive")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	//请求失败处理
	status := gojsonq.New().JSONString(string(body)).Find("code")
	if status == nil {
		return nil, err
	}
	if int(status.(float64)) != 1 {
		return nil, err
	}
	//拉取回答数据
	answer_json := gojsonq.New().JSONString(string(body)).Find("data.answer")
	//如果没有答案
	if answer_json == nil {
		return nil, err
	}
	//转换数据
	answer_result := []string{}
	for _, answer := range strings.Split(answer_json.(string), "#") {
		answer_result = append(answer_result, strings.TrimSpace(answer))
	}
	//赋值数据
	problem.Answers = answer_result

	return &qentity.ResultQuestion{
		Question: problem,
		Msg:      "请求成功",
		Replier:  "内置全能题库",
		Code:     200,
	}, nil
}

// spacex免费题库
func SpaceXQueRequestApi(problem qentity.Question, retry int, err error) (*qentity.ResultQuestion, error) {
	options := ""
	if problem.Options != nil {
		for i, v := range problem.Options {
			options = options + v
			if i != len(problem.Options)-1 {
				options += "#"
			}
		}
	}

	urlStr := "https://doc.spacex-api.com/Api/tiku?question=" + url.QueryEscape(problem.Content) + "&token=c5ecf55ebeb4b61d669713b54693ce09&options=" + url.QueryEscape(options)
	method := "GET"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}

	client := &http.Client{
		Transport: tr,
	}
	req, err := http.NewRequest(method, urlStr, nil)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "doc.spacex-api.com")
	req.Header.Add("Connection", "keep-alive")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	//请求失败处理
	status := gojsonq.New().JSONString(string(body)).Find("code")
	if status == nil {
		return nil, err
	}
	if int(status.(float64)) != 1 {
		return nil, err
	}
	//拉取回答数据
	answer_json := gojsonq.New().JSONString(string(body)).Find("answer")
	//如果没有答案
	if answer_json == nil {
		return nil, err
	}
	//转换数据
	answer_result := []string{}
	for _, answer := range strings.Split(answer_json.(string), "#") {
		answer_result = append(answer_result, strings.TrimSpace(answer))
	}
	//赋值数据
	problem.Answers = answer_result

	return &qentity.ResultQuestion{
		Question: problem,
		Msg:      "请求成功",
		Replier:  "内置SpaceX题库",
		Code:     200,
	}, nil
}

// Json API在线搜题
func JsonApiQuesRequestApi(problem qentity.Question, retry int, err error) (*qentity.ResultQuestion, error) {
	options := ""
	if problem.Options != nil {
		for i, v := range problem.Options {
			options = options + v
			if i != len(problem.Options)-1 {
				options += "#"
			}
		}
	}
	urlStr := "https://www.aiask.site/v1/question/precise?question=" + url.QueryEscape(problem.Content) + "&options=" + url.QueryEscape(options)
	method := "POST"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}

	client := &http.Client{
		Transport: tr,
	}
	req, err := http.NewRequest(method, urlStr, nil)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "www.aiask.site")
	req.Header.Add("Connection", "keep-alive")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	//请求失败处理
	status := gojsonq.New().JSONString(string(body)).Find("code")
	if status == nil {
		return nil, err
	}
	if int(status.(float64)) != 200 {
		return nil, err
	}
	//拉取回答数据
	answer_json := gojsonq.New().JSONString(string(body)).Find("data.answer")
	//如果没有答案
	if answer_json == nil {
		return nil, err
	}
	//转换数据
	answer_result := []string{}
	for _, answer := range strings.Split(answer_json.(string), "#") {
		answer_result = append(answer_result, strings.TrimSpace(answer))
	}
	//赋值数据
	problem.Answers = answer_result

	return &qentity.ResultQuestion{
		Question: problem,
		Msg:      "请求成功",
		Replier:  "内置Json API在线搜题题库",
		Code:     200,
	}, nil
}

func FreeQueRequest() {

}
