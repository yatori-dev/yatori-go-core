package cqie

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	ddddocr "github.com/Changbaiqi/ddddocr-go/utils"
	"github.com/thedevsaddam/gojsonq"
	ort "github.com/yalue/onnxruntime_go"
	cqieApi "github.com/yatori-dev/yatori-go-core/api/cqie"
	"github.com/yatori-dev/yatori-go-core/utils"
	"github.com/yatori-dev/yatori-go-core/utils/log"
)

type CqieCourse struct {
	Id              string    //这个是课程的id也就是courseId
	CourseName      string    //课程名称
	StudentCourseId string    //对应学生课程ID
	SumUnit         int       //一共多少单元
	HaveUnit        int       //以学单元
	SumTime         time.Time //总时长
	HaveTime        time.Time //以学时长
	Learned         string    //已学进度
	CoursewareId    string    //某个和课程相关的ID
	Version         string
}

type CqieVideo struct {
	VideoId          string //视屏Id
	CourseId         string //课程Id
	UnitId           string //单元Id
	VideoName        string //视屏名称
	TimeLength       int    //视屏时长
	StudentCourseId  string //学时课程ID
	CoursewareId     string //某个和课程相关的ID
	StudyId          string //学习视屏临时生成的ID
	MaxCurrentPos    int    //当前观看进度
	StudyTime        int    //以及学习到的时间点
	Version          string
	VideoSegmentNode []VideoSegmentNode
}

// 视屏中的段任务
type VideoSegmentNode struct {
	Id              string
	SegmentId       string
	KnowledgeNodeId string
	CourseId        string
	UnitId          string
	SegmentName     string
	StartTimeStr    string
	EndTimeStr      string
}

// 试卷
type WorkPaper struct {
	questions []WorkQuestion //题目列表
}

// 题目
type WorkQuestion struct {
	Id              string
	CourseId        string
	RootUnitId      string
	UnitId          string
	KnowledgeNodeId string
	QuestionType    int
	Content         string
	Thinking        string
	ReferenceAnswer string
	Status          int
	OptionsVos      []OptionsVos
}

// 选择题选项
type OptionsVos struct {
	Id            string
	ExercisesId   string
	OptionContent string
	AnswerFlag    int
}

// CqieLoginAction 登录API聚合整理
// {"refresh_code":1,"status":false,"msg":"账号密码不正确"}
// {"_code": 1, "status": false,"msg": "账号登录超时，请重新登录", "result": {}}
func CqieLoginAction(cache *cqieApi.CqieUserCache) error {
	for {
		path, cookie := cache.VerificationCodeApi() //获取验证码
		cache.SetCookie(cookie)
		img, _ := utils.ReadImg(path) //读取验证码图片
		codeResult := ddddocr.SemiOCRVerification(img, ort.NewShape(1, 26))
		utils.DeleteFile(path)         //删除验证码文件
		cache.SetVerCode(codeResult)   //填写验证码
		jsonStr, _ := cache.LoginApi() //执行登录
		log.Print(log.DEBUG, "["+cache.Account+"] "+"LoginAction---"+jsonStr)
		if gojsonq.New().JSONString(jsonStr).Find("msg") == "验证码有误！" {
			continue
		} else if int(gojsonq.New().JSONString(jsonStr).Find("code").(float64)) != 200 {
			return errors.New(gojsonq.New().JSONString(jsonStr).Find("msg").(string))
		}
		cache.SetAccess_Token(gojsonq.New().JSONString(jsonStr).Find("data.access_token").(string))
		cache.SetToken(gojsonq.New().JSONString(jsonStr).Find("data.user.token").(string))
		cache.SetAppId(gojsonq.New().JSONString(jsonStr).Find("data.user.appId").(string))
		cache.SetIpaddr(gojsonq.New().JSONString(jsonStr).Find("data.user.ipaddr").(string))
		cache.SetDeptId(gojsonq.New().JSONString(jsonStr).Find("data.user.deptId").(string))
		userJson, err := cache.UserDetailsApi(8, nil) //获取用户信息
		if err != nil {
			return err
		}
		cache.SetUserId(gojsonq.New().JSONString(userJson).Find("data.userId").(string))
		cache.SetDeptId(gojsonq.New().JSONString(userJson).Find("data.deptId").(string))
		cache.SetStudentId(gojsonq.New().JSONString(userJson).Find("data.id").(string))
		cache.SetUserName(gojsonq.New().JSONString(userJson).Find("data.userName").(string))
		cache.SetOrgId(gojsonq.New().JSONString(userJson).Find("data.orgId").(string))
		cache.SetUserId(gojsonq.New().JSONString(userJson).Find("data.userId").(string))
		if mobile, ok := gojsonq.New().JSONString(userJson).Find("data.mobile").(string); ok {
			cache.SetMobile(mobile)
		}
		cache.SetOrgMajorId(gojsonq.New().JSONString(userJson).Find("data.orgMajorId").(string))

		log.Print(log.DEBUG, "["+cache.Account+"] "+" 登录成功")
		break
	}
	return nil
}

