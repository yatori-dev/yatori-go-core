package xuexitong

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/yatori-dev/yatori-go-core/que-core/qentity"
	"github.com/yatori-dev/yatori-go-core/que-core/qtype"
	"github.com/yatori-dev/yatori-go-core/utils/qutils"
)

// 学习通题目
type XXTWorkQuestionSubmitEntity struct {
	CourseId             string
	WordId               string
	CurrentTime          string
	CurrentCpi           string
	CurrentUploadEnc     string
	Cfid                 string
	Index                string
	AddTimes             string
	LimitWorkSubmitTimes string
	MatchEnc             string
	RandomOptions        string
	QuestionDataType     string
	EncWork              string
	TestUserRelationId   string
	ClassId              string
	Type                 string
	IsPhone              string
	Imei                 string
	SubCount             string
	RemainTime           string
	TempSave             string
	TimeOver             string
	EncRemainTime        string
	EncLastUpdateTime    string
	Cpi                  string
	Enc                  string
	Source               string
	Score                string
	UserId               string
	Tid                  string
	EnterPageTime        string
	AnsweredView         string
	PaperGroupId         string
	QuestionId           string //题目ID
	ExamRelationId       string
	AnswerId             string
	RemainTimeParam      string
	QType                qtype.QType //题目类型
	TypeName             string
	QuestionTypeCode     string
	QuestionTypeStr      string
	Question             qentity.Question
}

// PullWorkListHtmlApi 拉取邮箱作业列表
func (cache *XueXiTUserCache) PullWorkListHtmlApi(courseId string, classId string, cpi string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	urlStr := "https://mooc1-api.chaoxing.com/work/task-list?courseId=" + courseId + "&classId=" + classId + "&cpi=" + cpi
	method := "GET"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}
	client := &http.Client{
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return errors.New("太多重定向")
			}

			// 复制 Cookie
			if len(via) > 0 {
				for _, c := range via[0].Cookies() {
					req.AddCookie(c)
				}
			}
			return nil // 允许重定向
		},
	}
	req, err := http.NewRequest(method, urlStr, nil)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("User-Agent", GetUA("mobile"))
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	req.Header.Add("Upgrade-Insecure-Requests", "1")
	req.Header.Add("accept-language", "zh_CN")
	req.Header.Add("X-Requested-With", "com.chaoxing.mobile")
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}
	req.Header.Add("Host", "mooc1-api.chaoxing.com")
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
	//fmt.Println(string(body))
	return string(body), nil
}

// PullWorkEnterInformHtmlApi 拉取进入作业页面的html
func (cache *XueXiTUserCache) PullWorkEnterInformHtmlApi(
	taskrefId, msgId, courseId, userId, clazzId, enterType, encTask string,
	retry int, lastErr error,
) (string, string, error) {

	urlStr := "https://mooc1-api.chaoxing.com/android/mtaskmsgspecial?taskrefId=" +
		taskrefId + "&msgId=" + msgId + "&courseId=" + courseId + "&userId=" + userId +
		"&clazzId=" + clazzId + "&type=" + enterType + "&enc_task=" + encTask

	method := "GET"

	var finalURL string // ⭐ 用于保存最终的有效 URL

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client := &http.Client{
		Transport: tr,

		// ⭐ 重定向处理，抓取重定向后的 URL
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// 限制次数
			if len(via) >= 5 {
				return errors.New("太多重定向")
			}
			// 每次重定向都更新 finalURL
			finalURL = req.URL.String()

			// ⭐ 手动携带 Cookie
			if len(via) > 0 {
				for _, c := range via[0].Cookies() {
					req.AddCookie(c)
				}
			}
			return nil
		},
	}

	req, err := http.NewRequest(method, urlStr, nil)
	if err != nil {
		return "", "", err
	}

	req.Header.Add("User-Agent", GetUA("mobile"))
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	req.Header.Add("Upgrade-Insecure-Requests", "1")
	req.Header.Add("accept-language", "zh_CN")
	req.Header.Add("X-Requested-With", "com.chaoxing.mobile")
	req.Header.Add("Host", "mooc1-api.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")

	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}

	res, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()

	// ⭐ 如果没有重定向，最终 URL 就是初始 URL
	if finalURL == "" {
		finalURL = res.Request.URL.String()
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", "", err
	}

	return string(body), finalURL, nil
}

