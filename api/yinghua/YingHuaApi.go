package yinghua

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/yatori-dev/yatori-go-core/entity"
	"github.com/yatori-dev/yatori-go-core/global"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/yatori-dev/yatori-go-core/utils"
)

type YingHuaUserCache struct {
	PreUrl   string //前置url
	Account  string //账号
	Password string //用户密码
	verCode  string //验证码
	cookie   string //验证码用的session
	token    string //保持会话的Token
	sign     string //签名
}

func (cache *YingHuaUserCache) GetVerCode() string {
	return cache.verCode
}
func (cache *YingHuaUserCache) SetVerCode(verCode string) {
	cache.verCode = verCode
}

func (cache *YingHuaUserCache) GetCookie() string {
	return cache.cookie
}
func (cache *YingHuaUserCache) SetCookie(cookie string) {
	cache.cookie = cookie
}

func (cache *YingHuaUserCache) GetToken() string {
	return cache.token
}
func (cache *YingHuaUserCache) SetToken(token string) {
	cache.token = token
}

func (cache *YingHuaUserCache) GetSign() string {
	return cache.token
}
func (cache *YingHuaUserCache) SetSign(sign string) {
	cache.sign = sign
}

func NewCache(input int) YingHuaUserCache {
	return YingHuaUserCache{
		PreUrl:   global.Config.Users[input].URL,
		Account:  global.Config.Users[input].Account,
		Password: global.Config.Users[input].Password,
	}
}

func NewCacheList() []YingHuaUserCache {
	var cacheList []YingHuaUserCache
	for i, user := range global.Config.Users {
		if user.AccountType != "YINGHUA" {
			continue
		}
		cacheList = append(cacheList, NewCache(i))
	}
	return cacheList
}

//func (cache YingHuaUserCache) String() string {
//
//	return ""
//}

// LoginApi 登录接口
func (cache *YingHuaUserCache) LoginApi(retry int, lastError error) (string, error) {
	if retry < 0 {
		return "", lastError
	}
	url := cache.PreUrl + "/user/login.json"
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("username", cache.Account)
	_ = writer.WriteField("password", cache.Password)
	_ = writer.WriteField("code", cache.verCode)
	_ = writer.WriteField("redirect", cache.PreUrl)
	err := writer.Close()
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("Cookie", cache.cookie)
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")

	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		time.Sleep(150 * time.Millisecond)
		return cache.LoginApi(retry-1, err)
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

// VerificationCodeApi 获取验证码和SESSION验证码,并返回文件路径和SESSION字符串
var randChar []string = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f", "A", "B", "C", "D", "E", "F"}

func (cache *YingHuaUserCache) VerificationCodeApi(retry int) (string, string) {
	if retry < 0 {
		return "", ""
	}
	url := cache.PreUrl + fmt.Sprintf("/service/code?r=%d", time.Now().Unix())
	method := "GET"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}

	req, err := http.NewRequest(method, url, nil)
	req.Header.Add("Cookie", cache.cookie)

	if err != nil {
		fmt.Println(err)
		return "", ""
	}
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")

	res, err := client.Do(req)
	if err != nil {
		time.Sleep(150 * time.Millisecond)
		if strings.Contains(err.Error(), "A connection attempt failed because the connected party did not properly respond after a period of time") {
			return cache.VerificationCodeApi(retry)
		}
		return cache.VerificationCodeApi(retry - 1)
	}
	defer res.Body.Close()

	codeFileName := "code" + randChar[rand.Intn(len(randChar))] //生成验证码文件名称
	for i := 0; i < 10; i++ {
		codeFileName += randChar[rand.Intn(len(randChar))]
	}
	codeFileName += ".png"
	utils.PathExistForCreate("./assets/code/") //检测是否存在路径，如果不存在则创建
	filepath := fmt.Sprintf("./assets/code/%s", codeFileName)
	file, err := os.Create(filepath)
	if err != nil {
		log.Println(err)
		return "", ""
	}
	defer file.Close()

	_, err = io.Copy(file, res.Body)
	if err != nil {
		log.Println(err)
		return "", ""
	}
	return filepath, res.Header.Get("Set-Cookie")
}

