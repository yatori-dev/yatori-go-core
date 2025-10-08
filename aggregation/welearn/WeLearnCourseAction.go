package welearn

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/api/welearn"
)

// 课程
type WeLearnCourse struct {
	Scid      string  `json:"scid"`
	Cid       string  `json:"cid"`
	Name      string  `json:"name"` //课程名字
	Img       string  `json:"img"`
	Per       float32 `json:"per"` //课程学习进度
	Type      string  `json:"type"`
	TaskCount string  `json:"taskcount"`
	Uid       string  `json:"uid"`
	ClassId   string  `json:"classid"`
}

// 课程章节
type WeLearnChapter struct {
	Unitname string `json:"unitname"`
	Id       string `json:"id"`
	UnitIdx  int    `json:"unitidx"`
	Visible  bool   `json:"visible"`
	Name     string `json:"name"`
}

// 对应章节的任务点
type WeLearnPoint struct {
	Name               string    `json:"name"`
	Number             int       `json:"number"`
	Crate              string    `json:"crate"`
	Id                 string    `json:"id"`         //id
	Location           string    `json:"location"`   //章节定位位置
	IsComplete         string    `json:"iscomplete"` //是否完成两种状态：未完成，已完成
	LearnTime          time.Time `json:"learntime"`  //学习时间
	IsVisible          bool      `json:"isvisible"`  //是否可见，一般可见才刷，true为可见
	VsTime             string    `json:"vstime"`
	VeTime             string    `json:"vetime"`
	Snmsg              string    `json:"snmsg"`
	IsLimited          bool      `json:"islimited"` //是否限制
	Enablereview       bool      `json:"enablereview"`
	LearnCount         int       `json:"learncount"`   //学习次数
	CompleteTime       time.Time `json:"completetime"` //完成时间
	DisplayAccessStyle string    `json:"displayaccessstyle"`
}

// 拉取学习中的课程列表
func WeLearnPullCourseListAction(cache *welearn.WeLearnUserCache) ([]WeLearnCourse, error) {
	courseList := make([]WeLearnCourse, 0)
	listJson, err := cache.PullCourseListApi(3, nil)
	if err != nil {
		//log.Println("PullCourseListApi err:", err)
		return []WeLearnCourse{}, err
	}
	ret := gojsonq.New().JSONString(listJson).Find("ret")
	if ret == nil {
		return []WeLearnCourse{}, errors.New(listJson)
	}
	//返回信息异常
	if int(ret.(float64)) != 0 {
		return []WeLearnCourse{}, errors.New(listJson)
	}

	//遍历拉取课程
	listObj := gojsonq.New().JSONString(listJson).Find("clist")
	if courses, ok1 := listObj.([]interface{}); ok1 {
		for _, item := range courses {
			if obj, ok2 := item.(map[string]interface{}); ok2 {
				course := WeLearnCourse{
					Scid:      fmt.Sprintf("%d", int(obj["scid"].(float64))),
					Cid:       fmt.Sprintf("%d", int(obj["cid"].(float64))),
					Name:      obj["name"].(string),
					Img:       obj["img"].(string),
					Per:       float32(obj["per"].(float64)),
					Type:      obj["type"].(string),
					TaskCount: obj["taskcount"].(string),
				}
				uid, classId, _ := WeLearnGetCourseInfoAction(cache, course)
				course.Uid = uid
				course.ClassId = classId
				courseList = append(courseList, course)
			}

		}
	}

	return courseList, nil
}

// 获取必要课程信息
func WeLearnGetCourseInfoAction(cache *welearn.WeLearnUserCache, course WeLearnCourse) (string, string, error) {
	var uid string
	var classid string
	html, err := cache.PullCourseInfoApi(course.Cid, 3, nil)
	if err != nil {
		return "", "", err
	}
	re := regexp.MustCompile(`"uid":\s*(\d+)`)
	matches := re.FindStringSubmatch(html)
	if len(matches) > 1 {
		uid = matches[1]
	}
	re1 := regexp.MustCompile(`classid=(\d+)`)
	matches1 := re1.FindStringSubmatch(html)
	if len(matches1) > 1 {
		classid = matches1[1]
	}
	return uid, classid, nil
}

