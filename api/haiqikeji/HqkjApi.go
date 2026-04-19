package haiqikeji

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type HqkjUserCache struct {
	PreUrl    string //前置url
	Account   string //账号
	Password  string //用户密码
	IpProxySW bool   //是否开启IP代理
	ProxyIP   string //代理IP
	verCode   string //验证码
	cookie    string //验证码用的session
	cookies   []*http.Cookie
	Token     string //保持会话的Token
	sign      string //签名
	UserId    string
	SchoolId  string //学校id
}

// 登录
func (cache *HqkjUserCache) LoginApi(retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}

	//如果开启了IP代理，那么就直接添加代理
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse("http://" + cache.ProxyIP) // 设置代理
		}
	}

	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	req, err := http.NewRequest("GET", cache.PreUrl+"/api/user/login?number="+url.QueryEscape(cache.Account)+"&password="+url.QueryEscape(cache.Password)+"&schoolId="+cache.SchoolId, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("accept-language", "zh-CN,zh;q=0.9,zh-TW;q=0.8,en;q=0.7")
	req.Header.Set("cache-control", "no-cache")
	req.Header.Set("pragma", "no-cache")
	req.Header.Set("priority", "u=1, i")
	req.Header.Set("sec-ch-ua", `"Not:A-Brand";v="99", "Google Chrome";v="145", "Chromium";v="145"`)
	req.Header.Set("sec-ch-ua-mobile", "?1")
	req.Header.Set("sec-ch-ua-platform", `"Android"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-origin")
	req.Header.Set("user-agent", "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Mobile Safari/537.36")
	req.Header.Set("cookie", "__root_domain_v=.haiqikeji.com; _qddaz=QD.247873038960618; _qdda=3-1.1; _qddab=3-vn3xs9.mmitm9t9")
	resp, err := client.Do(req)
	if err != nil {
		//log.Fatal(err)
		return cache.LoginApi(retry-1, err)
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Printf("%s\n", bodyText)
	return string(bodyText), nil
}

// 用于获取School账号信息
func (cache *HqkjUserCache) PullSchoolInfoApi(urlStr string, retry int, lastErr error) string {
	if retry < 0 {
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
			return url.Parse("http://" + cache.ProxyIP) // 设置代理
		}
	}

	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	//req, err := http.NewRequest("GET", "https://swxy.haiqikeji.com/api/course/selectdomain?domain=swxy.haiqikeji.com", nil)
	req, err := http.NewRequest("GET", cache.PreUrl+"/api/course/selectdomain?domain="+urlStr, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("accept-language", "zh-CN,zh;q=0.9,zh-TW;q=0.8,en;q=0.7")
	req.Header.Set("cache-control", "no-cache")
	req.Header.Set("pragma", "no-cache")
	req.Header.Set("priority", "u=1, i")
	req.Header.Set("referer", cache.PreUrl)
	req.Header.Set("sec-ch-ua", `"Not:A-Brand";v="99", "Google Chrome";v="145", "Chromium";v="145"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-origin")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36")
	req.Header.Set("cookie", "__root_domain_v=.haiqikeji.com; _qddaz=QD.247873038960618; _qdda=3-1.1; _qddab=3-b2u746.mmrwmvxj")
	resp, err := client.Do(req)
	if err != nil {
		//log.Fatal(err)
		return cache.PullSchoolInfoApi(urlStr, retry-1, err)
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Printf("%s\n", bodyText)
	return string(bodyText)
}

// 拉取用户个人信息
func (cache *HqkjUserCache) PullUserInfoApi(retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}

	//如果开启了IP代理，那么就直接添加代理
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse("http://" + cache.ProxyIP) // 设置代理
		}
	}

	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	req, err := http.NewRequest("GET", cache.PreUrl+"/api/user/yee_student_info", nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("accept-language", "zh-CN,zh;q=0.9,zh-TW;q=0.8,en;q=0.7")
	req.Header.Set("authorization", cache.Token)
	req.Header.Set("cache-control", "no-cache")
	req.Header.Set("pragma", "no-cache")
	req.Header.Set("priority", "u=1, i")
	req.Header.Set("sec-ch-ua", `"Not:A-Brand";v="99", "Google Chrome";v="145", "Chromium";v="145"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-origin")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36")
	req.Header.Set("cookie", "__root_domain_v=.haiqikeji.com; _qddaz=QD.247873038960618; _qddab=3-8f2fdp.mms0w07g")
	resp, err := client.Do(req)
	if err != nil {
		//log.Fatal(err)
		return cache.PullUserInfoApi(retry-1, err)
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	return string(bodyText), nil
}

func (cache *HqkjUserCache) PullCourseListApi(retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}

	//如果开启了IP代理，那么就直接添加代理
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse("http://" + cache.ProxyIP) // 设置代理
		}
	}

	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	req, err := http.NewRequest("GET", cache.PreUrl+"/api/user/yee_my_course_list?schoolId="+cache.SchoolId+"&studentId="+cache.UserId+"&type=0&pageNum=1&pageSize=10000&_t=1773594164210", nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("accept-language", "zh-CN,zh;q=0.9,zh-TW;q=0.8,en;q=0.7")
	req.Header.Set("authorization", cache.Token)
	req.Header.Set("cache-control", "no-cache")
	req.Header.Set("pragma", "no-cache")
	req.Header.Set("priority", "u=1, i")
	req.Header.Set("referer", "https://swxy.haiqikeji.com/student/home")
	req.Header.Set("sec-ch-ua", `"Not:A-Brand";v="99", "Google Chrome";v="145", "Chromium";v="145"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-origin")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36")
	req.Header.Set("cookie", "__root_domain_v=.haiqikeji.com; _qddaz=QD.247873038960618; _qddab=3-b2u746.mmrwmvxj")
	resp, err := client.Do(req)
	if err != nil {
		//log.Fatal(err)
		return cache.PullCourseListApi(retry-1, err)
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Printf("%s\n", bodyText)
	return string(bodyText), nil
}

// 拉取大章节
func (cache *HqkjUserCache) PullChapterListApi(courseId string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	form := new(bytes.Buffer)
	writer := multipart.NewWriter(form)
	formField, err := writer.CreateFormField("schoolId")
	if err != nil {
		log.Fatal(err)
	}
	_, err = formField.Write([]byte(cache.SchoolId))

	formField, err = writer.CreateFormField("courseId")
	if err != nil {
		log.Fatal(err)
	}
	//_, err = formField.Write([]byte("1012279"))
	_, err = formField.Write([]byte(courseId))

	writer.Close()

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}

	//如果开启了IP代理，那么就直接添加代理
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse("http://" + cache.ProxyIP) // 设置代理
		}
	}

	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	req, err := http.NewRequest("POST", cache.PreUrl+"/api/user/yee_chapter_select", form)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("accept-language", "zh-CN,zh;q=0.9,zh-TW;q=0.8,en;q=0.7")
	req.Header.Set("authorization", cache.Token)
	req.Header.Set("cache-control", "no-cache")
	req.Header.Set("origin", cache.PreUrl)
	req.Header.Set("pragma", "no-cache")
	req.Header.Set("priority", "u=1, i")
	req.Header.Set("sec-ch-ua", `"Not:A-Brand";v="99", "Google Chrome";v="145", "Chromium";v="145"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-origin")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36")
	req.Header.Set("cookie", "__root_domain_v=.haiqikeji.com; _qddaz=QD.247873038960618; _qdda=3-1.luvll; _qddab=3-8f2fdp.mms0w07g")
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := client.Do(req)
	if err != nil {
		//log.Fatal(err)
		return cache.PullChapterListApi(courseId, retry-1, err)
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Printf("%s\n", bodyText)
	return string(bodyText), nil
}

// 拉取大章节对应下的小节点
func (cache *HqkjUserCache) PullChapterNodeListApi(chapterId string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}

	//如果开启了IP代理，那么就直接添加代理
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse("http://" + cache.ProxyIP) // 设置代理
		}
	}

	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	//req, err := http.NewRequest("GET", "https://swxy.haiqikeji.com/api/user/yee_node_select?schoolId=15&chapterId=1101498", nil)
	req, err := http.NewRequest("GET", cache.PreUrl+"/api/user/yee_node_select?schoolId="+cache.SchoolId+"&chapterId="+chapterId, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("accept-language", "zh-CN,zh;q=0.9,zh-TW;q=0.8,en;q=0.7")
	req.Header.Set("authorization", cache.Token)
	req.Header.Set("cache-control", "no-cache")
	req.Header.Set("pragma", "no-cache")
	req.Header.Set("priority", "u=1, i")
	req.Header.Set("sec-ch-ua", `"Not:A-Brand";v="99", "Google Chrome";v="145", "Chromium";v="145"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-origin")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36")
	req.Header.Set("cookie", "__root_domain_v=.haiqikeji.com; _qddaz=QD.247873038960618; _qdda=3-1.luvll; _qddab=3-8f2fdp.mms0w07g")
	resp, err := client.Do(req)
	if err != nil {
		//log.Fatal(err)
		return cache.PullChapterNodeListApi(chapterId, retry-1, err)
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return string(bodyText), nil
}

