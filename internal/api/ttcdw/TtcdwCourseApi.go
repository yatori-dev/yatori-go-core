package ttcdw

import (
	"bytes"
	"crypto/des"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/yatori-dev/yatori-go-core/v3/utils"
)

// 拉取所有项目
func (cache *TtcdwUserCache) PullProjectApi(retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	urlStr := "https://www.ttcdw.cn/m/open/app/v1/memProject/list?state=1&pageNum=1&pageSize=100"
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
		Transport: tr,
	}
	req, err := http.NewRequest(method, urlStr, nil)

	if err != nil {
		//fmt.Println(err)
		return "", err
	}
	//设置Cookie
	for _, v := range cache.Cookies {
		req.AddCookie(v)
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "www.ttcdw.cn")
	req.Header.Add("Connection", "keep-alive")

	res, err := client.Do(req)
	if err != nil {
		time.Sleep(150 * time.Millisecond)
		return cache.PullProjectApi(retry-1, err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	utils.CookiesAddNoRepetition(&cache.Cookies, req.Cookies())
	return string(body), nil
}

// 拉取项目的课堂内容，比如必修或者非必修
func (cache *TtcdwUserCache) PullClassRoomApi(courseProjectId string, classId string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}

	urlStr := "https://www.ttcdw.cn/m/open/app/v2/member/project/" + courseProjectId + "/segment?classId=" + classId
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
		Transport: tr,
	}
	req, err := http.NewRequest(method, urlStr, nil)

	if err != nil {
		return "", err
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "www.ttcdw.cn")
	req.Header.Add("Connection", "keep-alive")

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	//fmt.Println(string(body))
	utils.CookiesAddNoRepetition(&cache.Cookies, req.Cookies())
	return string(body), nil
}
func (cache *TtcdwUserCache) PullCourseInfoApi(segmentId, courseId string, retry int, lastErr error) (string, error) {

	urlStr := "https://www.ttcdw.cn/m/open/app/v1/course/basic/" + courseId + "?segId=" + segmentId
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
		Transport: tr,
	}
	req, err := http.NewRequest(method, urlStr, nil)

	if err != nil {
		return "", nil
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "www.ttcdw.cn")
	req.Header.Add("Connection", "keep-alive")
	//设置Cookie
	for _, v := range cache.Cookies {
		req.AddCookie(v)
	}
	res, err := client.Do(req)
	if err != nil {
		return "", nil
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", nil
	}
	utils.CookiesAddNoRepetition(&cache.Cookies, req.Cookies())
	return string(body), nil
}

// 拉取对应项目的课程
func (cache *TtcdwUserCache) PullCourseApi(segmentId, itemId string, retry int, lastErr error) (string, error) {

	urlStr := "https://www.ttcdw.cn/m/open/app/v1/items/bxk/course/list?types=&segmentId=" + segmentId + "&itemId=" + itemId + "&moduleId=&pageNum=1&pageSize=100"
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
		Transport: tr,
	}
	req, err := http.NewRequest(method, urlStr, nil)

	if err != nil {
		return "", err
	}
	//设置Cookie
	for _, v := range cache.Cookies {
		req.AddCookie(v)
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "www.ttcdw.cn")
	req.Header.Add("Connection", "keep-alive")

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	utils.CookiesAddNoRepetition(&cache.Cookies, req.Cookies())
	return string(body), nil
}

// 拉取项目对应课程章节列表
func (cache *TtcdwUserCache) PullChapterListHtmlApi(cid string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	urlStr := "https://service.icourses.cn/hep-company/sword/company/shareChapter?cid=" + cid + "&shield="
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
		Transport: tr,
	}
	req, err := http.NewRequest(method, urlStr, nil)

	if err != nil {
		return "", err
	}
	//设置Cookie
	for _, v := range cache.Cookies {
		req.AddCookie(v)
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "service.icourses.cn")
	req.Header.Add("Connection", "keep-alive")

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	utils.CookiesAddNoRepetition(&cache.Cookies, req.Cookies())
	return string(body), nil
}

// 获取章节secId对应的子章节内容json
func (cache *TtcdwUserCache) PullGetResApi(sectionId string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	urlStr := "https://service.icourses.cn/hep-company//sword/company/getRess"
	method := "POST"

	payload := strings.NewReader("sectionId=" + sectionId)

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
		Transport: tr,
	}
	req, err := http.NewRequest(method, urlStr, payload)

	if err != nil {
		return "", err
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "service.icourses.cn")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	//设置Cookie
	for _, v := range cache.Cookies {
		req.AddCookie(v)
	}

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	utils.CookiesAddNoRepetition(&cache.Cookies, req.Cookies())
	return string(body), nil
}