// 直接用token登录，方便测试用
func CqieLoginTokenAction(cache *cqieApi.CqieUserCache, token string) error {
	cache.SetAccess_Token(token)
	userJson, err := cache.UserDetailsApi(8, nil) //获取用户信息
	if err != nil {
		return err
	}
	cache.SetUserId(gojsonq.New().JSONString(userJson).Find("data.userId").(string))
	cache.SetDeptId(gojsonq.New().JSONString(userJson).Find("data.deptId").(string))
	cache.SetStudentId(gojsonq.New().JSONString(userJson).Find("data.id").(string))
	cache.SetUserName(gojsonq.New().JSONString(userJson).Find("data.userName").(string))
	cache.SetOrgId(gojsonq.New().JSONString(userJson).Find("data.orgId").(string))
	cache.SetUserId(gojsonq.New().JSONString(userJson).Find("data.userId").(string))
	if mobile, ok := gojsonq.New().JSONString(userJson).Find("data.mobile").(string); ok {
		cache.SetMobile(mobile)
	}
	cache.SetOrgMajorId(gojsonq.New().JSONString(userJson).Find("data.orgMajorId").(string))

	log.Print(log.DEBUG, "["+cache.Account+"] "+" 登录成功")
	return nil
}

// CqiePullCourseListAction 拉取课程列表信息
func CqiePullCourseListAction(cache *cqieApi.CqieUserCache) ([]CqieCourse, error) {
	var courseList []CqieCourse
	courseApi, err := cache.PullCourseListApiNew(5, nil)
	if err != nil {
	}
	if gojsonq.New().JSONString(courseApi).Find("msg") != "操作成功" {
		return courseList, errors.New("获取数据失败：" + courseApi)
	}
	jsonList := gojsonq.New().JSONString(courseApi).Find("data.records")
	if items, ok := jsonList.([]interface{}); ok {
		for _, item := range items {
			// 每个 item 是 map[string]interface{} 类型
			if obj, ok := item.(map[string]interface{}); ok {
				if obj["sumTime"] == nil || obj["haveTime"] == nil { //没有视屏总时间的部分说明该部分不是视屏，那么直接跳过添加
					continue
				}
				sumTime, _ := time.Parse("15:04", obj["sumTime"].(string))
				haveTime, _ := time.Parse("15:04", obj["haveTime"].(string))
				courseList = append(courseList, CqieCourse{
					Id:              obj["id"].(string),
					CourseName:      obj["name"].(string),
					SumUnit:         int(obj["sumUnit"].(float64)),
					HaveUnit:        int(obj["haveUnit"].(float64)),
					SumTime:         sumTime,
					HaveTime:        haveTime,
					Learned:         obj["learned"].(string),
					StudentCourseId: obj["studentCourseId"].(string),
					CoursewareId:    obj["coursewareId"].(string),
					Version:         obj["version"].(string),
				})
			}
		}
	}
	return courseList, nil
}

