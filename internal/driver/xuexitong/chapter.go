package xuexitong

import (
	"context"
	"sort"
	"strings"

	"github.com/yatori-dev/yatori-go-core/v3/core"
	agg "github.com/yatori-dev/yatori-go-core/v3/internal/aggregation/xuexitong"
	api "github.com/yatori-dev/yatori-go-core/v3/internal/api/xuexitong"
	"github.com/yatori-dev/yatori-go-core/v3/internal/driver/driverutil"
	"github.com/yatori-dev/yatori-go-core/v3/models/ctype"
)

var _ core.ChapterTester = (*client)(nil)

// DoChapterTest 完成一个章节测验（章测）节点。
//
// Work.Raw 直接传 Tree 返回的 NodeWork.Raw，或者只传 NodeWork.ID，由 adapter
// 从 Tree 缓存里取回对应的 PointWorkDto。
func (c *client) DoChapterTest(ctx context.Context, req core.WorkRequest) (*core.AnswerResult, error) {
	unlock := c.Lock()
	defer unlock()
	if !c.loggedIn {
		return nil, core.NotLoggedIn(core.XueXiTong, "chapter_test")
	}

	point, err := c.resolveChapterTestPoint(req.Work)
	if err != nil {
		return nil, err
	}
	// 章节测验的 enc/ktoken 在手机端卡片里，先补拉一次卡片；顺便判断是否已完成。
	card, cardErr := fetchMobileCardCached(
		c.cache,
		driverutil.Int(point.ClassID),
		driverutil.Int(point.CourseID),
		point.KnowledgeID,
		point.CardIndex,
		driverutil.Int(point.Cpi),
	)
	if cardErr != nil {
		return nil, core.Wrap(core.XueXiTong, "chapter_test.card", cardErr)
	}
	unfinished, detectErr := point.AttachmentsDetection(card)
	if detectErr != nil {
		c.Logger.Debugf("xuexitong: chapter test attachment detection failed: %v", detectErr)
	} else if !unfinished {
		return &core.AnswerResult{
			TaskID:  req.Work.ID,
			Title:   firstNonEmpty(req.Work.Name, "章节测验"),
			Message: "章节测验已完成，跳过",
		}, nil
	}
	question, err := agg.ParseWorkQuestionAction(c.cache, point)
	if err != nil {
		return nil, core.Wrap(core.XueXiTong, "chapter_test.parse", err)
	}
	if !chapterTestHasQuestions(question) {
		return &core.AnswerResult{
			TaskID:  req.Work.ID,
			Title:   firstNonEmpty(question.Title, req.Work.Name, "章节测验"),
			Message: "章节测验没有题目，已跳过",
		}, nil
	}
	if err := c.answerChapterTest(&question, req.Source); err != nil {
		return nil, err
	}
	report, err := agg.WorkNewSubmitAnswerAction(c.cache, question, req.Submit)
	if err != nil {
		return nil, core.Wrap(core.XueXiTong, "chapter_test.submit", err)
	}

	return &core.AnswerResult{
		TaskID:    req.Work.ID,
		Title:     firstNonEmpty(question.Title, req.Work.Name, "章节测验"),
		Submitted: req.Submit,
		Answers:   chapterTestAnswers(question),
		Message:   report,
		Raw:       report,
	}, nil
}

func (c *client) resolveChapterTestPoint(work core.Work) (*api.PointWorkDto, error) {
	if point, ok := work.Raw.(*api.PointWorkDto); ok && point != nil {
		return point, nil
	}
	if node, ok := c.nodes[work.ID]; ok {
		if point, ok := node.Raw.(*api.PointWorkDto); ok && point != nil {
			return point, nil
		}
	}
	return nil, core.Invalid(core.XueXiTong, "chapter_test", "chapter test node is missing; call Tree first")
}

func chapterTestHasQuestions(question api.Question) bool {
	return len(question.Choice) > 0 ||
		len(question.Judge) > 0 ||
		len(question.Fill) > 0 ||
		len(question.Short) > 0 ||
		len(question.TermExplanation) > 0 ||
		len(question.Essay) > 0 ||
		len(question.Matching) > 0 ||
		len(question.Other) > 0
}

