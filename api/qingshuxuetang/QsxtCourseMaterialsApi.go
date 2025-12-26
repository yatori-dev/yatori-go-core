package qingshuxuetang

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
)

// 拉取课程资料列表
func (cache *QsxtUserCache) PullCourseMaterialsListApi(periodId, classId, schoolId, source, userClassId, courseId string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	//url := "https://api.qingshuxuetang.com/v25_10/course/getCourseDetail?periodId=27&classId=321&schoolId=1976&source=1&userClassId=0&courseId=599"
	url := "https://api.qingshuxuetang.com/v25_10/course/getCourseDetail?periodId=" + periodId + "&classId=" + classId + "&schoolId=" + schoolId + "&source=" + source + "&userClassId=" + userClassId + "&courseId=" + courseId
	method := "GET"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}
	client := &http.Client{
		Transport: tr,
	}
	req, err := http.NewRequest(method, url, nil)

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
		return cache.PullCourseMaterialsListApi(periodId, classId, schoolId, source, userClassId, courseId, retry, lastErr)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return string(body), nil
}
