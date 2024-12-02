package cqie

import (
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// UserDetailsApi 获取用户信息
func (cache *CqieUserCache) UserDetailsApi(retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	url := "https://study.cqie.edu.cn/gateway/system/sysUser/nowUserDetails"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		time.Sleep(time.Millisecond * 150)
		return cache.UserDetailsApi(retry-1, err)
	}
	req.Header.Add("Authorization", cache.access_token)
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "study.cqie.edu.cn")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Cookie", cache.cookie)

	res, err := client.Do(req)
	if err != nil {
		time.Sleep(time.Millisecond * 150)
		return cache.UserDetailsApi(retry-1, err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		time.Sleep(time.Millisecond * 150)
		return cache.UserDetailsApi(retry-1, err)
	}
	return string(body), nil
}

// PullCourseListApi 拉取课程列表
func (cache *CqieUserCache) PullCourseListApi(retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	url := "https://study.cqie.edu.cn/gateway/system/orgStudent/pagedMyCourse"
	method := "POST"

	payload := strings.NewReader(`{
   "filters": {
	 "courseNature": "",
	 "name": "",
	 "schoolYear": "",
	 "studentId": "` + cache.studentId + `",
	 "term": "",
	 "majorId": "` + cache.orgMajorId + `"
   },
   "pageIndex": 1,
   "pageSize": 200
 }`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		time.Sleep(time.Millisecond * 150)
		return cache.PullCourseListApi(retry-1, err)
	}
	req.Header.Add("Authorization", cache.access_token)
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "study.cqie.edu.cn")
	req.Header.Add("Connection", "keep-alive")

	res, err := client.Do(req)
	if err != nil {
		time.Sleep(time.Millisecond * 150)
		return cache.PullCourseListApi(retry-1, err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		time.Sleep(time.Millisecond * 150)
		return cache.PullCourseListApi(retry-1, err)
	}
	return string(body), nil
}

// GetVideoStudyIdApi 学习视屏前需要先调用此接口获取id才能学习
func (cache *CqieUserCache) GetVideoStudyIdApi(studentCourseId, videoId string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	url := "https://study.cqie.edu.cn/gateway/system/orgStudent/studyVideo?studentCourseId=" + studentCourseId + "&videoId=" + videoId + "&studentId=" + cache.studentId + "&majorId=" + cache.orgMajorId + "&version=ZSB_DSJ_24"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		time.Sleep(time.Millisecond * 150)
		return cache.GetVideoStudyIdApi(studentCourseId, videoId, retry-1, lastErr)
	}
	req.Header.Add("Authorization", cache.access_token)
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")

	res, err := client.Do(req)
	if err != nil {
		time.Sleep(time.Millisecond * 150)
		return cache.GetVideoStudyIdApi(studentCourseId, videoId, retry-1, lastErr)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		time.Sleep(time.Millisecond * 150)
		return cache.GetVideoStudyIdApi(studentCourseId, videoId, retry-1, lastErr)
	}
	return string(body), nil
}

// SubmitStudyTimeApi 提交学时
func (cache *CqieUserCache) SubmitStudyTimeApi(
	id string, //不知道是啥
	courseId string,
	studentCourseId string,
	unitId string,
	videoId string,
	studyTime time.Time, /**/
	coursewareId string,
	startPos int, /*开始点*/
	stopPos int, /*结束点*/
	maxPos int, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}

	url := "https://study.cqie.edu.cn/gateway/system/orgStudent/updateStudyVideoPlan"
	method := "POST"

	payload := strings.NewReader(`{
		"id": "` + id + `",
	   "orgId": "` + cache.orgId + `",
	   "deptId": "` + cache.deptId + `",
	   "majorId": "` + cache.orgMajorId + `",
	   "version": "` + "ZSB_DSJ_24" + `",
	   "courseId": "` + courseId + `",
	   "studentCourseId": "` + studentCourseId + `",
	   "unitId": "` + unitId + `",
	   "knowledgeId": null,
	   "videoId": "` + videoId + `",
	   "studentId": "` + cache.userId + `",
	   "studyTime": "` + studyTime.Format("2006-01-02 15:04:05") + `",
	   "startPos":` + strconv.Itoa(startPos) + `,
	   "stopPos": ` + strconv.Itoa(stopPos) + `,
	   "studyTime": "` + studyTime.Format("2006-01-02 15:04:05") + `",
	   "maxCurrentPos": ` + strconv.Itoa(maxPos) + `,
	   "coursewareId": "` + coursewareId + `"
	 }`)
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		time.Sleep(time.Millisecond * 150)
		return cache.SubmitStudyTimeApi(id, courseId, studentCourseId, unitId, videoId, studyTime, courseId, startPos, stopPos, maxPos, retry-1, err)
	}
	req.Header.Add("Authorization", cache.GetAccess_Token())
	req.Header.Add("Cookie", cache.GetCookie())
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		time.Sleep(time.Millisecond * 150)
		return cache.SubmitStudyTimeApi(id, courseId, studentCourseId, unitId, videoId, studyTime, courseId, startPos, stopPos, maxPos, retry-1, err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		time.Sleep(time.Millisecond * 150)
		return cache.SubmitStudyTimeApi(id, courseId, studentCourseId, unitId, videoId, studyTime, courseId, startPos, stopPos, maxPos, retry-1, err)
	}
	return string(body), nil
}

