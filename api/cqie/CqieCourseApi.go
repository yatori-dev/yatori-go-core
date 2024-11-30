package cqie

import (
	"fmt"
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

// PullCourseList 拉取课程列表
func (cache *CqieUserCache) PullCourseList(retry int, lastErr error) (string, error) {
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
		return cache.PullCourseList(retry-1, err)
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
		return cache.PullCourseList(retry-1, err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		time.Sleep(time.Millisecond * 150)
		return cache.PullCourseList(retry-1, err)
	}
	return string(body), nil
}

// SubmitStudyTimeApi 提交学时
func (cache *CqieUserCache) SubmitStudyTimeApi(
	id string,
	courseId string,
	studentCourseId string,
	unitId string,
	videoId string,
	studentId string, /*学生ID*/
	studyTime time.Time, /**/
	coursewareId string,
	startPos int, /*开始点*/
	stopPos int, /*结束点*/
	maxPos int) {

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
	// 	payload := strings.NewReader(`{
	//    "id": "ce70b5cd1765f698788c84d0eec5a95a",
	//   "orgId": "1",
	//   "deptId": "f8955096610de9b61f0e89762fe44448",
	//   "majorId": "966c92d8abcec472a717ca5af8c24a49",
	//   "version": "ZSB_DSJ_24",
	//   "courseId": "1cbcd0b22030fdf75117cb02d2c766ec",
	//   "studentCourseId": "e256c4c26b3e0dc27901029755396bdd",
	//   "unitId": "84ed9e4674d811bc10a9f55390231f5f",
	//   "knowledgeId": null,
	//   "videoId": "65da2135311dc5cc461d53c6c827cc45",
	//   "studentId": "1e49b8b4ea23d07eb0724d0c927ed2bc",
	//   "studyTime": "2024-11-27 19:49:24",
	//   "startPos":` + strconv.Itoa(startPos) + `,
	//   "stopPos": ` + strconv.Itoa(stopPos) + `,
	//   "studyTime": "2024-11-27 19:49:24",
	//   "maxCurrentPos": ` + strconv.Itoa(maxPos) + `,
	//   "coursewareId": "558bf1e99f2727d493dbe2ed05983115"
	// }`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Authorization", cache.GetAccess_Token())
	req.Header.Add("Cookie", cache.GetCookie())
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}
