package yinghua

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/yatori-dev/yatori-go-core/que-core/qentity"
	"github.com/yatori-dev/yatori-go-core/que-core/qtype"
	"github.com/yatori-dev/yatori-go-core/utils/qutils"

	"github.com/yatori-dev/yatori-go-core/utils"
)

type YingHuaUserCache struct {
	PreUrl    string //前置url
	Account   string //账号
	Password  string //用户密码
	IpProxySW bool   //是否开启IP代理
	ProxyIP   string //代理IP
	verCode   string //验证码
	cookie    string //验证码用的session
	cookies   []*http.Cookie
	token     string //保持会话的Token
	sign      string //签名
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

//func (cache YingHuaUserCache) String() string {
//
//	return ""
//}

// LoginApi 登录接口
func (cache *YingHuaUserCache) LoginApi(retry int, lastError error) (string, error) {
	if retry < 0 {
		return "", lastError
	}
	urlStr := cache.PreUrl + "/user/login.json"
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

	//如果开启了IP代理，那么就直接添加代理
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse(cache.ProxyIP) // 设置代理
		}
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	req, err := http.NewRequest(method, urlStr, payload)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	//req.Header.Add("Cookie", cache.cookie)
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)

	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		time.Sleep(150 * time.Millisecond)
		return cache.LoginApi(retry-1, err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		res.Body.Close() //立即释放
		fmt.Println(err)
		return "", err
	}
	//502情况进行重新请求
	if strings.Contains(string(body), "502 Bad Gateway") || strings.Contains(string(body), "504 Gateway Time-out") {
		res.Body.Close()                   //立即释放
		time.Sleep(time.Millisecond * 150) //延迟
		return cache.LoginApi(retry-1, err)
	}

	defer res.Body.Close()
	//fmt.Println(string(body))
	return string(body), nil
}

// VerificationCodeApi 获取验证码和SESSION验证码,并返回文件路径和SESSION字符串
var randChar []string = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f", "A", "B", "C", "D", "E", "F"}

func (cache *YingHuaUserCache) VerificationCodeApi(retry int) (string, string) {
	if retry < 0 {
		return "", ""
	}
	//1758349354572
	rand.Seed(time.Now().UnixNano())
	r := fmt.Sprintf("%.16f", rand.Float64())
	urlStr := fmt.Sprintf("%s/service/code?r=%s", cache.PreUrl, r)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}

	//如果开启了IP代理，那么就直接添加代理
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse(cache.ProxyIP) // 设置代理
		}
	}

	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	// 构建请求
	req, err := http.NewRequest("GET", urlStr, nil)

	if err != nil {
		fmt.Println(err)
		return "", ""
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Set("Connection", "keep-alive")
	res, err := client.Do(req)
	if err != nil {
		time.Sleep(150 * time.Millisecond)
		if strings.Contains(err.Error(), "A connection attempt failed because the connected party did not properly respond after a period of time") {
			return cache.VerificationCodeApi(retry)
		}
		return cache.VerificationCodeApi(retry - 1)
	}
	codeFileName := "code" + randChar[rand.Intn(len(randChar))] //生成验证码文件名称
	for i := 0; i < 10; i++ {
		codeFileName += randChar[rand.Intn(len(randChar))]
	}
	codeFileName += ".png"
	utils.PathExistForCreate("./assets/code/") //检测是否存在路径，如果不存在则创建
	filepath := fmt.Sprintf("./assets/code/%s", codeFileName)
	file, err := os.Create(filepath)
	if err != nil {
		res.Body.Close() //立即释放
		log.Println(err)
		return "", ""
	}

	_, err = io.Copy(file, res.Body)
	if err != nil {
		res.Body.Close() //立即释放
		log.Println(err)
		return "", ""
	}

	file.Close()
	if utils.IsBadImg(filepath) {
		res.Body.Close()           //立即释放
		utils.DeleteFile(filepath) //删除坏的文件
		return cache.VerificationCodeApi(retry - 1)
	}
	defer res.Body.Close()
	cache.cookies = res.Cookies()               //设置Cookie
	cache.cookie = res.Header.Get("Set-Cookie") //设置Cookie
	return filepath, res.Header.Get("Set-Cookie")
}

