package icve

import (
	"encoding/json"
	"fmt"
	"image"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/utils"
)

type CapUnionData struct {
	State       int    `json:"state"`
	Ticket      string `json:"ticket"`
	Capclass    string `json:"capclass"`
	Subcapclass string `json:"subcapclass"`
	Src1        string `json:"src_1"`
	Src2        string `json:"src_2"`
	Src3        string `json:"src_3"`
	Sess        string `json:"sess"`
	Randstr     string `json:"randstr"`
	Sid         string `json:"sid"`
	LogJs       string `json:"log_js"`
	Data        struct {
		CommCaptchaCfg struct {
			TdcPath     string `json:"tdc_path"`
			FeedbackUrl string `json:"feedback_url"`
			PowCfg      struct {
				Prefix string `json:"prefix"`
				Md5    string `json:"md5"`
			} `json:"pow_cfg"`
		} `json:"comm_captcha_cfg"`
		DynShowInfo struct {
			Lang        string `json:"lang"`
			Instruction string `json:"instruction"`
			BgElemCfg   struct {
				Size2D   []int `json:"size_2d"`
				ClickCfg struct {
					MarkStyle string   `json:"mark_style"`
					DataType  []string `json:"data_type"`
				} `json:"click_cfg"`
				ImgUrl string `json:"img_url"`
			} `json:"bg_elem_cfg"`
			VerifyTriggerCfg struct {
				VerifyIcon bool `json:"verify_icon"`
			} `json:"verify_trigger_cfg"`
			ColorScheme string `json:"color_scheme"`
		} `json:"dyn_show_info"`
	} `json:"data"`
	Uip string `json:"uip"`
}

// 拉取Aid数据
func (cache *IcveUserCache) PullAidApi() string {

	url := "https://sso.icve.com.cn/prod-api/captcha/encrypt"
	method := "POST"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return ""
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36 Edg/140.0.0.0")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "sso.icve.com.cn")
	req.Header.Add("Connection", "keep-alive")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	//fmt.Println(string(body))
	utils.CookiesAddNoRepetition(&cache.Cookies, res.Cookies())
	return gojsonq.New().JSONString(string((body))).Find("data").(string)
}

// 拉取验证码数据
func (cache *IcveUserCache) PullVerDataApi() (CapUnionData, error) {

	urlStr := "https://turing.captcha.qcloud.com/cap_union_prehandle?aid=196632980&protocol=https&accver=1&showtype=popup&ua=TW96aWxsYS81LjAgKFdpbmRvd3MgTlQgMTAuMDsgV2luNjQ7IHg2NCkgQXBwbGVXZWJLaXQvNTM3LjM2IChLSFRNTCwgbGlrZSBHZWNrbykgQ2hyb21lLzE0MC4wLjAuMCBTYWZhcmkvNTM3LjM2IEVkZy8xNDAuMC4wLjA%253D&noheader=1&fb=1&aged=0&enableAged=0&enableDarkMode=0&grayscale=1&clientype=2" + "&aidEncrypted=" + cache.PullAidApi() + "&cap_cd=&uid=&lang=zh-cn&entry_url=https%253A%252F%252Fsso.icve.com.cn%252Fsso%252Fauth&elder_captcha=0&js=%252FtgJCap.977ef8c3.js&login_appid=&wb=1&subsid=3&callback=_aq_55600&sess="
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, urlStr, nil)

	if err != nil {
		fmt.Println(err)
		return CapUnionData{}, err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36 Edg/140.0.0.0")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "turing.captcha.qcloud.com")
	req.Header.Add("Connection", "keep-alive")
	for _, cookie := range cache.Cookies {
		req.AddCookie(cookie)
	}

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return CapUnionData{}, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return CapUnionData{}, err
	}
	//fmt.Println(string(body))
	compile, err := regexp.Compile(`\(([\w\W]*)\)`)
	if err != nil {
		fmt.Println(err)
	}
	submatch := compile.FindStringSubmatch(string(body))
	if len(submatch) <= 0 {
		return CapUnionData{}, err
	}
	capUnionData := CapUnionData{}
	err = json.Unmarshal([]byte(submatch[1]), &capUnionData)
	if err != nil {
		fmt.Println(err)
		return CapUnionData{}, err
	}
	utils.CookiesAddNoRepetition(&cache.Cookies, res.Cookies())
	return capUnionData, nil
}