func (c *client) answerChapterTest(question *api.Question, source core.AnswerSource) error {
	switch source {
	case core.AnswerSourceManual:
		return nil
	case core.AnswerSourceBank:
		bank := c.Config.Options.QuestionBank
		if bank == nil || strings.TrimSpace(bank.URL) == "" {
			return core.Invalid(core.XueXiTong, "chapter_test.answer", "question bank config is required")
		}
		for i := range question.Choice {
			question.Choice[i].AnswerExternalGet(bank.URL)
		}
		for i := range question.Judge {
			question.Judge[i].AnswerExternalGet(bank.URL)
		}
		for i := range question.Fill {
			question.Fill[i].AnswerExternalGet(bank.URL)
		}
		for i := range question.Short {
			question.Short[i].AnswerExternalGet(bank.URL)
		}
		for i := range question.TermExplanation {
			question.TermExplanation[i].AnswerExternalGet(bank.URL)
		}
		for i := range question.Essay {
			question.Essay[i].AnswerExternalGet(bank.URL)
		}
		for i := range question.Matching {
			question.Matching[i].AnswerExternalGet(bank.URL)
		}
		for i := range question.Other {
			question.Other[i].AnswerExternalGet(bank.URL)
		}
		return nil
	case core.AnswerSourceAI:
		ai := c.Config.Options.AI
		if ai == nil || strings.TrimSpace(ai.APIKey) == "" {
			return core.Invalid(core.XueXiTong, "chapter_test.answer", "AI config is required")
		}
		aiType := ctype.AiType(ai.Type)
		title := firstNonEmpty(question.Title, "章节测验")
		for i := range question.Choice {
			message := agg.AIProblemMessage(title, question.Choice[i].Type.String(), api.ExamTurn{XueXChoiceQue: question.Choice[i]})
			question.Choice[i].AnswerAIGet(c.cache.UserID, ai.URL, ai.Model, aiType, message, ai.APIKey)
		}
		for i := range question.Judge {
			message := agg.AIProblemMessage(title, question.Judge[i].Type.String(), api.ExamTurn{XueXJudgeQue: question.Judge[i]})
			question.Judge[i].AnswerAIGet(c.cache.UserID, ai.URL, ai.Model, aiType, message, ai.APIKey)
		}
		for i := range question.Fill {
			message := agg.AIProblemMessage(title, question.Fill[i].Type.String(), api.ExamTurn{XueXFillQue: question.Fill[i]})
			question.Fill[i].AnswerAIGet(c.cache.UserID, ai.URL, ai.Model, aiType, message, ai.APIKey)
		}
		for i := range question.Short {
			message := agg.AIProblemMessage(title, question.Short[i].Type.String(), api.ExamTurn{XueXShortQue: question.Short[i]})
			question.Short[i].AnswerAIGet(c.cache.UserID, ai.URL, ai.Model, aiType, message, ai.APIKey)
		}
		for i := range question.TermExplanation {
			message := agg.AIProblemMessage(title, question.TermExplanation[i].Type.String(), api.ExamTurn{XueXTermExplanationQue: question.TermExplanation[i]})
			question.TermExplanation[i].AnswerAIGet(c.cache.UserID, ai.URL, ai.Model, aiType, message, ai.APIKey)
		}
		for i := range question.Essay {
			message := agg.AIProblemMessage(title, question.Essay[i].Type.String(), api.ExamTurn{XueXEssayQue: question.Essay[i]})
			question.Essay[i].AnswerAIGet(c.cache.UserID, ai.URL, ai.Model, aiType, message, ai.APIKey)
		}
		for i := range question.Matching {
			message := agg.AIProblemMessage(title, question.Matching[i].Type.String(), api.ExamTurn{XueXMatchingQue: question.Matching[i]})
			question.Matching[i].AnswerAIGet(c.cache.UserID, ai.URL, ai.Model, aiType, message, ai.APIKey)
		}
		for i := range question.Other {
			message := agg.AIProblemMessage(title, question.Other[i].Type.String(), api.ExamTurn{XueXOtherQue: question.Other[i]})
			question.Other[i].AnswerAIGet(c.cache.UserID, ai.URL, ai.Model, aiType, message, ai.APIKey)
		}
		return nil
	default:
		agg.AnswerFixedPattern(question.Choice, question.Judge)
		return nil
	}
}

func chapterTestAnswers(question api.Question) []core.Answer {
	out := make([]core.Answer, 0,
		len(question.Choice)+len(question.Judge)+len(question.Fill)+len(question.Short)+
			len(question.TermExplanation)+len(question.Essay)+len(question.Matching)+len(question.Other))

	for _, item := range question.Choice {
		out = append(out, core.Answer{
			QuestionID: item.Qid,
			Question:   item.Text,
			Type:       item.Type.String(),
			Options:    chapterTestOptionValues(item.Options),
			Value:      item.Answers,
			Raw:        item,
		})
	}
	for _, item := range question.Judge {
		out = append(out, core.Answer{
			QuestionID: item.Qid,
			Question:   item.Text,
			Type:       item.Type.String(),
			Options:    chapterTestOptionValues(item.Options),
			Value:      item.Answers,
			Raw:        item,
		})
	}
	for _, item := range question.Fill {
		out = append(out, core.Answer{
			QuestionID: item.Qid,
			Question:   item.Text,
			Type:       item.Type.String(),
			Value:      item.OpFromAnswer,
			Raw:        item,
		})
	}
	for _, item := range question.Short {
		out = append(out, core.Answer{
			QuestionID: item.Qid,
			Question:   item.Text,
			Type:       item.Type.String(),
			Raw:        item,
		})
	}
	for _, item := range question.TermExplanation {
		out = append(out, core.Answer{
			QuestionID: item.Qid,
			Question:   item.Text,
			Type:       item.Type.String(),
			Raw:        item,
		})
	}
	for _, item := range question.Essay {
		out = append(out, core.Answer{
			QuestionID: item.Qid,
			Question:   item.Text,
			Type:       item.Type.String(),
			Raw:        item,
		})
	}
	for _, item := range question.Matching {
		out = append(out, core.Answer{
			QuestionID: item.Qid,
			Question:   item.Text,
			Type:       item.Type.String(),
			Value:      item.Answers,
			Raw:        item,
		})
	}
	for _, item := range question.Other {
		out = append(out, core.Answer{
			QuestionID: item.Qid,
			Question:   item.Text,
			Type:       item.Type.String(),
			Raw:        item,
		})
	}
	return out
}

func chapterTestOptionValues(options map[string]string) []string {
	if len(options) == 0 {
		return nil
	}
	keys := make([]string, 0, len(options))
	for key := range options {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make([]string, 0, len(keys))
	for _, key := range keys {
		out = append(out, options[key])
	}
	return out
}