// KeepAliveApi 登录心跳保活
func KeepAliveApi(cache YingHuaUserCache, retry int) string {
	if retry < 0 {
		return ""
	}
	urlStr := cache.PreUrl + "/api/online.json"
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("platform", "Android")
	_ = writer.WriteField("version", "1.4.8")
	_ = writer.WriteField("token", cache.token)
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

	//如果开启了IP代理，那么就直接添加代理
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse(cache.ProxyIP) // 设置代理
		}
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	req, err := http.NewRequest(method, urlStr, payload)

	if err != nil {
		fmt.Println(err)
		return ""
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}
	res, err := client.Do(req)
	if err != nil {
		time.Sleep(time.Millisecond * 150) //延迟
		return KeepAliveApi(cache, retry)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		res.Body.Close() //立即释放
		fmt.Println(err)
		return ""
	}
	if strings.Contains(string(body), "502 Bad Gateway") || strings.Contains(string(body), "504 Gateway Time-out") {
		res.Body.Close()                   //立即释放
		time.Sleep(time.Millisecond * 150) //延迟
		return KeepAliveApi(cache, retry-1)
	}

	defer res.Body.Close()
	return string(body)
}

// CourseListApi 拉取课程列表API
func (cache *YingHuaUserCache) CourseListApi(retry int, lastError error) (string, error) {
	if retry < 0 {
		return "", lastError
	}
	urlStr := cache.PreUrl + "/api/course/list.json"
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

	//如果开启了IP代理，那么就直接添加代理
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse(cache.ProxyIP) // 设置代理
		}
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	req, err := http.NewRequest(method, urlStr, payload)
	//req.Header.Set("Cookie", cache.cookie)
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}
	if err != nil {
		return "", err
	}
	req.Header.Add("Cookie", "tgw_I7_route=3d5c4e13e7d88bb6849295ab943042a2")
	req.Header.Add("User-Agent", utils.DefaultUserAgent)

	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		time.Sleep(time.Millisecond * 150) //延迟
		if strings.Contains(err.Error(), "A connection attempt failed because the connected party did not properly respond after a period of time") {
			return cache.CourseListApi(retry, err)
		}
		return cache.CourseListApi(retry-1, err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		res.Body.Close()                   //立即释放
		time.Sleep(time.Millisecond * 150) //延迟
		return cache.CourseListApi(retry-1, err)
	}
	if strings.Contains(string(body), "502 Bad Gateway") || strings.Contains(string(body), "504 Gateway Time-out") {
		res.Body.Close()                   //立即释放
		time.Sleep(time.Millisecond * 150) //延迟
		return cache.CourseListApi(retry-1, err)
	}
	defer res.Body.Close()
	return string(body), nil
}

// CourseDetailApi 获取课程详细信息API
func (cache *YingHuaUserCache) CourseDetailApi(courseId string, retry int, lastError error) (string, error) {
	if retry < 0 {
		return "", lastError
	}
	urlStr := cache.PreUrl + "/api/course/detail.json"
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

	//如果开启了IP代理，那么就直接添加代理
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse(cache.ProxyIP) // 设置代理
		}
	}

	//如果开启了IP代理，那么就直接添加代理
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse(cache.ProxyIP) // 设置代理
		}
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	req, err := http.NewRequest(method, urlStr, payload)
	req.Header.Add("Cookie", cache.cookie)

	if err != nil {
		return "", err
	}
	//req.Header.Add("Cookie", "tgw_I7_route=3d5c4e13e7d88bb6849295ab943042a2")
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)

	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		time.Sleep(time.Millisecond * 150) //延迟
		if strings.Contains(err.Error(), "A connection attempt failed because the connected party did not properly respond after a period of time") {
			return cache.CourseDetailApi(courseId, retry, err)
		}
		return cache.CourseDetailApi(courseId, retry-1, err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		res.Body.Close()                   //立即释放
		time.Sleep(time.Millisecond * 150) //延迟
		return cache.CourseDetailApi(courseId, retry-1, err)
	}
	if strings.Contains(string(body), "502 Bad Gateway") || strings.Contains(string(body), "504 Gateway Time-out") {
		res.Body.Close()                   //立即释放
		time.Sleep(time.Millisecond * 150) //延迟
		return cache.CourseDetailApi(courseId, retry-1, err)
	}

	defer res.Body.Close()
	return string(body), err
}