// KeepAliveApi 登录心跳保活
func KeepAliveApi(UserCache YingHuaUserCache) string {

	url := UserCache.PreUrl + "/api/online.json"
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("platform", "Android")
	_ = writer.WriteField("version", "1.4.8")
	_ = writer.WriteField("token", UserCache.token)
	//_ = writer.WriteField("schoolId", "7")
	err := writer.Close()
	if err != nil {
		fmt.Println(err)
		return ""
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return ""
	}
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")

	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		time.Sleep(time.Millisecond * 150) //延迟
		return KeepAliveApi(UserCache)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	if strings.Contains(string(body), "502 Bad Gateway") {
		time.Sleep(time.Millisecond * 150) //延迟
		return KeepAliveApi(UserCache)
	}
	return string(body)
}

// CourseListApi 拉取课程列表API
func (cache *YingHuaUserCache) CourseListApi(retry int, lastError error) (string, error) {
	if retry < 0 {
		return "", lastError
	}
	url := cache.PreUrl + "/api/course/list.json"
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("platform", "Android")
	_ = writer.WriteField("version", "1.4.8")
	_ = writer.WriteField("type", "0")
	_ = writer.WriteField("token", cache.token)
	err := writer.Close()
	if err != nil {
		return "", err
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	req, err := http.NewRequest(method, url, payload)
	req.Header.Set("Cookie", cache.cookie)
	if err != nil {
		return "", err
	}
	req.Header.Add("Cookie", "tgw_I7_route=3d5c4e13e7d88bb6849295ab943042a2")
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")

	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		time.Sleep(time.Millisecond * 150) //延迟
		if strings.Contains(err.Error(), "A connection attempt failed because the connected party did not properly respond after a period of time") {
			return cache.CourseListApi(retry, err)
		}
		return cache.CourseListApi(retry-1, err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		time.Sleep(time.Millisecond * 150) //延迟
		return cache.CourseListApi(retry-1, err)
	}
	if strings.Contains(string(body), "502 Bad Gateway") {
		time.Sleep(time.Millisecond * 150) //延迟
		return cache.CourseListApi(retry, err)
	}
	return string(body), nil
}

// CourseDetailApi 获取课程详细信息API
func (cache *YingHuaUserCache) CourseDetailApi(courseId string, retry int, lastError error) (string, error) {
	if retry < 0 {
		return "", lastError
	}
	url := cache.PreUrl + "/api/course/detail.json"
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("platform", "Android")
	_ = writer.WriteField("version", "1.4.8")
	_ = writer.WriteField("courseId", courseId)
	_ = writer.WriteField("token", cache.token)
	err := writer.Close()
	if err != nil {
		return "", err
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	req, err := http.NewRequest(method, url, payload)
	req.Header.Add("Cookie", cache.cookie)

	if err != nil {
		return "", err
	}
	req.Header.Add("Cookie", "tgw_I7_route=3d5c4e13e7d88bb6849295ab943042a2")
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")

	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		time.Sleep(time.Millisecond * 150) //延迟
		if strings.Contains(err.Error(), "A connection attempt failed because the connected party did not properly respond after a period of time") {
			return cache.CourseDetailApi(courseId, retry, err)
		}
		return cache.CourseDetailApi(courseId, retry-1, err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		time.Sleep(time.Millisecond * 150) //延迟
		return cache.CourseDetailApi(courseId, retry-1, err)
	}
	if strings.Contains(string(body), "502 Bad Gateway") {
		time.Sleep(time.Millisecond * 150) //延迟
		return cache.CourseDetailApi(courseId, retry, err)
	}
	return string(body), err
}

// CourseVideListApi 对应课程的视屏列表
func CourseVideListApi(UserCache YingHuaUserCache, courseId string /*课程ID*/, retry int, lastError error) (string, error) {
	if retry < 0 {
		return "", lastError
	}
	url := UserCache.PreUrl + "/api/course/chapter.json"
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("platform", "Android")
	_ = writer.WriteField("version", "1.4.8")
	_ = writer.WriteField("token", UserCache.token)
	_ = writer.WriteField("courseId", courseId)
	err := writer.Close()
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	req, err := http.NewRequest(method, url, payload)
	req.Header.Set("Cookie", UserCache.cookie)
	if err != nil {
		time.Sleep(time.Millisecond * 150) //延迟
		return CourseVideListApi(UserCache, courseId, retry-1, err)
	}
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")

	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		time.Sleep(time.Millisecond * 150) //延迟
		if strings.Contains(err.Error(), "A connection attempt failed because the connected party did not properly respond after a period of time") {
			return CourseVideListApi(UserCache, courseId, retry, err)
		}
		return CourseVideListApi(UserCache, courseId, retry-1, err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		time.Sleep(time.Millisecond * 150) //延迟
		return CourseVideListApi(UserCache, courseId, retry-1, err)
	}
	if strings.Contains(string(body), "502 Bad Gateway") {
		time.Sleep(time.Millisecond * 150) //延迟
		return CourseVideListApi(UserCache, courseId, retry, err)
	}
	return string(body), nil
}

// SubmitStudyTimeApi 提交学时
func SubmitStudyTimeApi(UserCache YingHuaUserCache, nodeId string /*对应视屏节点ID*/, studyId string /*学习分配ID*/, studyTime int /*提交的学时*/, retry int, lastError error) (string, error) {
	if retry < 0 {
		return "", lastError
	}
	url := UserCache.PreUrl + "/api/node/study.json"
	method := "POST"
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("platform", "Android")
	_ = writer.WriteField("version", "1.4.8")
	_ = writer.WriteField("nodeId", nodeId)
	_ = writer.WriteField("token", UserCache.token)
	_ = writer.WriteField("terminal", "Android")
	_ = writer.WriteField("studyTime", strconv.Itoa(studyTime))
	_ = writer.WriteField("studyId", studyId)
	err := writer.Close()
	if err != nil {
		time.Sleep(time.Millisecond * 150)
		return SubmitStudyTimeApi(UserCache, nodeId, studyId, studyTime, retry-1, err)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // 跳过证书验证
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		//fmt.Println(err)
		return "", err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.13; rv:88.0) Gecko/20100101 Firefox/88.0")

	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		time.Sleep(time.Millisecond * 150)
		return SubmitStudyTimeApi(UserCache, nodeId, studyId, studyTime, retry-1, err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		time.Sleep(time.Millisecond * 150)
		return SubmitStudyTimeApi(UserCache, nodeId, studyId, studyTime, retry-1, err)
	}

	//避免502情况
	if strings.Contains(string(body), "502 Bad Gateway") {
		time.Sleep(time.Millisecond * 150) //延迟
		return SubmitStudyTimeApi(UserCache, nodeId, studyId, studyTime, retry-1, err)
	}

	return string(body), nil
}

// VideStudyTimeApi 获取单个视屏的学习进度
func VideStudyTimeApi(userEntity entity.UserEntity, nodeId string, retryNum int, lastError error) string {
	if retryNum < 0 {
		return ""
	}
	url := userEntity.PreUrl + "/api/node/video.json"
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("platform", "Android")
	_ = writer.WriteField("version", "1.4.8")
	_ = writer.WriteField("nodeId", nodeId)
	_ = writer.WriteField("token", userEntity.Token)
	err := writer.Close()
	if err != nil {
		fmt.Println(err)
		return ""
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		time.Sleep(time.Millisecond * 150) //延迟
		return VideStudyTimeApi(userEntity, nodeId, retryNum-1, lastError)
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.13; rv:88.0) Gecko/20100101 Firefox/88.0")

	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		time.Sleep(time.Millisecond * 150) //延迟
		return VideStudyTimeApi(userEntity, nodeId, retryNum-1, lastError)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		time.Sleep(time.Millisecond * 150) //延迟
		return VideStudyTimeApi(userEntity, nodeId, retryNum-1, lastError)
	}
	if strings.Contains(string(body), "502 Bad Gateway") {
		time.Sleep(time.Millisecond * 150) //延迟
		return VideStudyTimeApi(userEntity, nodeId, retryNum, lastError)
	}
	return string(body)
}

// VideWatchRecodeApi 获取指定课程视屏观看记录
func VideWatchRecodeApi(UserCache YingHuaUserCache, courseId string, page int, retry int, lastError error) (string, error) {
	if retry < 0 {
		return "", lastError
	}
	url := UserCache.PreUrl + "/api/record/video.json"
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("platform", "Android")
	_ = writer.WriteField("version", "1.4.8")
	_ = writer.WriteField("token", UserCache.token)
	_ = writer.WriteField("courseId", courseId)
	_ = writer.WriteField("page", strconv.Itoa(page))
	err := writer.Close()
	if err != nil {
		//fmt.Println(err)
		return VideWatchRecodeApi(UserCache, courseId, page, retry-1, err)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	req, err := http.NewRequest(method, url, payload)
	req.Header.Set("Cookie", UserCache.cookie)
	if err != nil {
		//fmt.Println(err)
		return VideWatchRecodeApi(UserCache, courseId, page, retry-1, err)
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.13; rv:88.0) Gecko/20100101 Firefox/88.0")

	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		//fmt.Println(err)
		return VideWatchRecodeApi(UserCache, courseId, page, retry-1, err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		//fmt.Println(err)
		return VideWatchRecodeApi(UserCache, courseId, page, retry-1, err)
	}
	if strings.Contains(string(body), "502 Bad Gateway") {
		time.Sleep(time.Millisecond * 150) //延迟
		return VideWatchRecodeApi(UserCache, courseId, page, retry, lastError)
	}
	return string(body), nil
}

// ExamDetailApi 获取考试信息
func ExamDetailApi(UserCache YingHuaUserCache, nodeId string, retryNum int, lastError error) (string, error) {
	if retryNum < 0 {
		return "", lastError
	}
	url := UserCache.PreUrl + "/api/node/exam.json?nodeId=" + nodeId
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("platform", "Android")
	_ = writer.WriteField("version", "1.4.8")
	_ = writer.WriteField("nodeId", nodeId)
	_ = writer.WriteField("token", UserCache.token)
	_ = writer.WriteField("terminal", "Android")
	err := writer.Close()
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	req, err := http.NewRequest(method, url, payload)
	req.Header.Add("Cookie", UserCache.cookie)

	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.13; rv:88.0) Gecko/20100101 Firefox/88.0")

	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		time.Sleep(time.Millisecond * 150) //延迟
		return ExamDetailApi(UserCache, nodeId, retryNum-1, err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		time.Sleep(time.Millisecond * 150) //延迟
		return ExamDetailApi(UserCache, nodeId, retryNum-1, err)
	}
	if strings.Contains(string(body), "502 Bad Gateway") {
		time.Sleep(time.Millisecond * 150) //延迟
		return ExamDetailApi(UserCache, nodeId, retryNum, err)
	}
	return string(body), nil
}

// StartExam 开始考试接口
// {"_code":9,"status":false,"msg":"考试测试时间还未开始","result":{}}
func StartExam(userCache YingHuaUserCache, courseId, nodeId, examId string, retryNum int, lastError error) (string, error) {
	if retryNum < 0 {
		return "", lastError
	}
	url := userCache.PreUrl + "/api/exam/start.json?nodeId=" + nodeId + "&courseId=" + courseId + "&token=" + userCache.token + "&examId=" + examId
	method := "GET"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		time.Sleep(100 * time.Millisecond)
		return StartExam(userCache, courseId, nodeId, examId, retryNum-1, err)
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.13; rv:88.0) Gecko/20100101 Firefox/88.0")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		time.Sleep(100 * time.Millisecond)
		return StartExam(userCache, courseId, nodeId, examId, retryNum-1, err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		time.Sleep(100 * time.Millisecond)
		return StartExam(userCache, courseId, nodeId, examId, retryNum-1, err)
	}
	if strings.Contains(string(body), "502 Bad Gateway") {
		time.Sleep(time.Millisecond * 150) //延迟
		return StartExam(userCache, courseId, nodeId, examId, retryNum, lastError)
	}
	return string(body), nil
}