// 拉取课程的大章节
func WeLearnPullCourseChapterAction(cache *welearn.WeLearnUserCache, course WeLearnCourse) ([]WeLearnChapter, error) {
	chapterList := make([]WeLearnChapter, 0)
	chapterListJson, err1 := cache.PullCourseChapterApi(course.Cid, course.Uid, course.ClassId, 3, nil)
	if err1 != nil {
		return chapterList, err1
	}

	ret := gojsonq.New().JSONString(chapterListJson).Find("ret")
	if ret == nil || int(ret.(float64)) != 0 {
		return chapterList, errors.New(chapterListJson)
	}

	chapterObj := gojsonq.New().JSONString(chapterListJson).Find("info")

	if chapters, ok1 := chapterObj.([]interface{}); ok1 {
		for unitIndx, item := range chapters {
			if obj, ok2 := item.(map[string]interface{}); ok2 {
				chapter := WeLearnChapter{
					Unitname: obj["unitname"].(string),
					Id:       obj["id"].(string),
					UnitIdx:  unitIndx,
					Name:     obj["name"].(string),
				}
				parseBool, err2 := strconv.ParseBool(obj["visible"].(string))
				if err2 == nil {
					chapter.Visible = parseBool
				}
				chapterList = append(chapterList, chapter)
			}
		}
	}
	//fmt.Println(chapterList)
	return chapterList, nil
}

// 拉取对应课程章节的小任务点
func WeLearnPullChapterPointAction(cache *welearn.WeLearnUserCache, course WeLearnCourse, chapter WeLearnChapter) ([]WeLearnPoint, error) {
	pointList := make([]WeLearnPoint, 0)
	fmt.Println(course.Cid, course.Uid, course.ClassId, chapter.Id)
	pointsJson, err := cache.PullCoursePointApi(course.Cid, course.Uid, course.ClassId, fmt.Sprintf("%d", chapter.UnitIdx), 3, nil)
	if err != nil {
		return pointList, err
	}

	ret := gojsonq.New().JSONString(pointsJson).Find("ret")
	if ret == nil || int(ret.(float64)) != 0 {
		return pointList, errors.New(pointsJson)
	}

	chapterObj := gojsonq.New().JSONString(pointsJson).Find("info")

	if points, ok1 := chapterObj.([]interface{}); ok1 {
		for _, item := range points {
			if obj, ok2 := item.(map[string]interface{}); ok2 {
				point := WeLearnPoint{
					Name:               obj["name"].(string),
					Number:             int(obj["number"].(float64)),
					Crate:              obj["crate"].(string),
					Id:                 obj["id"].(string),
					Location:           obj["location"].(string),
					IsComplete:         obj["iscomplete"].(string),
					VsTime:             obj["vstime"].(string),
					VeTime:             obj["vetime"].(string),
					Snmsg:              obj["snmsg"].(string),
					LearnCount:         int(obj["learncount"].(float64)),
					DisplayAccessStyle: obj["DisplayAccessStyle"].(string),
				}
				isVisible, err2 := strconv.ParseBool(obj["isvisible"].(string))
				if err2 == nil {
					point.IsVisible = isVisible
				}
				isLimited, err2 := strconv.ParseBool(obj["islimited"].(string))
				if err2 == nil {
					point.IsLimited = isLimited
				}
				enablereview, err2 := strconv.ParseBool(obj["enablereview"].(string))
				if err2 == nil {
					point.Enablereview = enablereview
				}
				learnTime, err3 := time.Parse("15:04:05", obj["learntime"].(string))
				if err3 == nil {
					point.LearnTime = learnTime
				}
				if obj["completetime"].(string) != "" {
					completeTime, err4 := time.Parse("2006-01-02 15:04:05", obj["completetime"].(string))
					if err4 == nil {
						point.CompleteTime = completeTime
					}
				}
				pointList = append(pointList, point)
			}
		}
	}
	return pointList, nil
}

