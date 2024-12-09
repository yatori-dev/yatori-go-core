package yinghua

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/yatori-dev/yatori-go-core/utils"
)

// ExamTopics holds a map of ExamTopic indexed by answerId
type YingHuaExamTopics struct {
	YingHuaExamTopics map[string]YingHuaExamTopic
}

// ExamTopic represents a single exam question
type YingHuaExamTopic struct {
	AnswerId string        `json:"answerId"`
	Index    string        `json:"index"`
	Source   string        `json:"source"`
	Content  string        `json:"content"`
	Type     string        `json:"type"`
	Selects  []TopicSelect `json:"selects"`
	Answers  string        `json:"answers"`
}

// TopicSelect represents a possible answer choice
type TopicSelect struct {
	Value string `json:"value"`
	Num   string `json:"num"`
	Text  string `json:"text"`
}

// 题目转换
func TurnExamTopic(examHtml string) YingHuaExamTopics {
	exchangeTopics := YingHuaExamTopics{
		YingHuaExamTopics: make(map[string]YingHuaExamTopic),
	}

	// Regular expression to extract the topic index and answerId
	topicPattern := `<li>[ \f\n\r\t\v]*<a data-id="([^\"]*)"[ \f\n\r\t\v]*href="[^\"]*"[ \f\n\r\t\v]*class="[^\"]*"[ \f\n\r\t\v]*id="[^\"]*"[ \f\n\r\t\v]*data-index="[^\"]*"[ \f\n\r\t\v]*onclick="[^\"]*">([^<]*)</a>[ \f\n\r\t\v]*</li>`
	topicRegexp := regexp.MustCompile(topicPattern)
	topicMap := make(map[string]string)

	// Extract answerId and index
	matches := topicRegexp.FindAllStringSubmatch(examHtml, -1)
	for _, match := range matches {
		answerId := match[1]
		index := match[2]
		topicMap[index] = answerId
	}

	// 匹配所有考试题目
	formPattern := `<form method="post" action="/api/[^/]*\/submit">([\w\W]*?)</form>`
	formRegexp := regexp.MustCompile(formPattern)

	// Extract the form contents
	formMatches := formRegexp.FindAllStringSubmatch(examHtml, -1)
	for _, formMatch := range formMatches {
		topicHtml := formMatch[1] //截取题目对应单个题目部分html

		// Extracting topic number, type, and source
		topicNumPattern := `<span class="num">[\D]*?([\d]+)`
		topicNumRegexp := regexp.MustCompile(topicNumPattern)
		topicNumMatcher := topicNumRegexp.FindStringSubmatch(topicHtml)

		var num, tag, source, content string
		var selects []TopicSelect

		if len(topicNumMatcher) > 0 {
			num = topicNumMatcher[1]
		}

		tagPattern := `<span class="tag">([\s\S]*?)</span>`
		tagRegexp := regexp.MustCompile(tagPattern)
		tagMatcher := tagRegexp.FindStringSubmatch(topicHtml)
		if len(tagMatcher) > 0 {
			tag = tagMatcher[1]
		}

		sourcePattern := `<span[ \f\n\r\t\v]*class="txt">[^\d]*([\d]*)[^分]*分[^<]*</span>`
		sourceRegexp := regexp.MustCompile(sourcePattern)
		sourceMatcher := sourceRegexp.FindStringSubmatch(topicHtml)
		if len(sourceMatcher) > 0 {
			source = sourceMatcher[1]
		}

		// Extract the question content based on the type of the question (Single choice, Multiple choice, Judgment)
		if tag == "单选" || tag == "多选" || tag == "判断" {
			contentPattern := `<div[ \f\n\r\t\v]*class="content"[ \f\n\r\t\v]*style="[^\"]*">([\s\S]*?)</div>`
			contentRegexp := regexp.MustCompile(contentPattern)
			contentMatcher := contentRegexp.FindStringSubmatch(topicHtml)
			if len(contentMatcher) > 0 {
				content = contentMatcher[1]
			}

			// Extract possible selections for the topic
			selectPattern := `<li>[^<]*<label>[^<]*<input type="([^"]*)"[^v]*value="([^"]*)"[ \f\n\r\t\v]*[checked="checked"]*[ \f\n\r\t\v]*class="[^"]*"[ \f\n\r\t\v]*name="[^"]*">[ \f\n\r\t\v]*<span class="num">([^<]*)</span>[ \f\n\r\t\v]*<span[ \f\n\r\t\v]*class="txt">([^<]*)</span>[ \f\n\r\t\v]*</label>[ \f\n\r\t\v]*</li>`
			selectRegexp := regexp.MustCompile(selectPattern)
			selectMatches := selectRegexp.FindAllStringSubmatch(topicHtml, -1)
			for _, selectMatch := range selectMatches {
				selectValue := selectMatch[2]
				selectNum := selectMatch[3]
				selectText := selectMatch[4]
				selects = append(selects, TopicSelect{
					Value: selectValue,
					Num:   selectNum,
					Text:  selectText,
				})
			}
			// Clean up content (strip illegal strings)
			content = strings.ReplaceAll(content, "<p>", "")
			content = strings.ReplaceAll(content, "</p>", "\n")
			content = strings.ReplaceAll(content, "&nbsp;", "")
		} else if tag == "填空" {
			contentPattern := `<div[ \f\n\r\t\v]*class="content"[ \f\n\r\t\v]*style="[^\"]*">([\s\S]*?)</div>`
			contentRegexp := regexp.MustCompile(contentPattern)
			contentMatcher := contentRegexp.FindStringSubmatch(topicHtml)
			if len(contentMatcher) > 0 {
				content = contentMatcher[1]
			}

			// Regular expression to extract fill-in-the-blank fields
			//fmt.Println(topicHtml)
			//fmt.Println("若打印出此数据请不要马上关闭，立即复制给作者。因为可能是傻逼英华引起的BUG，需要用户提供以上内容")
			//fillRegexp := regexp.MustCompile(`<input ((?<!answer).)+answer_(\d)+((?<!>).)+>`)
			fillRegexp := regexp.MustCompile(`<input class="[^"]*" autocomplete="[^"]*" autocomplete="[^"]*" type="[^"]*" style="[^"]*" name="answer_([^"]*)" value="[^"]*"/>`)
			fillMatches := fillRegexp.FindAllStringSubmatch(topicHtml, -1)
			for _, fillMatch := range fillMatches {
				answerId := fillMatch[1]
				selects = append(selects, TopicSelect{
					Value: answerId,
					Num:   answerId,
					Text:  "",
				})
			}

			// Replace fill-in-the-blank code
			//codePattern := "<code>((?<!answer).)+answer_(\\d)+((?<!</code>).)+</code>"
			codeRegexp := regexp.MustCompile(`<code> class="[^"]*" autocomplete="[^"]*" autocomplete="[^"]*" type="[^"]*" style="[^"]*" name="answer_([^"]*)" value="[^"]*"[^<]*</code>`)
			codeMatches := codeRegexp.FindAllStringSubmatch(content, -1)
			for _, codeMatch := range codeMatches {
				answerId := codeMatch[1]
				content = strings.ReplaceAll(content, codeMatch[0], fmt.Sprintf("（answer_%s）", answerId))
			}

			// Clean up content
			content = strings.ReplaceAll(content, "<p>", "")
			content = strings.ReplaceAll(content, "</p>", "\n")
			content = strings.ReplaceAll(content, "&nbsp;", "")
		} else if tag == "简答" {
			contentPattern := `<div[ \f\n\r\t\v]*class="content"[ \f\n\r\t\v]*style="[^\"]*">([\s\S]*?)</div>`
			contentRegexp := regexp.MustCompile(contentPattern)
			contentMatcher := contentRegexp.FindStringSubmatch(topicHtml)
			if len(contentMatcher) > 0 {
				content = contentMatcher[1]
			}

			//fmt.Println(topicHtml)
		}

		// Construct the ExamTopic
		examTopic := YingHuaExamTopic{
			AnswerId: topicMap[num],
			Index:    num,
			Source:   source,
			Content:  content,
			Type:     tag,
			Selects:  selects,
		}

		// Add the topic to the ExamTopics map
		exchangeTopics.YingHuaExamTopics[topicMap[num]] = examTopic
	}

	return exchangeTopics
}

