package enaea

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/api/enaea"
)

type EnaeaProject struct {
	CircleNameShort  string    //短标题名称
	ClusterId        string    //组ID?反正不知道啥玩意
	ClusterName      string    //期数名称
	PlanState        int       //计划状态
	CircleId         string    //班级ID?反正不知道啥玩意
	CircleName       string    //班级期数名称
	StartTime        time.Time //开始时间
	EndTime          time.Time //结束时间
	CircleCardNumber string    //卡片数字编号
}
type EnaeaCourse struct {
	TitleTag          string  //课程对应侧边栏标签
	CourseTitle       string  //课程名称
	Remark            string  //课程节点名称
	CourseContentType string  //课程内容类型
	StudyProgress     float32 //学习进度
	CourseId          string  //课程ID
	CircleId          string
	SyllabusId        string
}
type EnaeaVideo struct {
	TitleTag         string  //侧边栏标签，选修还是必修
	CourseName       string  //课程名称
	CourseContentStr string  //视屏标签名称
	FileName         string  //视频文件名称
	TccId            string  //视频的TccID
	StudyProgress    float32 //视频学习进度
	Id               string  //Id
	CourseId         string
	CircleId         string
	VideoLength      int    //视屏总时长，单位秒
	SCFUCKPKey       string //key
	SCFUCKPValue     string //value
}

// 学习公社Enaea考试结构体
type EnaeaExam struct {
	TitleTag              string    //侧边栏标签
	ExamTitle             string    //考试标题
	ExamId                string    //试卷ID
	StartTime             time.Time //开始时间
	EndTime               time.Time //结束时间
	Introduce             string    //考试介绍
	Score                 float32   //考试结果分数
	PassScore             float32   //通过分数，或者说及格分数
	TotalScore            float32   //试卷总分
	CommentCount          int       //评论次数
	SubmitTime            time.Time //提交试卷时间
	SyllabusResourceId    string    //不知道啥玩意的ID
	ResourceRandomPaperId string    //不知道啥玩意的ID
	Remark                string    //不知道啥玩意
	ResourceId            string    //不知道这玩意有啥用
}

// ProjectListAction 获取所需要学习的工程列表
func ProjectListAction(cache *enaea.EnaeaUserCache) ([]EnaeaProject, error) {
	var projects []EnaeaProject
	api, err := enaea.PullProjectsApi(cache)
	if err != nil {
		return nil, err
	}
	jsonList := gojsonq.New().JSONString(api).Find("result.list")
	// 断言为切片并遍历
	if items, ok := jsonList.([]interface{}); ok {
		for _, item := range items {
			// 每个 item 是 map[string]interface{} 类型
			if obj, ok := item.(map[string]interface{}); ok {
				startTime, _ := time.Parse("2006.01.02", strings.Split(obj["startEndTime"].(string), "-")[0])
				endTime, _ := time.Parse("2006.01.02", strings.Split(obj["startEndTime"].(string), "-")[1])
				projects = append(projects, EnaeaProject{
					CircleNameShort:  obj["circleNameShort"].(string),
					CircleId:         strconv.Itoa(int(obj["circleId"].(float64))),
					ClusterId:        strconv.Itoa(int(obj["clusterId"].(float64))),
					CircleName:       obj["circleName"].(string),
					ClusterName:      obj["clusterName"].(string),
					StartTime:        startTime,
					EndTime:          endTime,
					PlanState:        int(obj["planState"].(float64)),
					CircleCardNumber: obj["circleCardNumber"].(string),
				})
			}
		}
	}
	return projects, nil
}

