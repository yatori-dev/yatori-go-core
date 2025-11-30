package qingshuxuetang

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/api/qingshuxuetang"
)

type QsxtWork struct {
	Id           string
	Title        string
	Type         int
	TotalTime    int     //总时长
	TotalScore   float64 //总分
	AnswerStatus int     //答题状态，打过了的话这个是2，没答过这个是-1，答过但没提交过只保存了那么就是0
	StudentScore float64 //得分
	TimeSpend    int     //花费时间
	Free         bool
	OutOfOrder   bool
	ViewDetail   bool
	CourseName   string
	SchoolId     string
	ClassId      string
	CourseId     string
	PassScore    float64
	AuditStatus  int //从没答过或者答过但只保存这个是null，写过了并提交过的话这个是-1
	FinalExam    bool
	WebDetailUrl string
}

func PullWorkListAction(cache *qingshuxuetang.QsxtUserCache, course QsxtCourse) ([]QsxtWork, error) {
	workList := []QsxtWork{}
	workListJson, err := cache.PullWorkListApi(course.SemesterId, course.ClassId, course.SchoolId, course.CourseId, 3, nil)
	if err != nil {
		return workList, err
	}
	//异常处理
	pullStatus := gojsonq.New().JSONString(workListJson).Find("hr")
	if pullStatus == nil {
		return workList, fmt.Errorf(workListJson)
	}
	if int(pullStatus.(float64)) != 0 {
		return workList, fmt.Errorf(workListJson)
	}

	if worksJson, ok := gojsonq.New().JSONString(workListJson).Find("data.rows").([]interface{}); ok {
		for _, workJson := range worksJson {
			if workObj, ok := workJson.(map[string]interface{}); ok {
				work := QsxtWork{
					Id:           workObj["id"].(string),
					Title:        workObj["title"].(string),
					Type:         int(workObj["type"].(float64)),
					TotalTime:    int(workObj["totalTime"].(float64)),
					AnswerStatus: int(workObj["answerStatus"].(float64)),
					Free:         workObj["free"].(bool),
					OutOfOrder:   workObj["outOfOrder"].(bool),
					ViewDetail:   workObj["viewDetail"].(bool),
					CourseName:   workObj["courseName"].(string),
					SchoolId:     course.SchoolId,
					ClassId:      course.ClassId,
					CourseId:     workObj["courseId"].(string),
					PassScore:    workObj["passScore"].(float64),
					FinalExam:    workObj["finalExam"].(bool),
					WebDetailUrl: workObj["webDetailUrl"].(string),
				}
				if score, ok1 := workObj["studentScore"].(float64); ok1 {
					work.StudentScore = score
				}
				if timeSpend, ok1 := workObj["timeSpend"].(float64); ok1 {
					work.TimeSpend = int(timeSpend)
				}
				if auditStatus, ok1 := workObj["auditStatus"].(float64); ok1 {
					work.AuditStatus = int(auditStatus)
				}
				workList = append(workList, work)
			}
		}
	}
	return workList, nil
}

// 写作业
func WriteWorkAction(cache *qingshuxuetang.QsxtUserCache, work QsxtWork, isSubmit bool) (string, error) {
	workQuestionsJson, err := cache.PullWorkQuestionListApi(work.ClassId, work.Id, work.SchoolId, work.CourseId, 3, nil)
	if err != nil {
		return "", err
	}
	//异常处理
	pullStatus := gojsonq.New().JSONString(workQuestionsJson).Find("hr")
	if pullStatus == nil {
		return "", fmt.Errorf(workQuestionsJson)
	}
	if int(pullStatus.(float64)) != 0 {
		return "", fmt.Errorf(workQuestionsJson)
	}

	type QsxtAnswer struct {
		Answer     string  `json:"answer"`
		QuestionId string  `json:"questionId"`
		Score      float64 `json:"score"`
		TimeSpend  int     `json:"timeSpend"`
	}
	type SubmitAnswer struct {
		Action         int          `json:"action"`
		ClassId        int          `json:"classId"`
		CourseId       string       `json:"courseId"`
		IssForceSubmit bool         `json:"issForceSubmit"`
		QuizId         string       `json:"quizId"`
		SchoolId       string       `json:"schoolId"`
		TimeSpend      int          `json:"timeSpend"`
		StudentAnswers []QsxtAnswer `json:"studentAnswers"`
	}
	numClassId, _ := strconv.Atoi(work.ClassId)
	answers := []QsxtAnswer{}
	submitAnswers := SubmitAnswer{
		ClassId:        numClassId,
		CourseId:       work.CourseId,
		IssForceSubmit: false,
		QuizId:         work.Id,
		SchoolId:       work.SchoolId,
		TimeSpend:      work.TotalTime / 5,
	}
	if isSubmit {
		submitAnswers.Action = 1 //0代表保存，1代表直接提交
	} else {
		submitAnswers.Action = 0 //0代表保存，1代表直接提交
	}

	if questionsJson, ok := gojsonq.New().JSONString(workQuestionsJson).Find("data.studentQuestions").([]interface{}); ok {

		for _, questionJson := range questionsJson {
			if questionObj, ok := questionJson.(map[string]interface{}); ok {
				questionId := questionObj["questionId"].(string) //题目ID
				//questionType := int(questionObj["questionType"].(float64)) //题目类型，1代表单选题
				solution := questionObj["solution"].(string) //答案
				score := questionObj["score"].(float64)
				submitJson, err1 := cache.SubmitAnswerApi(solution, questionId, work.Id, work.SchoolId, 5, nil)
				if err1 != nil {
					return "", err1
				}
				//异常处理
				pullStatus = gojsonq.New().JSONString(submitJson).Find("hr")
				if pullStatus == nil {
					return "", fmt.Errorf(submitJson)
				}
				if int(pullStatus.(float64)) != 0 {
					return "", fmt.Errorf(submitJson)
				}
				answers = append(answers, QsxtAnswer{
					Answer:     solution,
					QuestionId: questionId,
					Score:      score,
					TimeSpend:  0,
				})
			}
		}

	}
	submitAnswers.StudentAnswers = answers
	marshal, err := json.Marshal(submitAnswers)
	if err != nil {
		return "", err
	}
	saveJson, err := cache.SaveAnswerApi(string(marshal), 3, nil)
	if err != nil {
		return "", err
	}
	return saveJson, nil
}
