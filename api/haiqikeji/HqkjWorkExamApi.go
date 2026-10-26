package haiqikeji

import (
	"crypto/tls"
	"encoding/json"
	"errors"
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
//     └→ work_answer_add / work_answer_addText / work_answer_addBlank    逐题暂存（作业）
//     └→ exam_answer_add / exam_answer_addText / exam_answer_addBlank    逐题暂存（考试）
//     └→ work_answer_finish / exam_answer_finish                         提交作业 / 交卷（state=2）
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

// requireAnswerRecordId recordId 是学生作答记录的唯一定位。
// 传 0 时服务端仍可能返回 code=200，但答案会写进无效记录，网页上看不到——
// 因此这里直接拒绝，避免“控制台提示提交成功，网站却没有答案”。
func requireAnswerRecordId(recordId string) (string, error) {
	recordId = strings.TrimSpace(recordId)
	if recordId == "" || recordId == "0" || recordId == "null" {
		return "", errors.New("recordId为空或为0，已拒绝提交（答案无法写入学生作答记录）")
	}
	return recordId, nil
}

// SubmitWorkAnswerApi 提交作业客观题答案（yee_work_answer_add，type=1/2/3）。
// payload 与真实抓包一致：8 字段、不带 wrId/waId。
func (cache *HqkjUserCache) SubmitWorkAnswerApi(courseId, workId, topicId, recordId, questionType string, answers []string, retry int, lastErr error) (string, error) {
	recordId, err := requireAnswerRecordId(recordId)
	if err != nil {
		return "", err
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

// SubmitWorkAnswerTextApi 提交作业简答题答案（yee_work_answer_addText，type=4）。
// 简答题的 answer 是字符串而不是数组；images/files 为附件数组，无附件时传空数组。
func (cache *HqkjUserCache) SubmitWorkAnswerTextApi(courseId, workId, topicId, recordId, questionType, text string, retry int, lastErr error) (string, error) {
	recordId, err := requireAnswerRecordId(recordId)
	if err != nil {
		return "", err
	}
	answerData, err := json.Marshal(text)
	if err != nil {
		return "", err
	}
	payload := `{"schoolId":` + cache.SchoolId + `,"courseId":` + courseId + `,"userId":` + cache.UserId +
		`,"workId":` + workId + `,"topicId":` + topicId + `,"answer":` + string(answerData) +
		`,"recordId":` + recordId + `,"type":` + questionType + `,"images":[],"files":[]}`
	return cache.hqkjRequestWithRetry("POST", cache.PreUrl+"/api/user/yee_work_answer_addText", "application/json", payload, retry, lastErr)
}

// SubmitWorkAnswerBlankApi 提交作业填空题答案（yee_work_answer_addBlank，type=5）。
// 填空题的 answer 是 {"a1":"...","a2":"..."} 形式的对象，按空位顺序编号。
func (cache *HqkjUserCache) SubmitWorkAnswerBlankApi(courseId, workId, topicId, recordId, questionType string, answers map[string]string, retry int, lastErr error) (string, error) {
	recordId, err := requireAnswerRecordId(recordId)
	if err != nil {
		return "", err
	}
	answerData, err := json.Marshal(answers)
	if err != nil {
		return "", err
	}
	payload := `{"schoolId":` + cache.SchoolId + `,"courseId":` + courseId + `,"userId":` + cache.UserId +
		`,"workId":` + workId + `,"topicId":` + topicId + `,"answer":` + string(answerData) +
		`,"recordId":` + recordId + `,"type":` + questionType + `}`
	return cache.hqkjRequestWithRetry("POST", cache.PreUrl+"/api/user/yee_work_answer_addBlank", "application/json", payload, retry, lastErr)
}

// SubmitExamAnswerApi 提交考试客观题答案（yee_exam_answer_add，type=1/2/3）。
// 字段与真实抓包一致：{schoolId,courseId,userId,examId,topicId,answer,recordId,type}，不带 wrId/waId。
func (cache *HqkjUserCache) SubmitExamAnswerApi(courseId, examId, topicId, recordId, questionType string, answers []string, retry int, lastErr error) (string, error) {
	recordId, err := requireAnswerRecordId(recordId)
	if err != nil {
		return "", err
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

// SubmitExamAnswerTextApi 提交考试简答题答案（yee_exam_answer_addText，type=4）。
func (cache *HqkjUserCache) SubmitExamAnswerTextApi(courseId, examId, topicId, recordId, questionType, text string, retry int, lastErr error) (string, error) {
	recordId, err := requireAnswerRecordId(recordId)
	if err != nil {
		return "", err
	}
	answerData, err := json.Marshal(text)
	if err != nil {
		return "", err
	}
	payload := `{"schoolId":` + cache.SchoolId + `,"courseId":` + courseId + `,"userId":` + cache.UserId +
		`,"examId":` + examId + `,"topicId":` + topicId + `,"answer":` + string(answerData) +
		`,"recordId":` + recordId + `,"type":` + questionType + `,"images":[],"files":[]}`
	return cache.hqkjRequestWithRetry("POST", cache.PreUrl+"/api/user/yee_exam_answer_addText", "application/json", payload, retry, lastErr)
}

// SubmitExamAnswerBlankApi 提交考试填空题答案（yee_exam_answer_addBlank，type=5）。
func (cache *HqkjUserCache) SubmitExamAnswerBlankApi(courseId, examId, topicId, recordId, questionType string, answers map[string]string, retry int, lastErr error) (string, error) {
	recordId, err := requireAnswerRecordId(recordId)
	if err != nil {
		return "", err
	}
	answerData, err := json.Marshal(answers)
	if err != nil {
		return "", err
	}
	payload := `{"schoolId":` + cache.SchoolId + `,"courseId":` + courseId + `,"userId":` + cache.UserId +
		`,"examId":` + examId + `,"topicId":` + topicId + `,"answer":` + string(answerData) +
		`,"recordId":` + recordId + `,"type":` + questionType + `}`
	return cache.hqkjRequestWithRetry("POST", cache.PreUrl+"/api/user/yee_exam_answer_addBlank", "application/json", payload, retry, lastErr)
}

// FinishWorkAnswerApi 提交作业（yee_work_answer_finish, state=2）。
// 逐题 yee_work_answer_add 只是暂存，调用此接口后作业才变成“已提交”。
// 参数与网页端一致：schoolId/studentId/userId/workId/state=2/recordId/courseId。
func (cache *HqkjUserCache) FinishWorkAnswerApi(courseId, workId, recordId string, retry int, lastErr error) (string, error) {
	recordId, err := requireAnswerRecordId(recordId)
	if err != nil {
		return "", err
	}
	reqUrl := cache.PreUrl + "/api/user/yee_work_answer_finish?schoolId=" + cache.SchoolId +
		"&studentId=" + cache.UserId + "&userId=" + cache.UserId + "&workId=" + workId +
		"&state=2&recordId=" + recordId + "&courseId=" + courseId
	return cache.hqkjRequestWithRetry("GET", reqUrl, "", "", retry, lastErr)
}

// FinishExamAnswerApi 交卷（yee_exam_answer_finish, state=2）。逐题 yee_exam_answer_add 仅暂存，
// 此接口才正式交卷出成绩。参数按真实抓包：workId=examId、studentId=userId=cache.UserId。
func (cache *HqkjUserCache) FinishExamAnswerApi(courseId, examId, recordId string, retry int, lastErr error) (string, error) {
	recordId, err := requireAnswerRecordId(recordId)
	if err != nil {
		return "", err
	}
	reqUrl := cache.PreUrl + "/api/user/yee_exam_answer_finish?schoolId=" + cache.SchoolId +
		"&studentId=" + cache.UserId + "&userId=" + cache.UserId + "&workId=" + examId +
		"&examId=" + examId + "&state=2&recordId=" + recordId + "&courseId=" + courseId
	return cache.hqkjRequestWithRetry("GET", reqUrl, "", "", retry, lastErr)
}
