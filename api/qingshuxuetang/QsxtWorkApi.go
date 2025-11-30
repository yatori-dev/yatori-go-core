package qingshuxuetang

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// 拉取作业列表
func (cache *QsxtUserCache) PullWorkListApi(periodId, classId, schoolId, courseId string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	//url := "https://api.qingshuxuetang.com/v25_10/quiz/search?periodId=24&classId=45&schoolId=114079&type=2&courseId=879"
	urlStr := fmt.Sprintf("https://api.qingshuxuetang.com/v25_10/quiz/search?periodId=%s&classId=%s&schoolId=%s&type=2&courseId=%s", periodId, classId, schoolId, courseId)
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
		return "", err
	}
	req.Header.Add("User-Agent", "okhttp/4.2.2")
	req.Header.Add("Connection", "Keep-Alive")
	req.Header.Add("Authorization-QS", cache.Token)
	req.Header.Add("Device-Trace-Id-QS", "b0afcf7e-a8ae-48f2-b438-66982a13dc16")
	req.Header.Add("Device-Info-QS", "{\"appType\":1,\"appVersion\":\"25.10.0\",\"clientType\":2,\"deviceName\":\"xiaomi MI 5X\",\"netType\":1,\"osVersion\":\"8.1.0\"}")
	req.Header.Add("User-Agent-QS", "QSXT")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "api.qingshuxuetang.com")

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

// 拉取试卷题目内容
func (cache *QsxtUserCache) PullWorkQuestionListApi(classId, quizId, schoolId, courseId string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	//urlStr := "https://api.qingshuxuetang.com/v25_10/quiz/detail?classId=45&quizId=2_69084c57c28a765d5459f6d7&schoolId=114079&courseId=879"
	urlStr := fmt.Sprintf("https://api.qingshuxuetang.com/v25_10/quiz/detail?classId=%s&quizId=%s&schoolId=%s&courseId=%s", classId, quizId, schoolId, courseId)
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
		return "", err
	}
	req.Header.Add("User-Agent", "okhttp/4.2.2")
	req.Header.Add("Connection", "Keep-Alive")
	req.Header.Add("Authorization-QS", cache.Token)
	req.Header.Add("Device-Trace-Id-QS", "b0afcf7e-a8ae-48f2-b438-66982a13dc16")
	req.Header.Add("Device-Info-QS", "{\"appType\":1,\"appVersion\":\"25.10.0\",\"clientType\":2,\"deviceName\":\"xiaomi MI 5X\",\"netType\":1,\"osVersion\":\"8.1.0\"}")
	req.Header.Add("User-Agent-QS", "QSXT")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "api.qingshuxuetang.com")

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

// 提交对应题目答案
func (cache *QsxtUserCache) SubmitAnswerApi(answer, questionId, quizId, schoolId string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}

	urlStr := "https://api.qingshuxuetang.com/v25_10/quiz/question/answer"
	method := "POST"

	//payload := strings.NewReader(`{"questionAnswers": [{"answer": "B","questionId": "68d350735c711a0e952fb656"}],quizId": "2_69084c57c28a765d5459f6d7","schoolId": 114079}`)
	payload := strings.NewReader(`{"questionAnswers": [{"answer": "` + answer + `","questionId": "` + questionId + `"}],"quizId": "` + quizId + `","schoolId": ` + schoolId + `}`)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}
	client := &http.Client{
		Transport: tr,
	}
	req, err := http.NewRequest(method, urlStr, payload)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("User-Agent", "okhttp/4.2.2")
	req.Header.Add("Connection", "Keep-Alive")
	req.Header.Add("Authorization-QS", cache.Token)
	req.Header.Add("Device-Trace-Id-QS", "b0afcf7e-a8ae-48f2-b438-66982a13dc16")
	req.Header.Add("Device-Info-QS", "{\"appType\":1,\"appVersion\":\"25.10.0\",\"clientType\":2,\"deviceName\":\"xiaomi MI 5X\",\"netType\":1,\"osVersion\":\"8.1.0\"}")
	req.Header.Add("User-Agent-QS", "QSXT")
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "api.qingshuxuetang.com")

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

// 保存答题
func (cache *QsxtUserCache) SaveAnswerApi(answers string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	url := "https://api.qingshuxuetang.com/v25_10/quiz/submit"
	method := "POST"

	payload := strings.NewReader(answers)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}
	client := &http.Client{
		Transport: tr,
	}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("User-Agent", "okhttp/4.2.2")
	req.Header.Add("Connection", "Keep-Alive")
	req.Header.Add("Authorization-QS", cache.Token)
	req.Header.Add("Device-Trace-Id-QS", "b0afcf7e-a8ae-48f2-b438-66982a13dc16")
	req.Header.Add("Device-Info-QS", "{\"appType\":1,\"appVersion\":\"25.10.0\",\"clientType\":2,\"deviceName\":\"xiaomi MI 5X\",\"netType\":1,\"osVersion\":\"8.1.0\"}")
	req.Header.Add("User-Agent-QS", "QSXT")
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "api.qingshuxuetang.com")

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