// 拉取试卷
func (cache *XueXiTUserCache) PullWorkPaperHtmlApi(courseId, classId, workId, source, msgId, cpi, workAnswerId, enc, keyboardDisplayRequiresUserAction string, retry int, lastErr error) (string, error) {
	//url := "https://mooc1-api.chaoxing.com/mooc-ans/work/phone/doHomeWork?courseId=258101827&workId=48731428&classId=134204187&oldWorkId&cpi=411545273&mooc=1&msgId=0&source=0&checkIntegrity=true&enc=737ad94cd5529ffa3ba68606eb91a124&keyboardDisplayRequiresUserAction=1"

	//url := "https://mooc1-api.chaoxing.com/mooc-ans/work/phone/doHomeWork?courseId=258101827&workId=49156357&cpi=411545273&workAnswerId=54758238&classId=134204187&oldWorkId&mooc=1&msgId=0&source=0&checkIntegrity=true&enc=737ad94cd5529ffa3ba68606eb91a124&keyboardDisplayRequiresUserAction=1"
	urlStr := "https://mooc1-api.chaoxing.com/mooc-ans/work/phone/doHomeWork?courseId=" + courseId + "&workId=" + workId + "&cpi=" + cpi + "&workAnswerId=" + workAnswerId + "&classId=" + classId + "&oldWorkId&mooc=1&msgId=" + msgId + "&source=" + source + "&checkIntegrity=true&enc=" + enc + "&keyboardDisplayRequiresUserAction=" + keyboardDisplayRequiresUserAction
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
	req.Header.Add("User-Agent", GetUA("mobile"))
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	req.Header.Add("Upgrade-Insecure-Requests", "1")
	req.Header.Add("accept-language", "zh_CN")
	req.Header.Add("Referer", "https://mooc1-api.chaoxing.com/mooc-ans/work/phone/task-work?taskrefId=49156357&courseId=258101827&classId=134204187&userId=346635955&role=&source=0&enc_task=f17cf2658668d00b935b2c218fefcf56&cpi=411545273&vx=0&fromGroup=0")
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
	return string(body), nil
}