// 组装AI问题消息
func AIProblemMessage(testPaperTitle string, examTopic YingHuaExamTopic) utils.AIChatMessages {
	problem := `试卷名称：` + testPaperTitle + `
题目类型：` + examTopic.Type + `
题目内容：` + examTopic.Content + "\n"

	//选择题
	if examTopic.Type == "单选" || examTopic.Type == "多选" {
		for _, v := range examTopic.Selects {
			problem += v.Num + v.Text + "\n"
		}
		return utils.AIChatMessages{Messages: []utils.Message{
			{
				Role:    "user",
				Content: `接下来你只需要回答选项对应内容即可，不能回答任何选项无关的任何内容，包括解释以及标点符也不需要。`,
			},
			{
				Role:    "user",
				Content: `就算你不知道选什么也随机选输出其选项内容，回答的格式为数组格式，比如：["选项1","选项二"]，注意回复的时候不要带选项字母，只需照抄选项对应的文字即可，`,
			},
			{
				Role:    "user",
				Content: problem,
			},
		}}
	} else if examTopic.Type == "判断" {
		return utils.AIChatMessages{Messages: []utils.Message{
			{
				Role:    "user",
				Content: `接下来你只需要回答“正确”或者“错误”即可，不能回答任何无关的内容，包括解释以及标点符也不需要。`,
			},
			{
				Role:    "user",
				Content: `就算你不知道选什么也随机选输出其选项内容，回答的格式为数组格式，比如：["正确"]，注意回复的时候不要带选项字母，只需照抄选项对应的文字即可，`,
			},
			{
				Role:    "user",
				Content: problem,
			},
		}}
	} else if examTopic.Type == "填空" { //填空题
		return utils.AIChatMessages{Messages: []utils.Message{
			{
				Role:    "user",
				Content: `其中，“（answer_数字）”相关字样的地方是你需要填写答案的地方，现在你只需要按顺序回复我对应每个填空项的答案即可，并且采用json格式的回复方式，比如["答案1","答案2"]其他不符合json格式的内容无需回复。你只需回复答案对应json，无需回答任何解释！！！`,
			},
			{
				Role:    "user",
				Content: problem,
			},
		}}
	} else if examTopic.Type == "简答" { //简答
		return utils.AIChatMessages{Messages: []utils.Message{
			{
				Role:    "user",
				Content: `这是一个简答题，现在你只需要回复我对应简答题答案即可，采用json格式的回复方式，比如["答案"]，但是注意你只需要把所有答案填写在一个元素项里面就行，别分开，比如你不能["xxx","zzz"]这样写，你只能["xxxzzz"]这样写，其他不符合json格式的内容无需回复。你只需回复答案对应json，无需回答任何解释！！！`,
			},
			{
				Role:    "user",
				Content: problem,
			},
		}}
	}
	return utils.AIChatMessages{Messages: []utils.Message{}}
}
