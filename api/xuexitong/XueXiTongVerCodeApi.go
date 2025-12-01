package xuexitong

import (
	"bytes"
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
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

// XueXiTSliderVerificationCodeApi 获取学习通滑块验证码相关信息，返回信息cx_captcha_function({"t":1764584640340,"captchaId":"Ew0z9skxsLzVKQjmeObQiRVLxkxbPkRF"})
func (cache *XueXiTUserCache) XueXiTSliderVerificationCodeApi(captchaId string, retry int, lastErr error) (string, error) {

	urlStr := "https://captcha.chaoxing.com/captcha/get/conf?callback=cx_captcha_function&captchaId=" + captchaId + "&_=" + fmt.Sprintf("%d", time.Now().UnixMilli())
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, urlStr, nil)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("User-Agent", GetUA("mobile"))
	req.Header.Add("Accept-Language", "zh-CN,en-US;q=0.9")
	req.Header.Add("X-Requested-With", "com.chaoxing.mobile")
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "captcha.chaoxing.com")
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

// 拉取验证码图片等信息
func (cache *XueXiTUserCache) XueXiTSliderVerificationImgApi(captchaId, serverTime, referer string, retry int, lastErr error) (string, error) {
	// 计算 MD5
	captchaKeyHash := md5.Sum([]byte(serverTime + uuid.New().String()))
	captchaKey := hex.EncodeToString(captchaKeyHash[:])

	//ivHash := md5.Sum([]byte(fmt.Sprintf("%s%s%d%s", captchaId, "slide", time.Now().UnixMilli(), uuid.New().String())))
	//iv := hex.EncodeToString(ivHash[:])

	// 计算token
	sum := md5.Sum([]byte(fmt.Sprintf("%s%s%s%s", serverTime, captchaId, "slide", captchaKey)))
	md5hex := hex.EncodeToString(sum[:])
	serverTimeInt, _ := strconv.ParseInt(serverTime, 10, 64)
	token := fmt.Sprintf("%s:%d", md5hex, serverTimeInt+300000)
	version := "1.1.20"
	urlStr := "https://captcha.chaoxing.com/captcha/get/verification/image?callback=cx_captcha_function&captchaId=" + captchaId + "&type=slide&version=" + version + "&captchaKey=" + captchaKey + "&token=" + token + "&referer=" + url.QueryEscape(referer)
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, urlStr, nil)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("User-Agent", GetUA("mobile"))
	req.Header.Add("Accept-Language", "zh-CN,en-US;q=0.9")
	req.Header.Add("X-Requested-With", "com.chaoxing.mobile")
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "captcha.chaoxing.com")
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

// 请求并获取图片
func (cache *XueXiTUserCache) PullSliderImgApi(imgUrl string) (image.Image, error) {
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, imgUrl, nil)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	req.Header.Add("User-Agent", GetUA("mobile"))
	req.Header.Add("Accept-Language", "zh-CN,en-US;q=0.9")
	req.Header.Add("X-Requested-With", "com.chaoxing.mobile")
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "captcha.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP 状态码异常: %d", resp.StatusCode)
	}

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("图片解码失败: %v", err)
	}

	return img, nil
}
