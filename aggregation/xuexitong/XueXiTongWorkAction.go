package xuexitong

import (
	"strings"

	ddddocr "github.com/Changbaiqi/ddddocr-go/utils"
	ort "github.com/yalue/onnxruntime_go"
	"github.com/yatori-dev/yatori-go-core/api/xuexitong"
	"github.com/yatori-dev/yatori-go-core/models/ctype"
	"github.com/yatori-dev/yatori-go-core/utils"
	log2 "github.com/yatori-dev/yatori-go-core/utils/log"
	"github.com/yatori-dev/yatori-go-core/utils/qutils"
)

// WorkNewSubmitAnswerAction 提交答题
func WorkNewSubmitAnswerAction(userCache *xuexitong.XueXiTUserCache, question xuexitong.Question, isSubmit bool) (string, error) {
	submitState := "1"
	if isSubmit {
		submitState = ""
	}
	answer, err := userCache.WorkNewSubmitAnswer(question.CourseId, question.ClassId, question.Knowledgeid, question.Cpi,
		question.JobId, question.TotalQuestionNum, question.AnswerId, question.WorkAnswerId, question.Api,
		question.FullScore, question.OldSchoolId, question.OldWorkId, question.WorkRelationId, question.Enc_work,
		question, submitState)
	//answer, err := userCache.WorkNewSubmitAnswerNew(question.CourseId, question.ClassId, question.Knowledgeid, question.Cpi,
	//	question.JobId, question.TotalQuestionNum, question.AnswerId, question.WorkAnswerId, question.Api,
	//	question.FullScore, question.OldSchoolId, question.OldWorkId, question.WorkRelationId, question.Enc_work,
	//	question, submitState)
	if err != nil {
		if err.Error() == "触发验证码" {
			log2.Print(log2.DEBUG, utils.RunFuncName(), "触发验证码，正在进行AI智能识别绕过.....")
			for {
				img, err1 := userCache.XueXiTVerificationCodeApi(7, nil)
				if err1 != nil {
					return "", err1
				}

				_, width, _ := utils.GetImageShape(img)

				var shape ort.Shape
				if width == 140 {
					shape = ort.NewShape(1, 23)
				} else {
					shape = ort.NewShape(1, 30)
				}
				codeResult := ddddocr.SemiOCRVerification(img, shape)
				status, err1 := userCache.XueXiTPassVerificationCode(codeResult, 7, nil)
				//fmt.Println(codeResult)
				//fmt.Println(status)
				if status {
					break
				}
			}
			answer, err = userCache.WorkNewSubmitAnswer(question.CourseId, question.ClassId, question.Knowledgeid, question.Cpi,
				question.JobId, question.TotalQuestionNum, question.AnswerId, question.WorkAnswerId, question.Api,
				question.FullScore, question.OldSchoolId, question.OldWorkId, question.WorkRelationId, question.Enc_work,
				question, submitState) //尝试重新拉取卡片信息
			log2.Print(log2.DEBUG, utils.RunFuncName(), "绕过成功")
		}
	}

	//fmt.Println(answer)
	return answer, nil
}

// 开始做题
func StartAIWorkAction(cache *xuexitong.XueXiTUserCache, userId, aiUrl, model, apiKey string, aiTYpe ctype.AiType, questionAction xuexitong.Question, isSubmit int) string {
	//选择题
	for i := range questionAction.Choice {
		q := &questionAction.Choice[i] // 获取对应选项
		message := AIProblemMessage(questionAction.Title, q.Text, xuexitong.ExamTurn{
			XueXChoiceQue: *q,
		})
		q.AnswerAIGet(userId, aiUrl, model, aiTYpe, message, apiKey)
	}
	//判断题
	for i := range questionAction.Judge {
		q := &questionAction.Judge[i] // 获取对应选项
		message := AIProblemMessage(q.Type.String(), q.Text, xuexitong.ExamTurn{
			XueXJudgeQue: *q,
		})

		q.AnswerAIGet(userId, aiUrl, model, aiTYpe, message, apiKey)
	}
	//填空题
	for i := range questionAction.Fill {
		q := &questionAction.Fill[i] // 获取对应选项
		message := AIProblemMessage(q.Type.String(), q.Text, xuexitong.ExamTurn{
			XueXFillQue: *q,
		})

		q.AnswerAIGet(userId, aiUrl, model, aiTYpe, message, apiKey)
	}

	//简答题
	for i := range questionAction.Short {
		q := &questionAction.Short[i]
		message := AIProblemMessage(q.Type.String(), q.Text, xuexitong.ExamTurn{
			XueXShortQue: *q,
		})
		q.AnswerAIGet(userId, aiUrl, model, aiTYpe, message, apiKey)
	}
	var resultStr string
	if isSubmit == 0 {
		resultStr, _ = WorkNewSubmitAnswerAction(cache, questionAction, false)
	} else if isSubmit == 1 {
		resultStr, _ = WorkNewSubmitAnswerAction(cache, questionAction, true)
	}
	return resultStr
}

// 外部题库答题
func StartExternalWorkAction(cache *xuexitong.XueXiTUserCache, exUrl string, questionAction xuexitong.Question, isSubmit int) string {
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
	//连线题
	for i := range questionAction.Matching {
		q := &questionAction.Matching[i]
		q.AnswerExternalGet(exUrl)
	}
	var resultStr string
	if isSubmit == 0 {
		resultStr, _ = WorkNewSubmitAnswerAction(cache, questionAction, false)
	} else if isSubmit == 1 {
		resultStr, _ = WorkNewSubmitAnswerAction(cache, questionAction, true)
	}

	return resultStr
}

// 答案修正匹配
func AnswerFixedPattern(choices []xuexitong.ChoiceQue, judges []xuexitong.JudgeQue) {
	//选择题修正
	for i, choice := range choices {
		if choice.Answers != nil {
			candidateSelects := []string{} //待选
			selectAnswers := []string{}
			for _, option := range choice.Options {
				candidateSelects = append(candidateSelects, option)
			}
			if len(candidateSelects) > 0 {
				for _, answer := range choice.Answers {
					selectAnswers = append(selectAnswers, qutils.SimilarityArrayAnswer(answer, candidateSelects))
				}
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
				answer = strings.ReplaceAll(answer, "对", "正确")
				answer = strings.ReplaceAll(answer, "√", "正确")
				answer = strings.ReplaceAll(answer, "×", "错误")
				selectAnswer = append(selectAnswer, qutils.SimilarityArrayAnswer(answer, []string{"正确", "错误"}))
			}
			judges[i].Answers = selectAnswer
		}
	}
}
