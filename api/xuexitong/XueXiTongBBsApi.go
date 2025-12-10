package xuexitong

import (
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/yatori-dev/yatori-go-core/utils"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// 拉取任务点链接数据
func (cache *XueXiTUserCache) PullBbsCircleIdApi(mid, jobid string, isPortal bool, knowledgeid, ut, clazzId, enc, utenc, courseId string, isJob bool) (string, string, error) {

	//<div class="PublicCardBox" id="topicMainDiv" data="https://groupweb.chaoxing.com/course/topic/v3/bbs/8893f73c548368ee1a6f632c21a9219d/197f09e0290c430a885db5cf5459987d/replysList?courseId=255126242&classId=127158759" onclick="openTopicUrl()">

	urlStr := "https://mooc1.chaoxing.com/mooc-ans/bbscircle/chapter?mtopicid=" + mid + "&jobid=" + jobid + "&isPortal=" + strconv.FormatBool(isPortal) + "&knowledgeid=" + knowledgeid + "&ut=" + ut + "&clazzId=" + clazzId + "&enc=" + enc + "&utenc=" + utenc + "&courseid=" + courseId + "&isJob=" + strconv.FormatBool(isJob)
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, urlStr, nil)

	if err != nil {
		fmt.Println(err)
		return "", "", nil
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
		fmt.Println(err)
		return "", "", nil
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", "", nil
	}
	//utils.CookiesAddNoRepetition(&cache.cookies, res.Cookies())

	compile := regexp.MustCompile("topic/v3/bbs/([^/]+)/([^/]+)/replysList")
	submatch := compile.FindStringSubmatch(string(body))
	if len(submatch) > 2 {
		return submatch[1], submatch[2], nil
	}

	return "", "", errors.New("无法截取bbs关键信息")
}

// 拉取utenc参数
func (cache *XueXiTUserCache) PullUtEnc(courseId, clazzid, chapterId, enc string) (string, error) {

	//urlStr := "https://mooc1.chaoxing.com/mycourse/studentstudy?chapterId=" + chapterId + "&courseId=" + courseId + "&clazzid=" + clazzid + "&cpi=" + cpi + "&enc=" + enc + "&mooc2=1"
	urlStr := "https://mooc1.chaoxing.com/mooc-ans/mycourse/studentstudy?chapterId=" + chapterId + "&courseId=" + courseId + "&clazzid=" + clazzid + "&enc=" + enc
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
		return "", err
	}

	//req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36 Edg/136.0.0.0")
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "groupweb.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer res.Body.Close()
	//替换cookie
	//utils.CookiesAddNoRepetition(&cache.cookies, res.Cookies())

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	if strings.Contains(string(body), "请输入验证码") || strings.Contains(string(body), "请输入图片中的验证码") {
		return "", errors.New("触发验证码")
	}
	compile := regexp.MustCompile(`var utEnc="([^"]+)";`)
	submatch := compile.FindStringSubmatch(string(body))
	if len(submatch) > 1 {
		return submatch[1], nil
	}
	return "", errors.New("无法获取utEnc参数")
}

// 拉取讨论关键参数信新城
func (cache *XueXiTUserCache) PullBbsInfoApi(id1, id2, courseId, classId string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}

	urlStr := "https://groupweb.chaoxing.com/course/topic/v3/bbs/" + id1 + "/" + id2 + "/replysList?courseId=" + courseId + "&classId=" + classId
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
		return "", err
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "groupweb.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	return string(body), nil
}

// 拉取讨论关键参数手机端。其中mtopid就是mid
func (cache *XueXiTUserCache) PullPhoneBbsInfoApi(mtopid, jobid, knowledgeid, courseId, clazzId string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}

	//url := "https://mooc1-api.chaoxing.com/mooc-ans/bbscircle/chapter?mtopicid=6126910848511765298213523&jobid=1765298213522842&isPortal=false&knowledgeid=1088037085&ut=s&clazzId=134204187&enc&utenc=undefined&courseid=258101827&isJob=true&isMobile=true"
	url := "https://mooc1-api.chaoxing.com/mooc-ans/bbscircle/chapter?mtopicid=" + mtopid + "&jobid=" + jobid + "&isPortal=false&knowledgeid=" + knowledgeid + "&ut=s&clazzId=" + clazzId + "&enc&utenc=undefined&courseid=" + courseId + "&isJob=true&isMobile=true"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("User-Agent", GetUA("mobile"))
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	req.Header.Add("Upgrade-Insecure-Requests", "1")
	//req.Header.Add("Referer", "https://mooc1-api.chaoxing.com/ananas/modules/insertbbs/index.html?v=2025-1128-0958")
	req.Header.Add("Accept-Language", "zh-CN,en-US;q=0.9")
	req.Header.Add("X-Requested-With", "com.chaoxing.mobile")
	req.Header.Add("Host", "mooc1-api.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")
	for _, cookie := range cache.cookies {
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

	return string(body), nil
}