// 拉取对应课程的所有视屏
func PullCourseVideoListAction(cache *cqieApi.CqieUserCache, course *CqieCourse) ([]CqieVideo, error) {
	var videoList []CqieVideo
	courseApi, err := cache.PullCourseDetailApi(course.Id, course.StudentCourseId, course.Version, 5, nil)
	if err != nil {
		return videoList, err
	}
	if gojsonq.New().JSONString(courseApi).Find("msg") != "操作成功" {
		return videoList, errors.New("获取数据失败：" + courseApi)
	}
	jsonList := gojsonq.New().JSONString(courseApi).Find("data.courseCatalogVos")
	if items, ok := jsonList.([]interface{}); ok {
		for _, item := range items {
			// 每个 item 是 map[string]interface{} 类型
			if obj, ok := item.(map[string]interface{}); ok { //进入到courseCatalogVos层，即章节层
				if nodes, ok := obj["children"].([]interface{}); ok { //如果有对应章节子节点那么继续
					for _, node := range nodes { //循环获取所有节点
						if nodeObj, ok := node.(map[string]interface{}); ok { //检查是否为节点对象
							if videos, ok := nodeObj["courseCatalogVideoVos"].([]interface{}); ok { //判断对应节点是否有视屏列表
								for _, video := range videos { //循环获取节点视屏列表
									// 每个 item 是 map[string]interface{} 类型
									if obj1, ok := video.(map[string]interface{}); ok {
										videoList = append(videoList, CqieVideo{
											VideoId:         obj1["id"].(string),
											CourseId:        obj1["courseId"].(string),
											UnitId:          obj1["unitId"].(string),
											VideoName:       obj1["name"].(string),
											TimeLength:      int(obj1["timeLength"].(float64)),
											StudentCourseId: course.StudentCourseId,
											CoursewareId:    course.CoursewareId,
											Version:         course.Version,
										})
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return videoList, nil
}

func PullCourseVideoListAndProgress(cache *cqieApi.CqieUserCache, course *CqieCourse) ([]CqieVideo, error) {
	var videoList []CqieVideo
	courseApi, err := cache.PullProgressDetailApi(course.Id, course.StudentCourseId, course.Version, 5, nil)
	if err != nil {
		return videoList, err
	}
	if gojsonq.New().JSONString(courseApi).Find("msg") != "操作成功" {
		return videoList, errors.New("获取数据失败：" + courseApi)
	}
	jsonList := gojsonq.New().JSONString(courseApi).Find("data")
	if items, ok := jsonList.([]interface{}); ok {
		for _, item := range items {
			// 每个 item 是 map[string]interface{} 类型
			if obj, ok := item.(map[string]interface{}); ok { //进入到courseCatalogVos层，即章节层
				if nodes, ok := obj["children"].([]interface{}); ok { //如果有对应章节子节点那么继续
					for _, node := range nodes { //循环获取所有节点
						if nodeObj, ok := node.(map[string]interface{}); ok { //检查是否为节点对象
							if videos, ok := nodeObj["courseCatalogVideoVos"].([]interface{}); ok { //判断对应节点是否有视屏列表
								for _, video := range videos { //循环获取节点视屏列表
									// 每个 item 是 map[string]interface{} 类型
									if obj1, ok := video.(map[string]interface{}); ok {
										studyTime := 0
										if obj1["haveTime"] != nil {
											timeSplit := strings.Split(obj1["haveTime"].(string), ":")
											hour, _ := strconv.Atoi(timeSplit[0])
											minute, _ := strconv.Atoi(timeSplit[1])
											second, _ := strconv.Atoi(timeSplit[2])
											studyTime += hour*60*60 + minute*60 + second
										}

										//如果有分段视频
										videSegmentList := []VideoSegmentNode{}
										if videoSegmentsList, ok := obj1["courseCatalogVideoSegments"].([]interface{}); ok {
											for _, videoSegment := range videoSegmentsList {
												if videoSegmentObj, ok := videoSegment.(map[string]interface{}); ok {
													if videoSegmentKnowledgeTimeRangesVos, ok := videoSegmentObj["videoSegmentKnowledgeTimeRangesVos"].([]interface{}); ok {
														for _, videoSegmentKnowledgeTimeRangesVosList := range videoSegmentKnowledgeTimeRangesVos {
															if videoSegmentKnowledgeTimeRangesVosObj, ok := videoSegmentKnowledgeTimeRangesVosList.(map[string]interface{}); ok {
																//fmt.Println(videoSegmentKnowledgeTimeRangesVosObj)
																segmentNode := VideoSegmentNode{
																	Id:           videoSegmentKnowledgeTimeRangesVosObj["id"].(string),
																	SegmentId:    videoSegmentKnowledgeTimeRangesVosObj["segmentId"].(string),
																	CourseId:     videoSegmentObj["courseId"].(string),
																	UnitId:       videoSegmentObj["unitId"].(string),
																	SegmentName:  videoSegmentObj["segmentName"].(string),
																	StartTimeStr: videoSegmentKnowledgeTimeRangesVosObj["startTimeStr"].(string),
																	EndTimeStr:   videoSegmentKnowledgeTimeRangesVosObj["endTimeStr"].(string),
																}
																if knowledgeId, ok1 := videoSegmentKnowledgeTimeRangesVosObj["knowledgeNodeId"].(string); ok1 {
																	segmentNode.KnowledgeNodeId = knowledgeId
																}
																videSegmentList = append(videSegmentList, segmentNode)
															}
														}
													}

												}
											}
										}

										videoList = append(videoList, CqieVideo{
											VideoId:          obj1["id"].(string),
											CourseId:         obj1["courseId"].(string),
											UnitId:           obj1["unitId"].(string),
											VideoName:        obj1["name"].(string),
											TimeLength:       int(obj1["timeLength"].(float64)),
											StudentCourseId:  course.StudentCourseId,
											CoursewareId:     course.CoursewareId,
											StudyTime:        studyTime,
											Version:          course.Version,
											VideoSegmentNode: videSegmentList,
										})
									}
								}
							}
						}
					}
				}
				if videos, ok := obj["courseCatalogVideoVos"].([]interface{}); ok { //判断对应节点是否有视屏列表
					for _, video := range videos { //循环获取节点视屏列表
						// 每个 item 是 map[string]interface{} 类型
						if obj1, ok := video.(map[string]interface{}); ok {
							studyTime := 0
							if obj1["haveTime"] != nil {
								timeSplit := strings.Split(obj1["haveTime"].(string), ":")
								hour, _ := strconv.Atoi(timeSplit[0])
								minute, _ := strconv.Atoi(timeSplit[1])
								second, _ := strconv.Atoi(timeSplit[2])
								studyTime += hour*60*60 + minute*60 + second
							}
							videoList = append(videoList, CqieVideo{
								VideoId:         obj1["id"].(string),
								CourseId:        obj1["courseId"].(string),
								UnitId:          obj1["unitId"].(string),
								VideoName:       obj1["name"].(string),
								TimeLength:      int(obj1["timeLength"].(float64)),
								StudentCourseId: course.StudentCourseId,
								CoursewareId:    course.CoursewareId,
								StudyTime:       studyTime,
								Version:         course.Version,
							})
						}
					}
				}
			}
		}
	}
	return videoList, nil
}

// 学习视屏前一定要先调用这个函数才能开始学习
func StartStudyVideoAction(cache *cqieApi.CqieUserCache, video *CqieVideo) error {
	api, err := cache.GetVideoStudyIdApi(video.StudentCourseId, video.VideoId, video.Version, 5, nil)
	if err != nil {
		return err
	}

	if gojsonq.New().JSONString(api).Find("msg") != "操作成功" {
		return errors.New("获取数据失败：" + api)
	}
	find := gojsonq.New().JSONString(api).Find("data")
	if obj, ok := find.(map[string]interface{}); ok {
		if obj["coursewareId"] == nil {
			return errors.New("无法正常获取学习ID，返回内容：" + api)
		}
		video.CoursewareId = obj["coursewareId"].(string)
		video.StudyId = obj["id"].(string)
		video.MaxCurrentPos = int(obj["maxCurrentPos"].(float64))
	}
	return nil
}

// 提交学时
func SubmitStudyTimeAction(cache *cqieApi.CqieUserCache, video *CqieVideo, studyTime time.Time, startPos int, stopPos int, maxPos int) error {
	api, err := cache.SubmitStudyTimeApi(video.StudyId, video.Version, video.CourseId, video.StudentCourseId, video.UnitId, video.VideoId, studyTime, video.CoursewareId, startPos, stopPos, maxPos, 5, nil)
	if err != nil {
		return err
	}
	if gojsonq.New().JSONString(api).Find("msg") != "操作成功" {
		return errors.New("提交学时异常：" + api)
	}
	return nil
}

// 保存视屏学习时间点，学习完一个视屏就保存一次
func SaveVideoStudyTimeAction(cache *cqieApi.CqieUserCache, video *CqieVideo, startPos, stopPos int) error {
	var api string
	var err error

	api, err = cache.SaveStudyTimeApi(video.CourseId, video.StudentCourseId, video.UnitId, video.VideoId, video.CoursewareId, video.Version, startPos, stopPos, 5, nil)
	//如果有段节点
	for _, segmentNode := range video.VideoSegmentNode {
		api, err = cache.SaveSegmentStudyTimeApi(segmentNode.CourseId, video.StudentCourseId, segmentNode.UnitId, video.VideoId, video.CoursewareId, segmentNode.Id, fmt.Sprintf("%d", video.MaxCurrentPos), video.Version, startPos, stopPos, 5, nil)
		if segmentNode.KnowledgeNodeId != "" { //如果有作业，那么就写
			paperJson, err1 := cache.PullVideoWorkPaperApi(segmentNode.Id, video.StudentCourseId, segmentNode.UnitId)
			if err1 != nil {
				return err1
			}
			questionList := []WorkQuestion{}
			if paperListObj, ok := gojsonq.New().JSONString(paperJson).Find("data").([]interface{}); ok {
				for _, questionJson := range paperListObj {
					if questionObj, ok := questionJson.(map[string]interface{}); ok {
						options := []OptionsVos{}
						if optionVosListObj, ok := questionObj["optionVos"].([]interface{}); ok {
							for _, optionVo := range optionVosListObj {
								if optionVoObj, ok := optionVo.(map[string]interface{}); ok {
									vos := OptionsVos{
										Id:            optionVoObj["id"].(string),
										ExercisesId:   optionVoObj["exercisesId"].(string),
										OptionContent: optionVoObj["optionContent"].(string),
										AnswerFlag:    int(optionVoObj["answerFlag"].(float64)),
									}
									options = append(options, vos)
								}
							}
						}
						question := WorkQuestion{
							Id:              questionObj["id"].(string),
							CourseId:        questionObj["courseId"].(string),
							UnitId:          questionObj["unitId"].(string),
							KnowledgeNodeId: questionObj["knowledgeNodeId"].(string),
							QuestionType:    int(questionObj["questionType"].(float64)),
							Content:         questionObj["content"].(string),
							ReferenceAnswer: questionObj["referenceAnswer"].(string),
							Status:          int(questionObj["status"].(float64)),
							OptionsVos:      options,
						}

						if thinkContent, ok1 := questionObj["thinking"].(string); ok1 {
							question.Thinking = thinkContent
						}
						questionList = append(questionList, question)
					}
				}
			}
			//提交答案----------------------------------------------
			type Po struct {
				SubmitAnswer       string `json:"submitAnswer"`
				ExercisesId        string `json:"exercisesId"`
				QuestionType       int    `json:"questionType"`
				ReferenceAnswer    string `json:"referenceAnswer"`
				RecordId           string `json:"recordId"`
				SegmentKnowledgeId string `json:"segmentKnowledgeId"`
				UpdateCount        int    `json:"updateCount"`
			}
			type Answer struct {
				PoList             []Po   `json:"poList"`
				StudentCourseId    string `json:"studentCourseId"`
				UnitId             string `json:"unitId"`
				CourseId           string `json:"courseId"`
				SegmentKnowledgeId string `json:"segmentKnowledgeId"`
				StudentId          string `json:"studentId"`
				VideoId            string `json:"videoId"`
				DeptId             string `json:"deptId"`
				MajorId            string `json:"majorId"`
				Version            string `json:"version"`
				OrgId              string `json:"orgId"`
			}
			answer := Answer{
				StudentCourseId:    video.StudentCourseId,
				UnitId:             segmentNode.UnitId,
				CourseId:           segmentNode.CourseId,
				SegmentKnowledgeId: segmentNode.Id,
				StudentId:          cache.GetStudentId(),
				VideoId:            video.VideoId,
				DeptId:             cache.GetDeptId(),
				MajorId:            cache.GetOrgMajorId(),
				Version:            video.Version,
				OrgId:              cache.GetOrgId(),
			}
			poList := []Po{}
			for _, question := range questionList {
				po := Po{
					SubmitAnswer:       question.ReferenceAnswer,
					ExercisesId:        question.Id,
					QuestionType:       question.QuestionType,
					ReferenceAnswer:    question.ReferenceAnswer,
					SegmentKnowledgeId: segmentNode.KnowledgeNodeId,
					UpdateCount:        0,
				}
				poList = append(poList, po)
			}
			answer.PoList = poList
			//提交答案
			answerJson, err1 := json.Marshal(answer)
			if err1 != nil {
				return err1
			}
			answerApi, err1 := cache.SubmitWorkAnswerApi(string(answerJson))
			if err1 != nil {
				return err1
			}
			if code, ok := gojsonq.New().JSONString(answerApi).Find("code").(float64); ok {
				if int(code) != 200 {
					log.Print(log.INFO, "自动写作业提交答案错误：", answerApi)
				}
			}

		}
	}

	if err != nil {
		return err
	}
	if gojsonq.New().JSONString(api).Find("msg") != "操作成功" {
		return errors.New("保存学习点异常：" + api)
	}
	video.StudyId = gojsonq.New().JSONString(api).Find("data.id").(string) //赋值分配的学习ID
	return nil
}
