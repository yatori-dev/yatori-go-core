package zxcpks

import (
	"fmt"
	"strconv"
	"time"

	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/api/zxcpks"
)

type ZxcpksCourse struct {
	Id            string      `json:"id"`
	Name          string      `json:"name"`
	Mode          int         `json:"mode"`
	CollegeId     string      `json:"collegeId"`
	CategoryId    interface{} `json:"categoryId"`
	Lecturers     string      `json:"lecturers"`
	StartDate     string      `json:"startDate"`
	EndDate       string      `json:"endDate"`
	Cover         string      `json:"cover"`
	Content       interface{} `json:"content"`
	Credit        float64     `json:"credit"`
	Allow         int         `json:"allow"`
	Intro         string      `json:"intro"`
	TeacherIntro  interface{} `json:"teacherIntro"`
	Code          string      `json:"code"`
	StuCount      int         `json:"stuCount"`
	Proclamation  interface{} `json:"proclamation"`
	ClusterId     int         `json:"clusterId"`
	PeriodName    string      `json:"periodName"`
	AddTime       string      `json:"addTime"`
	CreateId      int         `json:"createId"`
	SchoolId      int         `json:"schoolId"`
	CateBid       int         `json:"cateBid"`
	CateMid       int         `json:"cateMid"`
	SignStartTime interface{} `json:"signStartTime"`
	SignEndTime   interface{} `json:"signEndTime"`
	SignScope     int         `json:"signScope"`
	SignClass     string      `json:"signClass"`
	LecturerName  string      `json:"lecturerName"`
	Offline       int         `json:"offline"`
	Mission       int         `json:"mission"`
	SignLimit     int         `json:"signLimit"`
	LineLock      int         `json:"lineLock"`
	AddDate       string      `json:"addDate"`
	TplId         int         `json:"tplId"`
	TemplateId    int         `json:"templateId"`
}

type ZxcpksNode struct {
	Id            string      `json:"id"`
	Name          string      `json:"name"`
	Type          interface{} `json:"type"`
	ChapterId     string      `json:"chapterId"`
	CourseId      string      `json:"courseId"`
	VideoFile     interface{} `json:"videoFile"`
	VideoDuration int         `json:"videoDuration"`
	VotingPath    interface{} `json:"votingPath"`
	TabVideo      int         `json:"tabVideo"`
	TabFile       int         `json:"tabFile"`
	TabVote       int         `json:"tabVote"`
	TabWork       int         `json:"tabWork"`
	TabExam       int         `json:"tabExam"`
	Sort          int         `json:"sort"`
	VideoMode     int         `json:"videoMode"`
	LocalFile     string      `json:"localFile"`
	SchoolId      int         `json:"schoolId"`
	Lock          int         `json:"lock"`
	UnlockTime    int         `json:"unlockTime"`
}

// 拉取课程列表
func ZxcpksCourseListAction(cache *zxcpks.ZxcpksUserCache) ([]ZxcpksCourse, error) {
	courseList := make([]ZxcpksCourse, 0)
	coursesResult, err := cache.PullCourseListApi()
	if err != nil {
		return nil, err
	}
	if cslist, ok := gojsonq.New().JSONString(coursesResult).Find("data").([]any); ok {
		for _, course := range cslist {
			if cs, ok := course.(map[string]any); ok {
				zxcpksCourse := ZxcpksCourse{
					Id:           strconv.Itoa(int(cs["id"].(float64))),
					Name:         cs["name"].(string),
					Intro:        cs["intro"].(string),
					CollegeId:    strconv.Itoa(int(cs["collegeId"].(float64))),
					PeriodName:   cs["periodName"].(string),
					LecturerName: cs["lecturerName"].(string),
				}
				courseList = append(courseList, zxcpksCourse)
			}
		}
	}

	return courseList, nil
}