// 拉取讨论任务点详细关键信息
func (cache *XueXiTUserCache) PullPhoneBbsDetailApi(topicId string) (string, error) {
	_c_0 := ParamFor_c_0_Generete()
	_time := fmt.Sprintf("%d", time.Now().UnixMilli())
	puid := ""
	//获取puid
	if puid == "" {
		for _, cookie := range cache.cookies {
			if cookie.Name == "UID" { //获取puid
				puid = cookie.Value
				break
			}
		}
	}
	inf_enc := InfEncSign(map[string]string{
		"_c_0_": _c_0,
		"token": "4faa8662c59590c6f43ae9fe5b002b42",
		"_time": _time,
	}, []string{"_c_0_", "token", "_time"})
	url := "https://groupyd.chaoxing.com/apis/topic/getTopic?_c_0_=" + _c_0 + "&token=4faa8662c59590c6f43ae9fe5b002b42&_time=" + _time + "&inf_enc=" + inf_enc
	method := "POST"

	payload := strings.NewReader("puid=" + puid + "&maxW=1080&topicId=" + topicId)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("User-Agent", GetUA("mobile"))
	req.Header.Add("Connection", "Keep-Alive")
	req.Header.Add("Accept-Language", "zh_CN")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "groupyd.chaoxing.com")
	for _, cookie := range cache.cookies {
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
	return string(body), nil
}

// 回复讨论
func (cache *XueXiTUserCache) AnswerBbsApi(topicUUid, courseId, classId, topic_content, urlToken, bbsid string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	urlStr := "https://groupweb.chaoxing.com/pc/invitation/" + topicUUid + "/addReplys"
	method := "POST"
	newUUID, _ := uuid.NewUUID()

	payload := strings.NewReader("courseId=" + courseId + "&classId=" + classId + "&replyId=-1&uuid=" + newUUID.String() + "&topic_content=" + url.QueryEscape(topic_content) + "&anonymous=&urlToken=" + urlToken + "&bbsid=" + bbsid)

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
	req, err := http.NewRequest(method, urlStr, payload)

	if err != nil {
		return "", err
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "groupweb.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
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
	return string(body), nil
}

// 手机端回复讨论任务点
func (cache *XueXiTUserCache) AnswerPhoneBbsApi(classId, topicUUID, content string) (string, error) {
	_c_0 := ParamFor_c_0_Generete()
	_time := fmt.Sprintf("%d", time.Now().UnixMilli())
	puid := ""
	//获取puid
	if puid == "" {
		for _, cookie := range cache.cookies {
			if cookie.Name == "UID" { //获取puid
				puid = cookie.Value
				break
			}
		}
	}
	newUUID, _ := uuid.NewUUID()
	uuidV := newUUID.String()
	inf_enc := InfEncSign(map[string]string{
		"token":     "4faa8662c59590c6f43ae9fe5b002b42",
		"_time":     _time,
		"_c_0_":     _c_0,
		"puid":      puid,
		"uuid":      uuidV,
		"tag":       "classId" + classId,
		"maxW":      "1080",
		"topicUUID": topicUUID,
		"anonymous": "0",
	}, []string{"token", "_time", "_c_0_", "puid", "uuid", "tag", "maxW", "topicUUID", "anonymous"})
	urlStr := "https://groupyd.chaoxing.com/apis/invitation/addReply?token=4faa8662c59590c6f43ae9fe5b002b42&_time=" + _time + "&_c_0_=" + _c_0 + "&puid=" + puid + "&uuid=" + uuidV + "&tag=classId" + classId + "&maxW=1080&topicUUID=" + topicUUID + "&anonymous=0&inf_enc=" + inf_enc
	method := "POST"

	payload := strings.NewReader("content=" + url.QueryEscape(content))

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
	req, err := http.NewRequest(method, urlStr, payload)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("User-Agent", GetUA("mobile"))
	req.Header.Add("Connection", "Keep-Alive")
	req.Header.Add("Accept-Language", "zh_CN")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "groupyd.chaoxing.com")
	for _, cookie := range cache.cookies {
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
	return string(body), nil
}

// InfEncSign 移动端为参数添加 inf_enc 签名
func InfEncSign(params map[string]string, order []string) string {
	const DESKey = "Z(AfY@XS"

	parts := make([]string, 0, len(order))
	for _, k := range order {
		// 跳过不存在的 key（或你也可以要求都存在）
		v, ok := params[k]
		if !ok {
			continue
		}
		// 使用 url.QueryEscape 与 Python urlencode 行为兼容（空格 -> +）
		parts = append(parts, k+"="+url.QueryEscape(v))
	}

	// 拼接并加上 DESKey
	query := strings.Join(parts, "&") + "&DESKey=" + DESKey

	// md5
	sum := md5.Sum([]byte(query))
	return hex.EncodeToString(sum[:])
}

// 移动端_c_0参数生成
func ParamFor_c_0_Generete() string {
	u := uuid.New()
	c0 := strings.ReplaceAll(u.String(), "-", "")
	return c0
}
