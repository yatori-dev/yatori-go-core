package xuexitong

import (
	"github.com/yatori-dev/yatori-go-core/api/entity"
	"github.com/yatori-dev/yatori-go-core/api/xuexitong"
)

// WorkNewSubmitAnswerAction 提交答题
func WorkNewSubmitAnswerAction(userCache *xuexitong.XueXiTUserCache, question entity.Question, isSubmit bool) string {
	submitState := ""
	if isSubmit {
		submitState = "1"
	}
	answer, _ := userCache.WorkNewSubmitAnswer(question.CourseId, question.ClassId, question.Knowledgeid, question.Cpi,
		question.JobId, question.TotalQuestionNum, question.AnswerId, question.WorkAnswerId, question.Api,
		question.FullScore, question.OldSchoolId, question.OldWorkId, question.WorkRelationId, question.Enc_work,
		question, submitState)
	//fmt.Println(answer)
	return answer
}
