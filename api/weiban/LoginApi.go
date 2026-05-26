package weiban

import (
	"bytes"
	"crypto/aes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/yatori-dev/yatori-go-core/utils"
)

type WeiBanCache struct {
	School     string //学校
	TenantCode string //学校id
	VerifyTime string //验证码时间戳
	VerifyCode string //验证码
	Account    string //账号
	Password   string //密码
	IpProxySW  bool
	ProxyIP    string
	Cookies    []*http.Cookie
}

// PullTenantCodeApi 拉取学校列表
func (cache *WeiBanCache) PullTenantCodeApi() (string, error) {

	urlStr := "https://weiban.mycourse.cn/pharos/login/getTenantListWithLetter.do?timestamp=1758343815.004"
	method := "POST"

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
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "weiban.mycourse.cn")
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
	fmt.Println(string(body))
	return string(body), nil
}

// 拉取学校配置
func (cache *WeiBanCache) PullTenantConfigApi() (string, error) {

	urlStr := "https://weiban.mycourse.cn/pharos/tenantconfig/getSimpleConfig.do?timestamp=1758343997.34"
	method := "POST"

	payload := strings.NewReader("tenantCode=" + cache.TenantCode)

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
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "weiban.mycourse.cn")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	//req.Header.Add("Cookie", "SERVERID=960d3937b58d431003f75e175ffea128|1758344099|1758343911")

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

var randChar []string = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f", "A", "B", "C", "D", "E", "F"}

// 拉取验证码
func (cache *WeiBanCache) PullCapterApi(retry int, lastErr error) (string, error) {
	if lastErr != nil {
		return "", lastErr
	}
	randomTime := fmt.Sprintf("%d", time.Now().UnixMilli())
	cache.VerifyTime = randomTime
	urlStr := "https://weiban.mycourse.cn/pharos/login/randLetterImage.do?time=" + randomTime
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
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "weiban.mycourse.cn")
	req.Header.Add("Connection", "keep-alive")
	//req.Header.Add("Cookie", "SERVERID=960d3937b58d431003f75e175ffea128|1758344218|1758343911")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer res.Body.Close()

	//body, err := ioutil.ReadAll(res.Body)
	//if err != nil {
	//	fmt.Println(err)
	//	return "", err
	//}
	utils.CookiesAddNoRepetition(&cache.Cookies, res.Cookies())
	//fmt.Println(string(body))
	//下载验证码
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
		return "", err
	}

	_, err = io.Copy(file, res.Body)
	if err != nil {
		res.Body.Close() //立即释放
		log.Println(err)
		return "", err
	}

	file.Close()
	if utils.IsBadImg(filepath) {
		res.Body.Close()           //立即释放
		utils.DeleteFile(filepath) //删除坏的文件
		return cache.PullCapterApi(retry-1, nil)
	}
	return filepath, nil
}

// {"code":"0","data":{"token":"3f5872ab-a461-4f76-b04e-b777f065d6d0","userId":"f96f49f8-dde0-46e5-8c0d-7050a5af64d8","userName":"21712aa0b3324cdc9f96df95316ee23a","realName":"魏涛","userNameLabel":"考生号","uniqueValue":"25611030390001","isBind":"1","tenantCode":"710065","batchCode":"012","gender":1,"openid":"oeNCVuNPgOivWXnp83OSMkPMKXl8","unionid":"oQSZgv2oLyEsqmrvMBYYP5k3vlr0","switchGoods":1,"switchDanger":1,"switchNetCase":1,"preBanner":"https://h.mycourse.cn/pharosfile/resources/images/projectbanner/pre.png","normalBanner":"https://h.mycourse.cn/pharosfile/resources/images/projectbanner/normal.png","specialBanner":"https://h.mycourse.cn/pharosfile/resources/images/projectbanner/special.png","militaryBanner":"https://h.mycourse.cn/pharosfile/resources/images/projectbanner/military.png","contestIndexImage":"https://weibanstatic.mycourse.cn/pharos/resource/710065/image/contest/20220413/f03fbbd4-d378-429c-910e-1da389cca7d2.jpg","contestThemeImage":"https://weibanstatic.mycourse.cn/pharos/resource/710065/image/contest/20220413/fe59b829-2124-4969-9691-c23f432ea0f8.jpg","isLoginFromWechat":2,"tenantName":"西安文理学院","tenantType":1,"loginSide":1,"popForcedCompleted":2,"showGender":2,"showOrg":2,"orgLabel":"院系","nickName":"魏涛","imageUrl":"https://resource.mycourse.cn/mercury/resources/mercury/wb/images/portrait.jpg","defensePower":60,"knowledgePower":60,"safetyIndex":99},"detailCode":"0"}
func (cache *WeiBanCache) LoginApi() (string, error) {
	urlStr := "https://weiban.mycourse.cn/pharos/login/login.do"

	// 调用你之前写的 AES 加密函数
	//encryptData, err := Encrypt(string(jsonBytes))
	encryptData, err := Encrypt(`{"keyNumber":"` + cache.Account + `","password":"` + cache.Password + `","tenantCode":"` + cache.TenantCode + `","time":` + cache.VerifyTime + `,"verifyCode":"` + cache.VerifyCode + `"}`)
	if err != nil {
		return "", err
	}

	// 构造表单参数
	formData := []byte("data=" + encryptData)

	// 构造请求
	req, err := http.NewRequest("POST", urlStr+"?timestamp="+cache.VerifyTime, bytes.NewBuffer(formData))
	if err != nil {
		return "", err
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "weiban.mycourse.cn")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, cookie := range cache.Cookies {
		req.AddCookie(cookie)
	}

	// 发送请求
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
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// PKCS7 填充
func pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// AES ECB 加密
func encryptECB(data, key []byte) ([]byte, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, fmt.Errorf("invalid key size: %d", len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	bs := block.BlockSize()
	data = pkcs7Padding(data, bs)
	encrypted := make([]byte, len(data))

	// ECB 手动分块加密
	for start := 0; start < len(data); start += bs {
		block.Encrypt(encrypted[start:start+bs], data[start:start+bs])
	}
	return encrypted, nil
}

func Encrypt(data string) (string, error) {
	key, err := base64.URLEncoding.DecodeString("d2JzNTEyAAAAAAAAAAAAAA==")
	if err != nil {
		return "", err
	}

	encData, err := encryptECB([]byte(data), key)
	if err != nil {
		return "", err
	}

	// 返回base64 编码字符串
	return base64.URLEncoding.EncodeToString(encData), nil
}
