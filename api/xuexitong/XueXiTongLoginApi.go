package xuexitong

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// 注意Api类文件主需要写最原始的接口请求和最后的json的string形式返回，不需要用结构体序列化。
// 序列化和具体的功能实现请移步到Action代码文件中
const (
	ApiLoginWeb = "https://passport2.chaoxing.com/fanyalogin"

	ApiPullCourses = "https://mooc1-api.chaoxing.com/mycourse/backclazzdata"

	// ApiChapterPoint 接口-课程章节任务点状态
	ApiChapterPoint = "https://mooc1-api.chaoxing.com/job/myjobsnodesmap"
	ApiChapterCards = "https://mooc1-api.chaoxing.com/gas/knowledge"
	ApiPullChapter  = "https://mooc1-api.chaoxing.com/gas/clazz"

	// PageMobileChapterCard SSR页面-客户端章节任务卡片
	PageMobileChapterCard = "https://mooc1-api.chaoxing.com/knowledge/cards"

	// APIChapterCardResource 接口-课程章节卡片资源
	APIChapterCardResource = "https://mooc1-api.chaoxing.com/ananas/status"
	// APIVideoPlayReport 接口-视频播放上报
	APIVideoPlayReport  = "https://mooc1.chaoxing.com/mooc-ans/multimedia/log/a"
	APIVideoPlayReport2 = "https://mooc1-api.chaoxing.com/multimedia/log/a" // cxkitty的

	// ApiWorkCommit 接口-单元作业答题提交
	ApiWorkCommit = "https://mooc1-api.chaoxing.com/work/addStudentWorkNew"
	// ApiWorkCommitNew 接口-新的作业提交答案接口
	ApiWorkCommitNew = "https://mooc1.chaoxing.com/mooc-ans/work/addStudentWorkNew"

	// 接口-课程文档阅读上报
	ApiDocumentReadingReport = "https://mooc1.chaoxing.com/ananas/job/document"

	// PageMobileWork SSR页面-客户端单元测验答题页
	PageMobileWork  = "https://mooc1-api.chaoxing.com/android/mworkspecial"           // 这是个cxkitty中的
	PageMobileWorkY = "https://mooc1-api.chaoxing.com/mooc-ans/work/phone/doHomeWork" // 这个是自己爬的

	KEY = "u2oh6Vu^HWe4_AES" // 注意 Go 语言中字符串默认就是 UTF-8 编码
)

type XueXiTUserCache struct {
	Name     string //用户使用Phone
	Password string //用户密码

	UserID      string // 用户ID
	JsonContent map[string]interface{}
	cookies     []*http.Cookie
	cookie      string //验证码用的session
}

func (cache *XueXiTUserCache) GetCookie() string {
	return cache.cookie
}

// pad 确保数据长度是块大小的整数倍，以便符合块加密算法的要求
func pad(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padText...)
}

// LoginApi 登录Api
func (cache *XueXiTUserCache) LoginApi() (string, error) {
	key := []byte(KEY)
	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println("Error creating cipher:", err)
		return "", err
	}

	// 加密电话号码
	phonePadded := pad([]byte(cache.Name), block.BlockSize())
	phoneCipherText := make([]byte, len(phonePadded))
	mode := cipher.NewCBCEncrypter(block, key)
	mode.CryptBlocks(phoneCipherText, phonePadded)
	phoneEncrypted := base64.StdEncoding.EncodeToString(phoneCipherText)

	// 加密密码
	passwdPadded := pad([]byte(cache.Password), block.BlockSize())
	passwdCipherText := make([]byte, len(passwdPadded))
	mode = cipher.NewCBCEncrypter(block, key)
	mode.CryptBlocks(passwdCipherText, passwdPadded)
	passwdEncrypted := base64.StdEncoding.EncodeToString(passwdCipherText)

	// 发送请求
	resp, err := http.PostForm(ApiLoginWeb, url.Values{
		"fid":               {"-1"},
		"uname":             {phoneEncrypted},
		"password":          {passwdEncrypted},
		"t":                 {"true"},
		"forbidotherlogin":  {"0"},
		"validate":          {""},
		"doubleFactorLogin": {"0"},
		"independentId":     {"0"},
		"independentNameId": {"0"},
	})
	if err != nil {
		fmt.Println("Error sending request:", err)
		return "", err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var jsonContent map[string]interface{}
	err = json.Unmarshal(body, &jsonContent)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return "", err
	}

	if status, ok := jsonContent["status"].(bool); !ok || !status {
		return "", errors.New(string(body))
	}
	values := resp.Header.Values("Set-Cookie")
	for _, v := range values {
		cache.cookie += strings.ReplaceAll(strings.ReplaceAll(v, "HttpOnly", ""), "Path=/", "")
		//if strings.Contains(v, "UUID=") {
		//
		//}
	}
	cache.cookies = resp.Cookies() //赋值cookie

	cache.JsonContent = jsonContent
	return string(body), nil
}

// VerificationCodeApi 获取验证码和SESSION验证码,并返回文件路径和SESSION字符串
func (cache *XueXiTUserCache) VerificationCodeApi() (string, string) {
	//TODO 待完成
	return "", ""
}

// 学习通不知道用来干嘛的接口，但是每隔一段时间都会发送一次
func (cache *XueXiTUserCache) MonitorApi() (string, error) {
	fid := ""
	for _, cookie := range cache.cookies {
		if cookie.Name == "fid" {
			fid = cookie.Value
		}
	}
	url := fmt.Sprintf("https://detect.chaoxing.com/api/monitor?version=%s&refer=%s&from=&fid=%s&jsoncallback=jsonp%s&t=%d", "1748603971011", "http%%253A%252F%252Fi.mooc.chaoxing.com", fid, generate17DigitNumber(), time.Now().UnixMilli())
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36 Edg/136.0.0.0")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "detect.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")
	//req.Header.Add("Cookie", "JSESSIONID=965F54A70062952629CC029FA92F2AA1")
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}

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
	fmt.Println(string(body))
	return string(body), nil
}

// 随机生成17位置数字，带前导0
func generate17DigitNumber() string {
	rand.Seed(time.Now().UnixNano())

	result := ""
	for i := 0; i < 17; i++ {
		digit := rand.Intn(10) // 生成0-9的随机数字
		result += fmt.Sprintf("%d", digit)
	}
	return result
}
