package xuexitong

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/yatori-dev/yatori-go-core/utils"
)

var randChar []string = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f", "A", "B", "C", "D", "E", "F"}

// 获取学习通验证码
func (cache *XueXiTUserCache) XueXiTVerificationCodeApi(retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}

	urlStr := "https://mooc1-api.chaoxing.com/processVerifyPng.ac?t=" + fmt.Sprintf("%d", time.Now().UnixMilli())
	method := "GET"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
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
	}
	req, err := http.NewRequest(method, urlStr, nil)

	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}
	req.Header.Add("User-Agent", GetUA("mobile"))
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "mooc1-api.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer res.Body.Close()

	//body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	//utils.SaveImageAsJPEG(body,"")
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
		//return "", ""
		return "", err
	}

	_, err = io.Copy(file, res.Body)
	if err != nil {
		res.Body.Close() //立即释放
		log.Println(err)
		//return "", ""
		return "", err
	}

	file.Close()
	if utils.IsBadImg(filepath) {
		res.Body.Close()           //立即释放
		utils.DeleteFile(filepath) //删除坏的文件
		//return ""
		return cache.XueXiTVerificationCodeApi(retry-1, err)
	}
	//fmt.Println(string(body))
	return filepath, nil
}

// 提交学习通验证码
func (cache *XueXiTUserCache) XueXiTPassVerificationCode(code string, retry int, lastErr error) (bool, error) {
	if retry < 0 {
		return false, lastErr
	}
	urlStr := "https://mooc1-api.chaoxing.com/html/processVerify.ac"
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("app", "0")
	_ = writer.WriteField("ucode", code)
	err := writer.Close()
	if err != nil {
		fmt.Println(err)
		return false, err
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
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
		// 禁止自动重定向
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// 返回 http.ErrUseLastResponse 表示不要跟随重定向
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequest(method, urlStr, payload)

	if err != nil {
		fmt.Println(err)
		return false, err
	}
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}
	req.Header.Add("User-Agent", GetUA("mobile"))
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "mooc1-api.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Type", "multipart/form-data; boundary=--------------------------114712911338779046453834")

	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return false, err
	}
	defer res.Body.Close()

	//body, err := ioutil.ReadAll(res.Body)
	//if err != nil {
	//	fmt.Println(err)
	//	return false, err
	//}
	//fmt.Println(string(body))
	if res.StatusCode == 302 {
		return true, nil
	}
	return false, nil
}