// GetExamTopicApi 获取所有考试题目，但是HTML，建议配合TurnExamTopic函数使用将题目html转成结构体
func GetExamTopicApi(UserCache YingHuaUserCache, nodeId, examId string, retryNum int, lastError error) (string, error) {
	if retryNum < 0 {
		return "", lastError
	}
	// Creating a custom HTTP client with timeout and SSL context (skip SSL setup for simplicity)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // 跳过证书验证
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   30 * time.Second,
	}

	// Creating the request body (empty JSON object)
	body := []byte("{}")

	// Create the request
	url := fmt.Sprintf("%s/api/exam.json?nodeId=%s&examId=%s&token=%s", UserCache.PreUrl, nodeId, examId, UserCache.token)
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	// Set the headers
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.13; rv:88.0) Gecko/20100101 Firefox/88.0")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", UserCache.PreUrl)
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Type", "application/json; charset=utf-8")

	// Perform the request
	resp, err := client.Do(req)
	if err != nil {
		time.Sleep(100 * time.Millisecond)
		return GetExamTopicApi(UserCache, nodeId, examId, retryNum-1, err)
	}
	defer resp.Body.Close()

	// Read the response body
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		time.Sleep(100 * time.Millisecond)
		return GetExamTopicApi(UserCache, nodeId, examId, retryNum-1, err)
	}
	if strings.Contains(string(body), "502 Bad Gateway") {
		time.Sleep(time.Millisecond * 150) //延迟
		return GetExamTopicApi(UserCache, nodeId, examId, retryNum, err)
	}
	return string(bodyBytes), nil
}

