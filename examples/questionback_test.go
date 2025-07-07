package examples

import (
	"fmt"
	"github.com/yatori-dev/yatori-go-core/que-core/questionback"
	"testing"
)

// 测试题库缓存插入
func TestQuestionBackInsert(t *testing.T) {
	init, err := questionback.QuestionBackInit()
	if err != nil {
		t.Error(err)
	}
	// 插入题目缓存
	question := questionback.Question{
		Type:    "填空题",
		Content: "这是一个示例题目",
		Answers: `{"答案1", "答案2"}`,
	}
	content := question.QuestionBackSelectsForTypeAndContent(init)

	for _, v := range content {
		fmt.Println(v)
	}
	//question.QuestionBackInsert(init)

}

// 题目自动缓存逻辑
func TestQuestionBack(t *testing.T) {
	db, err := questionback.QuestionBackInit()
	if err != nil {
		t.Error(err)
	}
	question := questionback.Question{
		Type:    "填空题",
		Content: "这是一个示例题目",
		Answers: `{"答案1", "答案2"}`,
	}
	content := question.QuestionBackSelectsForTypeAndContent(db)
	if len(content) == 0 { //如果没有题目则触发缓存逻辑
		err := question.QuestionBackInsert(db)
		if err != nil {
			t.Error(err)
		}
	}

}

//// 转为英华转缓存用的Question
//func YingHuaTopicTurnQuestion(topic yinghua.YingHuaExamTopic) questionback.Question {
//	question := questionback.Question{
//		Type:    topic.Type,
//		Content: topic.Content,
//		Answers: topic.Answers,
//	}
//	return question
//}
//
//// Question转英华
//func QuestionTurnYingHuaTopic(qu questionback.Question) yinghua.YingHuaExamTopic {
//	return yinghua.YingHuaExamTopic{}
//}
