package ttcdw

import (
	"bytes"
	"crypto/des"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/yatori-dev/yatori-go-core/utils"
)

// 拉取所有项目
func (cache *TtcdwUserCache) PullProjectApi(retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	url := "https://www.ttcdw.cn/m/open/app/v1/memProject/list?state=1&pageNum=1&pageSize=100"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

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

	url := "https://www.ttcdw.cn/m/open/app/v2/member/project/" + courseProjectId + "/segment?classId=" + classId
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return "", err
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "www.ttcdw.cn")
	req.Header.Add("Connection", "keep-alive")
	//req.Header.Add("Cookie", "HWWAFSESID=c92a3799bef8ba22d2; HWWAFSESTIME=1734968345848; passport=https://www.ttcdw.cn/p/passport; u-lastLoginTime=1734968374078; u-activeState=1; u-mobileState=0; u-mobile=18837922277; u-preLoginTime=1734964883966; u-token=eyJhbGciOiJIUzI1NiJ9.eyJqdGkiOiI0Mjk0ZmUzOC0zMTBiLTRlMmQtOWQwMS0xN2EyYzZjNjA2ZmYiLCJpYXQiOjE3MzQ5NjgzNzQsInN1YiI6IjUyNzI4Mzc0NTk0NTcwMjQwMCIsImlzcyI6Imd1b3JlbnQiLCJhdHRlc3RTdGF0ZSI6MCwic3JjIjoid2ViIiwiYWN0aXZlU3RhdGUiOjEsIm1vYmlsZSI6IjE4ODM3OTIyMjc3IiwicGxhdGZvcm1JZCI6IjEzMTQ1ODU0OTgzMzExIiwiYWNjb3VudCI6IjE4ODM3OTIyMjc3IiwiZXhwIjoxNzM1MDA0Mzc0fQ.sln7IZEkCDgZNqVSOXHvZXj-EklcTkEoRHMgVinPFh4; u-token-legacy=eyJhbGciOiJIUzI1NiJ9.eyJqdGkiOiI0Mjk0ZmUzOC0zMTBiLTRlMmQtOWQwMS0xN2EyYzZjNjA2ZmYiLCJpYXQiOjE3MzQ5NjgzNzQsInN1YiI6IjUyNzI4Mzc0NTk0NTcwMjQwMCIsImlzcyI6Imd1b3JlbnQiLCJhdHRlc3RTdGF0ZSI6MCwic3JjIjoid2ViIiwiYWN0aXZlU3RhdGUiOjEsIm1vYmlsZSI6IjE4ODM3OTIyMjc3IiwicGxhdGZvcm1JZCI6IjEzMTQ1ODU0OTgzMzExIiwiYWNjb3VudCI6IjE4ODM3OTIyMjc3IiwiZXhwIjoxNzM1MDA0Mzc0fQ.sln7IZEkCDgZNqVSOXHvZXj-EklcTkEoRHMgVinPFh4; u-id=527283745945702400; u-account=18837922277; ufo-urn=MTg4Mzc5MjIyNzc=; ufo-un=5LqO5pif5qKF; ufo-id=527283745945702400; u-name=yx_user_tjAvr5JL")

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

	url := "https://www.ttcdw.cn/m/open/app/v1/course/basic/" + courseId + "?segId=" + segmentId
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return "", nil
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "www.ttcdw.cn")
	req.Header.Add("Connection", "keep-alive")
	//req.Header.Add("Cookie", "HWWAFSESID=c92a3799bef8ba22d2; HWWAFSESTIME=1734968345848; passport=https://www.ttcdw.cn/p/passport; u-lastLoginTime=1734968374078; u-activeState=1; u-mobileState=0; u-mobile=18837922277; u-preLoginTime=1734964883966; u-token=eyJhbGciOiJIUzI1NiJ9.eyJqdGkiOiI0Mjk0ZmUzOC0zMTBiLTRlMmQtOWQwMS0xN2EyYzZjNjA2ZmYiLCJpYXQiOjE3MzQ5NjgzNzQsInN1YiI6IjUyNzI4Mzc0NTk0NTcwMjQwMCIsImlzcyI6Imd1b3JlbnQiLCJhdHRlc3RTdGF0ZSI6MCwic3JjIjoid2ViIiwiYWN0aXZlU3RhdGUiOjEsIm1vYmlsZSI6IjE4ODM3OTIyMjc3IiwicGxhdGZvcm1JZCI6IjEzMTQ1ODU0OTgzMzExIiwiYWNjb3VudCI6IjE4ODM3OTIyMjc3IiwiZXhwIjoxNzM1MDA0Mzc0fQ.sln7IZEkCDgZNqVSOXHvZXj-EklcTkEoRHMgVinPFh4; u-token-legacy=eyJhbGciOiJIUzI1NiJ9.eyJqdGkiOiI0Mjk0ZmUzOC0zMTBiLTRlMmQtOWQwMS0xN2EyYzZjNjA2ZmYiLCJpYXQiOjE3MzQ5NjgzNzQsInN1YiI6IjUyNzI4Mzc0NTk0NTcwMjQwMCIsImlzcyI6Imd1b3JlbnQiLCJhdHRlc3RTdGF0ZSI6MCwic3JjIjoid2ViIiwiYWN0aXZlU3RhdGUiOjEsIm1vYmlsZSI6IjE4ODM3OTIyMjc3IiwicGxhdGZvcm1JZCI6IjEzMTQ1ODU0OTgzMzExIiwiYWNjb3VudCI6IjE4ODM3OTIyMjc3IiwiZXhwIjoxNzM1MDA0Mzc0fQ.sln7IZEkCDgZNqVSOXHvZXj-EklcTkEoRHMgVinPFh4; u-id=527283745945702400; u-account=18837922277; ufo-urn=MTg4Mzc5MjIyNzc=; ufo-un=5LqO5pif5qKF; ufo-id=527283745945702400; u-name=yx_user_tjAvr5JL")
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

	url := "https://www.ttcdw.cn/m/open/app/v1/items/bxk/course/list?types=&segmentId=" + segmentId + "&itemId=" + itemId + "&moduleId=&pageNum=1&pageSize=100"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

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
	url := "https://service.icourses.cn/hep-company/sword/company/shareChapter?cid=" + cid + "&shield="
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

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
	url := "https://service.icourses.cn/hep-company//sword/company/getRess"
	method := "POST"

	payload := strings.NewReader("sectionId=" + sectionId)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

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