// CourseVideListApi 对应课程的视屏列表
func CourseVideListApi(cache YingHuaUserCache, courseId string /*课程ID*/, retry int, lastError error) (string, error) {
	if retry < 0 {
		return "", lastError
	}
	urlStr := cache.PreUrl + "/api/course/chapter.json"
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("platform", "Android")
	_ = writer.WriteField("version", "1.4.8")
	_ = writer.WriteField("token", cache.token)
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

	//如果开启了IP代理，那么就直接添加代理
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse(cache.ProxyIP) // 设置代理
		}
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	req, err := http.NewRequest(method, urlStr, payload)
	req.Header.Set("Cookie", cache.cookie)
	if err != nil {
		time.Sleep(time.Millisecond * 150) //延迟
		return CourseVideListApi(cache, courseId, retry-1, err)
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}
	res, err := client.Do(req)
	if err != nil {
		time.Sleep(time.Millisecond * 150) //延迟
		if strings.Contains(err.Error(), "A connection attempt failed because the connected party did not properly respond after a period of time") {
			return CourseVideListApi(cache, courseId, retry, err)
		}
		return CourseVideListApi(cache, courseId, retry-1, err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		res.Body.Close()                   //立即释放
		time.Sleep(time.Millisecond * 150) //延迟
		return CourseVideListApi(cache, courseId, retry-1, err)
	}
	if strings.Contains(string(body), "502 Bad Gateway") || strings.Contains(string(body), "504 Gateway Time-out") {
		res.Body.Close()                   //立即释放
		time.Sleep(time.Millisecond * 150) //延迟
		return CourseVideListApi(cache, courseId, retry-1, err)
	}

	defer res.Body.Close()
	return string(body), nil
}

// SubmitStudyTimeApi 提交学时
func SubmitStudyTimeApi(cache YingHuaUserCache, nodeId string /*对应视屏节点ID*/, studyId string /*学习分配ID*/, studyTime int /*提交的学时*/, retry int, lastError error) (string, error) {
	if retry < 0 {
		return "", lastError
	}
	urlStr := cache.PreUrl + "/api/node/study.json"
	method := "POST"
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("platform", "Android")
	_ = writer.WriteField("version", "1.4.8")
	_ = writer.WriteField("nodeId", nodeId)
	_ = writer.WriteField("token", cache.token)
	_ = writer.WriteField("terminal", "Android")
	_ = writer.WriteField("studyTime", strconv.Itoa(studyTime))
	_ = writer.WriteField("studyId", studyId)
	err := writer.Close()
	if err != nil {
		time.Sleep(time.Millisecond * 150)
		return SubmitStudyTimeApi(cache, nodeId, studyId, studyTime, retry-1, err)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}

	//如果开启了IP代理，那么就直接添加代理
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse(cache.ProxyIP) // 设置代理
		}
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	req, err := http.NewRequest(method, urlStr, payload)

	if err != nil {
		//fmt.Println(err)
		return "", err
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}
	res, err := client.Do(req)
	if err != nil {
		time.Sleep(time.Millisecond * 150)
		return SubmitStudyTimeApi(cache, nodeId, studyId, studyTime, retry-1, err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		res.Body.Close() //立即释放
		time.Sleep(time.Millisecond * 150)
		return SubmitStudyTimeApi(cache, nodeId, studyId, studyTime, retry-1, err)
	}

	//避免502情况
	if strings.Contains(string(body), "502 Bad Gateway") || strings.Contains(string(body), "504 Gateway Time-out") {
		res.Body.Close()                   //立即释放
		time.Sleep(time.Millisecond * 150) //延迟
		return SubmitStudyTimeApi(cache, nodeId, studyId, studyTime, retry-1, err)
	}
	defer res.Body.Close()
	return string(body), nil
}