// 拉取项目对应课程
func CourseListAction(cache *enaea.EnaeaUserCache, circleId string) ([]EnaeaCourse, error) {
	var courses []EnaeaCourse
	courseHTML, err := enaea.PullStudyCourseHTMLApi(cache, circleId)
	if err != nil {
		return nil, err
	}
	//<li  class="left20">
	//<a title="课程学习" href="circleIndexRedirect.do?action=toNewMyClass&type=courseCategory4jwu&circleId=304591&syllabusId=1814144&isRequired=false&studentProgress=11">课程学习</a>
	//</li>
	// Use regex to find the syllabusId in the response body
	regexPattern := fmt.Sprintf(`<a title="([^"]*)" href="circleIndexRedirect.do\?action=toNewMyClass&type=course([^&]{0,50})&circleId=%s&syllabusId=([^&]*?)&isRequired=[^&]*&studentProgress=([\d]+)+">[^<]*</a>`, circleId)
	re := regexp.MustCompile(regexPattern)
	matches := re.FindAllStringSubmatch(courseHTML, -1)
	for _, v := range matches {
		api, err := enaea.PullStudyCourseListApi(cache, circleId, v[3], v[2])
		if err != nil {
			return nil, err
		}
		jsonList := gojsonq.New().JSONString(api).Find("result.list")
		// 断言为切片并遍历
		if items, ok := jsonList.([]interface{}); ok {
			for _, item := range items {
				// 每个 item 是 map[string]interface{} 类型
				if obj, ok := item.(map[string]interface{}); ok {
					remark := obj["remark"].(string)
					centerDTO := obj["studyCenterDTO"].(map[string]interface{})
					studyProgress, _ := strconv.ParseFloat(centerDTO["studyProgress"].(string), 64)
					courses = append(courses, EnaeaCourse{
						TitleTag:          v[1],
						CircleId:          circleId,
						SyllabusId:        v[3],
						Remark:            remark,
						StudyProgress:     float32(studyProgress),
						CourseId:          strconv.Itoa(int(centerDTO["courseId"].(float64))),
						CourseTitle:       centerDTO["courseTitle"].(string),
						CourseContentType: centerDTO["coursecontentType"].(string),
					},
					)
				}
			}
		}

	}
	return courses, nil
}

// 拉取对应课程的视频
func VideoListAction(cache *enaea.EnaeaUserCache, course *EnaeaCourse) ([]EnaeaVideo, error) {
	var videos []EnaeaVideo
	api, err := enaea.PullCourseVideoListApi(cache, course.CircleId, course.CourseId)
	if err != nil {
		return nil, err
	}
	jsonList := gojsonq.New().JSONString(api).Find("result.list")
	if items, ok := jsonList.([]interface{}); ok {
		for _, item := range items {
			if obj, ok := item.(map[string]interface{}); ok {

				studyProgress, _ := strconv.ParseFloat(obj["studyProgress"].(string), 64)
				if obj["filename"] == nil {
					fmt.Println("空")
				}
				courseContentStr, _ := url.QueryUnescape(obj["courseContentStr"].(string))
				//统计视屏时长
				videoTime := 0
				if obj["length"] != nil {
					length := strings.Split(obj["length"].(string), ":")
					hours, hoursErr := strconv.Atoi(length[0])
					if hoursErr == nil {
						videoTime += hours * 60 * 60
					}

					minus, minusErr := strconv.Atoi(length[1])
					if minusErr == nil {
						videoTime += minus * 60
					}

					seconds, secondsErr := strconv.Atoi(length[2])
					if secondsErr == nil {
						videoTime += seconds
					}
				}

				videos = append(videos, EnaeaVideo{
					TitleTag:         course.TitleTag,
					CourseName:       course.Remark,
					TccId:            obj["tccId"].(string),
					FileName:         obj["filename"].(string),
					CourseContentStr: courseContentStr,
					StudyProgress:    float32(studyProgress),
					VideoLength:      videoTime,
					Id:               strconv.Itoa(int(obj["id"].(float64))),
					CourseId:         course.CourseId,
					CircleId:         course.CircleId,
				})
			}
		}
	}
	return videos, nil
}

// 开始学习适配，首次学习视频前必须先调用这个函数接口
func StatisticTicForCCVideAction(cache *enaea.EnaeaUserCache, video *EnaeaVideo) error {
	json, K, V, err := enaea.StatisticTicForCCVideApi(cache, video.CourseId, video.Id, video.CircleId)
	if err != nil {
		return err
	}
	if gojsonq.New().JSONString(json).Find("success") == false {
		return errors.New(gojsonq.New().JSONString(json).Find("message").(string))
	}
	video.SCFUCKPKey = K
	video.SCFUCKPValue = V
	return nil
}

