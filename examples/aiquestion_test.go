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
func TestSingleChoiceAIQuestion(t *testing.T) {
	utils.YatoriCoreInit()
	setup()
	aiSetting := global.Config.Setting.AiSetting
	matchingTurn := entity.ExamTurn{
		XueXChoiceQue: entity.ChoiceQue{
			Type:    qtype.SingleChoice,
			Qid:     "",
			Text:    "21.[单选题] 下列国民党右派制造的反革命活动（事变），其先后顺序是(       )①“七·一五”反革命政变②“四·一二”反革命政变③中山舰事件④西山会议",
			Options: map[string]string{"A": "①②③④", "B": "②③①④", "C": "④①②③", "D": "④③②①"},
		},
	}
	aiMessage := xuexitong.AIProblemMessage("考试", qtype.SingleChoice.String(), matchingTurn)
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

// 学习通AI答题统一转换测试
func TestXXTAIQuestion(t *testing.T) {
	utils.YatoriCoreInit()
	setup()
	aiSetting := global.Config.Setting.AiSetting
	matchingTurn := entity.ExamTurn{
		XueXEssayQue: entity.EssayQue{
			Type: qtype.Essay,
			Qid:  "",
			Text: "电影《八佰》是展现1937年沪会战期间四行仓库保卫战的一部历史战争片！电影中表现了国防动员的重要性。请在本讨论主题中谈谈你对电影中展示出的，国防动员各要素重要性的认识！！",
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