// VideStudyTimeApi 获取单个视屏的学习进度
func (cache *YingHuaUserCache) VideStudyTimeApi(nodeId string, retryNum int, lastError error) string {
	if retryNum < 0 {
		return ""
	}
	urlStr := cache.PreUrl + "/api/node/video.json"
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("platform", "Android")
	_ = writer.WriteField("version", "1.4.8")
	_ = writer.WriteField("nodeId", nodeId)
	_ = writer.WriteField("token", cache.token)
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

	//如果开启了IP代理，那么就直接添加代理
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse(cache.ProxyIP) // 设置代理
		}
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	req, err := http.NewRequest(method, urlStr, payload)

	if err != nil {
		time.Sleep(time.Millisecond * 150) //延迟
		return cache.VideStudyTimeApi(nodeId, retryNum-1, lastError)
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}
	res, err := client.Do(req)
	if err != nil {
		time.Sleep(time.Millisecond * 150) //延迟
		return cache.VideStudyTimeApi(nodeId, retryNum-1, lastError)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		time.Sleep(time.Millisecond * 150) //延迟
		return cache.VideStudyTimeApi(nodeId, retryNum-1, lastError)
	}
	if strings.Contains(string(body), "502 Bad Gateway") || strings.Contains(string(body), "504 Gateway Time-out") {
		time.Sleep(time.Millisecond * 150) //延迟
		return cache.VideStudyTimeApi(nodeId, retryNum-1, lastError)
	}
	return string(body)
}

// VideWatchRecodeApi 获取指定课程视屏观看记录
func VideWatchRecodeApi(cache YingHuaUserCache, courseId string, page int, retry int, lastError error) (string, error) {
	if retry < 0 {
		return "", lastError
	}
	urlStr := cache.PreUrl + "/api/record/video.json"
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("platform", "Android")
	_ = writer.WriteField("version", "1.4.8")
	_ = writer.WriteField("token", cache.token)
	_ = writer.WriteField("courseId", courseId)
	_ = writer.WriteField("page", strconv.Itoa(page))
	err := writer.Close()
	if err != nil {
		//fmt.Println(err)
		return VideWatchRecodeApi(cache, courseId, page, retry-1, err)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}

	//如果开启了IP代理，那么就直接添加代理
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse(cache.ProxyIP) // 设置代理
		}
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	req, err := http.NewRequest(method, urlStr, payload)
	req.Header.Set("Cookie", cache.cookie)
	if err != nil {
		//fmt.Println(err)
		return VideWatchRecodeApi(cache, courseId, page, retry-1, err)
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}
	res, err := client.Do(req)
	if err != nil {
		//fmt.Println(err)
		return VideWatchRecodeApi(cache, courseId, page, retry-1, err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		//fmt.Println(err)
		return VideWatchRecodeApi(cache, courseId, page, retry-1, err)
	}
	if strings.Contains(string(body), "502 Bad Gateway") || strings.Contains(string(body), "504 Gateway Time-out") {
		res.Body.Close()                   //立即释放
		time.Sleep(time.Millisecond * 150) //延迟
		return VideWatchRecodeApi(cache, courseId, page, retry-1, lastError)
	}
	defer res.Body.Close()
	return string(body), nil
}