// 获取视频学习进度
func (cache *HqkjUserCache) PullLastProgressApi(nodeId string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}

	//如果开启了IP代理，那么就直接添加代理
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse("http://" + cache.ProxyIP) // 设置代理
		}
	}

	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	req, err := http.NewRequest("GET", cache.PreUrl+"/api/user/last_progress?nodeId="+nodeId+"&userId="+cache.UserId+"&schoolId="+cache.SchoolId, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("accept-language", "zh-CN,zh;q=0.9,zh-TW;q=0.8,en;q=0.7")
	req.Header.Set("authorization", cache.Token)
	req.Header.Set("cache-control", "no-cache")
	req.Header.Set("pragma", "no-cache")
	req.Header.Set("priority", "u=1, i")
	req.Header.Set("sec-ch-ua", `"Not:A-Brand";v="99", "Google Chrome";v="145", "Chromium";v="145"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-origin")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36")
	//req.Header.Set("cookie", "__root_domain_v=.haiqikeji.com; _qddaz=QD.247873038960618; _qdda=3-1.luvll; _qddab=3-8f2fdp.mms0w07g")
	resp, err := client.Do(req)
	if err != nil {
		//log.Fatal(err)
		return cache.PullLastProgressApi(nodeId, retry-1, err)
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return string(bodyText), nil
}

// 打点开始学习，获取学习id
func (cache *HqkjUserCache) StartStudyApi(nodeId, courseId string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}

	//如果开启了IP代理，那么就直接添加代理
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse("http://" + cache.ProxyIP) // 设置代理
		}
	}

	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	//var data = strings.NewReader(`{"schoolId":`+cache.SchoolId+`,"userId":1257795,"nodeId":"1473241","courseId":"1012279","terminal":"web"}`)
	var data = strings.NewReader(`{"schoolId":` + cache.SchoolId + `,"userId":` + cache.UserId + `,"nodeId":"` + nodeId + `","courseId":"` + courseId + `","terminal":"web"}`)
	req, err := http.NewRequest("POST", cache.PreUrl+"/api/user/study_session_start", data)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("accept-language", "zh-CN,zh;q=0.9,zh-TW;q=0.8,en;q=0.7")
	req.Header.Set("authorization", cache.Token)
	req.Header.Set("cache-control", "no-cache")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("origin", cache.PreUrl)
	req.Header.Set("pragma", "no-cache")
	req.Header.Set("priority", "u=1, i")
	req.Header.Set("sec-ch-ua", `"Not:A-Brand";v="99", "Google Chrome";v="145", "Chromium";v="145"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-origin")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36")
	//req.Header.Set("cookie", "__root_domain_v=.haiqikeji.com; _qddaz=QD.247873038960618; _qdda=3-1.luvll; _qddab=3-8f2fdp.mms0w07g")
	resp, err := client.Do(req)
	if err != nil {
		//log.Fatal(err)
		return cache.StartStudyApi(nodeId, courseId, retry-1, err)
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Printf("%s\n", bodyText)
	return string(bodyText), nil
}

// 提交学时
func (cache *HqkjUserCache) SubmitStudyTimeApi(sessionId string, progress int, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}

	//如果开启了IP代理，那么就直接添加代理
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse("http://" + cache.ProxyIP) // 设置代理
		}
	}

	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	//var data = strings.NewReader(`{"sessionId":"20410346-558e-4b8c-8166-48fa58d76abf","progress":"3"}`)
	var data = strings.NewReader(`{"sessionId":"` + sessionId + `","progress":"` + fmt.Sprintf("%d", progress) + `"}`)
	req, err := http.NewRequest("POST", cache.PreUrl+"/api/user/study_session_heartbeat", data)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("accept-language", "zh-CN,zh;q=0.9,zh-TW;q=0.8,en;q=0.7")
	req.Header.Set("authorization", cache.Token)
	req.Header.Set("cache-control", "no-cache")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("origin", cache.PreUrl)
	req.Header.Set("pragma", "no-cache")
	req.Header.Set("priority", "u=1, i")
	req.Header.Set("sec-ch-ua", `"Not:A-Brand";v="99", "Google Chrome";v="145", "Chromium";v="145"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-origin")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		//log.Fatal(err)
		return cache.SubmitStudyTimeApi(sessionId, progress, retry-1, err)
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	return string(bodyText), nil
}

func (cache *HqkjUserCache) PullWorkInfoApi(courseId, nodeId string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}

	//如果开启了IP代理，那么就直接添加代理
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse("http://" + cache.ProxyIP) // 设置代理
		}
	}

	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	req, err := http.NewRequest("GET", cache.PreUrl+"/api/user/yee_course_student_work_list?schoolId="+cache.SchoolId+"&studentId="+cache.UserId+"&courseId="+courseId+"&nodeId="+nodeId, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("accept-language", "zh-CN,zh;q=0.9,zh-TW;q=0.8,en;q=0.7")
	req.Header.Set("authorization", cache.Token)
	req.Header.Set("cache-control", "no-cache")
	req.Header.Set("pragma", "no-cache")
	req.Header.Set("priority", "u=1, i")

	req.Header.Set("sec-ch-ua", `"Not:A-Brand";v="99", "Google Chrome";v="145", "Chromium";v="145"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-origin")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Printf("%s\n", bodyText)
	return string(bodyText), nil
}

