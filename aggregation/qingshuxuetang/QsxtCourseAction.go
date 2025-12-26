package qingshuxuetang

import (
	"fmt"
	"strconv"

	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/api/qingshuxuetang"
)

type QsxtCourse struct {
	ClassId                       string  `json:"classId"`
	ProjectName                   string  `json:"projectName"`
	SchoolId                      string  `json:"schoolId"`
	SchoolName                    string  `json:"schoolName"`
	SemesterId                    string  `json:"semesterId"` //学期ID
	SemesterYear                  string  `json:"semesterYear"`
	SemesterName                  string  `json:"semesterName"`
	CourseId                      string  `json:"courseId"`
	CourseName                    string  `json:"courseName"`
	CourseCoverImg                string  `json:"courseCoverImg"`
	HasCourseWare                 bool    `json:"hasCourseWare"`
	HasNewCourse                  bool    `json:"hasNewCourse"`
	StudyStatus                   int     `json:"studyStatus"`
	StudyStatusName               string  `json:"studyStatusName"`
	AllowLearn                    bool    `json:"allowLearn"`
	ClassCredit                   float32 `json:"classCredit"`
	CoursewareLearnGainScore      float64 `json:"coursewareGainLearnScore"`      //课件学习当前获分
	CoursewareLearnTotalScore     float64 `json:"coursewareLearnTotalScore"`     //课件学习总分
	CourseWorkGainScore           float64 `json:"courseWorkGainScore"`           //课程作业获得的分数
	CourseWorkTotalScore          float64 `json:"courseWorkTotalScore"`          //课程作业总分
	CourseMaterialsLearnGainCore  float64 `json:"courseMaterialsLearnGainCore"`  //课程资料学习当前获分
	CourseMaterialsLearnTotalCore float64 `json:"courseMaterialsLearnTotalCore"` //课程资料学习总分
}

type QsxtNode struct {
	ClassId            string `json:"classId"`
	SchoolId           string `json:"schoolId"`
	CourseId           string `json:"courseId"`
	SemesterId         string `json:"semesterId"`
	BigId              string `json:"BigId"`
	BigName            string `json:"BigName"`
	BigCoverImgUrl     string `json:"coverImgUrl"`
	NodeId             string `json:"NodeId"`
	NodeName           string `json:"NodeName"`
	Size               string `json:"Size"`
	NodeType           string `json:"NodeType"` //有不同类型，chapter为章节节点一般不学，html类似于文档类型，video为视屏
	Duration           int    `json:"Duration"` //视频总时长（ms）,如果是课件之类的这个值一般为0
	NodeSize           int    `json:"NodeSize"`
	StudyTimes         int    `json:"StudyTimes"`         //不知道是个啥
	TotalStudyDuration int    `json:"totalstudyDuration"` //一共学习了多久
	LastStudyTime      int    `json:"lastStudyTime"`      //最后学习的时间点
	MaxStudyPosition   int    `json:"maxStudyPosition"`
}

// 拉取课程列表
func PullCourseListAction(cache *qingshuxuetang.QsxtUserCache) ([]QsxtCourse, error) {
	courseList := []QsxtCourse{}
	coursesJson, err := cache.QsxtPullCourseApi(3, nil)
	if err != nil {
		return nil, err
	}
	//异常处理
	pullStatus := gojsonq.New().JSONString(coursesJson).Find("hr")
	if pullStatus == nil {
		return nil, fmt.Errorf(coursesJson)
	}
	if int(pullStatus.(float64)) != 0 {
		return nil, fmt.Errorf(coursesJson)
	}

	courseListJson := gojsonq.New().JSONString(coursesJson).Find("data")
	if projects, ok := courseListJson.([]interface{}); ok {
		for _, projectJson := range projects {
			if project, ok1 := projectJson.(map[string]interface{}); ok1 {
				classId := strconv.Itoa(int(project["classId"].(float64)))
				projectName := project["name"].(string)
				schoolName := project["schoolName"].(string)
				schoolId := strconv.Itoa(int(project["schoolId"].(float64)))
				if periods, ok2 := project["periods"].([]interface{}); ok2 {
					for _, periodJson := range periods {
						if period, ok3 := periodJson.(map[string]interface{}); ok3 {
							semesterId := strconv.Itoa(int(period["id"].(float64))) //学期ID
							semesterYear := period["year"].(string)
							semesterName := period["name"].(string)
							//这里才是课程
							if csListJson, ok4 := period["courses"].([]interface{}); ok4 {
								for _, cJson := range csListJson {
									if courseJson, ok5 := cJson.(map[string]interface{}); ok5 {
										courseName := courseJson["name"].(string)
										studyStatus := int(courseJson["studyStatus"].(float64))
										studyStatusName := courseJson["studyStatusName"].(string)
										allowLearn := courseJson["allowLearn"].(bool)
										course := QsxtCourse{
											ClassId:         classId,
											ProjectName:     projectName,
											SchoolId:        schoolId,
											SchoolName:      schoolName,
											SemesterId:      semesterId,
											CourseId:        courseJson["id"].(string),
											SemesterYear:    semesterYear,
											SemesterName:    semesterName,
											CourseName:      courseName,
											StudyStatus:     studyStatus,
											StudyStatusName: studyStatusName,
											AllowLearn:      allowLearn,
										}
										if coverImg, ok := courseJson["coverImg"].(string); ok {
											course.CourseCoverImg = coverImg
										}
										//更新分数数据
										_, err1 := UpdateCourseScore(cache, &course)
										if err1 != nil {
											return nil, err1
										}

										courseList = append(courseList, course)

									}
								}
							}
						}
					}
				}
			}
		}
	}
	return courseList, nil
}

