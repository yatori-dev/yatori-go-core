package xuexitong

import (
	"github.com/yatori-dev/yatori-go-core/api/entity"
	"github.com/yatori-dev/yatori-go-core/api/xuexitong"
	"github.com/yatori-dev/yatori-go-core/models/ctype"
)

// WorkNewSubmitAnswerAction 提交答题
func WorkNewSubmitAnswerAction(userCache *xuexitong.XueXiTUserCache, question entity.Question, isSubmit bool) string {
	submitState := "1"
	if isSubmit {
		submitState = ""
	}
	answer, _ := userCache.WorkNewSubmitAnswer(question.CourseId, question.ClassId, question.Knowledgeid, question.Cpi,
		question.JobId, question.TotalQuestionNum, question.AnswerId, question.WorkAnswerId, question.Api,
		question.FullScore, question.OldSchoolId, question.OldWorkId, question.WorkRelationId, question.Enc_work,
		question, submitState)
	//fmt.Println(answer)
	return answer
}

// 开始做题
func StartAIWorkAction(cache *xuexitong.XueXiTUserCache, userId, aiUrl, model, apiKey string, aiTYpe ctype.AiType, questionAction entity.Question, isSubmit int) string {
	//选择题
	for i := range questionAction.Choice {
		q := &questionAction.Choice[i] // 获取对应选项
		message := AIProblemMessage(q.Type.String(), q.Text, entity.ExamTurn{
			XueXChoiceQue: *q,
		})
		q.AnswerAIGet(userId, aiUrl, model, aiTYpe, message, apiKey)
	}
	//判断题
	for i := range questionAction.Judge {
		q := &questionAction.Judge[i] // 获取对应选项
		message := AIProblemMessage(q.Type.String(), q.Text, entity.ExamTurn{
			XueXJudgeQue: *q,
		})

		q.AnswerAIGet(userId, aiUrl, model, aiTYpe, message, apiKey)
	}
	//填空题
	for i := range questionAction.Fill {
		q := &questionAction.Fill[i] // 获取对应选项
		message := AIProblemMessage(q.Type.String(), q.Text, entity.ExamTurn{
			XueXFillQue: *q,
		})

		q.AnswerAIGet(userId, aiUrl, model, aiTYpe, message, apiKey)
	}

	//简答题
	for i := range questionAction.Short {
		q := &questionAction.Short[i]
		message := AIProblemMessage(q.Type.String(), q.Text, entity.ExamTurn{
			XueXShortQue: *q,
		})
		q.AnswerAIGet(userId, aiUrl, model, aiTYpe, message, apiKey)
	}
	var resultStr string
	if isSubmit == 0 {
		resultStr = WorkNewSubmitAnswerAction(cache, questionAction, false)
	} else if isSubmit == 1 {
		resultStr = WorkNewSubmitAnswerAction(cache, questionAction, true)
	}
	return resultStr
}
