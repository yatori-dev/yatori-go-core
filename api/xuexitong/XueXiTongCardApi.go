package xuexitong

import (
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/yatori-dev/yatori-go-core/api/entity"
	"github.com/yatori-dev/yatori-go-core/utils"
	log2 "github.com/yatori-dev/yatori-go-core/utils/log"
)

// PageMobileChapterCard 客户端章节任务卡片 原始html数据返回
func (cache *XueXiTUserCache) PageMobileChapterCard(
	classId, courseId, knowledgeId, cardIndex, cpi int, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	method := "GET"

	params := url.Values{}
	params.Add("clazzid", strconv.Itoa(classId))
	params.Add("courseid", strconv.Itoa(courseId))
	params.Add("knowledgeid", strconv.Itoa(knowledgeId))
	params.Add("num", strconv.Itoa(cardIndex))
	params.Add("isPhone", "1")
	params.Add("control", "true")
	params.Add("cpi", strconv.Itoa(cpi))
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
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
	req, err := http.NewRequest(method, PageMobileChapterCard+"?"+params.Encode(), nil)

	if err != nil {
		//fmt.Println(err)
		log2.Print(log2.INFO, err.Error())
		return "", err
	}
	//req.Header.Add("Cookie", cache.cookie)
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}
	req.Header.Add("User-Agent", GetUA("mobile"))
	//req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36 Edg/136.0.0.0")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "mooc1-api.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")

	res, err := client.Do(req)
	if err != nil {
		//fmt.Println(err)
		//log2.Print(log2.INFO, err.Error())
		return cache.PageMobileChapterCard(classId, courseId, knowledgeId, cardIndex, cpi, retry-1, err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	utils.CookiesAddNoRepetition(&cache.cookies, res.Cookies()) //赋值cookie
	return string(body), nil
}

type APIError struct {
	Message string
}

func (e *APIError) Error() string {
	return e.Message
}

// VideoDtoFetch 视频数据
func (cache *XueXiTUserCache) VideoDtoFetch(p *entity.PointVideoDto, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	params := url.Values{}
	params.Set("k", strconv.Itoa(p.FID))
	params.Set("flag", "normal")
	params.Set("_dc", strconv.FormatInt(time.Now().UnixNano()/1e6, 10))
	method := "GET"
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
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
	resp, err := http.NewRequest(method, fmt.Sprintf("%s/%s?%s", APIChapterCardResource, p.ObjectID, params.Encode()), nil)
	// resp, err := p.Session.Client.Get(fmt.Sprintf("%s/%s?%s", APIChapterCardResource, p.ObjectID, params.Encode()))
	if err != nil {
		return "", err
	}
	resp.Header.Add("Host", " mooc1-api.chaoxing.com")
	resp.Header.Add("Connection", " keep-alive")
	//resp.Header.Add("User-Agent", " Mozilla/5.0 (Linux; Android 12; SM-N9006 Build/V417IR; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/95.0.4638.74 Mobile Safari/537.36 (schild:e9b05c3f9fb49fef2f516e86ac3c4ff1) (device:SM-N9006) Language/zh_CN com.chaoxing.mobile/ChaoXingStudy_3_6.3.7_android_phone_10822_249 (@Kalimdor)_4627cad9c4b6415cba5dc6cac39e6c96")
	resp.Header.Add("User-Agent", GetUA("mobile"))
	resp.Header.Add("X-Requested-With", " XMLHttpRequest")
	resp.Header.Add("Accept", " */*")
	resp.Header.Add("Sec-Fetch-Site", " same-origin")
	resp.Header.Add("Sec-Fetch-Mode", " cors")
	resp.Header.Add("Sec-Fetch-Dest", " empty")
	resp.Header.Add("Referer", " https://mooc1-api.chaoxing.com/ananas/modules/video/index_wap.html?v=372024-1121-1947")
	resp.Header.Add("Accept-Language", " zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7")
	//resp.Header.Add("Cookie", cache.cookie)
	for _, cookie := range cache.cookies {
		resp.AddCookie(cookie)
	}
	res, err := client.Do(resp)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return cache.VideoDtoFetch(p, retry-1, fmt.Errorf("status code: %d", res.StatusCode))
	}
	body, err := ioutil.ReadAll(res.Body)
	utils.CookiesAddNoRepetition(&cache.cookies, res.Cookies()) //赋值cookie
	return string(body), nil
}

