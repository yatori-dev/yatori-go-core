package entity

import (
	"github.com/yatori-dev/yatori-go-core/models/ctype"
	"github.com/yatori-dev/yatori-go-core/utils"
	"github.com/yatori-dev/yatori-go-core/utils/log"
	"os"
)

// XueXiTCourse 课程所有信息
type XueXiTCourseJson struct {
	Result           int           `json:"result"`
	Msg              string        `json:"msg"`
	ChannelList      []ChannelItem `json:"channelList"`
	Mcode            string        `json:"mcode"`
	Createcourse     int           `json:"createcoursed"`
	TeacherEndCourse int           `json:"teacherEndCourse"`
	ShowEndCourse    int           `json:"showEndCourse"`
	HasMore          bool          `json:"hasMore"`
	StuEndCourse     int           `json:"stuEndCourse"`
}

// ChannelItem 课程列表
type ChannelItem struct {
	Cfid     int    `json:"cfid"`
	Norder   int    `json:"norder"`
	CataName string `json:"cataName"`
	Cataid   string `json:"cataid"`
	Id       int    `json:"id"`
	Cpi      int    `json:"cpi"`
	Key      any    `json:"key"`
	Content  struct {
		Studentcount int    `json:"studentcount"`
		Chatid       string `json:"chatid"`
		IsFiled      int    `json:"isFiled"`
		Isthirdaq    int    `json:"isthirdaq"`
		Isstart      bool   `json:"isstart"`
		Isretire     int    `json:"isretire"`
		Name         string `json:"name"`
		Course       struct {
			Data []struct {
				BelongSchoolId     string `json:"belongSchoolId"`
				Coursestate        int    `json:"coursestate"`
				Teacherfactor      string `json:"teacherfactor"`
				IsCourseSquare     int    `json:"isCourseSquare"`
				Schools            string `json:"schools"`
				CourseSquareUrl    string `json:"courseSquareUrl"`
				Imageurl           string `json:"imageurl"`
				AppInfo            string `json:"appInfo"`
				Name               string `json:"name"`
				DefaultShowCatalog int    `json:"defaultShowCatalog"`
				Id                 int    `json:"id"`
				AppData            int    `json:"appData"`
			} `json:"data"`
		} `json:"course"`
		Roletype int    `json:"roletype"`
		Id       int    `json:"id"`
		State    int    `json:"state"`
		Cpi      int    `json:"cpi"`
		Bbsid    string `json:"bbsid"`
		IsSquare int    `json:"isSquare"`
	} `json:"content"`
	Topsign int `json:"topsign"`
}

// XueXiTCourse 关键信息过滤截取最终的实体
type XueXiTCourse struct {
	CourseName string //课程名称
	ClassId    string //classId
	CourseId   string //课程Id
	Cpi        string //不知道是啥玩意，反正需要
	PersonId   string //个人Id
	UserId     string //UserId
}

// ExamTopics holds a map of ExamTopic indexed by answerId
type YingHuaExamTopics struct {
	YingHuaExamTopics map[string]YingHuaExamTopic
}

// ExamTopic represents a single exam question
type YingHuaExamTopic struct {
	AnswerId string        `json:"answerId"`
	Index    string        `json:"index"`
	Source   string        `json:"source"`
	Content  string        `json:"content"`
	Type     string        `json:"type"`
	Selects  []TopicSelect `json:"selects"`
	Answers  string        `json:"answers"`
}

// TopicSelect represents a possible answer choice
type TopicSelect struct {
	Value string `json:"value"`
	Num   string `json:"num"`
	Text  string `json:"text"`
}

type ChoiceQue struct {
	Type    ctype.QueType
	Text    string
	Options map[string]string
	Answer  string // 答案
}

// Question TODO 这里考虑是否在其中直接将答案做出 直接上报提交 或 保存提交
type Question struct {
	Choice []ChoiceQue //选择类型
}

type ExamTurn struct {
	ChoiceQue
	YingHuaExamTopic
}

func (q *ChoiceQue) AnswerAIGet(userID string,
	url,
	model string,
	aiType ctype.AiType,
	aiChatMessages utils.AIChatMessages,
	apiKey string) {
	aiAnswer, err := utils.AggregationAIApi(url, model, aiType, aiChatMessages, apiKey)
	if err != nil {
		log.Print(log.INFO, `[`, userID, `] `, log.BoldRed, "Ai异常，返回信息：", err.Error())
		os.Exit(0)
	}
	q.Answer = aiAnswer
}

// 转标准题目格式
func (q *YingHuaExamTopic) turnProblem() utils.Problem {
	problem := utils.Problem{
		Hash:    "",
		Type:    q.Type,
		Content: q.Content,
		Answer:  []string{},
		Json:    "",
	}
	return problem
}