// 结束session并保存记录
func (cache HqkjUserCache) EndStudyApi(sessionId string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}

	//如果开启了IP代理，那么就直接添加代理
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse("http://" + cache.ProxyIP) // 设置代理
		}
	}

	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	var data = strings.NewReader(`{"sessionId":"` + sessionId + `"}`)
	req, err := http.NewRequest("POST", cache.PreUrl+"/api/user/study_session_end", data)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("accept-language", "zh-CN,zh;q=0.9,zh-TW;q=0.8,en;q=0.7")
	req.Header.Set("authorization", cache.Token)
	req.Header.Set("cache-control", "no-cache")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("origin", cache.PreUrl)
	req.Header.Set("pragma", "no-cache")
	req.Header.Set("priority", "u=1, i")
	req.Header.Set("sec-ch-ua", `"Not:A-Brand";v="99", "Google Chrome";v="145", "Chromium";v="145"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-origin")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Printf("%s\n", bodyText)
	return string(bodyText), nil
}

// 拉取章节中的考试列表
func (cache HqkjUserCache) PullExamListApi(courseId, nodeId string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}

	//如果开启了IP代理，那么就直接添加代理
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse("http://" + cache.ProxyIP) // 设置代理
		}
	}

	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	//urlStr := "https://whxyart.haiqikeji.com/api/user/yee_course_student_exam_list?schoolId=" + cache.SchoolId + "&studentId=1270478&courseId=1009407&nodeId=1508247"
	urlStr := "https://whxyart.haiqikeji.com/api/user/yee_course_student_exam_list?schoolId=" + cache.SchoolId + "&studentId=" + cache.UserId + "&courseId=" + courseId + "&nodeId=" + nodeId
	method := "GET"

	req, err := http.NewRequest(method, urlStr, nil)

	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	req.Header.Add("authorization", cache.Token)
	req.Header.Add("pragma", "no-cache")
	req.Header.Add("priority", "u=1, i")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Connection", "keep-alive")

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
	//fmt.Println(string(body))
	return string(body), nil
}

// 拉取考试详细信息
func (cache HqkjUserCache) PullExamDetailInformApi(examId, courseId, title string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	//url := "https://whxyart.haiqikeji.com/api/user/yee_course_student_exam_detail?schoolId=25&studentId=1270478&examId=1003087&courseId=1009407&title=%25E5%2588%259B%25E4%25B8%259A%25E5%259F%25BA%25E7%25A1%2580%25E8%2580%2583%25E8%25AF%2595"
	urlStr := "https://whxyart.haiqikeji.com/api/user/yee_course_student_exam_detail?schoolId=" + cache.SchoolId + "&studentId=" + cache.UserId + "&examId=" + examId + "&courseId=" + courseId + "&title=" + url.QueryEscape(title)
	method := "GET"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}

	//如果开启了IP代理，那么就直接添加代理
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse("http://" + cache.ProxyIP) // 设置代理
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
	req.Header.Add("authorization", cache.Token)
	req.Header.Add("pragma", "no-cache")
	req.Header.Add("priority", "u=1, i")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Connection", "keep-alive")

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
	//fmt.Println(string(body))
	return string(body), nil
}

// 进入考试
func (cache HqkjUserCache) PullEnterExamApi(examId, courseId, title string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	//url := "https://whxyart.haiqikeji.com/api/user/yee_course_student_exam_detail?schoolId=25&studentId=1270478&examId=1003087&courseId=1009407&title=%25E5%2588%259B%25E4%25B8%259A%25E5%259F%25BA%25E7%25A1%2580%25E8%2580%2583%25E8%25AF%2595"
	urlStr := "https://whxyart.haiqikeji.com/api/user/yee_course_student_exam_detail?schoolId=" + cache.SchoolId + "&studentId=" + cache.UserId + "&examId=" + examId + "&courseId=" + courseId + "&title=" + url.QueryEscape(title)
	method := "GET"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}

	//如果开启了IP代理，那么就直接添加代理
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse("http://" + cache.ProxyIP) // 设置代理
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
	req.Header.Add("authorization", cache.Token)
	req.Header.Add("pragma", "no-cache")
	req.Header.Add("priority", "u=1, i")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Connection", "keep-alive")

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
	//fmt.Println(string(body))
	return string(body), nil
}

// 拉取题目列表
func (cache HqkjUserCache) PullExamQuestionsApi(examId, courseId, title string, retry int, lastErr error) (string, error) {

	url := "https://whxyart.haiqikeji.com/api/user/yee_course_student_exam_start?schoolId=25&studentId=1270478&courseId=1009407&examId=1003087&platform=pc&createUserId=1249456&state=1&classId=1016378&paperId=6196&random=&randData=%257B%257D&randNumber="
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	req.Header.Add("authorization", cache.Token)
	req.Header.Add("pragma", "no-cache")
	req.Header.Add("priority", "u=1, i")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Connection", "keep-alive")

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
	//fmt.Println(string(body))
	return string(body), nil
}

// 提交答案
func (cache HqkjUserCache) AnswerApi(courseId, examId, topicId, recordId, questionType string, answers []string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	//url := "https://whxyart.haiqikeji.com/api/user/yee_exam_answer_add"
	urlStr := "https://whxyart.haiqikeji.com/api/user/yee_exam_answer_add"
	method := "POST"

	//payload := strings.NewReader(`{"schoolId":25,"courseId":1009407,"userId":1270478,"examId":1003087,"topicId":1253487,"answer":["B"],"recordId":57198,"type":1}`)
	answersData, err := json.Marshal(answers)
	if err != nil {
		panic(err)
	}

	payload := strings.NewReader(`{"schoolId":` + cache.SchoolId + `,"courseId":` + courseId + `,"userId":` + cache.UserId + `,"examId":` + examId + `,"topicId":` + topicId + `,"answer":` + string(answersData) + `,"recordId":` + recordId + `,"type":` + questionType + `}`)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}

	//如果开启了IP代理，那么就直接添加代理
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse("http://" + cache.ProxyIP) // 设置代理
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
	req.Header.Add("authorization", cache.Token)
	req.Header.Add("pragma", "no-cache")
	req.Header.Add("priority", "u=1, i")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36")
	req.Header.Add("content-type", "application/json")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Connection", "keep-alive")

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
	//fmt.Println(string(body))
	return string(body), nil
}