func (cache *XueXiTUserCache) VideoSubmitStudyTimeApi(p *entity.PointVideoDto, playingTime int, isdrag int /*提交模式，0代表正常视屏播放提交，2代表暂停播放状态，3代表着点击开始播放状态*/, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	clipTime := fmt.Sprintf("0_%d", p.Duration)
	hash := md5.Sum([]byte(fmt.Sprintf("[%s][%s][%s][%s][%d][%s][%d][%s]",
		p.ClassID, cache.UserID, p.JobID, p.ObjectID, playingTime*1000, "d_yHJ!$pdA~5", p.Duration*1000, clipTime)))
	enc := hex.EncodeToString(hash[:])
	//
	urlStr := "https://mooc1.chaoxing.com/mooc-ans/multimedia/log/a/" + p.Cpi + "/" + p.DToken + "?clazzId=" + p.ClassID + "&playingTime=" + strconv.Itoa(playingTime) + "&duration=" + strconv.Itoa(p.Duration) + "&clipTime=" + clipTime + "&objectId=" + p.ObjectID + "&otherInfo=" + p.OtherInfo + "&courseId=" + p.CourseID + "&jobid=" + p.JobID + "&userid=" + cache.UserID + "&isdrag=" + strconv.Itoa(isdrag) + "&view=pc&enc=" + enc + "&rt=" + fmt.Sprintf("%.2f", p.RT) + "&videoFaceCaptureEnc=" + p.VideoFaceCaptureEnc + "&dtype=Video&_t=" + strconv.FormatInt(time.Now().UnixMilli(), 10) + "&attDuration=" + strconv.Itoa(p.Duration) + "&attDurationEnc=" + p.AttDurationEnc

	method := "GET"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
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

	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36 Edg/139.0.0.0")
	//req.Header.Add("User-Agent", GetUA("mobile"))
	req.Header.Add("Sec-Ch-Ua-Platform", "Windows")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "mooc1.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Sec-Ch-Ua-Platform", "Windows")
	req.Header.Add("Sec-Fetch-Site", "same-origin")
	req.Header.Add("Sec-Fetch-Mode", " cors")
	req.Header.Add("Sec-Fetch-Dest", " empty")
	req.Header.Add("Pragam", "no-cache")
	//fid,k8s,route,_uid,UID,vc3,uf,cx_p_token,p_auth_token,xxtenc,DSSTASH_LOG,jrose
	//fanyamoocs,videos_id,thirdRegist
	cache.cookies = append(cache.cookies, &http.Cookie{Name: "fanyamoocs", Value: "11401F839C536D9E"})
	cache.cookies = append(cache.cookies, &http.Cookie{Name: "thirdRegist", Value: "0"})
	cache.cookies = append(cache.cookies, &http.Cookie{Name: "videojs_id", Value: "1778753"})
	//filterCookies := utils.CookiesFiltration([]string{"fid", "k8s", "route", "fanyamoocs", "_uid", "UID", "vc3", "uf", "cx_p_token", "p_auth_token", "xxtenc", "DSSTASH_LOG", "jrose", "thirdRegist", "videojs_id"}, cache.cookies)
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}

	//req.AddCookie(&http.Cookie{Name: "fanyamoocs", Value: "11401F839C536D9E"})
	//req.AddCookie(&http.Cookie{Name: "thirdRegist", Value: "0"})
	//req.AddCookie(&http.Cookie{Name: "videojs_id", Value: "1778753"})

	res, err := client.Do(req)
	if err != nil {
		log2.Print(log2.DEBUG, err.Error())
		return cache.VideoSubmitStudyTimeApi(p, playingTime, isdrag, retry-1, err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", nil
	}

	if res.StatusCode != http.StatusOK {
		return cache.VideoSubmitStudyTimeApi(p, playingTime, isdrag, retry-1, fmt.Errorf("failed to fetch video, status code: %d", res.StatusCode))
	}
	//fmt.Println(string(body))
	utils.CookiesAddNoRepetition(&cache.cookies, res.Cookies()) //赋值cookie
	return string(body), nil
}

