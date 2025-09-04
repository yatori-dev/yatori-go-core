package mooc

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/utils"
)

type MOOCUserCache struct {
	Account   string //账号
	Password  string //用户密码
	TK        string //通过GT获取的参数
	Sid       string
	X         string
	T         int
	Puzzle    string
	Mod       string
	MinTime   int64
	MaxTime   int64
	IpProxySW bool   //是否开启IP代理
	ProxyIP   string //代理IP
	cookies   []*http.Cookie
}

// 用于初始化Cookie参数
func (cache *MOOCUserCache) InitCookiesApi() {

	url := "https://www.icourse163.org/member/login.htm"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "www.icourse163.org")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Referer", "https://www.icourse163.org/member/login.htm")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()
	utils.CookiesAddNoRepetition(&cache.cookies, res.Cookies())

	encStr2 := MOOCEncMS4(BuildDLInitParams("imooc", "cjJVGQM", "www.icourse163.org", 1, "https://www.icourse163.org/member/login.htm", BuildRtId()))
	url2 := "https://reg.icourse163.org/dl/zj/yd/ini"
	method2 := "POST"

	payload2 := strings.NewReader(`{"encParams":"` + encStr2 + `"}`)

	client2 := &http.Client{}
	req2, err2 := http.NewRequest(method2, url2, payload2)
	for _, cookie := range cache.cookies {
		req2.AddCookie(cookie)
	}
	if err2 != nil {
		fmt.Println(err2)
		return
	}
	req2.Header.Add("Origin", "https://reg.icourse163.org")
	req2.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36")
	req2.Header.Add("Content-Type", "application/json")
	req2.Header.Add("Accept", "*/*")
	req2.Header.Add("Host", "reg.icourse163.org")
	req2.Header.Add("Connection", "keep-alive")

	res2, err2 := client2.Do(req2)
	if err2 != nil {
		fmt.Println(err2)
		return
	}
	defer res2.Body.Close()

	body2, err2 := ioutil.ReadAll(res2.Body)
	if err2 != nil {
		fmt.Println(err2)
		return
	}
	fmt.Println(string(body2))
	utils.CookiesAddNoRepetition(&cache.cookies, res2.Cookies())
}

// powGetP 接口
func (cache *MOOCUserCache) PowGetPApi() {

	url := "https://reg.icourse163.org/dl/zj/yd/powGetP"
	method := "POST"
	encStr := MOOCEncMS4(BuildPowGetPParams("imooc", "cjJVGQM", "18973485974", "5722fb36-7665-4510-8281-c202f414978c", 1, "https://www.icourse163.org/member/login.htm?returnUrl=aHR0cHM6Ly93d3cuaWNvdXJzZTE2My5vcmcvaW5kZXguaHRt#/webLoginIndex", BuildRtId()))
	payload := strings.NewReader(`{"encParams":"` + encStr + `"}`)
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}

	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "reg.icourse163.org")
	req.Header.Add("Connection", "keep-alive")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	if gojsonq.New().JSONString(string(body)).Find("ret").(string) == "201" {
		cache.MaxTime = int64(gojsonq.New().JSONString(string(body)).Find("pVInfo.maxTime").(float64))
		cache.MinTime = int64(gojsonq.New().JSONString(string(body)).Find("pVInfo.minTime").(float64))
		cache.Sid = gojsonq.New().JSONString(string(body)).Find("pVInfo.sid").(string)
		cache.Puzzle = gojsonq.New().JSONString(string(body)).Find("pVInfo.args.puzzle").(string)
		cache.X = gojsonq.New().JSONString(string(body)).Find("pVInfo.args.x").(string)
		cache.T = int(gojsonq.New().JSONString(string(body)).Find("pVInfo.args.t").(float64))
		cache.Mod = gojsonq.New().JSONString(string(body)).Find("pVInfo.args.mod").(string)
	}
	fmt.Println(string(body))
}

func (cache *MOOCUserCache) GtApi() {

	url := "https://reg.icourse163.org/dl/zj/yd/gt"
	method := "POST"
	encStr := MOOCEncMS4(BuildGTParams(cache.Account, 1, "imooc", "cjJVGQM", "https://www.icourse163.org/member/login.htm", BuildRtId()))
	payload := strings.NewReader(`{"encParams":"` + encStr + `"}`)
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "reg.icourse163.org")
	req.Header.Add("Connection", "keep-alive")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
	if gojsonq.New().JSONString(string(body)).Find("ret").(string) == "201" {
		cache.TK = gojsonq.New().JSONString(string(body)).Find("tk").(string)
	}
	utils.CookiesAddNoRepetition(&cache.cookies, res.Cookies())
}

func (cache *MOOCUserCache) LoginApi() {

	url := "https://reg.icourse163.org/dl/zj/yd/pwd/l"
	method := "POST"

	runTimes, spendTime, T, X, sign := VdfAsync(Data{
		NeedCheck: true,
		Sid:       cache.Sid,
		HashFunc:  "VDF_FUNCTION",
		MaxTime:   cache.MaxTime,
		MinTime:   cache.MinTime,
		Args: Args{
			Mod:    cache.Mod,
			T:      cache.T,
			Puzzle: cache.Puzzle,
			X:      cache.X,
		},
	})
	buildParams := BuildLParams(1, 10, cache.Account, MOOCRSA(cache.Password), "imooc", "cjJVGQM", cache.TK, "", cache.Puzzle, int(spendTime), runTimes, cache.Sid, X, T, int(sign), 1, "https://www.icourse163.org/member/login.htm", BuildRtId())
	encStr := MOOCEncMS4(buildParams)
	payload := strings.NewReader(`{"encParams":"` + encStr + `"}`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}

	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "reg.icourse163.org")
	req.Header.Add("Connection", "keep-alive")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
	utils.CookiesAddNoRepetition(&cache.cookies, res.Cookies())
}
