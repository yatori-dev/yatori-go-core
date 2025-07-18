package external

import (
	"encoding/json"
	"github.com/yatori-dev/yatori-go-core/que-core/qentity"
	"io"
	"net/http"
	"strings"
	"time"
)

// 用于请求外部题库接口使用
func ApiQueRequest(problem qentity.Question, url string, retry int, err error) (*qentity.ResultQuestion, error) {
	if retry <= 0 {
		return nil, err
	}
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	data, _ := json.Marshal(problem)
	resp, err := client.Post(url, "application/json", strings.NewReader(string(data)))
	if err != nil {
		return ApiQueRequest(problem, url, retry-1, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var answers qentity.ResultQuestion
	err1 := json.Unmarshal(body, &answers)
	if err1 != nil {
		return nil, err1
	}
	return &answers, nil
}

// 用于检测是否能够正常访问题库接口
func CheckApiQueRequest(url string, retry int, err error) error {
	if retry <= 0 {
		return err
	}
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	problem := qentity.Question{
		Type:    "多选题",
		Content: "1、According to the successful salesperson Summer, what are the principles\n\nwe should follow in business writing?",
		Options: []string{
			"A.politeness",
			"B.correct",
			"C.clear",
			"D.concise",
		},
		Answers: []string{
			//"A", "B", "C", "D",
		},
	}
	data, _ := json.Marshal(problem)
	resp, err := client.Post(url, "application/json", strings.NewReader(string(data)))
	if err != nil {
		return CheckApiQueRequest(url, retry-1, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var answers qentity.ResultQuestion
	err1 := json.Unmarshal(body, &answers)
	if err1 != nil {
		return err1
	}
	return nil
}