// 拉取视频列表
func (cache *TtcdwUserCache) PullVideoListApi(courseId, itemId, segId, projectId, orgId string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	urlStr := "https://www.ttcdw.cn/p/course/services/course/public/course/lesson/" + courseId + "?ddtab=true&itemId=" + itemId + "&segId=" + segId + "&projectId=" + projectId + "&orgId=" + orgId + "&orgId=" + orgId + "&type=1&courseType=1&courseId=" + courseId + "&id=" + courseId + "&isContent=false&sourceType=1&_=" + fmt.Sprintf("%d", time.Now().UnixMilli())
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
		Transport: tr,
	}
	req, err := http.NewRequest(method, urlStr, nil)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Add("accept-language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6")
	req.Header.Add("cache-control", "no-cache")
	req.Header.Add("pragma", "no-cache")
	req.Header.Add("priority", "u=1, i")
	//req.Header.Add("referer", "https://www.ttcdw.cn/p/course/videorevision/v_895438431542083584?ddtab=true&itemId=1033511805027008512&segId=1033511477548335104&projectId=1033502195012517888&orgId=171864496496529408&type=1&courseType=1")
	req.Header.Add("sec-ch-ua", "\"Microsoft Edge\";v=\"141\", \"Not?A_Brand\";v=\"8\", \"Chromium\";v=\"141\"")
	req.Header.Add("sec-ch-ua-mobile", "?0")
	req.Header.Add("sec-ch-ua-platform", "\"Windows\"")
	req.Header.Add("sec-fetch-dest", "empty")
	req.Header.Add("sec-fetch-mode", "cors")
	req.Header.Add("sec-fetch-site", "same-origin")
	req.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Safari/537.36 Edg/141.0.0.0")
	req.Header.Add("x-requested-with", "XMLHttpRequest")
	req.Header.Add("content-type", "application/x-www-form-urlencoded")
	req.Header.Add("Host", "www.ttcdw.cn")
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

// 提交学时接口
func (cache *TtcdwUserCache) StudyTimeSubmitApi(orgId, courseId, itemId, videoId string, playProgress int, segId string, isFinish bool, typeNum, tjzj, clockInDot, sourceId, clockInRule, eventType string, retry int, lastErr error) (string, error) {

	urlStr := "https://www.ttcdw.cn/p/course/services/member/study/progress?orgId=" + orgId
	method := "POST"

	payload := strings.NewReader("courseId=" + courseId + "&itemId=" + itemId + "&videoId=" + videoId + "&playProgress=" + fmt.Sprintf("%d", playProgress) + "&segId=" + segId + "&isFinish=" + fmt.Sprintf("%t", isFinish) + "&type=1&tjzj=1&clockInDot=599&sourceId=1033502195012517888&clockInRule=0&timeLimit=-1&eventType=" + eventType)

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
		Transport: tr,
	}
	req, err := http.NewRequest(method, urlStr, payload)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Add("accept-language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6")
	req.Header.Add("cache-control", "no-cache")
	req.Header.Add("encryptionvalue", "589318420329488384")
	req.Header.Add("isencryption", "true")
	req.Header.Add("origin", "https://www.ttcdw.cn")
	req.Header.Add("pragma", "no-cache")
	req.Header.Add("priority", "u=0, i")
	req.Header.Add("referer", "https://www.ttcdw.cn/p/course/videorevision/v_589318420329488384?ddtab=true&itemId=1033512033289420800&segId=1033511477548335104&projectId=1033502195012517888&orgId=171864496496529408&type=1&courseType=2")
	req.Header.Add("sec-ch-ua", "\"Microsoft Edge\";v=\"141\", \"Not?A_Brand\";v=\"8\", \"Chromium\";v=\"141\"")
	req.Header.Add("sec-ch-ua-mobile", "?0")
	req.Header.Add("sec-ch-ua-platform", "\"Windows\"")
	req.Header.Add("sec-fetch-dest", "empty")
	req.Header.Add("sec-fetch-mode", "cors")
	req.Header.Add("sec-fetch-site", "same-origin")
	req.Header.Add("u-platformid", "13145854983311")
	req.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Safari/537.36 Edg/141.0.0.0")
	req.Header.Add("x-requested-with", "XMLHttpRequest")
	req.Header.Add("content-type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("Host", "www.ttcdw.cn")
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

// PKCS7 填充
func pkcs7Padding(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padText...)
}

// 加密函数
func encrypt(message string, key string) (string, error) {
	// 将key转化为8字节的切片（DES需要8字节的密钥）
	keyBytes := []byte(key)
	if len(keyBytes) < 8 {
		return "", errors.New("key must be at least 8 bytes long")
	}
	if len(keyBytes) > 8 {
		keyBytes = keyBytes[:8] // 截断为8字节
	}

	// 创建DES cipher块
	c, err := des.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}

	// 填充消息到8字节倍数，使用PKCS7填充方式
	messageBytes := []byte(message)
	messageBytes = pkcs7Padding(messageBytes, des.BlockSize)

	// 创建ECB模式下的加密器
	cipherText := make([]byte, len(messageBytes))
	for i := 0; i < len(messageBytes); i += des.BlockSize {
		c.Encrypt(cipherText[i:i+des.BlockSize], messageBytes[i:i+des.BlockSize])
	}

	// 返回Base64编码的加密结果
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

// 数据分组函数
func group(str string, step int) []string {
	var result []string
	for i := 0; i < len(str); i += step {
		end := i + step
		if end > len(str) {
			end = len(str)
		}
		result = append(result, str[i:end])
	}
	return result
}

// 加密数据函数
func EncData(dataStr string) (string, error) {
	// 按照100字符分组
	arr := group(dataStr, 100)
	var rulArr []string

	// 对每个分组进行加密
	for _, item := range arr {
		// 使用给定的密钥加密
		encryptedValue, err := encrypt(item, "MK49ICOURSES1102")
		if err != nil {
			return "", err
		}
		rulArr = append(rulArr, encryptedValue)
	}

	// 返回JSON字符串
	return fmt.Sprintf("[%s]", strings.Join(rulArr, ",")), nil
}