// SubmitExamApi 提交考试答案接口
func SubmitExamApi(UserCache YingHuaUserCache, examId, answerId string, answers YingHuaAnswer, finish string, retryNum int, lastError error) (string, error) {
	if retryNum < 0 {
		return "", lastError
	}
	// Creating the HTTP client with a timeout (30 seconds)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // 跳过证书验证
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}

	// Create a buffer to hold the multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add form fields to the multipart data
	writer.WriteField("platform", "Android")
	writer.WriteField("version", "1.4.8")
	writer.WriteField("examId", examId)
	writer.WriteField("terminal", "Android")
	writer.WriteField("answerId", answerId)
	writer.WriteField("finish", finish)
	writer.WriteField("token", UserCache.token)

	// Add the answer fields
	if answers.Type == "单选" || answers.Type == "判断" || answers.Type == "简答" {
		writer.WriteField("answer", answers.Answers[0])
	} else if answers.Type == "多选" {
		for _, v := range answers.Answers {
			writer.WriteField("answer[]", v)
		}
	} else if answers.Type == "填空" {
		for i, v := range answers.Answers {
			writer.WriteField("answer_"+strconv.Itoa(i+1), v)
		}
	}

	// Close the writer to finalize the multipart form data
	writer.Close()

	// Create the request with the necessary headers
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/exam/submit.json", UserCache.PreUrl), body)
	if err != nil {
		time.Sleep(100 * time.Millisecond)
		return SubmitExamApi(UserCache, examId, answerId, answers, finish, retryNum-1, err)
	}

	// Set the headers
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.13; rv:88.0) Gecko/20100101 Firefox/88.0")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", UserCache.PreUrl)
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Type", writer.FormDataContentType())

	// Perform the request
	resp, err := client.Do(req)
	if err != nil {
		time.Sleep(100 * time.Millisecond)
		return SubmitExamApi(UserCache, examId, answerId, answers, finish, retryNum-1, err)
	}
	defer resp.Body.Close()

	// Read the response body (we're not using the body here, just ensuring the request goes through)
	bodyStr, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if strings.Contains(string(bodyStr), "502 Bad Gateway") {
		time.Sleep(time.Millisecond * 150) //延迟
		return SubmitExamApi(UserCache, examId, answerId, answers, finish, retryNum, err)
	}
	return string(bodyStr), nil
}

