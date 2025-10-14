package xuexitong

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/yatori-dev/yatori-go-core/utils"
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
