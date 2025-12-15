package xuexitong

import (
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/yatori-dev/yatori-go-core/que-core/qentity"
	"github.com/yatori-dev/yatori-go-core/que-core/qtype"
	"github.com/yatori-dev/yatori-go-core/utils/qutils"
)

// 学习通题目
type XXTExamQuestionSubmitEntity struct {
	CourseId           string
	TestPaperId        string
	TestUserRelationId string
	ClassId            string
	Type               string
	IsPhone            string
	Imei               string
	SubCount           string
	RemainTime         string
	TempSave           string
	TimeOver           string
	EncRemainTime      string
	EncLastUpdateTime  string
	Cpi                string
	Enc                string
	Source             string
	UserId             string
	Tid                string
	EnterPageTime      string
	AnsweredView       string
	ExitdTime          string
	PaperGroupId       string
	QuestionId         string //题目ID
	ExamRelationId     string
	AnswerId           string
	RemainTimeParam    string
	QType              qtype.QType //题目类型
	TypeName           string
	QuestionTypeCode   string
	QuestionTypeStr    string
	Question           qentity.Question
}

// PullExamListHtmlApi 拉取邮箱考试列表
func (cache *XueXiTUserCache) PullExamListHtmlApi(courseId string, classId string, cpi string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	urlStr := "https://mooc1-api.chaoxing.com/mooc-ans/exam/phone/task-list?courseId=" + courseId + "&classId=" + classId + "&cpi=" + cpi
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
	req, err := http.NewRequest("GET", urlStr, nil)

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

// PullExamEnterInformHtmlApi 拉取进入考试页面的html（这里面会有携带考试是否有滑块验证码等信息）
func (cache *XueXiTUserCache) PullExamEnterInformHtmlApi(
	taskrefId, msgId, courseId, userId, clazzId, enterType, encTask string,
	retry int, lastErr error,
) (string, string, error) {

	urlStr := "https://mooc1-api.chaoxing.com/exam-ans/android/mtaskmsgspecial?taskrefId=" +
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

// 拉取试卷（注意需要先过滑块验证码获取到验证码参数）
func (cache *XueXiTUserCache) PullExamPaperHtmlApi(courseId, classId, examId, source, examAnswerId, cpi, keyboardDisplayRequiresUserAction, imei, captchavalidate, jt string, retry int, lastErr error) (string, error) {
	//url := "https://mooc1-api.chaoxing.com/exam-ans/exam/phone/start?courseId=258101827&classId=134204187&examId=8186945&source=0&examAnswerId=167217517&cpi=411545273&keyboardDisplayRequiresUserAction=1&imei=76c82452584d47e39ab79aa54ea86554&faceDetectionResult&captchavalidate=validate_Ew0z9skxsLzVKQjmeObQiRVLxkxbPkRF_22DD053D736E6AC527CE57149BFE2534&jt=0&_v=0.3868294515418076&cxcid&cxtime&signt&_signcode=3&_signc=0&_signe=3-1&signk"
	urlStr := "https://mooc1-api.chaoxing.com/exam-ans/exam/phone/start?courseId=" + courseId + "&classId=" + classId + "&examId=" + examId + "&source=" + source + "&examAnswerId=" + examAnswerId + "&cpi=" + cpi + "&keyboardDisplayRequiresUserAction=" + keyboardDisplayRequiresUserAction + "&imei=" + imei + "&faceDetectionResult&captchavalidate=" + captchavalidate + "&jt=" + jt + "&_v=0.3868294515418076&cxcid&cxtime&signt&_signcode=3&_signc=0&_signe=3-1&signk"
	method := "GET"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
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
	//req.Header.Add("Referer", "https://mooc1-api.chaoxing.com/exam-ans/exam/phone/task-exam?taskrefId=8186945&courseId=258101827&classId=134204187&userId=346635955&role=&source=0&enc_task=e8a0e0f5b2faa978194ba2b19eef6371&cpi=411545273&vx=0")
	req.Header.Add("Accept-Language", "zh-CN,en-US;q=0.9")
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

// 拉取考试题目
func (cache *XueXiTUserCache) PullExamQuestionApi(courseId, classId, tId, id, cpi, remainTimeParam, enc, relationAnswerLastUpdateTime string, index int) (string, error) {

	//url := "https://mooc1-api.chaoxing.com/exam-ans/exam/test/reVersionTestStartNew?keyboardDisplayRequiresUserAction=1&courseId=258101827&classId=134204187&source=0&imei=76c82452584d47e39ab79aa54ea86554&tId=8201158&id=167239306&p=1&start=1&cpi=411545273&isphone=true&monitorStatus=0&monitorOp=-1&remainTimeParam=521941&relationAnswerLastUpdateTime=1764773341430&enc=40c8c154db29fb3ff6f01dfeade8a4fb"
	urlStr := "https://mooc1-api.chaoxing.com/exam-ans/exam/test/reVersionTestStartNew?keyboardDisplayRequiresUserAction=1&courseId=" + courseId + "&classId=" + classId + "&source=0&imei=" + IMEI + "&tId=" + tId + "&id=" + id + "&p=1&start=" + fmt.Sprintf("%d", index) + "&cpi=" + cpi + "&isphone=true&monitorStatus=0&monitorOp=-1&remainTimeParam=" + remainTimeParam + "&relationAnswerLastUpdateTime=" + relationAnswerLastUpdateTime + "&enc=" + enc
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, urlStr, nil)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("User-Agent", GetUA("mobile"))
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	req.Header.Add("Upgrade-Insecure-Requests", "1")
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

// 提交答题
func (cache *XueXiTUserCache) SubmitExamAnswerApi(question *XXTExamQuestionSubmitEntity, tempSave /*是否交卷*/ bool) (string, error) {

	//url := "https://mooc1-api.chaoxing.com/exam-ans/exam/test/reVersionSubmitTestNew?classId=134204187&courseId=258101827&testPaperId=8186945&testUserRelationId=167217517&cpi=411545273&version=1&tempSave=false&pos=90129888fd267cdf6604435c1b&rd=0.4715233554422915&value=%2528NaN%257CNaN%2529&qid=885532434&_edt=1764581639515265&_csign=1&_signcode=3&_signc=0&_signe=3-1&_signk&_cxcid&_cxtime&_signt"
	sig := GetExamSignature(question.UserId, question.QuestionId, rand.Intn(100)+900, rand.Intn(900)+100)
	urlStr := "https://mooc1-api.chaoxing.com/exam-ans/exam/test/reVersionSubmitTestNew?classId=" + question.ClassId + "&courseId=" + question.CourseId + "&testPaperId=" + question.TestPaperId + "&testUserRelationId=" + question.TestUserRelationId + "&cpi=" + question.Cpi + "&version=1&tempSave=" + fmt.Sprintf("%v", tempSave) + "&pos=" + sig["pos"].(string) + "&rd=" + fmt.Sprintf("%.16f", sig["rd"]) + "&value=" + url.QueryEscape(sig["value"].(string)) + "&qid=" + question.QuestionId + "&_edt=" + sig["_edt"].(string) + "&_csign=1&_signcode=3&_signc=0&_signe=3-1&_signk&_cxcid&_cxtime&_signt"

	method := "POST"

	values := url.Values{}

	values.Set("courseId", question.CourseId)
	values.Set("testPaperId", question.TestPaperId)
	values.Set("testUserRelationId", question.TestUserRelationId)
	values.Set("classId", question.ClassId)
	values.Set("type", "0")
	values.Set("isphone", "true")
	values.Set("imei", IMEI)
	values.Set("subCount", "")
	values.Set("remainTime", question.RemainTime)
	values.Set("tempSave", strconv.FormatBool(tempSave))
	values.Set("timeOver", "false")
	values.Set("encRemainTime", question.EncRemainTime)
	values.Set("encLastUpdateTime", question.EncLastUpdateTime)
	values.Set("enc", question.Enc)
	values.Set("userId", question.UserId)
	values.Set("start", "0")
	values.Set("enterPageTime", question.EnterPageTime)
	values.Set("randomOptions", "false")

	values.Set("score"+question.QuestionId, question.Source)
	values.Add("questionId", question.QuestionId) // 这个字段你原来是重复两次
	values.Add("questionId", question.QuestionId)

	values.Set("monitorforcesubmit", "0")
	values.Set("answeredView", "0")
	values.Set("exitdtime", "0")
	values.Set("paperGroupId", "0")

	if question.QuestionTypeStr == "单选题" || question.QuestionTypeStr == "多选题" {
		values.Set("type"+question.QuestionId, question.QuestionTypeCode)
		values.Set("typeName"+question.QuestionId, question.QuestionTypeStr)
		values.Set("hidetext", "")
		answerStr := ""
		for _, answer := range question.Question.Answers {
			answerStr += qutils.SimilarityArraySelect(answer, question.Question.Options)
		}
		values.Set("answer"+question.QuestionId, answerStr)
	} else if question.QuestionTypeStr == "判断题" {
		values.Set("type"+question.QuestionId, question.QuestionTypeCode)
		values.Set("typeName"+question.QuestionId, question.QuestionTypeStr)
		values.Set("hidetext", "")
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
		values.Set("typeName"+question.QuestionId, question.QuestionTypeStr)
		values.Set("hidetext", "")
		blankNum := ""
		for i, answer := range question.Question.Answers {
			values.Set("answerEditor"+question.QuestionId+fmt.Sprintf("%d", i+1), answer)
			blankNum += fmt.Sprintf("%d,", i+1)
		}
		values.Set("blankNum"+question.QuestionId, blankNum)
	} else if question.QuestionTypeStr == "简答题" || question.QuestionTypeStr == "论述题" {
		values.Set("type"+question.QuestionId, question.QuestionTypeCode)
		values.Set("typeName"+question.QuestionId, question.QuestionTypeStr)
		values.Set("hidetext", "")
		answerStr := ""
		if len(question.Question.Answers) > 0 {
			answerStr = question.Question.Answers[0]
		}
		values.Set("answer"+question.QuestionId, answerStr)
	}

	payload := strings.NewReader(values.Encode())

	client := &http.Client{}
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
	req.Header.Add("Referer", "https://mooc1-api.chaoxing.com/exam-ans/exam/test/reVersionTestStartNew?keyboardDisplayRequiresUserAction=1&courseId="+question.CourseId+"&classId="+question.ClassId+"&source=0&imei="+IMEI+"&tId="+question.Tid+"&id="+question.AnswerId+"&p=1&start=1&cpi="+question.Cpi+"&isphone=true&monitorStatus=0&monitorOp=-1&remainTimeParam="+question.RemainTimeParam+"&relationAnswerLastUpdateTime="+fmt.Sprintf("%d", time.Now().UnixMilli())+"&enc="+question.Enc)
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

func init() {
	rand.Seed(time.Now().UnixNano())
}

// 等价于 Python secrets.token_hex(n)
func tokenHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// 等价于 get_ts()
func getTs() string {
	return strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
}

func GetExamSignature(uid string, qid string, x int, y int) map[string]interface{} {
	ts := getTs()

	r1 := rand.Intn(9)
	r2 := rand.Intn(9)

	a := fmt.Sprintf("%s%s%d%d",
		tokenHex(16),
		ts[4:],
		r1,
		r2,
	)
	if qid != "" {
		a += qid
	}

	var temp int64 = 0
	for _, ch := range a {
		temp = (temp << 5) - temp + int64(ch)
	}

	salt := fmt.Sprintf("%d%d%d",
		r1,
		r2,
		(int64(0x7fffffff)&temp)%10,
	)

	encVal := uid
	if qid != "" {
		encVal += "_" + qid
	}
	encVal += "|" + salt

	var sb strings.Builder
	for _, c := range encVal {
		sb.WriteString(strconv.Itoa(int(c)))
	}
	encVal2 := sb.String()

	b := len(encVal2) / 5

	cStr := string(encVal2[b]) +
		string(encVal2[2*b]) +
		string(encVal2[3*b]) +
		string(encVal2[4*b])
	c, _ := strconv.Atoi(cStr)

	d := len(encVal)/2 + 1

	first10, _ := strconv.Atoi(encVal2[:10])
	e := (int64(c)*int64(first10) + int64(d)) % 0x7FFFFFFF

	pos := fmt.Sprintf("(%d|%d)", x, y)

	var result strings.Builder
	for _, ch := range pos {
		key := int(math.Floor(float64(e) / float64(0x7FFFFFFF) * 0xFF))
		v := int(ch) ^ key
		result.WriteString(fmt.Sprintf("%02x", v))
		e = (int64(c)*e + int64(d)) % 0x7FFFFFFF
	}

	return map[string]interface{}{
		"pos":   result.String() + tokenHex(4),
		"rd":    rand.Float64(),
		"value": pos,
		"_edt":  ts + salt,
	}
}