// VideoSubmitStudyTimePE 手机端学时提交
func (cache *XueXiTUserCache) VideoSubmitStudyTimePEApi(p *entity.PointVideoDto, playingTime int, isdrag int /*提交模式，0代表正常视屏播放提交，2代表暂停播放状态，3代表着点击开始播放状态*/, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	clipTime := fmt.Sprintf("0_%d", p.Duration)
	hash := md5.Sum([]byte(fmt.Sprintf("[%s][%s][%s][%s][%d][%s][%d][%s]",
		p.ClassID, cache.UserID, p.JobID, p.ObjectID, playingTime*1000, "d_yHJ!$pdA~5", p.Duration*1000, clipTime)))
	enc := hex.EncodeToString(hash[:])
	//
	urlStr := "https://mooc1.chaoxing.com/mooc-ans/multimedia/log/a/" + p.Cpi + "/" + p.DToken + "?clazzId=" + p.ClassID + "&playingTime=" + strconv.Itoa(playingTime) + "&duration=" + strconv.Itoa(p.Duration) + "&clipTime=" + clipTime + "&objectId=" + p.ObjectID + "&otherInfo=" + p.OtherInfo + "&courseId=" + p.CourseID + "&jobid=" + p.JobID + "&userid=" + cache.UserID + "&isdrag=" + strconv.Itoa(isdrag) + "&view=json&enc=" + enc + "&rt=" + fmt.Sprintf("%.2f", p.RT) + "&videoFaceCaptureEnc=" + p.VideoFaceCaptureEnc + "&dtype=Video&_t=" + strconv.FormatInt(time.Now().UnixMilli(), 10) + "&attDuration=" + strconv.Itoa(p.Duration) + "&attDurationEnc=" + p.AttDurationEnc

	method := "GET"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
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
		//fmt.Println(err)
		log2.Print(log2.INFO, err.Error())
		return "", nil
	}

	//req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36 Edg/139.0.0.0")
	req.Header.Add("User-Agent", GetUA("mobile"))
	req.Header.Add("Sec-Ch-Ua-Platform", "Windows")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "mooc1.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Sec-Ch-Ua-Platform", "Windows")
	req.Header.Add("Sec-Fetch-Site", "same-origin")
	req.Header.Add("Sec-Fetch-Mode", " cors")
	req.Header.Add("Sec-Fetch-Dest", " empty")
	req.Header.Add("Pragam", "no-cache")
	//fid,k8s,route,_uid,UID,vc3,uf,cx_p_token,p_auth_token,xxtenc,DSSTASH_LOG,jrose
	//fanyamoocs,videos_id,thirdRegist
	cache.cookies = append(cache.cookies, &http.Cookie{Name: "fanyamoocs", Value: "11401F839C536D9E"})
	cache.cookies = append(cache.cookies, &http.Cookie{Name: "thirdRegist", Value: "0"})
	cache.cookies = append(cache.cookies, &http.Cookie{Name: "videojs_id", Value: "1778753"})
	filterCookies := utils.CookiesFiltration([]string{"fid", "k8s", "route", "fanyamoocs", "_uid", "UID", "vc3", "uf", "cx_p_token", "p_auth_token", "xxtenc", "DSSTASH_LOG", "jrose", "thirdRegist", "videojs_id"}, cache.cookies)
	for _, cookie := range filterCookies {
		req.AddCookie(cookie)
	}

	//req.AddCookie(&http.Cookie{Name: "fanyamoocs", Value: "11401F839C536D9E"})
	//req.AddCookie(&http.Cookie{Name: "thirdRegist", Value: "0"})
	//req.AddCookie(&http.Cookie{Name: "videojs_id", Value: "1778753"})

	res, err := client.Do(req)
	if err != nil {
		//fmt.Println(err)
		log2.Print(log2.DEBUG, err.Error())
		return cache.VideoSubmitStudyTimePEApi(p, playingTime, isdrag, retry-1, err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		//fmt.Println(err)
		log2.Print(log2.INFO, err.Error())
		return "", err
	}

	if res.StatusCode != http.StatusOK {
		return cache.VideoSubmitStudyTimePEApi(p, playingTime, isdrag, retry-1, fmt.Errorf("failed to fetch video, status code: %d", res.StatusCode))
	}
	//fmt.Println(string(body))
	utils.CookiesAddNoRepetition(&cache.cookies, res.Cookies()) //赋值cookie
	return string(body), nil
}