// 用于保存学习点时间
func (cache *CqieUserCache) SaveStudyTimeApi(courseId, studentCourseId, unitId, videoId, coursewareId string, startPos, stopPos int, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	url := "https://study.cqie.edu.cn/gateway/system/orgStudent/saveStudyVideoPlan"
	method := "POST"

	payload := strings.NewReader(`{
	"courseId": "` + courseId + `",
	"majorId": "` + cache.orgMajorId + `",
	"startPos": "` + strconv.Itoa(startPos) + `",
	"stopPos": "` + strconv.Itoa(stopPos) + `",
	"studentCourseId": "` + studentCourseId + `",
	"studentId": "` + cache.studentId + `",
	"unitId": "` + unitId + `",
	"videoId": "` + videoId + `",
	"coursewareId": "` + coursewareId + `",
	"version": "ZSB_DSJ_24"
}`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		time.Sleep(time.Millisecond * 150)
		return cache.SaveStudyTimeApi(courseId, studentCourseId, unitId, videoId, coursewareId, startPos, stopPos, retry-1, err)
	}
	req.Header.Add("Authorization", cache.access_token)
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		time.Sleep(time.Millisecond * 150)
		return cache.SaveStudyTimeApi(courseId, studentCourseId, unitId, videoId, coursewareId, startPos, stopPos, retry-1, err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		time.Sleep(time.Millisecond * 150)
		return cache.SaveStudyTimeApi(courseId, studentCourseId, unitId, videoId, coursewareId, startPos, stopPos, retry-1, err)
	}
	return string(body), nil
}

// PullCourseDetailApi 拉取对应课程详细信息，一般用于获取对应视屏列表的
func (cache *CqieUserCache) PullCourseDetailApi(courseId, studentCourseId string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	url := "https://study.cqie.edu.cn/gateway/system/orgStudent/myCourseDetails?studentId=" + cache.studentId + "&id=" + courseId + "&majorId=" + cache.orgMajorId + "&studentCourseId=" + studentCourseId + "&version=ZSB_DSJ_24"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		time.Sleep(time.Millisecond * 150)
		return cache.PullCourseDetailApi(courseId, studentCourseId, retry-1, lastErr)
	}
	req.Header.Add("Authorization", cache.access_token)
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")

	res, err := client.Do(req)
	if err != nil {
		time.Sleep(time.Millisecond * 150)
		return cache.PullCourseDetailApi(courseId, studentCourseId, retry-1, lastErr)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		time.Sleep(time.Millisecond * 150)
		return cache.PullCourseDetailApi(courseId, studentCourseId, retry-1, lastErr)
	}
	return string(body), nil
}

// PullProgressDetailApi 拉取进度
func (cache *CqieUserCache) PullProgressDetailApi(courseId, studentCourseId string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	url := "https://study.cqie.edu.cn/gateway/system/orgStudent/progressDetails?studentId=" + cache.studentId + "&id=" + courseId + "&majorId=" + cache.orgMajorId + "&studentCourseId=" + studentCourseId + "&version=ZSB_DSJ_24"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		time.Sleep(time.Millisecond * 150)
		return cache.PullProgressDetailApi(courseId, studentCourseId, retry-1, lastErr)
	}
	req.Header.Add("Authorization", cache.access_token)
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")

	res, err := client.Do(req)
	if err != nil {
		time.Sleep(time.Millisecond * 150)
		return cache.PullProgressDetailApi(courseId, studentCourseId, retry-1, lastErr)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		time.Sleep(time.Millisecond * 150)
		return cache.PullProgressDetailApi(courseId, studentCourseId, retry-1, lastErr)
	}
	return string(body), nil
}
