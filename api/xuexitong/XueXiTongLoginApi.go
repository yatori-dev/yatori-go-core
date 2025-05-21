package xuexitong

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
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
		return "", err
	}
	values := resp.Header.Values("Set-Cookie")
	for _, v := range values {
		cache.cookie += strings.ReplaceAll(strings.ReplaceAll(v, "HttpOnly", ""), "Path=/", "")
		//if strings.Contains(v, "UUID=") {
		//
		//}
	}

	cache.JsonContent = jsonContent
	return string(body), nil
}

// VerificationCodeApi 获取验证码和SESSION验证码,并返回文件路径和SESSION字符串
func (cache *XueXiTUserCache) VerificationCodeApi() (string, string) {
	//TODO 待完成
	return "", ""
}
