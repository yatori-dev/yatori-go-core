package examples

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/yatori-dev/yatori-go-core/aggregation/xuexitong"
	"github.com/yatori-dev/yatori-go-core/api/entity"
	"github.com/yatori-dev/yatori-go-core/global"
	"github.com/yatori-dev/yatori-go-core/que-core/aiq"
	"github.com/yatori-dev/yatori-go-core/que-core/qtype"
	"github.com/yatori-dev/yatori-go-core/utils"
)

// 学习通AI答题统一转换测试
func TestXXTAIQuestion(t *testing.T) {
	utils.YatoriCoreInit()
	setup()
	aiSetting := global.Config.Setting.AiSetting
	matchingTurn := entity.ExamTurn{
		XueXEssayQue: entity.EssayQue{
			Type: qtype.Essay,
			Qid:  "",
			Text: "1.[论述题] 如何认识富国与强军的辩证统一关系？",
		},
		//XueXMatchingQue: entity.MatchingQue{
		//	Type:    qtype.Matching,
		//	Qid:     "",
		//	Text:    "短语与释义连线",
		//	Options: []string{"abide by", "account for", "come across"},
		//	Selects: []string{"偶然遇见，碰到", "遵守，遵循", "导致，引起"},
		//	Answers: nil,
		//},
	}
	aiMessage := xuexitong.AIProblemMessage("考试", qtype.Essay.String(), matchingTurn)
	fmt.Println(aiMessage)
	aiResult, err := aiq.AggregationAIApi(aiSetting.AiUrl, aiSetting.Model, aiSetting.AiType, aiMessage, aiSetting.APIKEY)
	if err != nil {
		t.Error(err)
	}
	var resultJson []string
	err = json.Unmarshal([]byte(aiResult), &resultJson)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(resultJson)
}