// 拉取课程详细信息
func PullCourseNodeListAction(cache *qingshuxuetang.QsxtUserCache, course QsxtCourse) ([]QsxtNode, error) {
	nodeList := []QsxtNode{}
	//拉取课程详细信息------------------------
	courseDetailJson, err := cache.QsxtPullCourseDetailApi(course.SemesterId, course.ClassId, course.SchoolId, course.CourseId, 3, nil)
	if err != nil {
		return nil, err
	}
	//异常处理
	pullStatus := gojsonq.New().JSONString(courseDetailJson).Find("hr")
	if pullStatus == nil {
		return nil, fmt.Errorf(courseDetailJson)
	}
	if int(pullStatus.(float64)) != 0 {
		return nil, fmt.Errorf(courseDetailJson)
	}

	detailJson := gojsonq.New().JSONString(courseDetailJson).Find("data")
	if _, ok := detailJson.(map[string]interface{}); !ok {
		return nil, fmt.Errorf(courseDetailJson)
	}

	//拉取任务点---------------------------------------
	nodesJson, err := cache.QsxtPullNodeApi(detailJson.(map[string]interface{})["coursewareUrl"].(string), 3, nil)
	if err != nil {
		return nil, err
	}
	//异常处理
	pullStatus = gojsonq.New().JSONString(nodesJson).Find("hr")
	if pullStatus == nil {
		return nil, fmt.Errorf(courseDetailJson)
	}
	if int(pullStatus.(float64)) != 0 {
		return nil, fmt.Errorf(courseDetailJson)
	}
	nodeListJson := gojsonq.New().JSONString(nodesJson).Find("data")

	if objJson, ok := nodeListJson.(map[string]interface{}); ok {
		bigId := objJson["id"].(string)
		bigName := objJson["name"].(string)
		bigCoverImgUrl := objJson["coverImgUrl"].(string)

		if ndJson, ok1 := objJson["nodes"].([]interface{}); ok1 {
			nodes := pullChapterAction(ndJson, course, bigId, bigName, bigCoverImgUrl)
			nodeList = append(nodeList, nodes...)
		}
	}

	// 拉取任务进度---------------------
	recordResult, err := cache.PullStudyRecordApi(course.SemesterId, course.ClassId, course.SchoolId, course.CourseId, 3, nil)
	if err != nil {
		return nil, err
	}
	//异常处理
	pullStatus = gojsonq.New().JSONString(courseDetailJson).Find("hr")
	if pullStatus == nil {
		return nil, fmt.Errorf(recordResult)
	}
	if int(pullStatus.(float64)) != 0 {
		return nil, fmt.Errorf(recordResult)
	}
	recordListJson := gojsonq.New().JSONString(recordResult).Find("data")
	if objJson, ok := recordListJson.([]interface{}); ok {
		for _, ndJson := range objJson {
			if node, ok1 := ndJson.(map[string]interface{}); ok1 {
				contentId := node["contentId"].(string)
				for i, nd := range nodeList {
					if nd.NodeId == contentId {
						nodeList[i].TotalStudyDuration = int(node["totalstudyDuration"].(float64))
						nodeList[i].StudyTimes = int(node["studyTimes"].(float64))
						nodeList[i].LastStudyTime = int(node["lastStudyTime"].(float64))
						nodeList[i].MaxStudyPosition = int(node["maxStudyPosition"].(float64))
					}

				}
			}
		}
	}
	return nodeList, nil
}