// VideoWatchRecodePCListApi 获取指定课程视屏观看记录接口2，PC端
func VideoWatchRecodePCListApi(cache YingHuaUserCache, courseId string, page int, retry int, lastError error) (string, error) {
	if retry < 0 {
		return "", lastError
	}

	urlStr := cache.PreUrl + "/user/study_record/video.json?courseId=" + courseId + "&_=" + fmt.Sprintf("%d", time.Now().Unix()) + "&page=" + fmt.Sprintf("%d", page)
	method := "GET"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}

	//如果开启了IP代理，那么就直接添加代理
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse(cache.ProxyIP) // 设置代理
		}
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	req, err := http.NewRequest(method, urlStr, nil)
	//req.Header.Set("Cookie", cache.cookie)
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}
	if err != nil {
		//fmt.Println(err)
		return VideoWatchRecodePCListApi(cache, courseId, page, retry-1, err)
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)

	res, err := client.Do(req)
	if err != nil {
		//fmt.Println(err)
		return VideoWatchRecodePCListApi(cache, courseId, page, retry-1, err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		//fmt.Println(err)
		return VideoWatchRecodePCListApi(cache, courseId, page, retry-1, err)
	}
	if strings.Contains(string(body), "502 Bad Gateway") || strings.Contains(string(body), "504 Gateway Time-out") {
		res.Body.Close()                   //立即释放
		time.Sleep(time.Millisecond * 150) //延迟
		return VideoWatchRecodePCListApi(cache, courseId, page, retry-1, lastError)
	}
	defer res.Body.Close()
	return string(body), nil
}

// ExamDetailApi 获取考试信息
func ExamDetailApi(cache YingHuaUserCache, nodeId string, retryNum int, lastError error) (string, error) {
	if retryNum < 0 {
		return "", lastError
	}
	urlStr := cache.PreUrl + "/api/node/exam.json?nodeId=" + nodeId
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("platform", "Android")
	_ = writer.WriteField("version", "1.4.8")
	_ = writer.WriteField("nodeId", nodeId)
	_ = writer.WriteField("token", cache.token)
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

	//如果开启了IP代理，那么就直接添加代理
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse(cache.ProxyIP) // 设置代理
		}
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	req, err := http.NewRequest(method, urlStr, payload)
	//req.Header.Add("Cookie", cache.cookie)
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}
	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)

	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		time.Sleep(time.Millisecond * 150) //延迟
		return ExamDetailApi(cache, nodeId, retryNum-1, err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		res.Body.Close()                   //立即释放
		time.Sleep(time.Millisecond * 150) //延迟
		return ExamDetailApi(cache, nodeId, retryNum-1, err)
	}
	if strings.Contains(string(body), "502 Bad Gateway") || strings.Contains(string(body), "504 Gateway Time-out") {
		res.Body.Close()                   //立即释放
		time.Sleep(time.Millisecond * 150) //延迟
		return ExamDetailApi(cache, nodeId, retryNum-1, err)
	}
	defer res.Body.Close()
	return string(body), nil
}

// StartExam 开始考试接口
// {"_code":9,"status":false,"msg":"考试测试时间还未开始","result":{}}
func StartExam(cache YingHuaUserCache, courseId, nodeId, examId string, retryNum int, lastError error) (string, error) {
	if retryNum < 0 {
		return "", lastError
	}
	urlStr := cache.PreUrl + "/api/exam/start.json?nodeId=" + nodeId + "&courseId=" + courseId + "&token=" + cache.token + "&examId=" + examId
	method := "GET"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}

	//如果开启了IP代理，那么就直接添加代理
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse(cache.ProxyIP) // 设置代理
		}
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	req, err := http.NewRequest(method, urlStr, nil)

	if err != nil {
		fmt.Println(err)
		time.Sleep(100 * time.Millisecond)
		return StartExam(cache, courseId, nodeId, examId, retryNum-1, err)
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		time.Sleep(100 * time.Millisecond)
		return StartExam(cache, courseId, nodeId, examId, retryNum-1, err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		res.Body.Close() //立即释放
		fmt.Println(err)
		time.Sleep(100 * time.Millisecond)
		return StartExam(cache, courseId, nodeId, examId, retryNum-1, err)
	}
	if strings.Contains(string(body), "502 Bad Gateway") || strings.Contains(string(body), "504 Gateway Time-out") {
		res.Body.Close()                   //立即释放
		time.Sleep(time.Millisecond * 150) //延迟
		return StartExam(cache, courseId, nodeId, examId, retryNum-1, lastError)
	}
	defer res.Body.Close()
	return string(body), nil
}

