package qingshuxuetang

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// 拉取课程
func (cache *QsxtUserCache) QsxtPullCourseApi(retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	url := "https://api.qingshuxuetang.com/v25_10/course/mine"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

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

// 拉取课程进度接口
func (cache *QsxtUserCache) QsxtPullCourseProcessApi(retry int, lastErr error) (string, error) {

	url := "https://api.qingshuxuetang.com/v25_10/course/extendInfo?courseIds%255B%255D=879,954,929,820&classId=45&period=24&schoolId=114079"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

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

// 拉取课程详细信息
func (cache *QsxtUserCache) QsxtPullCourseDetailApi(periodId, classId, schoolId, courseId string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}

	url := "https://api.qingshuxuetang.com/v25_10/course/getCourseDetail?periodId=" + periodId + "&classId=" + classId + "&schoolId=" + schoolId + "&source=1&userClassId=0&courseId=" + courseId
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

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

// 拉取课程任务点
func (cache *QsxtUserCache) QsxtPullNodeApi(urlStr string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}

	//url := "https://api.qingshuxuetang.com/v25_10/course/coursewareTree?sign=3952ef54d22df17936dac7aa6ef7ab20&id=5d4aafff9da4191eac39e616&timestamp=1761927389725"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, urlStr, nil)

	if err != nil {
		fmt.Println(err)
		return "", nil
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
		return "", nil
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", nil
	}

	return string(body), nil
}

// 拉取对应课程的对应活动分数达标分布
func (cache *QsxtUserCache) QsxtPullCourseScoreApi(periodId, classId, schoolId, courseId string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}

	url := "https://api.qingshuxuetang.com/v25_10/score/studentScore?periodId=" + periodId + "&classId=" + classId + "&schoolId=" + schoolId + "&courseId=" + courseId
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

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

// 拉取课程任务点学习时间记录
func (cache *QsxtUserCache) PullStudyRecordApi(periodId, classId, schoolId, courseId string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	url := "https://api.qingshuxuetang.com/v25_10/behavior/downloadStudyRecord?periodId=" + periodId + "&classId=" + classId + "&schoolId=" + schoolId + "&contentType=11&courseId=" + courseId
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

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

// 开始学习接口，每次学习前都要先调用
func (cache *QsxtUserCache) StartStudyApi(classId, contentId, courseId, periodId, schoolId string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	url := "https://api.qingshuxuetang.com/v25_10/behavior/studyRecordStart"
	method := "POST"

	payload := strings.NewReader(`{
  "classId": ` + classId + `,
  "clientType": 2,
  "contentId": "` + contentId + `",
  "contentType": 11,
  "courseId": "` + courseId + `",
  "learnPlanId": 0,
  "pageType": 2,
  "periodId": ` + periodId + `,
  "position": 0,
  "schoolId": ` + schoolId + `,
  "userClassId": 0
}`)

	client := &http.Client{}
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
	req.Header.Add("Cookie", "__environment=production; AccessToken=eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1SWQiOjQ3NTAxNTE3LCJyb2xlIjoxMDAsIm1vY2tJZGVudGl0eSI6ZmFsc2UsImNsaWVudCI6InBjd2ViIiwib3JncyI6WyJ7XCJ0eXBlXCI6XCJkZ1wiLFwicm9sZXNcIjpbMV0sXCJpZFwiOjExNDA3OSxcIm91SWRcIjo3MzcsXCJvdVJvbGVzXCI6WzFdfSJdLCJleHAiOjE3NzIxNjc2NTEsImp0aSI6Imp3dDAxMzg3MzdmYjJkMzRiYTNhMDExZGMxMzcyYjY5NDcyIiwicGxhdGZvcm0iOiJxc3h0In0.x0p2RrY9oEUFRVeyKbasHxBCbL6mWpBQiC7rp3024VA")

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

// 提交学时接口，在使用该接口前必须要前使用StartStudyApi才能正常提交学时
func (cache *QsxtUserCache) SubmitStudyTimeApi(schoolId, serverRecordId string, position int, isEnd bool, retry int, lastErr error) (string, error) {

	url := "https://api.qingshuxuetang.com/v25_10/behavior/studyRecordContinue"
	method := "POST"
	payloadStr := `{
  "clientType": 2,
  "contentType": 11,
  "detectId": 0,
  "position": ` + fmt.Sprintf("%d", position) + `,
  "schoolId": ` + schoolId + `,
  "serverRecordId": "` + serverRecordId + `"
}`
	//如果是结束状态的提交学时
	if isEnd {
		payloadStr = `{
  "clientType": 2,
  "contentType": 11,
  "detectId": 0,
  "end": true,
  "position": ` + fmt.Sprintf("%d", position) + `,
  "schoolId": ` + schoolId + `,
  "serverRecordId": "` + serverRecordId + `"
}`
	}
	payload := strings.NewReader(payloadStr)

	client := &http.Client{}
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
	if res.StatusCode != 200 {
		return "", errors.New(res.Status)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	return string(body), nil
}
