package haiqikeji

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ============================================================================
// 海旗科技（HQKJ）作业 / 考试相关接口
//
// 原先这些接口实现在 yatori-go-console 的 logic/haiqikeji 里，现按框架分层收敛到核心库：
//   work_list / exam_list（HqkjApi.go 已有）
//     └→ work_detail / exam_detail     拉详情，取 paperId / classId / recordId
//     └→ work_start / exam_start       开考，拉题面 + 选项
//     └→ work_record_consult_list / exam_record_consult_list  查看已答记录（老式作业回退拉题）
//     └→ work_answer_add / exam_answer_add                    逐题提交
//     └→ exam_answer_finish                                   交卷（state=2）
//
// 所有接口统一复用 cache 的会话信息（Token / SchoolId / UserId / PreUrl / IP 代理）。
// ============================================================================

// hqkjRequest 统一的海旗请求执行器：自动携带会话 Token，并复用账号的 IP 代理设置。
func (cache *HqkjUserCache) hqkjRequest(method, reqUrl, contentType, payload string) (string, error) {
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	if cache.IpProxySW { // 复用账号的 IP 代理设置
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse("http://" + cache.ProxyIP)
		}
	}
	client := &http.Client{Timeout: 30 * time.Second, Transport: tr}
	var body io.Reader
	if payload != "" {
		body = strings.NewReader(payload)
	}
	req, err := http.NewRequest(method, reqUrl, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("authorization", cache.Token)
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36")
	if contentType != "" {
		req.Header.Set("content-type", contentType)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(bodyText), nil
}

// hqkjRequestWithRetry 带重试的请求封装，保持与核心库其他接口一致的 retry/lastErr 语义。
func (cache *HqkjUserCache) hqkjRequestWithRetry(method, reqUrl, contentType, payload string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	var result string
	var err error
	for i := 0; i <= retry; i++ {
		result, err = cache.hqkjRequest(method, reqUrl, contentType, payload)
		if err == nil {
			return result, nil
		}
	}
	return "", err
}

// PullWorkDetailApi 拉取作业详情（yee_course_student_work_detail），返回 data.workDetail[]。
func (cache *HqkjUserCache) PullWorkDetailApi(workId, courseId, title string, retry int, lastErr error) (string, error) {
	reqUrl := cache.PreUrl + "/api/user/yee_course_student_work_detail?schoolId=" + cache.SchoolId +
		"&studentId=" + cache.UserId + "&workId=" + workId + "&courseId=" + courseId +
		"&title=" + url.QueryEscape(title)
	return cache.hqkjRequestWithRetry("GET", reqUrl, "", "", retry, lastErr)
}

// PullExamDetailApi 拉取考试详情（yee_course_student_exam_detail），返回 data.examDetail[]。
func (cache *HqkjUserCache) PullExamDetailApi(examId, courseId, title string, retry int, lastErr error) (string, error) {
	reqUrl := cache.PreUrl + "/api/user/yee_course_student_exam_detail?schoolId=" + cache.SchoolId +
		"&studentId=" + cache.UserId + "&examId=" + examId + "&courseId=" + courseId +
		"&title=" + url.QueryEscape(title)
	return cache.hqkjRequestWithRetry("GET", reqUrl, "", "", retry, lastErr)
}

// PullWorkStartApi 开始作业（yee_course_student_work_start），返回题面 data.workTopics[] 与整卷 recordId。
// paperId=0 的老式作业也需显式传 "0"（传空会被平台拒绝“缺少必要参数: paperId”）。
func (cache *HqkjUserCache) PullWorkStartApi(workId, courseId, paperId, classId string, retry int, lastErr error) (string, error) {
	reqUrl := cache.PreUrl + "/api/user/yee_course_student_work_start?schoolId=" + cache.SchoolId +
		"&studentId=" + cache.UserId + "&courseId=" + courseId + "&workId=" + workId +
		"&platform=pc&createUserId=" + cache.UserId + "&state=1&classId=" + classId + "&paperId=" + paperId +
		"&random=&randData=%257B%257D&randNumber="
	return cache.hqkjRequestWithRetry("GET", reqUrl, "", "", retry, lastErr)
}

// PullExamStartApi 开始考试（yee_course_student_exam_start），返回题面 data.examTopics[] 与整卷 recordId。
func (cache *HqkjUserCache) PullExamStartApi(examId, courseId, paperId, classId string, retry int, lastErr error) (string, error) {
	reqUrl := cache.PreUrl + "/api/user/yee_course_student_exam_start?schoolId=" + cache.SchoolId +
		"&studentId=" + cache.UserId + "&courseId=" + courseId + "&examId=" + examId +
		"&platform=pc&createUserId=" + cache.UserId + "&state=1&classId=" + classId + "&paperId=" + paperId +
		"&random=&randData=%257B%257D&randNumber="
	return cache.hqkjRequestWithRetry("GET", reqUrl, "", "", retry, lastErr)
}

// PullWorkConsultApi 查看作业作答记录（yee_work_record_consult_list），返回题目在 data.workResult[] 下。
// 参数 recordId 传 work_detail 里的记录 id（老式 paperId=0 作业用它回退拉题）。
func (cache *HqkjUserCache) PullWorkConsultApi(recordId, courseId string, retry int, lastErr error) (string, error) {
	reqUrl := cache.PreUrl + "/api/user/yee_work_record_consult_list?schoolId=" + cache.SchoolId +
		"&userId=" + cache.UserId + "&workId=" + recordId + "&courseId=" + courseId
	return cache.hqkjRequestWithRetry("GET", reqUrl, "", "", retry, lastErr)
}

// PullExamConsultApi 查看考试作答记录（yee_exam_record_consult_list），返回题目在 data.workResult[] 下。
func (cache *HqkjUserCache) PullExamConsultApi(recordId, courseId string, retry int, lastErr error) (string, error) {
	reqUrl := cache.PreUrl + "/api/user/yee_exam_record_consult_list?schoolId=" + cache.SchoolId +
		"&userId=" + cache.UserId + "&examId=" + recordId + "&courseId=" + courseId
	return cache.hqkjRequestWithRetry("GET", reqUrl, "", "", retry, lastErr)
}

// SubmitWorkAnswerApi 提交作业答案（yee_work_answer_add）。
// payload 与真实抓包一致：8 字段、不带 wrId/waId。
func (cache *HqkjUserCache) SubmitWorkAnswerApi(courseId, workId, topicId, recordId, questionType string, answers []string, retry int, lastErr error) (string, error) {
	if recordId == "" { // 兜底，避免拼出非法 JSON（正常 recordId 来自 work_start 响应的 recordId）
		recordId = "0"
	}
	answersData, err := json.Marshal(answers)
	if err != nil {
		return "", err
	}
	payload := `{"schoolId":` + cache.SchoolId + `,"courseId":` + courseId + `,"userId":` + cache.UserId +
		`,"workId":` + workId + `,"topicId":` + topicId + `,"answer":` + string(answersData) +
		`,"recordId":` + recordId + `,"type":` + questionType + `}`
	return cache.hqkjRequestWithRetry("POST", cache.PreUrl+"/api/user/yee_work_answer_add", "application/json", payload, retry, lastErr)
}

// SubmitExamAnswerApi 提交考试答案（yee_exam_answer_add）。
// 字段与真实抓包一致：{schoolId,courseId,userId,examId,topicId,answer,recordId,type}，不带 wrId/waId。
func (cache *HqkjUserCache) SubmitExamAnswerApi(courseId, examId, topicId, recordId, questionType string, answers []string, retry int, lastErr error) (string, error) {
	if recordId == "" { // 兜底，避免拼出非法 JSON（正常 recordId 来自 exam_start 响应的 examTopics[].recordId）
		recordId = "0"
	}
	answersData, err := json.Marshal(answers)
	if err != nil {
		return "", err
	}
	payload := `{"schoolId":` + cache.SchoolId + `,"courseId":` + courseId + `,"userId":` + cache.UserId +
		`,"examId":` + examId + `,"topicId":` + topicId + `,"answer":` + string(answersData) +
		`,"recordId":` + recordId + `,"type":` + questionType + `}`
	return cache.hqkjRequestWithRetry("POST", cache.PreUrl+"/api/user/yee_exam_answer_add", "application/json", payload, retry, lastErr)
}

// FinishExamAnswerApi 交卷（yee_exam_answer_finish, state=2）。逐题 yee_exam_answer_add 仅暂存，
// 此接口才正式交卷出成绩。参数按真实抓包：workId=examId、studentId=userId=cache.UserId。
func (cache *HqkjUserCache) FinishExamAnswerApi(courseId, examId, recordId string, retry int, lastErr error) (string, error) {
	if recordId == "" { // 兜底，避免拼出非法 URL
		recordId = "0"
	}
	reqUrl := cache.PreUrl + "/api/user/yee_exam_answer_finish?schoolId=" + cache.SchoolId +
		"&studentId=" + cache.UserId + "&userId=" + cache.UserId + "&workId=" + examId +
		"&examId=" + examId + "&state=2&recordId=" + recordId + "&courseId=" + courseId
	return cache.hqkjRequestWithRetry("GET", reqUrl, "", "", retry, lastErr)
}
