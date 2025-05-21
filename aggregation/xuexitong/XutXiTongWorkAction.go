package xuexitong

import (
	"fmt"
	"github.com/yatori-dev/yatori-go-core/api/entity"
	"github.com/yatori-dev/yatori-go-core/api/xuexitong"
)

// WorkNewSubmitAnswerAction 提交答题
func WorkNewSubmitAnswerAction(userCache *xuexitong.XueXiTUserCache, question entity.Question) {
	answer, _ := userCache.WorkNewSubmitAnswer(question.CourseId, question.ClassId, question.Knowledgeid, question.Cpi,
		question.JobId, question.TotalQuestionNum, question.AnswerId, question.WorkAnswerId, question.Api,
		question.FullScore, question.OldSchoolId, question.OldWorkId, question.WorkRelationId, question.Enc_work,
		question, "")
	fmt.Println(answer)
}
