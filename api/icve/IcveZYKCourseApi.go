package icve

import (
	"bytes"
	"crypto/aes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/yatori-dev/yatori-go-core/utils"
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
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "mooc.icve.com.cn")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Authorization", "Bearer "+cache.AccessToken)
	for _, v := range cache.Cookies {
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

// 资源库课程接口1
func (cache *IcveUserCache) PullZykCourse1Api() (string, error) {
	url := "https://www.icve.com.cn/prod-api/zyk/myLearn?pageSize=100&pageNum=1&queryType=2&selectType=1"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("Authorization", "Bearer "+cache.AccessToken)

	for _, cookie := range cache.Cookies {
		req.AddCookie(cookie)
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "www.icve.com.cn")
	req.Header.Add("Connection", "keep-alive")

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

// 资源库课程接口2
func (cache *IcveUserCache) PullZykCourse2Api() (string, error) {
	url := "https://zyk.icve.com.cn/prod-api/teacher/courseList/myCourseList?pageSize=100&pageNum=1&flag=1"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("Authorization", "Bearer "+cache.ZYKAccessToken)

	for _, cookie := range cache.Cookies {
		req.AddCookie(cookie)
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "www.icve.com.cn")
	req.Header.Add("Connection", "keep-alive")

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

// 拉取课程根节点列表
func (cache *IcveUserCache) PullRootNodeListApi(courseInfo string) (string, error) {

	url := "https://zyk.icve.com.cn/prod-api/teacher/courseContent/studyMoudleList?courseInfoId=" + courseInfo
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "zyk.icve.com.cn")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Authorization", "Bearer "+cache.ZYKAccessToken)
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

// 拉取课程根节点列表
func (cache *IcveUserCache) PullZykNodeListApi(level int, parentId, courseInfo string) (string, error) {

	url := "https://zyk.icve.com.cn/prod-api/teacher/courseContent/studyList?level=" + fmt.Sprintf("%d", level) + "&parentId=" + parentId + "&courseInfoId=" + courseInfo
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "zyk.icve.com.cn")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Authorization", "Bearer "+cache.ZYKAccessToken)
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
	return string(body), nil
}

// 拉取任务点详细信息
func (cache *IcveUserCache) PullZykNodeInfoApi(sourceId string) (string, error) {

	url := "https://zyk.icve.com.cn/prod-api/teacher/courseContent/" + sourceId
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "zyk.icve.com.cn")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Authorization", "Bearer "+cache.ZYKAccessToken)
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
	return string(body), nil

}

// 获取时长Api
func (cache *IcveUserCache) PullZykNodeDurationApi(fileUrl string) (string, error) {

	url := "https://upload.icve.com.cn/" + fileUrl + "/status"
	method := "POST"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "upload.icve.com.cn")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Cookie", "acw_tc=1a0c380717597778746754908e0d86367b753f74c585e515faa32ee572a5e1; SERVERID=27b9eb8d6e551a69b5bfbbc79124ceca|1759777888|1759777874")

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

// 资源库提交学习
func (cache *IcveUserCache) SubmitZYKStudyTimeApi(courseInfo string, id string, parentId string, studyTime int, sourceId string, studentId string, actualNum int, lastNum int, totalNum int) (string, error) {

	url := "https://zyk.icve.com.cn/prod-api/teacher/studyRecord"
	method := "PUT"

	params := map[string]interface{}{
		"courseInfoId": courseInfo,
		"id":           id, // 可以留空或不传
		"parentId":     parentId,
		"studyTime":    fmt.Sprintf("%d", studyTime),
		"sourceId":     sourceId,
		"studentId":    studentId,
		"actualNum":    fmt.Sprintf("%d", actualNum),
		"lastNum":      fmt.Sprintf("%d", lastNum),
		"totalNum":     fmt.Sprintf("%d", totalNum),
	}

	jsonBytes, _ := json.Marshal(params)
	key := []byte("djekiytolkijduey") //秘钥

	encData, err := aesEncryptECB(key, string(jsonBytes))
	if err != nil {
		panic(err)
	}
	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewBuffer([]byte(encData)))

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "zyk.icve.com.cn")
	req.Header.Add("Connection", "keep-alive")
	for _, cookie := range cache.Cookies {
		req.AddCookie(cookie)
	}
	req.Header.Add("Authorization", "Bearer "+cache.ZYKAccessToken)

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

// AES ECB 加密（PKCS7 填充）
func aesEncryptECB(key []byte, plaintext string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// PKCS7 padding
	padding := block.BlockSize() - len(plaintext)%block.BlockSize()
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	padded := append([]byte(plaintext), padtext...)

	// ECB 加密
	ciphertext := make([]byte, len(padded))
	for bs, be := 0, block.BlockSize(); bs < len(padded); bs, be = bs+block.BlockSize(), be+block.BlockSize() {
		block.Encrypt(ciphertext[bs:be], padded[bs:be])
	}

	// 返回 Base64
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}
