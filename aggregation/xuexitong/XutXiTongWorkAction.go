package xuexitong

import (
	"github.com/yatori-dev/yatori-go-core/api/entity"
	"github.com/yatori-dev/yatori-go-core/api/xuexitong"
	"github.com/yatori-dev/yatori-go-core/models/ctype"
	"github.com/yatori-dev/yatori-go-core/utils/qutils"
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

// 外部题库答题
func StartExternalWorkAction(cache *xuexitong.XueXiTUserCache, exUrl string, questionAction entity.Question, isSubmit int) string {
	//选择题
	for i := range questionAction.Choice {
		q := &questionAction.Choice[i] // 获取对应选项
		q.AnswerExternalGet(exUrl)
	}
	//判断题
	for i := range questionAction.Judge {
		q := &questionAction.Judge[i] // 获取对应选项
		q.AnswerExternalGet(exUrl)
	}
	//填空题
	for i := range questionAction.Fill {
		q := &questionAction.Fill[i] // 获取对应选项
		q.AnswerExternalGet(exUrl)
	}

	//简答题
	for i := range questionAction.Short {
		q := &questionAction.Short[i]
		q.AnswerExternalGet(exUrl)
	}
	var resultStr string
	if isSubmit == 0 {
		resultStr = WorkNewSubmitAnswerAction(cache, questionAction, false)
	} else if isSubmit == 1 {
		resultStr = WorkNewSubmitAnswerAction(cache, questionAction, true)
	}

	return resultStr
}

// 答案修正匹配
func AnswerFixedPattern(choices []entity.ChoiceQue, judges []entity.JudgeQue, fills []entity.FillQue, shorts []entity.ShortQue) {
	//选择题修正
	for i, choice := range choices {
		if choice.Answers != nil {
			candidateSelects := []string{} //待选
			selectAnswers := []string{}
			for _, option := range choice.Options {
				candidateSelects = append(candidateSelects, option)
			}
			for _, answer := range choice.Answers {
				selectAnswers = append(selectAnswers, qutils.SimilarityArrayAnswer(answer, candidateSelects))
			}
			if selectAnswers != nil {
				choices[i].Answers = selectAnswers
			}
		}
	}
	for i, judge := range judges {
		if judge.Answers != nil {
			selectAnswer := []string{}
			for _, answer := range judge.Answers {
				selectAnswer = append(selectAnswer, qutils.SimilarityArrayAnswer(answer, []string{"正确", "错误"}))
			}
			judges[i].Answers = selectAnswer
		}
	}
}
