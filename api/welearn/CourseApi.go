package welearn

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strings"

	"github.com/yatori-dev/yatori-go-core/utils"
)

// 拉取课程列表json
func (cache *WeLearnUserCache) PullCourseListApi(retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	urlStr := "https://welearn.sflep.com/ajax/authCourse.aspx?action=gmc&nocache=" + fmt.Sprintf("%.16f", rand.Float32())
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, urlStr, nil)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Add("Referer", "https://welearn.sflep.com/student/index.aspx")
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Host", "welearn.sflep.com")
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
	return string(body), nil
}

// 拉取课程必要的信息，用于后续请求
// 必要信息有，uid,classid
func (cache *WeLearnUserCache) PullCourseInfoApi(cid string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	url := "https://welearn.sflep.com/student/course_info.aspx?cid=" + cid
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("Referer", "https://welearn.sflep.com/student/index.aspx")
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "welearn.sflep.com")
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

	return string(body), nil
}

// 拉取大章节
func (cache *WeLearnUserCache) PullCourseChapterApi(cid, stuid, classid string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	params := url.Values{}
	params.Set("action", "courseunits")
	params.Set("cid", cid)
	params.Set("stuid", stuid)
	params.Set("classid", classid)
	params.Set("nocache", fmt.Sprintf("%.16f", rand.Float32()))

	urlStr := "https://welearn.sflep.com/ajax/StudyStat.aspx?" + params.Encode()
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, urlStr, nil)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("Referer", "https://welearn.sflep.com/student/index.aspx")
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "welearn.sflep.com")
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
	utils.CookiesAddNoRepetition(&cache.Cookies, res.Cookies())
	return string(body), nil
}

// 拉取大章节点对应的任务点
func (cache *WeLearnUserCache) PullCoursePointApi(cid, stuid, classid, unitidx string /*BYD，这玩意就是个位置索引，不需要上id值引入，踏马的居然演我*/, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	params := url.Values{}
	params.Set("action", "scoLeaves")
	params.Set("cid", cid)
	params.Set("stuid", stuid)
	params.Set("unitidx", unitidx)
	params.Set("classid", classid)
	params.Set("nocache", fmt.Sprintf("%.16f", rand.Float32()))

	urlStr := "https://welearn.sflep.com/ajax/StudyStat.aspx?" + params.Encode()
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, urlStr, nil)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("Referer", "https://welearn.sflep.com/student/index.aspx")
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "welearn.sflep.com")
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
	utils.CookiesAddNoRepetition(&cache.Cookies, res.Cookies())
	return string(body), nil
}