// 拉取验证码图片
func PullCapImgApi(data CapUnionData) (image.Image, error) {
	url := "https://turing.captcha.qcloud.com" + data.Data.DynShowInfo.BgElemCfg.ImgUrl
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36 Edg/140.0.0.0")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "turing.captcha.qcloud.com")
	req.Header.Add("Connection", "keep-alive")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer res.Body.Close()

	// 检查响应状态码
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("请求失败，状态码: %d", res.StatusCode)
	}

	// 解码图片
	img, _, err := image.Decode(res.Body)
	if err != nil {
		return nil, fmt.Errorf("图片解码失败: %w", err)
	}

	return img, nil

}

// 拉取验证码VM
func (cache *IcveUserCache) PullVMApi(data CapUnionData, path string) (string, error) {
	urlStr := "https://turing.captcha.qcloud.com" + data.Data.CommCaptchaCfg.TdcPath
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, urlStr, nil)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36 Edg/140.0.0.0")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "turing.captcha.qcloud.com")
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
	utils.SaveTextToFile(path, string(body), false, 0644)
	utils.CookiesAddNoRepetition(&cache.Cookies, res.Cookies())
	return "", nil
}

func (cache *IcveUserCache) SubmitVerApi(data CapUnionData, collect string, eks string, posData string) (string, error) {

	urlStr := "https://turing.captcha.qcloud.com/cap_union_new_verify"
	method := "POST"
	powAnswer, powCalcTime := powSolve(data, 30)

	payload := strings.NewReader("collect=" + url.QueryEscape(collect) + "&tlg=" + fmt.Sprintf("%d", len(collect)) + "&eks=" + url.QueryEscape(eks) + "&sess=" + url.QueryEscape(data.Sess) + "&ans=" + url.QueryEscape(posData) + "&pow_answer=" + url.QueryEscape(powAnswer) + "&pow_calc_time=" + powCalcTime)

	client := &http.Client{}
	req, err := http.NewRequest(method, urlStr, payload)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36 Edg/140.0.0.0")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "turing.captcha.qcloud.com")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
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
	utils.CookiesAddNoRepetition(&cache.Cookies, res.Cookies())
	//fmt.Println(string(body))
	return string(body), nil
}

// 拉取powSolve
func powSolve(data CapUnionData, timeout int64) (string, string) {
	cfgPrefix := data.Data.CommCaptchaCfg.PowCfg.Prefix //获取cfg值
	cfgMd5 := data.Data.CommCaptchaCfg.PowCfg.Md5       //获取目标MD5值
	cmd := exec.Command("./assets/tencentPowSolve.exe", cfgPrefix, cfgMd5, strconv.Itoa(int(timeout)))
	output, err := cmd.Output()
	if err != nil {
		fmt.Println(err)
	}
	//output1 := ""
	result := []string{}
	err1 := json.Unmarshal(output, &result)
	if err1 != nil {
		fmt.Println(err)
	}
	return result[0], result[1]
}

// 拉取Collect和eks参数
func GetCollectAndEKS(fileName string) (string, string) {
	cmd := exec.Command(`assets\tencentCollect.exe`, fileName)
	output, err := cmd.Output()
	if err != nil {
		fmt.Println(err)
	}
	cmd1 := exec.Command(`assets\tencentEks.exe`, fileName)
	output1, err1 := cmd1.Output()
	if err1 != nil {
		fmt.Println(err1)
	}
	fmt.Println(string(output))
	fmt.Println(string(output1))
	//output1 := ""
	return strings.ReplaceAll(string(output), "\n", ""), strings.ReplaceAll(string(output1), "\n", "")
}