// WorkDetailApi 获取作业信息
func WorkDetailApi(userCache YingHuaUserCache, nodeId string, retryNum int, lastError error) (string, error) {
	if retryNum < 0 {
		return "", lastError
	}
	url := userCache.PreUrl + "/api/node/work.json?nodeId=" + nodeId
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("platform", "Android")
	_ = writer.WriteField("version", "1.4.8")
	_ = writer.WriteField("nodeId", nodeId)
	_ = writer.WriteField("token", userCache.token)
	_ = writer.WriteField("terminal", "Android")
	err := writer.Close()
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // 跳过证书验证
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	req, err := http.NewRequest(method, url, payload)
	req.Header.Add("Cookie", userCache.cookie)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.13; rv:88.0) Gecko/20100101 Firefox/88.0")

	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		time.Sleep(100 * time.Millisecond)
		return WorkDetailApi(userCache, nodeId, retryNum-1, err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		time.Sleep(100 * time.Millisecond)
		return WorkDetailApi(userCache, nodeId, retryNum-1, err)
	}
	if strings.Contains(string(body), "502 Bad Gateway") {
		time.Sleep(time.Millisecond * 150) //延迟
		return WorkDetailApi(userCache, nodeId, retryNum, err)
	}
	return string(body), nil
}

// StartWork 开始做作业接口
// {"_code":9,"status":false,"msg":"您已完成作业，该作业仅可答题1次","result":{}}
func StartWork(userCache YingHuaUserCache, courseId, nodeId, workId string, retryNum int, lastError error) (string, error) {
	if retryNum < 0 {
		return "", lastError
	}
	url := userCache.PreUrl + "/api/work/start.json?nodeId=" + nodeId + "&courseId=" + courseId + "&token=" + userCache.token + "&workId=" + workId
	method := "GET"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.13; rv:88.0) Gecko/20100101 Firefox/88.0")

	res, err := client.Do(req)
	if err != nil {
		time.Sleep(100 * time.Millisecond)
		return StartWork(userCache, courseId, nodeId, workId, retryNum-1, err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	if strings.Contains(string(body), "502 Bad Gateway") {
		time.Sleep(time.Millisecond * 150) //延迟
		return StartWork(userCache, courseId, nodeId, workId, retryNum, err)
	}
	return string(body), nil
}

// GetWorkApi 获取所有作业题目
func GetWorkApi(UserCache YingHuaUserCache, nodeId, workId string, retryNum int, lastError error) (string, error) {
	if retryNum < 0 {
		return "", lastError
	}
	url := UserCache.PreUrl + "/api/work.json?nodeId=" + nodeId + "&workId=" + workId + "&token=" + UserCache.token
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	err := writer.Close()
	if err != nil {
		fmt.Println(err)
		return "", nil
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.13; rv:88.0) Gecko/20100101 Firefox/88.0")

	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		time.Sleep(100 * time.Millisecond)
		return GetWorkApi(UserCache, nodeId, workId, retryNum-1, err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	if strings.Contains(string(body), "502 Bad Gateway") {
		time.Sleep(time.Millisecond * 150) //延迟
		return GetWorkApi(UserCache, nodeId, workId, retryNum, err)
	}

	return string(body), nil
}

type YingHuaAnswer struct {
	Type    string   //题目类型
	Answers []string //回答内容
}

// SubmitWorkApi 提交作业答案接口
func SubmitWorkApi(UserCache YingHuaUserCache, workId, answerId string, answers YingHuaAnswer, finish string /*finish代表是否是最后提交并且结束考试，0代表不是，1代表是*/, retryNum int, lastError error) (string, error) {
	if retryNum < 0 {
		return "", lastError
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // 跳过证书验证
	}
	// Creating the HTTP client with a timeout (30 seconds)
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}

	// Create a buffer to hold the multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add form fields to the multipart data
	writer.WriteField("platform", "Android")
	writer.WriteField("version", "1.4.8")
	writer.WriteField("workId", workId)
	writer.WriteField("terminal", "Android")
	writer.WriteField("answerId", answerId)
	writer.WriteField("finish", finish)
	writer.WriteField("token", UserCache.token)
	if answers.Type == "单选" || answers.Type == "判断" || answers.Type == "简答" {
		writer.WriteField("answer", answers.Answers[0])
	} else if answers.Type == "多选" {
		for _, v := range answers.Answers {
			writer.WriteField("answer[]", v)
		}
	} else if answers.Type == "填空" {
		for i, v := range answers.Answers {
			writer.WriteField("answer_"+strconv.Itoa(i+1), v)
		}
	}

	// Close the writer to finalize the multipart form data
	writer.Close()

	// Create the request with the necessary headers
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/work/submit.json", UserCache.PreUrl), body)
	if err != nil {
		return "", err
	}

	// Set the headers
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.13; rv:88.0) Gecko/20100101 Firefox/88.0")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", UserCache.PreUrl)
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Type", writer.FormDataContentType())
	req.Header.Add("Cookie", UserCache.cookie)

	// Perform the request
	resp, err := client.Do(req)
	if err != nil {
		time.Sleep(100 * time.Millisecond)
		return SubmitWorkApi(UserCache, workId, answerId, answers, finish, retryNum-1, err)
	}
	defer resp.Body.Close()
	// Optionally, read the response body
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	if strings.Contains(string(bodyBytes), "502 Bad Gateway") {
		time.Sleep(time.Millisecond * 150) //延迟
		return SubmitWorkApi(UserCache, workId, answerId, answers, finish, retryNum, err)
	}
	return string(bodyBytes), nil
}