// GetExamTopicApi 获取所有考试题目，但是HTML，建议配合TurnExamTopic函数使用将题目html转成结构体
func GetExamTopicApi(cache YingHuaUserCache, nodeId, examId string, retryNum int, lastError error) (string, error) {
	if retryNum < 0 {
		return "", lastError
	}
	// Creating a custom HTTP client with timeout and SSL context (skip SSL setup for simplicity)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}

	//如果开启了IP代理，那么就直接添加代理
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse(cache.ProxyIP) // 设置代理
		}
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   30 * time.Second,
	}

	// Creating the request body (empty JSON object)
	body := []byte("{}")

	// Create the request
	url := fmt.Sprintf("%s/api/exam.json?nodeId=%s&examId=%s&token=%s", cache.PreUrl, nodeId, examId, cache.token)
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	// Set the headers
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", cache.PreUrl)
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}
	// Perform the request
	resp, err := client.Do(req)
	if err != nil {
		time.Sleep(100 * time.Millisecond)
		return GetExamTopicApi(cache, nodeId, examId, retryNum-1, err)
	}

	// Read the response body
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close() //立即释放
		time.Sleep(100 * time.Millisecond)
		return GetExamTopicApi(cache, nodeId, examId, retryNum-1, err)
	}
	if strings.Contains(string(body), "502 Bad Gateway") || strings.Contains(string(body), "504 Gateway Time-out") {
		resp.Body.Close()                  //立即释放
		time.Sleep(time.Millisecond * 150) //延迟
		return GetExamTopicApi(cache, nodeId, examId, retryNum-1, err)
	}
	defer resp.Body.Close()
	return string(bodyBytes), nil
}

// SubmitExamApi 提交考试答案接口
func SubmitExamApi(cache YingHuaUserCache, examId, answerId string, answers qentity.Question, finish string, retryNum int, lastError error) (string, error) {
	if retryNum < 0 {
		return "", lastError
	}
	// Creating the HTTP client with a timeout (30 seconds)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}

	//如果开启了IP代理，那么就直接添加代理
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse(cache.ProxyIP) // 设置代理
		}
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
	writer.WriteField("token", cache.token)

	// Add the answer fields
	if answers.Type == "单选" || answers.Type == "判断" || answers.Type == "简答" {
		//writer.WriteField("answer", answers.Answers[0])
		writer.WriteField("answer", qutils.SimilarityArraySelect(answers.Answers[0], answers.Answers))
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
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/exam/submit.json", cache.PreUrl), body)
	if err != nil {
		time.Sleep(100 * time.Millisecond)
		return SubmitExamApi(cache, examId, answerId, answers, finish, retryNum-1, err)
	}

	// Set the headers
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", cache.PreUrl)
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Type", writer.FormDataContentType())
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}
	// Perform the request
	resp, err := client.Do(req)
	if err != nil {
		time.Sleep(100 * time.Millisecond)
		return SubmitExamApi(cache, examId, answerId, answers, finish, retryNum-1, err)
	}

	// Read the response body (we're not using the body here, just ensuring the request goes through)
	bodyStr, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close() //立即释放
		return "", err
	}

	if strings.Contains(string(bodyStr), "502 Bad Gateway") || strings.Contains(string(bodyStr), "504 Gateway Time-out") {
		resp.Body.Close()                  //立即释放
		time.Sleep(time.Millisecond * 150) //延迟
		return SubmitExamApi(cache, examId, answerId, answers, finish, retryNum-1, err)
	}
	defer resp.Body.Close()
	return string(bodyStr), nil
}

