package xuexitong

import (
	"github.com/yatori-dev/yatori-go-core/core"
	agg "github.com/yatori-dev/yatori-go-core/internal/aggregation/xuexitong"
	api "github.com/yatori-dev/yatori-go-core/internal/api/xuexitong"
	"github.com/yatori-dev/yatori-go-core/internal/driver/driverutil"
	"github.com/yatori-dev/yatori-go-core/models/ctype"
	qentity "github.com/yatori-dev/yatori-go-core/que-core/qentity"
)

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func progressPercent(done, total int) float64 {
	return driverutil.Progress(done, total)
}

func videoProgress(p *api.PointVideoDto) float64 {
	if p == nil || p.Duration <= 0 {
		return 0
	}
	value := float64(p.PlayTime) / float64(p.Duration) * 100
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}

type questionWriter interface {
	WriteQuestionForAIAction(cache *api.XueXiTUserCache, aiUrl, model string, aiType ctype.AiType, apiKey string) error
	WriteQuestionForExternalAction(exUrl string) error
}

func questionOf(writer any) *qentity.Question {
	switch value := writer.(type) {
	case *agg.XXTWorkQuestion:
		return &value.Question
	case *agg.XXTExamQuestion:
		return &value.Question
	default:
		return nil
	}
}

func (c *client) answerQuestion(writer any, source core.AnswerSource) error {
	question := questionOf(writer)
	if question == nil {
		return core.Invalid(core.XueXiTong, "answer", "unsupported question type")
	}
	answerWriter, _ := writer.(questionWriter)
	switch source {
	case core.AnswerSourceManual:
		return nil
	case core.AnswerSourceBank:
		bank := c.Config.Options.QuestionBank
		if bank == nil || bank.URL == "" {
			return core.Invalid(core.XueXiTong, "answer", "question bank config is required")
		}
		if answerWriter == nil {
			return core.Invalid(core.XueXiTong, "answer", "question does not support external answers")
		}
		return core.Wrap(core.XueXiTong, "answer.bank", answerWriter.WriteQuestionForExternalAction(bank.URL))
	case core.AnswerSourceAI:
		ai := c.Config.Options.AI
		if ai == nil || ai.APIKey == "" {
			return core.Invalid(core.XueXiTong, "answer", "AI config is required")
		}
		if answerWriter == nil {
			return core.Invalid(core.XueXiTong, "answer", "question does not support AI answers")
		}
		return core.Wrap(core.XueXiTong, "answer.ai", answerWriter.WriteQuestionForAIAction(c.cache, ai.URL, ai.Model, ctype.AiType(ai.Type), ai.APIKey))
	default:
		builtinAnswer(question)
		return nil
	}
}

func builtinAnswer(question *qentity.Question) {
	if len(question.Answers) > 0 {
		return
	}
	question.Answers = []string{"A"}
}

func answerFromQuestion(question qentity.Question) core.Answer {
	return core.Answer{
		Question: question.Content,
		Type:     question.Type,
		Options:  []string(question.Options),
		Value:    []string(question.Answers),
	}
}