// 完成任务点
func WeLearnCompletePointAction(cache *welearn.WeLearnUserCache, course WeLearnCourse, point WeLearnPoint) error {
	_, err := cache.StartStudyApi(course.Cid, point.Id, course.Uid, "100", course.ClassId, true, 3, nil)
	if err != nil {
		return err
	}

	//fmt.Println(api)
	submitApi1, err1 := cache.SubmitStudyPlan1Api(course.Cid, point.Id, course.Uid, "100", course.ClassId, 3, nil)
	if err1 != nil {
		return err1
	}
	//第一种答题方式
	submitStatus := gojsonq.New().JSONString(submitApi1).Find("ret")
	if submitStatus == nil {
		return errors.New(submitApi1)
	}

	//如果第一种提完成任务点方案正常
	if int(submitStatus.(float64)) == 0 {
		return nil
	}

	submitApi2, err2 := cache.SubmitStudyPlan2Api(course.Cid, point.Id, course.Uid, "100", course.ClassId, 100, "completed", 3, nil)
	if err2 != nil {
		return err2
	}
	submitStatus2 := gojsonq.New().JSONString(submitApi2).Find("ret")
	if submitStatus2 == nil {
		return errors.New(submitApi2)
	}

	if int(submitStatus2.(float64)) != 0 {
		return errors.New(submitApi2)
	}
	return nil
}

// 提交学时前进行进行调用
func WeLearnSubmitStudyTimeAction(cache *welearn.WeLearnUserCache, course WeLearnCourse, point WeLearnPoint) (string, int, int, int, string, error) {
	//第一种提交学时方式
	submitApi, err := cache.SubmitStudyTimeApi(course.Uid, course.Cid, course.ClassId, point.Id, 3, nil)
	if err != nil {
		return "", 0, 0, 0, "", err
	}
	submitStatus := gojsonq.New().JSONString(submitApi).Find("ret")
	if submitStatus == nil {
		return "", 0, 0, 0, "", errors.New(submitApi)
	}

	//第二方案，可能是超时原因，所以先进行首次数据拉取再学习
	_, err1 := cache.StartStudyApi(course.Cid, point.Id, course.Uid, "100", course.ClassId, false, 3, nil)
	if err1 != nil {
		return "", 0, 0, 0, "", err1
	}
	if int(submitStatus.(float64)) == -1 {
		submitApi, err = cache.SubmitStudyTimeApi(course.Uid, course.Cid, course.ClassId, point.Id, 3, nil)
		if err != nil {
			return "", 0, 0, 0, "", err
		}
		submitStatus = gojsonq.New().JSONString(submitApi).Find("ret")
		if submitStatus == nil {
			return "", 0, 0, 0, "", errors.New(submitApi)
		}
		if int(submitStatus.(float64)) != 0 {
			return "", 0, 0, 0, "", errors.New(submitApi)
		}
	}

	comment := gojsonq.New().JSONString(submitApi).Find("comment")
	if comment == nil {
		return "", 0, 0, 0, "", errors.New(submitApi)
	}
	completionStatus := ""
	progressMeasure := 0
	sessionTime := 0
	totalTime := 0
	scaled := ""
	completionStatusRes, ok1 := gojsonq.New().JSONString(comment.(string)).Find("cmi.completion_status").(string)
	if ok1 {
		completionStatus = completionStatusRes
	}
	progressMeasureRes, ok2 := gojsonq.New().JSONString(comment.(string)).Find("cmi.progress_measure").(string)
	if ok2 {
		if progressMeasureRes != "" {
			turnProgress, err2 := strconv.Atoi(progressMeasureRes)
			if err2 == nil {
				progressMeasure = turnProgress
			}

		}
	}
	sessionTimeRes, ok3 := gojsonq.New().JSONString(comment.(string)).Find("cmi.session_time").(float64)
	if ok3 {
		sessionTime = int(sessionTimeRes)
	}
	totalTimeRes, ok4 := gojsonq.New().JSONString(comment.(string)).Find("cmi.total_time").(float64)
	if ok4 {
		totalTime = int(totalTimeRes)
	}
	scaledRes, ok5 := gojsonq.New().JSONString(comment.(string)).Find("cmi.score.scaled").(string)
	if ok5 {
		scaled = scaledRes
	}

	//cache.KeepPointSessionPlan1Api(course.Cid, point.Id, course.Uid, course.ClassId)
	return completionStatus, progressMeasure, sessionTime, totalTime, scaled, nil
}
