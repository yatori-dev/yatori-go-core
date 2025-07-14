package examples

import (
	"crypto/md5"
	"fmt"
	"github.com/yatori-dev/yatori-go-core/que-core/qtype"
	"github.com/yatori-dev/yatori-go-core/que-core/questionbank"
	"testing"
)

// 测试题库缓存插入
func TestQuestionBackInsert(t *testing.T) {
	init, err := questionbank.QuestionBackInit()
	if err != nil {
		t.Error(err)
	}

	qtype := qtype.FillInTheBlank
	qContent := "这是一个填空题"
	md5 := md5.Sum([]byte(fmt.Sprintf("%s-%s", qtype.String(), qContent)))
	// 插入题目缓存
	question := questionbank.Question{

		Md5:     fmt.Sprintf("%x", md5),
		Type:    qtype.String(),
		Content: qContent,
		Answers: `{"答案1", "答案2"}`,
	}

	result := question.SelectsForTypeAndContent(init)
	for _, v := range result {
		fmt.Println(v)
	}

	err = question.InsertIfNot(init)
	if err != nil {
		panic(err)
	}
	fmt.Println(question)
}

// 题目自动缓存逻辑
func TestQuestionBack(t *testing.T) {
	db, err := questionbank.QuestionBackInit()
	if err != nil {
		t.Error(err)
	}
	question := questionbank.Question{
		Type:    "填空题",
		Content: "这是一个示例题目",
		Answers: `{"答案1", "答案2"}`,
	}
	content := question.SelectsForTypeAndContent(db)
	if len(content) == 0 { //如果没有题目则触发缓存逻辑
		err := question.Insert(db)
		if err != nil {
			t.Error(err)
		}
	}

}

//// 转为英华转缓存用的Question
//func YingHuaTopicTurnQuestion(topic yinghua.YingHuaExamTopic) questionbank.Question {
//	question := questionbank.Question{
//		Type:    topic.Type,
//		Content: topic.Content,
//		Answers: topic.Answers,
//	}
//	return question
//}
//
//// Question转英华
//func QuestionTurnYingHuaTopic(qu questionbank.Question) yinghua.YingHuaExamTopic {
//	return yinghua.YingHuaExamTopic{}
//}
