package xuexitong

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

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
	url := "https://mooc1-api.chaoxing.com/exam-ans/exam/phone/start?courseId=" + courseId + "&classId=" + classId + "&examId=" + examId + "&source=" + source + "&examAnswerId=" + examAnswerId + "&cpi=" + cpi + "&keyboardDisplayRequiresUserAction=" + keyboardDisplayRequiresUserAction + "&imei=" + imei + "&faceDetectionResult&captchavalidate=" + captchavalidate + "&jt=" + jt + "&_v=0.3868294515418076&cxcid&cxtime&signt&_signcode=3&_signc=0&_signe=3-1&signk"
	method := "GET"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}
	client := &http.Client{
		Transport: tr,
	}
	req, err := http.NewRequest(method, url, nil)

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
func (cache *XueXiTUserCache) PullExamQuestionApi(courseId, classId, tId, id, cpi, remainTimeParam, enc string) (string, error) {

	//url := "https://mooc1-api.chaoxing.com/exam-ans/exam/test/reVersionTestStartNew?keyboardDisplayRequiresUserAction=1&courseId=258101827&classId=134204187&source=0&imei=76c82452584d47e39ab79aa54ea86554&tId=8201158&id=167239306&p=1&start=1&cpi=411545273&isphone=true&monitorStatus=0&monitorOp=-1&remainTimeParam=521941&relationAnswerLastUpdateTime=1764773341430&enc=40c8c154db29fb3ff6f01dfeade8a4fb"
	url := "https://mooc1-api.chaoxing.com/exam-ans/exam/test/reVersionTestStartNew?keyboardDisplayRequiresUserAction=1&courseId=" + courseId + "&classId=" + classId + "&source=0&imei=" + IMEI + "&tId=" + tId + "&id=" + id + "&p=1&start=1&cpi=" + cpi + "&isphone=true&monitorStatus=0&monitorOp=-1&remainTimeParam=" + remainTimeParam + "&relationAnswerLastUpdateTime=" + fmt.Sprintf("%d", time.Now().UnixMilli()) + "&enc=" + enc
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
func (cache *XueXiTUserCache) SubmitExamAnswerApi(classId, courseId, testPaperId, testUserRelationId, cpi, tempSave, pos, value, qid, edt string) (string, error) {

	//url := "https://mooc1-api.chaoxing.com/exam-ans/exam/test/reVersionSubmitTestNew?classId=134204187&courseId=258101827&testPaperId=8186945&testUserRelationId=167217517&cpi=411545273&version=1&tempSave=false&pos=90129888fd267cdf6604435c1b&rd=0.4715233554422915&value=%2528NaN%257CNaN%2529&qid=885532434&_edt=1764581639515265&_csign=1&_signcode=3&_signc=0&_signe=3-1&_signk&_cxcid&_cxtime&_signt"
	url := "https://mooc1-api.chaoxing.com/exam-ans/exam/test/reVersionSubmitTestNew?classId=" + classId + "&courseId=" + courseId + "&testPaperId=" + testPaperId + "&testUserRelationId=" + testUserRelationId + "&cpi=" + cpi + "&version=1&tempSave=" + tempSave + "&pos=" + pos + "&rd=0.4715233554422915&value=" + value + "&qid=" + qid + "&_edt=" + edt + "&_csign=1&_signcode=3&_signc=0&_signe=3-1&_signk&_cxcid&_cxtime&_signt"
	method := "POST"

	payload := strings.NewReader("courseId=" + courseId + "&testPaperId=" + testPaperId + "&testUserRelationId=" + testUserRelationId + "&classId=" + classId + "&type=0&isphone=true&imei=" + IMEI + "&subCount=&remainTime=3586&tempSave=false&timeOver=false&encRemainTime=3599&encLastUpdateTime=1764581621536&enc=fb34089d61c53db4caba284f366df017&userId=346635955&score885532434=5.0&questionId=885532434&questionId=885532434&start=0&enterPageTime=1764581621536&monitorforcesubmit=0&answeredView=0&exitdtime=0&paperGroupId=0&type885532434=0&typeName885532434=%E5%8D%95%E9%80%89%E9%A2%98&hidetext=&answer885532434=B")

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

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
