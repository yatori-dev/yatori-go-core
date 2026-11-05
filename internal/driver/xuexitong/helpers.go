package xuexitong

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/v3/core"
	agg "github.com/yatori-dev/yatori-go-core/v3/internal/aggregation/xuexitong"
	api "github.com/yatori-dev/yatori-go-core/v3/internal/api/xuexitong"
	"github.com/yatori-dev/yatori-go-core/v3/internal/driver/driverutil"
	"github.com/yatori-dev/yatori-go-core/v3/models/ctype"
	qentity "github.com/yatori-dev/yatori-go-core/v3/que-core/qentity"
)

// validateReadReport 校验文档/外链完成接口返回的 status 字段。
func validateReadReport(report string) error {
	status := gojsonq.New().JSONString(report).Find("status")
	if status == nil {
		return fmt.Errorf("学习提交返回异常：%s", summarizeReport(report))
	}
	if flag, ok := status.(bool); !ok || !flag {
		return fmt.Errorf("学习提交失败：%s", summarizeReport(report))
	}
	return nil
}

func summarizeReport(report string) string {
	text := strings.Join(strings.Fields(report), " ")
	if len(text) > 200 {
		text = text[:200] + "..."
	}
	return text
}

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

// videoFaceCaptureDue 判断当前播放位置是否已到视频随机人脸抓拍时间。
// randomFaceCaptureTimeList 为空时，只有播放到结尾才尝试采集。
func videoFaceCaptureDue(p *api.PointVideoDto, position int) bool {
	if p == nil || position <= 0 {
		return false
	}
	if len(p.RandomFaceCaptureTimes) == 0 {
		return p.Duration > 0 && position >= p.Duration
	}
	for _, captureTime := range p.RandomFaceCaptureTimes {
		if captureTime > 0 && position >= captureTime {
			return true
		}
	}
	return false
}

const (
	mobileCardTTL         = 2 * time.Minute
	mobileCardMinInterval = 3 * time.Second
)

type mobileCardCacheEntry struct {
	card any
	at   time.Time
}

var (
	mobileCardCache     sync.Map
	mobileCardRateMu    sync.Mutex
	mobileCardLastFetch = map[string]time.Time{}
)

// fetchMobileCardCached 缓存并限速手机端章节卡片请求。
//
// 学习通在多任务点模式下会同时开多个 Client，每个任务点都去拉一次
// 同一个卡片时很容易触发图片验证码。这里按账号串行化卡片请求，
// 并对同一张卡片做短 TTL 缓存，显著降低触发验证码的概率。
func fetchMobileCardCached(cache *api.XueXiTUserCache, classID, courseID, knowledgeID, cardIndex, cpi int) (any, error) {
	key := fmt.Sprintf("%s|%d|%d|%d|%d", cache.Name, classID, courseID, knowledgeID, cardIndex)
	if card, ok := loadMobileCard(key); ok {
		return card, nil
	}

	mobileCardRateMu.Lock()
	defer mobileCardRateMu.Unlock()

	if last, ok := mobileCardLastFetch[cache.Name]; ok {
		if wait := mobileCardMinInterval - time.Since(last); wait > 0 {
			time.Sleep(wait)
		}
	}
	mobileCardLastFetch[cache.Name] = time.Now()

	// 等待期间别的 worker 可能已经拉过同一张卡片。
	if card, ok := loadMobileCard(key); ok {
		return card, nil
	}

	card, _, err := agg.PageMobileChapterCardAction(cache, classID, courseID, knowledgeID, cardIndex, cpi)
	if err != nil {
		return nil, err
	}
	mobileCardCache.Store(key, mobileCardCacheEntry{card: card, at: time.Now()})
	return card, nil
}

func loadMobileCard(key string) (any, bool) {
	value, ok := mobileCardCache.Load(key)
	if !ok {
		return nil, false
	}
	entry, ok := value.(mobileCardCacheEntry)
	if !ok || time.Since(entry.at) >= mobileCardTTL {
		mobileCardCache.Delete(key)
		return nil, false
	}
	return entry.card, true
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
