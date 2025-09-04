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

var CCC = BuildRtId()

func (cache *MOOCUserCache) InitCookies() {

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

	encStr2 := MOOCEncMS4(BuildDLInitParams("imooc", "cjJVGQM", "www.icourse163.org", 1, "https://www.icourse163.org/member/login.htm", CCC))
	//url2 := "https://reg.icourse163.org/dl/zj/mail/ini"
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
	//req2.Header.Add("Referer", "https://reg.icourse163.org/webzj/v1.0.1/pub/index_reg2_new.html?cd=%2F%2Fcmc.stu.126.net%2Fu%2Fcss%2Fcms%2F&cf=mooc_urs_login_css.css&MGID=1756928408006.7036&wdaId=UA1438236666413&pkid=cjJVGQM&product=imooc&cdnhostname=webzj.netstatic.net")
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
func (cache *MOOCUserCache) PowGetP() {

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
	//req.Header.Add("Cookie", "EDUWEBDEVICE=59b79143b0fa444a8587e0e737c89274; hb_MA-A976-948FFA05E931_source=www.bing.com; Hm_lvt_77dc9a9d49448cf5e629e5bebaa5500b=1756875024; HMACCOUNT=FB705416A914C520; utid=XIYFPxIZ3m62RPI0hsB2mE9XQWOODej4; NTES_WEB_FP=758f73b0e5ff302587d2f3ba4232550b; ntes_zc_cid=6479c3c6-b452-4a3b-a83c-4d9ae88e3076; ntes_zc_yd_cjJVGQM=55894076A71C69C5C056693A6972CF712A5006CAFD5FDEF009C18D28F5D930BCF8C2698C9D077AC1667D782C350143EB571E692D3B33FEB37197F7050AAB8345260CBEC3F462FE61D35E8BF21EB1FCA3; THE_LAST_LOGIN_MOBILE=18973485974; l_yd_sign=-020170531CBfYiOSV7DxLNBhBbvybk64fLypxQfvgTnxQNUtqo6xRc-S9gi_yDpiKGAIA1pccpB6zUg6H6ejkcIOeIaaAkRlpSyoahWlhom46Ol5GbXxw..; NTESSTUDYSI=f9fb20fcca1045c293ef1d605d0ff372; Hm_lpvt_77dc9a9d49448cf5e629e5bebaa5500b=1756907442; l_yd_s_imooccjJVGQM=55894076A71C69C5C056693A6972CF712A5006CAFD5FDEF009C18D28F5D930BCF8C2698C9D077AC1667D782C350143EB91849A0DA3ACE9921AE68B42133AC0D25AECFD0CA2978EBE98596FD5F149EE71044A0A0AD0631F2E6C626FA4613C81128B37FD43756EB47C1BEC521815303416; l_s_imooccjJVGQM=55894076A71C69C5C056693A6972CF712A5006CAFD5FDEF009C18D28F5D930BCF8C2698C9D077AC1667D782C350143EB79BD053BA8F36FCEBB3A00261CBF0560C781BAAE1399CFF34439A80483839D7A54D9B229A414BF434FDE56D9B5F2D2EA2CFE2A1FE48E77C66987715C71BDB668; S_INFO=1756919327|0|0&60##|18973485974; P_INFO=18973485974|1756919327|1|imooc|00&99|gux&1756916214&imooc#gux&450400#10#0#0|&0||18973485974; NTES_YD_SESS=Kxy71vWlJFySKi16l8QfgmKkSMuOoqjiVHfalLMGXfLRUGQvUZBPoCjMck6AmxKJ0Yx7.8XHw5gOZOdL619UUXltlse8SkujSkKr7dIhphS2bR2NETai6n2jBN7tG4FAU75R6jnF3GA0s4PMf72A19ZM10Ww7AB.03lNAJcom9TRFydVs2OtG85gCB3ynMceRzVzAgixnXdXGlp56NVXAxab9T6.YUAJNL.XfLQJsCwuj; l_yd_sign=-020170531-FtSFJtwyf7gqVJw6OqGIv0l5RFyq4-3WST5IEDfo32Pjda4ZHDWXvcBdPOqdJseoAyL7_BHtlsl2-zTjeBn80M36WRb9zJ7-7kLMa9cOf9g..; NTES_YD_PASSPORT=KOhd.7.mOASL29mLzE4Z2k2uzd9Jhn3bNlrWB1.I2ubRP_OdPs6CJlGQb1jwkL4MTmLcF0hKgfp_r0lXQPg6PXxWdCdBNWcAbMyRtV3WdcmTLC3YshGT.gw1.XZAH2.jx9dW2Vmv6I4AWAunqtV6hR0lbgdIrzcnEfmk1QdYM2oZwijnEgAgMHjc24AUzfXJ4QneYlJvjp6l8tVhm2HOdrpQj")
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

func (cache *MOOCUserCache) Gt() {

	url := "https://reg.icourse163.org/dl/zj/yd/gt"
	method := "POST"
	encStr := MOOCEncMS4(BuildGTParams(cache.Account, 1, "imooc", "cjJVGQM", "https://www.icourse163.org/member/login.htm", BuildRtId()))
	//encStr := MOOCEncMS4(BuildGTParams(cache.Account, 1, "imooc", "cjJVGQM", "https://www.icourse163.org/member/login.htm?returnUrl=aHR0cHM6Ly93d3cuaWNvdXJzZTE2My5vcmcvaW5kZXguaHRt#/webLoginIndex", "ZiujIta50WCCrsu30y3566YgVb1pUwHE"))
	//encStr := `004424aa55dee76a0fbc5c5bbf0de86e4014e594e10bf6fef5993dade6d63ae0479bc5cf0c1fa1663c19c2260cc2bc3809cd8275d3c9e705bb2f4b2dd5e7a085610282b6c723b9c395537ecc58e5b08f193de0921d53436ab4ba9aa2553576bfb6980b4a8b4d6b0bc891222e3486133291967f4559bffe5c056ed13373f02e72d59d18489b55783756a4762dc0f73778a73146c856177121ee1db5b70d2ea3a6c7b5deb54bbbe5f1ce2989f001f8ab763a29fb9afc422047670093d25b35bf0468e71ab98225e300142f88361fc875b9bd4da574021f1aabe08cd8aab7a2030beb2e0940bec5f347a5378c5f5fe097e1964448f2f08cb20f5513337f47f4cad872a124eae117974f96fa80b162539f2658e4c4bdeae6a18f5c7b5a5c7918385bad97d9fc108b257c453655f478801e105d2b9a9d0bc5b4228e1c567e8bab24e3e3324762247a972a438b9db93564f42bce95e6bcddf1f09d8cced12f50ca690886ba3fa754117985c61b9202b6f6cb428db691939f7ad4ef5bab458b12d8aeabf54fd9ac64b4e2e3b8966f86e1638dea9bf9fc98a62bf99165ab56e36272c6de9395b5e221d50a8fba4370474d7176835ddd1991e794db36f33984b0f2c0258455f5c4752273a2449a1142cb2d443a174adf8d26fd1daba43ae85263df72cb7962ee2c0ef36209968c37d407d43fe01627d20cb95ea19af37d32ac9d8e727801ceededebcac46c6e20a95c310397bd976727d24ca9c6a01758307e0a11dd38e057eeb6698860d8e8496920f3861d02295a77634538fca6582be25091b18cbb4d7562e6f94873518545a0f3b4e3fd9e3f6acb25a590ed4bbf1e55bd015564932c77b098367a0d5d92b290f1b4ff8b3011f997698b0af9fa3ec8c011d48e4e595cf20e0bf52640eec0915f31d58b491ea9d45404e323984194c5b797a11dea1825c528b68929a51659e641da6f52be8c0ecd33eb9a2de377c18f196609230335f3109d773b643b3d41d01ca2ed42d231fc9f6ad0305b07e8c59252fcaa318179cdb1adb72db3831a035686e9e00c8fae4548f1c62d525467fff8b12d36d16a560d04ab5757720aff88c611819fea38b735d18038b3f7114249fb98175394724bcc9331beaf4a412cbfca5d6a2c6ae9c93402d99b279ce74493c4ba6213027b5ce76f88b6ca6e4c249912db35b0fbc53411bdea14fe4bf0b3cb13221b192cf99615c9cc6f97fdd9289cdf936f086bf0ca47cdda76424cad86974a77a2e1044bff2cff48ff6702d6478c4608a8ec733eb96918095656532b87eb1591d3c9f503d70359b3befd5ad00a0db9ad7d89c14d95ba3642d7d32895389ecb9b11b6df4281daec4bda4db8fc6ef798d62c449f311645`
	payload := strings.NewReader(`{"encParams":"` + encStr + `"}`)
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	//req.Header.Add("Cookie", "EDUWEBDEVICE=59b79143b0fa444a8587e0e737c89274; hb_MA-A976-948FFA05E931_source=www.bing.com; Hm_lvt_77dc9a9d49448cf5e629e5bebaa5500b=1756875024; HMACCOUNT=FB705416A914C520; utid=XIYFPxIZ3m62RPI0hsB2mE9XQWOODej4; NTES_WEB_FP=758f73b0e5ff302587d2f3ba4232550b; ntes_zc_cid=6479c3c6-b452-4a3b-a83c-4d9ae88e3076; ntes_zc_yd_cjJVGQM=55894076A71C69C5C056693A6972CF712A5006CAFD5FDEF009C18D28F5D930BCF8C2698C9D077AC1667D782C350143EB571E692D3B33FEB37197F7050AAB8345260CBEC3F462FE61D35E8BF21EB1FCA3; THE_LAST_LOGIN_MOBILE=18973485974; l_yd_sign=-020170531-CWp_QGrNbcAceRFhEabgiwOxQB71AtIA1yE-c0kRUWU0zBp-cioavf-QHmMt_1t_k4X5PVosiXEZETLGKkmGM_TUAVobz_ZtUEi1ljPG3Tw..; NTESSTUDYSI=d03f415423f546c6b5e3b3528684490a; Hm_lpvt_77dc9a9d49448cf5e629e5bebaa5500b=1756916869; l_s_imooccjJVGQM=55894076A71C69C5C056693A6972CF712A5006CAFD5FDEF009C18D28F5D930BCF8C2698C9D077AC1667D782C350143EBF75315A894F6DD40AA8AD04D4D22B61DDB6703324848E52312463259DA31FA0054D9B229A414BF434FDE56D9B5F2D2EA2CFE2A1FE48E77C66987715C71BDB668; l_yd_s_imooccjJVGQM=55894076A71C69C5C056693A6972CF712A5006CAFD5FDEF009C18D28F5D930BCF8C2698C9D077AC1667D782C350143EB9E5E60B423C564636D9D18440C7547B62E3E46983CB95F4F7EE4407C9D63368B044A0A0AD0631F2E6C626FA4613C81128B37FD43756EB47C1BEC521815303416; S_INFO=1756919327|0|0&60##|18973485974; P_INFO=18973485974|1756919327|1|imooc|00&99|gux&1756916214&imooc#gux&450400#10#0#0|&0||18973485974; NTES_YD_SESS=Kxy71vWlJFySKi16l8QfgmKkSMuOoqjiVHfalLMGXfLRUGQvUZBPoCjMck6AmxKJ0Yx7.8XHw5gOZOdL619UUXltlse8SkujSkKr7dIhphS2bR2NETai6n2jBN7tG4FAU75R6jnF3GA0s4PMf72A19ZM10Ww7AB.03lNAJcom9TRFydVs2OtG85gCB3ynMceRzVzAgixnXdXGlp56NVXAxab9T6.YUAJNL.XfLQJsCwuj; l_yd_sign=-020170531-FtSFJtwyf7gqVJw6OqGIv0l5RFyq4-3WST5IEDfo32Pjda4ZHDWXvcBdPOqdJseoAyL7_BHtlsl2-zTjeBn80M36WRb9zJ7-7kLMa9cOf9g..; NTES_YD_PASSPORT=KOhd.7.mOASL29mLzE4Z2k2uzd9Jhn3bNlrWB1.I2ubRP_OdPs6CJlGQb1jwkL4MTmLcF0hKgfp_r0lXQPg6PXxWdCdBNWcAbMyRtV3WdcmTLC3YshGT.gw1.XZAH2.jx9dW2Vmv6I4AWAunqtV6hR0lbgdIrzcnEfmk1QdYM2oZwijnEgAgMHjc24AUzfXJ4QneYlJvjp6l8tVhm2HOdrpQj")
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

func (cache *MOOCUserCache) Login() {

	url := "https://reg.icourse163.org/dl/zj/yd/pwd/l"
	method := "POST"

	//runTimes, spendTime, T, X := PowGetPTurnLData(cache.Puzzle, cache.Mod, cache.X, cache.T, cache.MinTime, cache.MaxTime)
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
	req.Header.Add("Referer", "https://reg.icourse163.org/webzj/v1.0.1/pub/index_dl2_new.html?cd=%2F%2Fcmc.stu.126.net%2Fu%2Fcss%2Fcms%2F&cf=mooc_urs_login_css.css&MGID=1756950144549.6956&wdaId=UA1438236666413&pkid=cjJVGQM&product=imooc&cdnhostname=webzj.netstatic.net")
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
