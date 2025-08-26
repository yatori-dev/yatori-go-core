package que_core

import (
	"encoding/json"
	"github.com/yatori-dev/yatori-go-core/que-core/qtype"
	"github.com/yatori-dev/yatori-go-core/utils/log"
	"io/ioutil"
	"net/http"
	"net/url"
)

// 题库答题
type QuesBank struct{}

// GetQuestionAnswers 获取题库中的对应题目答案
func (QuesBank) GetQuestionAnswers(queType qtype.QueType, content string, options []string) []string {
	// TODO 这里对应QuesBank的url 但是目前console 设置没有读取配置题库的url设置
	resp, err := http.PostForm("", url.Values{
		"type":    {queType.String()},
		"content": {content},
		"options": options,
	})
	if err != nil {
		log.Print(log.DEBUG, "Error sending request:", err)
		return nil
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var jsonContent map[string]interface{}
	err = json.Unmarshal(body, &jsonContent)
	if err != nil {
		log.Print(log.DEBUG, "Error parsing JSON:", err)
		return nil
	}
	return jsonContent["Answers"].([]string)
}
