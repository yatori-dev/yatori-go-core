package utils

import (
	"encoding/json"
	"gorm.io/gorm"
	"io"
	"net/http"
	"strings"
	"time"
)

// 题目结构体
type Problem struct {
	gorm.Model
	Hash    string   //题目信息的Hash
	Type    string   //题目类型，比如单选，多选，简答题等
	Content string   //题目内容
	Options []string //题目选项，一般选择题才会有该字段
	Answer  []string //答案
	Json    string   //json形式原内容
}
type Answer struct {
	Type    string
	Answers []string
}

// 用于请求外部题库接口使用
func (problem *Problem) ApiQueRequest(url string, retry int, err error) (Answer, error) {
	if retry <= 0 {
		return Answer{}, err
	}
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	data, _ := json.Marshal(problem)
	resp, err := client.Post(url, "application/json", strings.NewReader(string(data)))
	if err != nil {
		return problem.ApiQueRequest(url, retry-1, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var answers Answer
	err1 := json.Unmarshal(body, &answers)
	if err != nil {
		problem.ApiQueRequest(url, retry-1, err1)
	}
	return answers, nil
}

// 用于检测是否能够正常访问题库接口
func CheckApiQueRequest(url string, retry int, err error) error {
	if retry <= 0 {
		return err
	}
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	problem := Problem{
		Hash:    "",
		Type:    "多选",
		Content: "1、According to the successful salesperson Summer, what are the principles\n\nwe should follow in business writing?",
		Options: []string{
			"A.politeness",
			"B.correct",
			"C.clear",
			"D.concise",
		},
		Answer: []string{
			//"A", "B", "C", "D",
		},
		Json: "null",
	}
	data, _ := json.Marshal(problem)
	resp, err := client.Post(url, "application/json", strings.NewReader(string(data)))
	if err != nil {
		return CheckApiQueRequest(url, retry-1, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var answers Answer
	err1 := json.Unmarshal(body, &answers)
	if err1 != nil {
		return CheckApiQueRequest(url, retry-1, err1)
	}
	return nil
}