// WorkedDetail 获取最后作业得分接口
// {"_code":9,"status":false,"msg":"您已完成作业，该作业仅可答题1次","result":{}}
func WorkedFinallyDetailApi(userCache YingHuaUserCache, courseId, nodeId, workId string, retryNum int, lastError error) (string, error) {
	if retryNum < 0 {
		return "", lastError
	}
	url := userCache.PreUrl + "/api/work.json?nodeId=" + nodeId + "&workId=" + workId + "&token=" + userCache.token
	method := "GET"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("Cookie", userCache.cookie)
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.13; rv:88.0) Gecko/20100101 Firefox/88.0")

	res, err := client.Do(req)
	if err != nil {
		time.Sleep(100 * time.Millisecond)
		return WorkedFinallyDetailApi(userCache, courseId, nodeId, workId, retryNum-1, err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	if strings.Contains(string(body), "502 Bad Gateway") {
		time.Sleep(time.Millisecond * 150) //延迟
		return WorkedFinallyDetailApi(userCache, courseId, nodeId, workId, retryNum, err)
	}
	return string(body), nil
}

// WorkedDetail 获取最后作业得分接口
// {"_code":9,"status":false,"msg":"您已完成作业，该作业仅可答题1次","result":{}}
func ExamFinallyDetailApi(userCache YingHuaUserCache, courseId, nodeId, workId string, retryNum int, lastError error) (string, error) {
	if retryNum < 0 {
		return "", lastError
	}
	url := userCache.PreUrl + "/api/exam.json?nodeId=" + nodeId + "&examId=" + workId + "&token=" + userCache.token
	method := "GET"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("Cookie", userCache.cookie)
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.13; rv:88.0) Gecko/20100101 Firefox/88.0")

	res, err := client.Do(req)
	if err != nil {
		time.Sleep(100 * time.Millisecond)
		return ExamFinallyDetailApi(userCache, courseId, nodeId, workId, retryNum-1, err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	if strings.Contains(string(body), "502 Bad Gateway") {
		time.Sleep(time.Millisecond * 150) //延迟
		return ExamFinallyDetailApi(userCache, courseId, nodeId, workId, retryNum, err)
	}
	return string(body), nil
}