// 提交作业
func (cache *XueXiTUserCache) SubmitWorkAnswerApi(question *XXTWorkQuestionSubmitEntity, tempSave bool) (string, error) {

	urlStr := "https://mooc1-api.chaoxing.com/mooc-ans/work/phone/doNormalHomeWorkSubmit?tempSave=" + fmt.Sprintf("%v", tempSave)
	method := "POST"

	values := url.Values{}

	values.Set("workExamUploadUrl", "")
	values.Set("workExamUploadCrcUrl", "")
	values.Set("workRelationAnswerId", question.AnswerId)
	values.Set("knowledgeid", "0")
	values.Set("enc", question.Enc)
	values.Set("source", question.Source)
	values.Set("encWork", question.EncWork)
	//values.Set("encWork", "ace56cb8a1c65f68339d3cd452757caa")

	// 下面这三个在原字符串中是【重复出现】的
	values.Add("courseId", question.CourseId)
	values.Add("workRelationId", question.WordId)
	values.Add("classId", question.ClassId)

	values.Add("courseId", question.CourseId)
	values.Add("workRelationId", question.WordId)
	values.Add("classId", question.ClassId)

	values.Set("workTimesEnc", "")

	//values.Set("questionId", "405139692")
	values.Set("questionId", question.QuestionId)
	values.Set("index", question.Index)
	values.Set("tempSave", fmt.Sprintf("%v", tempSave))

	if question.QuestionTypeStr == "单选题" {
		values.Set("type"+question.QuestionId, question.QuestionTypeCode)
		values.Set("score"+question.QuestionId, question.Score)
		answerStr := ""
		for _, answer := range question.Question.Answers {
			answerStr += qutils.SimilarityArraySelect(answer, question.Question.Options)
		}
		values.Set("answer"+question.QuestionId, answerStr)
	} else if question.QuestionTypeStr == "多选题" {
		values.Set("type"+question.QuestionId, question.QuestionTypeCode)
		values.Set("score"+question.QuestionId, question.Score)
		answerStr := ""
		for _, answer := range question.Question.Answers {
			answerStr += qutils.SimilarityArraySelect(answer, question.Question.Options)
		}
		values.Set("answers"+question.QuestionId, answerStr)
	} else if question.QuestionTypeStr == "判断题" {
		values.Set("type"+question.QuestionId, question.QuestionTypeCode)
		values.Set("score"+question.QuestionId, question.Score)
		answerStr := ""
		for _, answer := range question.Question.Answers {
			arraySelect := qutils.SimilarityArraySelect(answer, question.Question.Options)
			if arraySelect == "A" {
				answerStr = "true"
			} else {
				answerStr += "false"
			}
		}
		values.Set("answer"+question.QuestionId, answerStr)
	} else if question.QuestionTypeStr == "填空题" {
		values.Set("type"+question.QuestionId, question.QuestionTypeCode)
		values.Set("score"+question.QuestionId, question.Score)
		blankNum := ""
		for i, answer := range question.Question.Answers {
			values.Set("answer"+question.QuestionId+fmt.Sprintf("%d", i+1), answer)
			blankNum += fmt.Sprintf("%d,", i+1)
		}
		values.Set("blankNum"+question.QuestionId, blankNum)
	} else if question.QuestionTypeStr == "简答题" || question.QuestionTypeStr == "论述题" {
		//values.Set("isAccessibleCustomFid","0")
		values.Set("type"+question.QuestionId, question.QuestionTypeCode)
		values.Set("score"+question.QuestionId, question.Score)
		answerStr := ""
		if len(question.Question.Answers) > 0 {
			answerStr = question.Question.Answers[0]
		}
		values.Set("answer"+question.QuestionId, answerStr)
		//values.Set("editorValue",answerStr)
	}

	payload := strings.NewReader(values.Encode())

	//payload := strings.NewReader("workExamUploadUrl=&workExamUploadCrcUrl=&workRelationAnswerId=54657628&knowledgeid=0&enc=737ad94cd5529ffa3ba68606eb91a124&source=0&encWork=ace56cb8a1c65f68339d3cd452757caa&courseId=258101827&workRelationId=48731428&classId=134204187&workTimesEnc=&courseId=258101827&workRelationId=48731428&classId=134204187&answer405139692=A&type405139692=0&score405139692=100.0&questionId=405139692&index=0&tempSave=false")

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
	req, err := http.NewRequest(method, urlStr, payload)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("User-Agent", GetUA("mobile"))
	req.Header.Add("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Add("Origin", "https://mooc1-api.chaoxing.com")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Add("Accept-Language", "zh-CN,en-US;q=0.9")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
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

// 拉取试卷
func (cache *XueXiTUserCache) PullWorkQuestionApi(courseId, classId, workId, source, msgId, cpi, workAnswerId, enc, keyboardDisplayRequiresUserAction string, index /*第几道题*/, retry int, lastErr error) (string, error) {
	//url := "https://mooc1-api.chaoxing.com/mooc-ans/work/phone/doHomeWork?courseId=258101827&workId=48731428&classId=134204187&oldWorkId&cpi=411545273&mooc=1&msgId=0&source=0&checkIntegrity=true&enc=737ad94cd5529ffa3ba68606eb91a124&keyboardDisplayRequiresUserAction=1"

	//url := "https://mooc1-api.chaoxing.com/mooc-ans/work/phone/doHomeWork?courseId=258101827&workId=49156357&cpi=411545273&workAnswerId=54758238&classId=134204187&oldWorkId&mooc=1&msgId=0&source=0&checkIntegrity=true&enc=737ad94cd5529ffa3ba68606eb91a124&keyboardDisplayRequiresUserAction=1"
	urlStr := "https://mooc1-api.chaoxing.com/mooc-ans/work/phone/doHomeWork?courseId=" + courseId + "&workId=" + workId + "&cpi=" + cpi + "&workAnswerId=" + workAnswerId + "&classId=" + classId + "&mooc=1" + "&source=" + source + "&enc=" + enc + "&keyboardDisplayRequiresUserAction=" + keyboardDisplayRequiresUserAction + "&index=" + fmt.Sprintf("%d", index)
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
	req.Header.Add("User-Agent", GetUA("mobile"))
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	req.Header.Add("Upgrade-Insecure-Requests", "1")
	req.Header.Add("accept-language", "zh_CN")
	req.Header.Add("Referer", "https://mooc1-api.chaoxing.com/mooc-ans/work/phone/task-work?taskrefId=49156357&courseId=258101827&classId=134204187&userId=346635955&role=&source=0&enc_task=f17cf2658668d00b935b2c218fefcf56&cpi=411545273&vx=0&fromGroup=0")
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
	return string(body), nil
}