// 节点列表
func ZxcpksNodeListAction(cache *zxcpks.ZxcpksUserCache, course ZxcpksCourse) ([]ZxcpksNode, error) {
	nodeList := make([]ZxcpksNode, 0)
	chapterResult, err := cache.PullChapterListApi(course.Id)
	if err != nil {
		return nil, err
	}
	if cslist, ok := gojsonq.New().JSONString(chapterResult).Find("data").([]any); ok {
		for _, chapter := range cslist {
			if cp, ok := chapter.(map[string]any); ok {
				chapterNodeResult, err := cache.PullChapterNodeListApi(strconv.Itoa(int(cp["id"].(float64))))
				if err != nil {
					return nil, err
				}
				if ndlist, ok := gojsonq.New().JSONString(chapterNodeResult).Find("data").([]any); ok {
					for _, node := range ndlist {
						if nd, ok := node.(map[string]any); ok {
							zxcpksNode := ZxcpksNode{
								Id:            strconv.Itoa(int(nd["id"].(float64))),
								Name:          nd["name"].(string),
								ChapterId:     strconv.Itoa(int(nd["chapterId"].(float64))),
								CourseId:      strconv.Itoa(int(nd["courseId"].(float64))),
								VideoDuration: int(nd["videoDuration"].(float64)),
							}
							nodeList = append(nodeList, zxcpksNode)
						}
					}
				}

			}
		}
	}

	return nodeList, nil
}

// 提交学时，秒刷
func ZxcpksSubmitFastSutdyTimeAction(cache *zxcpks.ZxcpksUserCache, node ZxcpksNode) (string, error) {
	startResult, err := cache.StartStudyApi(node.Id, node.CourseId)
	if err != nil {
		return "", err
	}

	//拉取当前适配的观看进度
	nowProgressResult, err := cache.PullLastProgressApi(node.Id)
	nowProgress := gojsonq.New().JSONString(nowProgressResult).Find("data").(string)
	fmt.Println("当前视频进度：", nowProgress)
	if err != nil {
		return "", err
	}
	sessionId := gojsonq.New().JSONString(startResult).Find("data").(string)
	time.Sleep(30 * time.Second) //先隔30s
	submitResult, err := cache.SubmitStudyTimeApi(sessionId, 100)
	if err != nil {
		return "", err
	}
	code := int(gojsonq.New().JSONString(submitResult).Find("code").(float64))
	if code != 200 {
		fmt.Println("提交学时失败：", submitResult)

	}
	return submitResult, nil
}

// 获取节点进度
func ZxcpksGetNodeProgressAction(cache *zxcpks.ZxcpksUserCache, node ZxcpksNode) (int, error) {

	//拉取当前适配的观看进度
	nowProgressResult, err := cache.PullLastProgressApi(node.Id)
	nowProgress := gojsonq.New().JSONString(nowProgressResult).Find("data").(string)
	//fmt.Println("当前视频进度：", nowProgress)
	if err != nil {
		return 0, err
	}

	code := int(gojsonq.New().JSONString(nowProgress).Find("code").(float64))
	if code != 200 {
		fmt.Println("获取进度失败：", nowProgressResult)

	}
	floatProgress, err := strconv.ParseFloat(nowProgress, 64)
	return int(floatProgress * 100), nil
}

// 开始学习时访问的接口，获取session
func ZxcpksStartSutdyAction(cache *zxcpks.ZxcpksUserCache, node ZxcpksNode) (string, error) {
	startResult, err := cache.StartStudyApi(node.Id, node.CourseId)
	if err != nil {
		return "", err
	}

	sessionId := gojsonq.New().JSONString(startResult).Find("data").(string)
	return sessionId, nil
}

// 提交学时
func ZxcpksSubmitSutdyTimeAction(cache *zxcpks.ZxcpksUserCache, node ZxcpksNode, sessionId string, progress int) (string, error) {
	submitResult, err := cache.SubmitStudyTimeApi(sessionId, progress)
	if err != nil {
		return "", err
	}
	code := int(gojsonq.New().JSONString(submitResult).Find("code").(float64))
	if code != 200 {
		fmt.Println("提交学时失败：", submitResult)

	}
	return submitResult, nil
}