// 遍历截取所有节点
func pullChapterAction(nodes []interface{}, course QsxtCourse, bigId, bigName, coverImgUrl string) []QsxtNode {
	nodeList := []QsxtNode{}
	for _, nodeJson := range nodes {
		if node, ok := nodeJson.(map[string]interface{}); ok {
			nd := QsxtNode{
				ClassId:        course.ClassId,
				CourseId:       course.CourseId,
				SchoolId:       course.SchoolId,
				SemesterId:     course.SemesterId,
				BigId:          bigId,
				BigName:        bigName,
				BigCoverImgUrl: coverImgUrl,
				NodeId:         node["id"].(string),
				NodeName:       node["name"].(string),
				NodeType:       node["type"].(string),
			}

			duration, ok1 := node["duration"].(float64)
			if ok1 {
				nd.Duration = int(duration)
			}

			nodeSize, ok2 := node["nodeSize"].(float64)
			if ok2 {
				nd.NodeSize = int(nodeSize)
			}
			nodeList = append(nodeList, nd)
			if nds, ok1 := node["nodes"].([]interface{}); ok1 {
				nodeList = append(nodeList, pullChapterAction(nds, course, bigId, bigName, coverImgUrl)...)
			}
		}
	}
	return nodeList
}

// 更新获取对应课程的分数信息
func UpdateCourseScore(cache *qingshuxuetang.QsxtUserCache, course *QsxtCourse) (string, error) {
	scoreJson, err := cache.QsxtPullCourseScoreApi(course.SemesterId, course.ClassId, course.SchoolId, course.CourseId, 3, nil)
	if err != nil {
		return "", err
	}
	pullStatus := gojsonq.New().JSONString(scoreJson).Find("hr")
	if pullStatus == nil {
		return "", fmt.Errorf(scoreJson)
	}
	if int(pullStatus.(float64)) != 0 {
		return "", fmt.Errorf(scoreJson)
	}
	if scoresArrayJson, ok := gojsonq.New().JSONString(scoreJson).Find("data.rows").([]interface{}); ok {
		for _, courseJson := range scoresArrayJson {
			if courseObj, ok1 := courseJson.(map[string]interface{}); ok1 {
				//过滤对应不上的课程
				if fmt.Sprintf("%d", int(courseObj["id"].(float64))) != course.CourseId ||
					fmt.Sprintf("%d", int(courseObj["periodId"].(float64))) != course.SemesterId {
					continue
				}
				if usualScoresListJson, ok2 := courseObj["usualScores"].([]interface{}); ok2 {
					for _, usualScoreJson := range usualScoresListJson {
						if usualScoreObj, ok3 := usualScoreJson.(map[string]interface{}); ok3 {
							usualScoreType := int(usualScoreObj["type"].(float64))
							maxScore := usualScoreObj["maxScore"].(float64)
							score := usualScoreObj["score"].(float64)
							switch usualScoreType {
							case 2: //课件观看
								course.CoursewareLearnTotalScore = maxScore
								course.CoursewareLearnGainScore = score
								break
							case 4: //作业
								course.CourseWorkTotalScore = maxScore
								course.CourseWorkGainScore = score
							case 6: //课程资料
								course.CourseMaterialsLearnTotalCore = maxScore
								course.CourseMaterialsLearnGainCore = score
							}

						}
					}
				}

			}
		}
	}
	return scoreJson, nil
}

// 开始学习
func (node QsxtNode) StartStudyTimeAction(cache *qingshuxuetang.QsxtUserCache) (string, error) {
	startResult, err := cache.StartStudyApi(node.ClassId, node.NodeId, "11", node.CourseId, node.SemesterId, node.SchoolId, 3, nil)
	if err != nil {
		return "", err
	}
	//异常处理
	pullStatus := gojsonq.New().JSONString(startResult).Find("hr")
	if pullStatus == nil {
		return "", fmt.Errorf(startResult)
	}
	if int(pullStatus.(float64)) != 0 {
		return "", fmt.Errorf(startResult)
	}

	startId := gojsonq.New().JSONString(startResult).Find("data").(string)
	return startId, nil
}

// 提交学时
func (node QsxtNode) SubmitStudyTimeAction(cache *qingshuxuetang.QsxtUserCache, startId string, isEnd bool) (string, error) {
	submitResult, err1 := cache.SubmitStudyTimeApi(node.SchoolId, startId, 0, isEnd, 3, nil)
	if err1 != nil {
		return "", err1
	}
	return submitResult, nil
}