// 提交学时
// {"success": false,"message":"nologin"}
func SubmitStudyTimeAction(cache *enaea.EnaeaUserCache, video *EnaeaVideo, time int64 /*普通模式下填time2.Now().UnixMilli()，暴力模式下填学习的分钟数*/, model int64 /*0为普通模式，1为暴力可自定义*/) error {
	var api string
	var err error
	if model == 0 { //普通模式
		api, err = enaea.SubmitStudyTimeApi(cache, video.CircleId, video.SCFUCKPKey, video.SCFUCKPValue, video.Id, time)
	} else if model == 1 {
		api, err = enaea.SubmitStudyTimeFastApi(cache, video.CircleId, video.SCFUCKPKey, video.SCFUCKPValue, video.Id, time)
	}

	if err != nil {
		return err
	}
	if gojsonq.New().JSONString(api).Find("success") == false {
		return errors.New(gojsonq.New().JSONString(api).Find("message").(string))
	}
	if gojsonq.New().JSONString(api).Find("progress") == nil {
		return errors.New("提交学时时服务器端返回消息异常：" + api)
	}
	video.StudyProgress = float32(gojsonq.New().JSONString(api).Find("progress").(float64))
	return nil
}

// 拉取项目对应考试列表
func ExamListAction(cache *enaea.EnaeaUserCache, circleId string) ([]EnaeaExam, error) {
	var exams []EnaeaExam
	courseHTML, err := enaea.PullStudyCourseHTMLApi(cache, circleId)
	if err != nil {
		return nil, err
	}
	//<li class="left20">
	//<a title="在线考试" href="circleIndexRedirect.do?action=toNewMyClass&amp;type=exam&amp;circleId=339403&amp;syllabusId=1949144&amp;isRequired=false&amp;studentProgress=100">在线考试</a>
	//</li>
	// Use regex to find the syllabusId in the response body
	regexPattern := fmt.Sprintf(`<a title="([^"]*)" href="circleIndexRedirect.do\?action=toNewMyClass&type=exam([^&]{0,50})&circleId=%s&syllabusId=([^&]*?)&isRequired=[^&]*&studentProgress=([\d]+)+">[^<]*</a>`, circleId)
	re := regexp.MustCompile(regexPattern)
	matches := re.FindAllStringSubmatch(courseHTML, -1)
	for _, v := range matches {
		api, err := enaea.PullExamListApi(cache, circleId, v[3], v[2])
		if err != nil {
			return nil, err
		}
		jsonList := gojsonq.New().JSONString(api).Find("result.list")
		// 断言为切片并遍历
		if items, ok := jsonList.([]interface{}); ok {
			for _, item := range items {
				// 每个 item 是 map[string]interface{} 类型
				if obj, ok := item.(map[string]interface{}); ok {
					remark := obj["remark"].(string)
					//centerDTO := obj["studyCenterDTO"].(map[string]interface{})
					//studyProgress, _ := strconv.ParseFloat(centerDTO["studyProgress"].(string), 64)
					exams = append(exams, EnaeaExam{
						TitleTag:              v[1],
						ExamTitle:             obj["title"].(string),
						ExamId:                strconv.Itoa(obj["id"].(int)),
						Remark:                remark,
						Introduce:             obj["introduce"].(string),
						ResourceId:            strconv.Itoa(obj["resourceId"].(int)),
						ResourceRandomPaperId: strconv.Itoa(obj["resourceRandomPaperId"].(int)),
						SyllabusResourceId:    obj["syllabusResourceId"].(string),
						PassScore:             (float32)(obj["passScore"].(float64)),
						TotalScore:            (float32)(obj["totalScore"].(float64)),
						CommentCount:          obj["commentCount"].(int),
						//StartTime:             obj[""],
						//EndTime:               obj,
					},
					)
				}
			}
		}

	}
	return exams, nil
}