// Deprecated: 此方法有BUG不推荐使用，将会在未来版本删除
func (cache *XueXiTUserCache) VideoDtoPlayReport(p *entity.PointVideoDto, playingTime int, isdrag int /*提交模式，0代表正常视屏播放提交，2代表暂停播放状态，3代表着点击开始播放状态，4代表播放结束*/, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	clipTime := fmt.Sprintf("0_%d", p.Duration)
	hash := md5.Sum([]byte(fmt.Sprintf("[%s][%s][%s][%s][%d][%s][%d][%s]",
		p.ClassID, cache.UserID, p.JobID, p.ObjectID, playingTime*1000, "d_yHJ!$pdA~5", p.Duration*1000, clipTime)))
	enc := hex.EncodeToString(hash[:])
	//fmt.Println(enc)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
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
	params := url.Values{}
	params.Set("otherInfo", p.OtherInfo)
	params.Set("playingTime", strconv.Itoa(playingTime))
	params.Set("duration", strconv.Itoa(p.Duration))
	params.Set("jobid", p.JobID)
	params.Set("clipTime", clipTime)
	params.Set("clazzId", p.ClassID)
	params.Set("objectId", p.ObjectID)
	params.Set("userid", cache.UserID)
	params.Set("isdrag", strconv.Itoa(isdrag)) //0为正常播放，2为点击暂停播放状态，3为点击开始播放
	params.Set("enc", enc)
	params.Set("rt", fmt.Sprintf("%f", p.RT))
	//params.Set("retry", "0.9")
	params.Set("dtype", "Video")
	params.Set("view", "pc")
	params.Set("rt", "0.9")
	params.Set("courseId", p.CourseID)
	params.Set("videoFaceCaptureEnc", p.VideoFaceCaptureEnc)
	params.Set("attDuration", strconv.Itoa(p.Duration))
	params.Set("attDurationEnc", p.AttDurationEnc)
	params.Set("_t", strconv.FormatInt(time.Now().UnixMilli(), 10))

	// 自定义编码函数以保留 & 和 =
	encodedParams := encodeWithSafeChars(params)
	method := "GET"

	resp, err := http.NewRequest(method, fmt.Sprintf("%s/%s/%s?%s", APIVideoPlayReport, p.Cpi, p.DToken, encodedParams), nil)
	if err != nil {
		return "", err
	}

	//resp.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36 Edg/139.0.0.0")
	resp.Header.Add("User-Agent", GetUA("mobile"))
	resp.Header.Add("Sec-Ch-Ua-Platform", "Windows")
	resp.Header.Add("Accept", "*/*")
	resp.Header.Add("Host", "mooc1.chaoxing.com")
	resp.Header.Add("Connection", "keep-alive")
	for _, cookie := range cache.cookies {
		resp.AddCookie(cookie)
	}
	resp.Header.Add("Referer", "https://mooc1.chaoxing.com/ananas/modules/video/index.html?v=2023-1110-1610")
	resp.Header.Add("Content-Type", " application/json")
	res, err := client.Do(resp)
	if err != nil {
		time.Sleep(time.Duration(retry*5) * time.Second)
		return cache.VideoDtoPlayReport(p, playingTime, isdrag, retry-1, err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch video, status code: %d", res.StatusCode)
	}
	body, err := ioutil.ReadAll(res.Body)
	utils.CookiesAddNoRepetition(&cache.cookies, res.Cookies()) //赋值cookie
	return string(body), nil
}

// encodeWithSafeChars 自定义编码函数，保留 & 和 =
func encodeWithSafeChars(values url.Values) string {
	var result []string
	for key, list := range values {
		for _, value := range list {
			// 手动编码键和值，但不编码 & 和 =
			encodedKey := url.QueryEscape(key)
			encodedValue := url.QueryEscape(value)
			// 替换 %3D (等号) 和 %26 (与号) 回原字符
			encodedKey = replaceSpecialChars(encodedKey)
			encodedValue = replaceSpecialChars(encodedValue)
			result = append(result, encodedKey+"="+encodedValue)
		}
	}
	return strings.Join(result, "&")
}

// replaceSpecialChars 将 %3D 和 %26 替换回等号和与号
func replaceSpecialChars(s string) string {
	return strings.NewReplacer("%3D", "=", "%26", "&").Replace(s)
}

// WorkFetchQuestion 获取作业题目
func (cache *XueXiTUserCache) WorkFetchQuestion(p *entity.PointWorkDto, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	method := "GET"

	params := url.Values{}
	params.Add("courseid", p.CourseID)
	var SorW func(string, string) string
	SorW = func(s string, w string) string {
		if s != "0" {
			return fmt.Sprintf("%s-%s", s, w)
		}
		return w
	}
	params.Add("workid", SorW(p.SchoolID, p.WorkID))
	params.Add("jobid", p.JobID)
	params.Add("needRedirect", "true")
	params.Add("knowledgeid", strconv.Itoa(p.KnowledgeID))
	params.Add("userid", p.PUID)
	params.Add("ut", "s")
	params.Add("clazzId", p.ClassID)
	params.Add("cpi", p.Cpi)
	params.Add("ktoken", p.KToken)
	params.Add("enc", p.Enc)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
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
	req, err := http.NewRequest(method, PageMobileWork+"?"+params.Encode(), nil)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	//req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36 Edg/139.0.0.0")
	req.Header.Add("User-Agent", GetUA("mobile"))
	req.Header.Add("Sec-Ch-Ua-Platform", "Windows")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "mooc1-api.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")
	//req.Header.Add("Cookie", cache.cookie)
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}

	res, err := client.Do(req)
	if err != nil {
		//fmt.Println(err)
		time.Sleep(time.Duration(retry*5) * time.Second)
		return cache.WorkFetchQuestion(p, retry-1, err)
		//return "", err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	utils.CookiesAddNoRepetition(&cache.cookies, res.Cookies()) //赋值cookie
	return string(body), nil
}

