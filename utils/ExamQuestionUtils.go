package utils

import (
	"encoding/json"
	"gorm.io/gorm"
	"io"
	"net/http"
	"strings"
)

// 题目结构体
type Problem struct {
	gorm.Model
	Hash    string   //题目信息的Hash
	Type    string   //题目类型，比如单选，多选，简答题等
	Content string   //题目内容
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
	data, _ := json.Marshal(problem)
	resp, _ := http.Post(url, "application/json", strings.NewReader(string(data)))
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var answers Answer
	err1 := json.Unmarshal(body, &answers)
	if err != nil {
		problem.ApiQueRequest(url, retry-1, err1)
	}
	return answers, nil
}