// 每次开始学习前先进行一次访问
func (cache *WeLearnUserCache) StartStudyApi(cid, scoId, uid, crate, classId string, isCompleted bool, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	client := &http.Client{}

	form1 := url.Values{}
	form1.Set("action", "startsco160928")
	form1.Set("cid", cid)
	form1.Set("scoid", scoId)
	form1.Set("uid", uid)
	if isCompleted {
		form1.Set("progress", "100")
		form1.Set("crate", crate)
		form1.Set("status", "unknown")
		form1.Set("cstatus", "completed")
		form1.Set("trycount", "0")
	}
	form1.Set("nocache", fmt.Sprintf("%.16f", rand.Float32()))

	req, err := http.NewRequest("POST", "https://welearn.sflep.com/Ajax/SCO.aspx?uid="+uid, strings.NewReader(form1.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if isCompleted {
		req.Header.Set("Referer", fmt.Sprintf("https://welearn.sflep.com/Student/StudyCourse.aspx?cid=%s&classid=%s&sco=%s", cid, classId, scoId))
	} else {
		req.Header.Set("Referer", "https://welearn.sflep.com/Student/StudyCourse.aspx")
	}

	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	for _, cookie := range cache.Cookies {
		req.AddCookie(cookie)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	utils.CookiesAddNoRepetition(&cache.Cookies, resp.Cookies())
	return string(body), nil
}

// 提交学习时间的接口
func (cache *WeLearnUserCache) SubmitStudyTimeApi(uid, cid, classId, scoId string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	client := &http.Client{}

	form1 := url.Values{}
	form1.Set("action", "getscoinfo_v7")
	form1.Set("cid", cid)
	form1.Set("scoid", scoId)
	form1.Set("uid", uid)
	form1.Set("nocache", fmt.Sprintf("%.16f", rand.Float32()))

	req, err := http.NewRequest("POST", "https://welearn.sflep.com/Ajax/SCO.aspx?"+uid, strings.NewReader(form1.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", fmt.Sprintf("https://welearn.sflep.com/Student/StudyCourse.aspx?cid=%s&classid=%s&sco=%s", cid, classId, scoId))
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	for _, cookie := range cache.Cookies {
		req.AddCookie(cookie)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	utils.CookiesAddNoRepetition(&cache.Cookies, resp.Cookies())
	return string(body), nil
}

// 点击任务点进去后保持会话用的,不过似乎没卵用
func (cache *WeLearnUserCache) KeepPointSessionPlan1Api(cid, scoId, uid, classId string, sessionTime, totalTime int, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	client := &http.Client{}

	form1 := url.Values{}
	form1.Set("action", "keepsco_with_getticket_with_updatecmitime")
	form1.Set("uid", uid)
	form1.Set("cid", cid)
	form1.Set("scoid", scoId)
	form1.Set("session_time", fmt.Sprintf("%d", sessionTime)) //会话已经保持了多久
	form1.Set("total_time", fmt.Sprintf("%d", totalTime))     //会话已经保持了多久,和session_time一样的值
	//form1.Set("timelimitsec", "1800")                         //最大限制时间，到达或超过这个时间会自动退出答题
	//form1.Set("endcaltime", "false")                          //不知道干嘛的
	form1.Set("nocache", fmt.Sprintf("%.16f", rand.Float32()))

	req, err := http.NewRequest("POST", "https://welearn.sflep.com/Ajax/SCO.aspx?uid="+uid, strings.NewReader(form1.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", fmt.Sprintf("https://welearn.sflep.com/Student/StudyCourse.aspx?cid=%s&classid=%s&sco=%s", cid, classId, scoId))
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	for _, cookie := range cache.Cookies {
		req.AddCookie(cookie)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	utils.CookiesAddNoRepetition(&cache.Cookies, req.Cookies())
	return string(body), nil
}

// 直接完成学习任务点，接口1
func (cache *WeLearnUserCache) SubmitStudyPlan1Api(cid, scoId, uid, crate, classId string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	client := &http.Client{}

	form1 := url.Values{}
	form1.Set("action", "setscoinfo")
	form1.Set("cid", cid)
	form1.Set("scoid", scoId)
	form1.Set("uid", uid)
	form1.Set("data", `{"cmi":{"completion_status":"completed","interactions":[],"launch_data":"","progress_measure":"1","score":{"scaled":"`+crate+`","raw":"100"},"session_time":"0","success_status":"unknown","total_time":"0","mode":"normal"},"adl":{"data":[]},"cci":{"data":[],"service":{"dictionary":{"headword":"","short_cuts":""},"new_words":[],"notes":[],"writing_marking":[],"record":{"files":[]},"play":{"offline_media_id":"9999"}},"retry_count":"0","submit_time":""}}[INTERACTIONINFO]`)
	form1.Set("isend", "False")
	form1.Set("nocache", fmt.Sprintf("%.16f", rand.Float32()))

	req, err := http.NewRequest("POST", "https://welearn.sflep.com/Ajax/SCO.aspx?uid="+uid, strings.NewReader(form1.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", fmt.Sprintf("https://welearn.sflep.com/Student/StudyCourse.aspx?cid=%s&classid=%s&sco=%s", cid, classId, scoId))
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	for _, cookie := range cache.Cookies {
		req.AddCookie(cookie)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	utils.CookiesAddNoRepetition(&cache.Cookies, req.Cookies())
	return string(body), nil
}

// 直接完成学习任务点，接口2
func (cache *WeLearnUserCache) SubmitStudyPlan2Api(cid, scoId, uid, crate, classId string, progress int, cstatus string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	client := &http.Client{}

	form1 := url.Values{}
	form1.Set("action", "savescoinfo160928")
	form1.Set("cid", cid)
	form1.Set("scoid", scoId)
	form1.Set("uid", uid)
	form1.Set("progress", fmt.Sprintf("%d", progress))
	form1.Set("crate", crate)
	form1.Set("status", "unknown")
	form1.Set("cstatus", cstatus) //完成状态，比如completed
	form1.Set("trycount", "0")
	form1.Set("nocache", fmt.Sprintf("%.16f", rand.Float32()))

	req, err := http.NewRequest("POST", "https://welearn.sflep.com/Ajax/SCO.aspx?uid="+uid, strings.NewReader(form1.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", fmt.Sprintf("https://welearn.sflep.com/Student/StudyCourse.aspx?cid=%s&classid=%s&sco=%s", cid, classId, scoId))
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	for _, cookie := range cache.Cookies {
		req.AddCookie(cookie)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	utils.CookiesAddNoRepetition(&cache.Cookies, req.Cookies())
	return string(body), nil
}
