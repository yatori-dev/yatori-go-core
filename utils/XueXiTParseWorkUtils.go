package utils

import (
	"fmt"

	"github.com/PuerkitoBio/goquery"
)

// QuestionSet 结构体用于存储每个 div.Py-mian1 的原始 HTML 内容
type QuestionSet struct {
	ID   string // 可以根据需要添加其他字段，例如题目ID
	HTML string
}

// ParseQuestionSets 函数用于提取所有 div.Py-mian1 节点的原始 HTML 内容
func ParseQuestionSets(doc *goquery.Document) []QuestionSet {
	questionNodes := doc.Find("div.Py-mian1")
	var questionSets []QuestionSet

	questionNodes.Each(func(i int, questionNode *goquery.Selection) {
		// 获取 div.Py-mian1 的 data 属性（如果有）
		dataAttr, exists := questionNode.Attr("data")
		if !exists {
			dataAttr = fmt.Sprintf("question_%d", i+1)
		}

		// 提取原始 HTML 内容
		htmlContent, _ := questionNode.Html()

		// 添加到结果切片
		questionSets = append(questionSets, QuestionSet{
			ID:   dataAttr,
			HTML: htmlContent,
		})
	})

	return questionSets
}

// 用于截取作业信息中用于回答回答题目的关键参数
func ParseWorkInform(doc *goquery.Document) (map[string]interface{}, error) {
	dataMap := make(map[string]interface{})
	findList := []string{"userId", "courseId", "courseId", "classId", "api", "workAnswerId",
		"answerId", "totalQuestionNum", "fullScore", "knowledgeid", "oldSchoolId",
		"oldWorkId", "jobid", "workRelationId", "enc", "enc_work", "userId", "cpi",
		"workTimesEnc", "randomOptions", "cfid", "uploadEnc", "workId"}
	for _, find := range findList {
		doc.Find("input#" + find).Each(func(i int, s *goquery.Selection) {
			val, exists := s.Attr("value")
			if exists {
				dataMap[find] = val
			}
		})
	}
	//截取试卷标题
	doc.Find(".chapter-title").Each(func(i int, s *goquery.Selection) {
		val, exists := s.Attr("workname")
		if exists {
			dataMap["title"] = val
		}
	})
	return dataMap, nil
}
