package entity

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/yatori-dev/yatori-go-core/models/ctype"
	"github.com/yatori-dev/yatori-go-core/que-core/aiq"
	"github.com/yatori-dev/yatori-go-core/que-core/qentity"
	"github.com/yatori-dev/yatori-go-core/que-core/qtype"
	"github.com/yatori-dev/yatori-go-core/utils/log"
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
	AnswerId string `json:"answerId"`
	Index    string `json:"index"`
	Source   string `json:"source"`
	qentity.Question
	//Content  string        `json:"content"`
	//Type     string        `json:"type"`
	//Selects  []TopicSelect `json:"selects"`
	//Answers  string        `json:"answers"`
}

// TopicSelect represents a possible answer choice
//type TopicSelect struct {
//	Value string `json:"value"`
//	Num   string `json:"num"`
//	Text  string `json:"text"`
//}

// ChoiceQue 选择类型
type ChoiceQue struct {
	Type    qtype.QueType
	Qid     string //题目ID
	Text    string
	Options map[string]string
	Answers []string // 答案
}

// JudgeQue 判断类型
type JudgeQue struct {
	Type    qtype.QueType
	Qid     string //题目ID
	Text    string
	Options map[string]string
	Answers []string // 答案
}

// FillQue 填空类型
type FillQue struct {
	Type         qtype.QueType
	Qid          string
	Text         string
	OpFromAnswer map[string][]string // 位置与答案
}

// 简答类型
type ShortQue struct {
	Type         qtype.QueType
	Qid          string
	Text         string
	OpFromAnswer map[string][]string
}

// 阅读理解
type ReadQue struct {
	Type         qtype.QueType
	Qid          string
	Text         string
	OpFormAnswer map[string]struct {
		ChildType  int
		ChildId    string
		DataItemID string
		Text       map[string]string // 内部选项的 选择 和 具体内容
	} //内部题目和 选项
}

// Question TODO 这里考虑是否在其中直接将答案做出 直接上报提交 或 保存提交
type Question struct {
	Title            string //试卷标题
	Cpi              string
	JobId            string
	WorkId           string
	ClassId          string
	CourseId         string
	Ua               string
	FormType         string
	SaveStatus       string
	Version          string
	Tempsave         string
	PyFlag           string
	UserId           string
	Knowledgeid      string
	OldWorkId        string //最原始作业id
	FullScore        string //满分是多少
	OldSchoolId      string //原始作业单位id
	Api              string //api值
	WorkRelationId   string
	Enc_work         string
	Isphone          string
	RandomOptions    string
	WorkAnswerId     string
	AnswerId         string
	TotalQuestionNum string
	Choice           []ChoiceQue //选择类型
	Judge            []JudgeQue  //判断类型
	Fill             []FillQue   //填空类型
	Short            []ShortQue  //简答类型
	Read             []ReadQue   //阅读类型
}
type ExamTurn struct {
	XueXChoiceQue ChoiceQue
	XueXJudgeQue  JudgeQue
	XueXFillQue   FillQue
	XueXShortQue  ShortQue
	YingHuaExamTopic
}

type AnswerSetter interface {
	SetAnswers([]string)
}

func (q *ChoiceQue) SetAnswers(answers []string) {
	q.Answers = answers
}

func (q *JudgeQue) SetAnswers(answers []string) {
	q.Answers = answers
}

func (q *FillQue) SetAnswers(answers []string) {
	if len(answers) == 0 {
		return
	}

	for key := range q.OpFromAnswer {
		// 提取键中的序号（假设格式为"0第X空"）
		index := extractIndexFromKey(key)
		if index >= 0 && index < len(answers) {
			q.OpFromAnswer[key] = []string{answers[index]}
		} else {
			// 默认使用第一个答案或空列表
			if len(answers) > 0 {
				q.OpFromAnswer[key] = []string{answers[0]}
			} else {
				q.OpFromAnswer[key] = []string{}
			}
		}
	}
}

func (q *ShortQue) SetAnswers(answers []string) {
	q.OpFromAnswer["简答"] = answers
}

// 从键中提取序号（例如："0第3空" → 2，注意索引从0开始）
func extractIndexFromKey(key string) int {
	// 简单实现，实际可能需要更复杂的字符串处理
	// 这里假设key格式为"0第X空"，提取X并转换为整数
	// 示例实现，需要根据实际格式调整
	regex := regexp.MustCompile(`第(\d+)空`)
	matches := regex.FindStringSubmatch(key)
	if len(matches) >= 2 {
		if idx, err := strconv.Atoi(matches[1]); err == nil {
			return idx - 1 // 转换为0-based索引
		}
	}
	return -1 // 无效索引
}

func GetAIAnswer(as AnswerSetter, userID string, url, model string, aiType ctype.AiType, aiChatMessages aiq.AIChatMessages, apiKey string) {
	aiAnswer, err := aiq.AggregationAIApi(url, model, aiType, aiChatMessages, apiKey)
	if err != nil {
		log.Print(log.INFO, `[`, userID, `] `, log.BoldRed, "Ai异常，返回信息：", err.Error())
		os.Exit(0)
	}
	var answers []string
	err = json.Unmarshal([]byte(aiAnswer), &answers)
	if err != nil {
		answers = []string{"A"}
		fmt.Println("AI回复解析错误:", err)
	}
	as.SetAnswers(answers)
}

// AnswerAIGet ChoiceQue的AI回答获取方法
func (q *ChoiceQue) AnswerAIGet(userID,
	url, model string, aiType ctype.AiType, aiChatMessages aiq.AIChatMessages, apiKey string) {
	GetAIAnswer(q, userID, url, model, aiType, aiChatMessages, apiKey)
}

// AnswerAIGet JudgeQue的AI回答获取方法
func (q *JudgeQue) AnswerAIGet(userID,
	url, model string, aiType ctype.AiType, aiChatMessages aiq.AIChatMessages, apiKey string) {
	GetAIAnswer(q, userID, url, model, aiType, aiChatMessages, apiKey)
}

// AnswerAIGet FillQue的AI回答获取方法
func (q *FillQue) AnswerAIGet(userID,
	url, model string, aiType ctype.AiType, aiChatMessages aiq.AIChatMessages, apiKey string) {
	GetAIAnswer(q, userID, url, model, aiType, aiChatMessages, apiKey)
}

// AnswerAIGet ShortQue的AI回答获取方法
func (q *ShortQue) AnswerAIGet(userID,
	url, model string, aiType ctype.AiType, aiChatMessages aiq.AIChatMessages, apiKey string) {
	GetAIAnswer(q, userID, url, model, aiType, aiChatMessages, apiKey)
}

// TurnProblem 转标准题目格式
//func (q *YingHuaExamTopic) TurnProblem() utils.Problem {
//	problem := utils.Problem{
//		Hash:    "",
//		Type:    q.Type,
//		Content: q.Content,
//		Options: []string{},
//		Answer:  []string{},
//		Json:    "",
//	}
//	for _, topicSelect := range q.Selects {
//		problem.Options = append(problem.Options, topicSelect.Num+topicSelect.Text)
//	}
//	return problem
//}