// WorkDetailApi 获取作业信息
func WorkDetailApi(cache YingHuaUserCache, nodeId string, retryNum int, lastError error) (string, error) {
	if retryNum < 0 {
		return "", lastError
	}
	urlStr := cache.PreUrl + "/api/node/work.json?nodeId=" + nodeId
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("platform", "Android")
	_ = writer.WriteField("version", "1.4.8")
	_ = writer.WriteField("nodeId", nodeId)
	_ = writer.WriteField("token", cache.token)
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

	//如果开启了IP代理，那么就直接添加代理
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse(cache.ProxyIP) // 设置代理
		}
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	req, err := http.NewRequest(method, urlStr, payload)
	//req.Header.Add("Cookie", cache.cookie)
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)

	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		time.Sleep(100 * time.Millisecond)
		return WorkDetailApi(cache, nodeId, retryNum-1, err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		time.Sleep(100 * time.Millisecond)
		return WorkDetailApi(cache, nodeId, retryNum-1, err)
	}
	if strings.Contains(string(body), "502 Bad Gateway") || strings.Contains(string(body), "504 Gateway Time-out") {
		time.Sleep(time.Millisecond * 150) //延迟
		return WorkDetailApi(cache, nodeId, retryNum-1, err)
	}
	return string(body), nil
}

// StartWork 开始做作业接口
// {"_code":9,"status":false,"msg":"您已完成作业，该作业仅可答题1次","result":{}}
func StartWork(userCache YingHuaUserCache, courseId, nodeId, workId string, retryNum int, lastError error) (string, error) {
	if retryNum < 0 {
		return "", lastError
	}
	urlStr := userCache.PreUrl + "/api/work/start.json?nodeId=" + nodeId + "&courseId=" + courseId + "&token=" + userCache.token + "&workId=" + workId
	method := "GET"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}

	//如果开启了IP代理，那么就直接添加代理
	if userCache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse(userCache.ProxyIP) // 设置代理
		}
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	req, err := http.NewRequest(method, urlStr, nil)

	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	for _, cookie := range userCache.cookies {
		req.AddCookie(cookie)
	}
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
	if strings.Contains(string(body), "502 Bad Gateway") || strings.Contains(string(body), "504 Gateway Time-out") {
		time.Sleep(time.Millisecond * 150) //延迟
		return StartWork(userCache, courseId, nodeId, workId, retryNum-1, err)
	}
	return string(body), nil
}

// GetWorkApi 获取所有作业题目
func GetWorkApi(UserCache YingHuaUserCache, nodeId, workId string, retryNum int, lastError error) (string, error) {
	if retryNum < 0 {
		return "", lastError
	}
	urlStr := UserCache.PreUrl + "/api/work.json?nodeId=" + nodeId + "&workId=" + workId + "&token=" + UserCache.token
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

	//如果开启了IP代理，那么就直接添加代理
	if UserCache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse(UserCache.ProxyIP) // 设置代理
		}
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	req, err := http.NewRequest(method, urlStr, payload)

	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	for _, cookie := range UserCache.cookies {
		req.AddCookie(cookie)
	}
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
	if strings.Contains(string(body), "502 Bad Gateway") || strings.Contains(string(body), "504 Gateway Time-out") {
		time.Sleep(time.Millisecond * 150) //延迟
		return GetWorkApi(UserCache, nodeId, workId, retryNum-1, err)
	}

	return string(body), nil
}

type YingHuaAnswer struct {
	Type    string   //题目类型
	Answers []string //回答内容
}