func (cache *XueXiTUserCache) WorkCommit(p *entity.PointWorkDto, fields []entity.WorkInputField, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	method := "POST"

	//TODO 此处需要对答案进行分析后提交 具体body模板 在 examples 中
	payload := strings.NewReader("")

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
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
	req, err := http.NewRequest(method, ApiWorkCommit, payload)

	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	//req.Header.Add("Cookie", cache.cookie)
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}
	//req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36 Edg/139.0.0.0")
	req.Header.Add("User-Agent", GetUA("mobile"))
	req.Header.Add("Sec-Ch-Ua-Platform", "Windows")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "mooc1-api.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")

	res, err := client.Do(req)
	if err != nil {
		//fmt.Println(err)
		time.Sleep(time.Duration(retry*5) * time.Second)
		return cache.WorkCommit(p, fields, retry, err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	utils.CookiesAddNoRepetition(&cache.cookies, res.Cookies()) //赋值cookie
	return string(body), nil
}

func (cache *XueXiTUserCache) DocumentDtoReadingReport(p *entity.PointDocumentDto, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	method := "GET"
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
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
	params := url.Values{}

	params.Add("jobid", p.JobID)
	params.Add("knowledgeid", strconv.Itoa(p.KnowledgeID))
	params.Add("courseid", p.CourseID)
	params.Add("clazzid", p.ClassID)
	params.Add("jtoken", p.Jtoken)
	params.Add("_dc", strconv.FormatInt(time.Now().UnixMilli(), 10))

	resp, err := http.NewRequest(method, ApiDocumentReadingReport+"?"+params.Encode(), nil)
	if err != nil {
		return "", err
	}

	//resp.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36 Edg/139.0.0.0")
	resp.Header.Add("User-Agent", GetUA("mobile"))
	resp.Header.Add("Sec-Ch-Ua-Platform", "Windows")
	resp.Header.Add("Accept", "*/*")
	resp.Header.Add("Host", "mooc1.chaoxing.com")
	resp.Header.Add("Connection", "keep-alive")
	//resp.Header.Add("Cookie", cache.cookie)
	for _, cookie := range cache.cookies {
		resp.AddCookie(cookie)
	}
	resp.Header.Add("Content-Type", " application/json")

	res, err := client.Do(resp)
	if err != nil {
		time.Sleep(time.Duration(retry*5) * time.Second)
		return cache.DocumentDtoReadingReport(p, retry-1, err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return cache.DocumentDtoReadingReport(p, retry-1, fmt.Errorf("status code: %d", res.StatusCode))
	}

	body, err := ioutil.ReadAll(res.Body)
	utils.CookiesAddNoRepetition(&cache.cookies, res.Cookies()) //赋值cookie
	return string(body), nil
}

// 另一个文档完成接口
func (cache *XueXiTUserCache) DocumentDtoReadingBookReport(p *entity.PointDocumentDto, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	method := "GET"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
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
	params := url.Values{}

	params.Add("jobid", p.JobID)
	params.Add("knowledgeid", strconv.Itoa(p.KnowledgeID))
	params.Add("courseid", p.CourseID)
	params.Add("clazzid", p.ClassID)
	params.Add("jtoken", p.Jtoken)
	params.Add("_dc", strconv.FormatInt(time.Now().UnixMilli(), 10))

	resp, err := http.NewRequest(method, "https://mooc1.chaoxing.com/ananas/job?"+params.Encode(), nil)
	if err != nil {
		return "", err
	}

	//resp.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36 Edg/139.0.0.0")
	resp.Header.Add("User-Agent", GetUA("mobile"))
	resp.Header.Add("Sec-Ch-Ua-Platform", "Windows")
	resp.Header.Add("Accept", "*/*")
	resp.Header.Add("Host", "mooc1.chaoxing.com")
	resp.Header.Add("Connection", "keep-alive")
	//resp.Header.Add("Cookie", cache.cookie)
	for _, cookie := range cache.cookies {
		resp.AddCookie(cookie)
	}
	resp.Header.Add("Content-Type", " application/json")

	res, err := client.Do(resp)
	if err != nil {
		time.Sleep(time.Duration(retry*5) * time.Second)
		return cache.DocumentDtoReadingBookReport(p, retry-1, err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		time.Sleep(time.Duration(retry*5) * time.Second)
		return cache.DocumentDtoReadingBookReport(p, retry-1, fmt.Errorf("status code: %d", res.StatusCode))
	}

	body, err := ioutil.ReadAll(res.Body)
	utils.CookiesAddNoRepetition(&cache.cookies, res.Cookies()) //赋值cookie
	return string(body), nil
}

// 外链完成接口
func (cache *XueXiTUserCache) HyperlinkDtoCompleteReport(p *entity.PointHyperlinkDto, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	urlStr := "https://mooc1.chaoxing.com/ananas/job/hyperlink?jobid=" + p.JobID + "&knowledgeid=" + strconv.Itoa(p.KnowledgeID) + "&courseid=" + p.CourseID + "&clazzid=" + p.ClassID + "&jtoken=" + p.Jtoken + "&checkMicroTopic=true&microTopicId=undefined&_dc=" + strconv.FormatInt(time.Now().UnixMilli(), 10)
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
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "mooc1.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}

	res, err := client.Do(req)
	if err != nil {
		//fmt.Println(err)
		time.Sleep(time.Duration(retry*5) * time.Second)
		return cache.HyperlinkDtoCompleteReport(p, retry-1, lastErr)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return cache.HyperlinkDtoCompleteReport(p, retry-1, fmt.Errorf("status code: %d", res.StatusCode))
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	//fmt.Println(string(body))
	return string(body), nil
}

// 拉取u参数
//func (cache *XueXiTUserCache) PullLiveUParam(liveId string, retry int, lastErr error) (string, int) {
//	if retry < 0 {
//		return "", lastErr
//	}
//	urlStr := "https://zhibo.chaoxing.com/" + liveId + "?courseId=251085317&classId=128238814&knowledgeId=967705955&jobId=live-6000256327632944&userId=221172669&rt=0.9&livesetenc=6b70119b3792fc81816f8ca1f4ba54c8&isjob=true&watchingInCourse=1&customPara1=128238814_251085317&customPara2=92401b9a2dce6d2e49c0706a186247c3&jobfs=0&isNotDrag=1&livedragenc=3a828d58949143863af938a250d4026c&sw=0&ds=0&liveswdsenc=b2f601fb4b8e506dd3e823232106918b"
//	method := "GET"
//
//	client := &http.Client{}
//	req, err := http.NewRequest(method, urlStr, nil)
//
//	if err != nil {
//		fmt.Println(err)
//		return "", 0
//	}
//	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")
//	req.Header.Add("Accept", "*/*")
//	req.Header.Add("Host", "zhibo.chaoxing.com")
//	req.Header.Add("Connection", "keep-alive")
//	for _, cookie := range cache.cookies {
//		req.AddCookie(cookie)
//	}
//
//	res, err := client.Do(req)
//	if err != nil {
//		fmt.Println(err)
//		return "", 0
//	}
//	defer res.Body.Close()
//
//	body, err := ioutil.ReadAll(res.Body)
//	if err != nil {
//		fmt.Println(err)
//		return "", 0
//	}
//	//fmt.Println(string(body))
//	r, _ := regexp.Compile("var uInfo = '([\\w\\W]*?)';")
//	match := r.FindStringSubmatch(string(body))
//	if len(match) <= 0 {
//		return "", 0
//	}
//
//	r1, _ := regexp.Compile("var watchMoment = ([\\d]*?);")
//	match1 := r1.FindStringSubmatch(string(body))
//	if len(match1) <= 0 {
//		return "", 0
//	}
//	atoi, err := strconv.Atoi(match1[1])
//	if err != nil {
//		return "", 0
//	}
//	utils.CookiesAddNoRepetition(&cache.cookies, res.Cookies())
//	return match[1], atoi
//}

// 拉取直播数据
func (cache *XueXiTUserCache) PullLiveInfoApi(p *entity.PointLiveDto, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	urlStr := "https://mooc1.chaoxing.com/ananas/live/liveinfo?liveid=" + p.LiveId + "&userid=" + p.UserId + "&clazzid=" + p.ClassID + "&knowledgeid=" + fmt.Sprintf("%d", p.KnowledgeID) + "&courseid=" + p.CourseID + "&jobid=" + p.JobID + "&ut=s"
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
		return "", err
	}

	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "mooc1.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}

	res, err := client.Do(req)
	if err != nil {
		time.Sleep(time.Duration(retry*5) * time.Second)
		return cache.PullLiveInfoApi(p, retry-1, err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		time.Sleep(time.Duration(retry*5) * time.Second)
		return cache.PullLiveInfoApi(p, retry-1, fmt.Errorf("status code: %d", res.StatusCode))
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	//fmt.Println(string(body))
	return string(body), nil
}

// 看直播前先建立连接
func (cache *XueXiTUserCache) LiveRelationReport(p *entity.PointLiveDto, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	urlStr := "https://mooc1.chaoxing.com/mooc-ans/live/relation?courseid=" + p.CourseID + "&knowledgeid=" + fmt.Sprintf("%d", p.KnowledgeID) + "&ut=s&jobid=" + p.JobID + "&aid=" + p.Aid
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
		return "", err
	}

	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "mooc1.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}
	res, err := client.Do(req)
	if err != nil {
		time.Sleep(time.Duration(retry*5) * time.Second)
		return cache.LiveRelationReport(p, retry-1, err)
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		time.Sleep(time.Duration(retry*5) * time.Second)
		return cache.LiveRelationReport(p, retry-1, fmt.Errorf("status code: %d", res.StatusCode))
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	//fmt.Println(string(body))
	utils.CookiesAddNoRepetition(&cache.cookies, res.Cookies())
	return string(body), nil
}

// 直播上报接口学时存储接口
func (cache *XueXiTUserCache) LiveWatchMomentReport(p *entity.PointLiveDto, UParam string, watchMoment float64, retry int, lastErr error) (string, error) {
	urlStr := "https://zhibo.chaoxing.com/apis/live/put/watchMoment?liveId=" + p.LiveId + "&streamName=" + p.StreamName + "&vdoid=" + p.Vdoid + "&watchMoment=" + fmt.Sprintf("%.6f", watchMoment) + "&t=" + strconv.FormatInt(time.Now().UnixMilli(), 10) + "&u=" + UParam
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
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "zhibo.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		time.Sleep(time.Duration(retry*5) * time.Second)
		return cache.LiveWatchMomentReport(p, UParam, watchMoment, retry-1, fmt.Errorf("status code: %d", res.StatusCode))
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	//fmt.Println(string(body))
	utils.CookiesAddNoRepetition(&cache.cookies, res.Cookies())
	return string(body), nil
}

// 直播尚博接口学时提交接口
func (cache *XueXiTUserCache) LiveSaveTimePcReport(p *entity.PointLiveDto, retry int, lastErr error) (string, error) {

	urlStr := "https://zhibo.chaoxing.com/saveTimePc?streamName=" + p.StreamName + "&vdoid=" + p.Vdoid + "&userId=" + p.UserId + "&isStart=1&t=" + strconv.FormatInt(time.Now().UnixMilli(), 10) + "&courseId=" + p.CourseID
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

	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "zhibo.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		time.Sleep(time.Duration(retry*5) * time.Second)
		return cache.LiveSaveTimePcReport(p, retry-1, fmt.Errorf("status code: %d", res.StatusCode))
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	//fmt.Println(string(body))
	utils.CookiesAddNoRepetition(&cache.cookies, res.Cookies())
	return string(body), nil
}