// SubmitWorkApi 提交作业答案接口
func SubmitWorkApi(cache YingHuaUserCache, workId, answerId string, answers qentity.Question, finish string /*finish代表是否是最后提交并且结束考试，0代表不是，1代表是*/, retryNum int, lastError error) (string, error) {
	if retryNum < 0 {
		return "", lastError
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}

	//如果开启了IP代理，那么就直接添加代理
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse(cache.ProxyIP) // 设置代理
		}
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
	writer.WriteField("token", cache.token)

	//单选，判断
	if answers.Type == qtype.SingleChoice.String() || answers.Type == qtype.TrueOrFalse.String() {

	}
	//多选题
	if answers.Type == qtype.MultipleChoice.String() {

	}
	//填空题
	if answers.Type == qtype.SingleChoice.String() {
		for i, v := range answers.Answers {
			writer.WriteField("answer_"+strconv.Itoa(i+1), v)
		}
	}
	//if answers.Type == "单选" || answers.Type == "判断" || answers.Type == "简答" {
	//	writer.WriteField("answer", answers.Answers[0])
	//} else if answers.Type == "多选" {
	//	for _, v := range answers.Answers {
	//		writer.WriteField("answer[]", v)
	//	}
	//} else if answers.Type == "填空" {
	//	for i, v := range answers.Answers {
	//		writer.WriteField("answer_"+strconv.Itoa(i+1), v)
	//	}
	//}

	// Close the writer to finalize the multipart form data
	writer.Close()

	// Create the request with the necessary headers
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/work/submit.json", cache.PreUrl), body)
	if err != nil {
		return "", err
	}

	// Set the headers
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", cache.PreUrl)
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Type", writer.FormDataContentType())
	//req.Header.Add("Cookie", cache.cookie)
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}

	// Perform the request
	resp, err := client.Do(req)
	if err != nil {
		time.Sleep(100 * time.Millisecond)
		return SubmitWorkApi(cache, workId, answerId, answers, finish, retryNum-1, err)
	}
	defer resp.Body.Close()
	// Optionally, read the response body
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	if strings.Contains(string(bodyBytes), "502 Bad Gateway") || strings.Contains(string(bodyBytes), "504 Gateway Time-out") {
		time.Sleep(time.Millisecond * 150) //延迟
		return SubmitWorkApi(cache, workId, answerId, answers, finish, retryNum-1, err)
	}
	return string(bodyBytes), nil
}

// WorkedDetail 获取最后作业得分接口
// {"_code":9,"status":false,"msg":"您已完成作业，该作业仅可答题1次","result":{}}
func WorkedFinallyDetailApi(userCache YingHuaUserCache, courseId, nodeId, workId string, retryNum int, lastError error) (string, error) {
	if retryNum < 0 {
		return "", lastError
	}
	urlStr := userCache.PreUrl + "/api/work.json?nodeId=" + nodeId + "&workId=" + workId + "&token=" + userCache.token
	method := "GET"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}

	//如果开启了IP代理，那么就直接添加代理
	if userCache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse(userCache.ProxyIP) // 设置代理
		}
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	req, err := http.NewRequest(method, urlStr, nil)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	//req.Header.Add("Cookie", userCache.cookie)
	for _, cookie := range userCache.cookies {
		req.AddCookie(cookie)
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)

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
	if strings.Contains(string(body), "502 Bad Gateway") || strings.Contains(string(body), "504 Gateway Time-out") {
		time.Sleep(time.Millisecond * 150) //延迟
		return WorkedFinallyDetailApi(userCache, courseId, nodeId, workId, retryNum-1, err)
	}
	return string(body), nil
}

// WorkedDetail 获取最后作业得分接口
// {"_code":9,"status":false,"msg":"您已完成作业，该作业仅可答题1次","result":{}}
func ExamFinallyDetailApi(userCache YingHuaUserCache, courseId, nodeId, workId string, retryNum int, lastError error) (string, error) {
	if retryNum < 0 {
		return "", lastError
	}
	urlStr := userCache.PreUrl + "/api/exam.json?nodeId=" + nodeId + "&examId=" + workId + "&token=" + userCache.token
	method := "GET"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}

	//如果开启了IP代理，那么就直接添加代理
	if userCache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse(userCache.ProxyIP) // 设置代理
		}
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	req, err := http.NewRequest(method, urlStr, nil)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	//req.Header.Add("Cookie", userCache.cookie)
	for _, cookie := range userCache.cookies {
		req.AddCookie(cookie)
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)

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
	if strings.Contains(string(body), "502 Bad Gateway") || strings.Contains(string(body), "504 Gateway Time-out") {
		time.Sleep(time.Millisecond * 150) //延迟
		return ExamFinallyDetailApi(userCache, courseId, nodeId, workId, retryNum-1, err)
	}
	return string(body), nil
}
