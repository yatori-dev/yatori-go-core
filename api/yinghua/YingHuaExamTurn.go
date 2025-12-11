package yinghua

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/yatori-dev/yatori-go-core/api/xuexitong"
	"github.com/yatori-dev/yatori-go-core/que-core/aiq"
	"github.com/yatori-dev/yatori-go-core/que-core/qentity"
)

// 题目转换
func TurnExamTopic(examHtml string) []xuexitong.YingHuaExamTopic {
	var topics = make([]xuexitong.YingHuaExamTopic, 0)

	//exchangeTopics := entity.YingHuaExamTopics{
	//	YingHuaExamTopics: make(map[string]entity.YingHuaExamTopic),
	//}

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
		//var selects []entity.TopicSelect
		var selects []string

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
				//selectValue := selectMatch[2]
				//selectNum := selectMatch[3]
				selectText := selectMatch[4]
				//selects = append(selects, entity.TopicSelect{
				//	Value: selectValue,
				//	Num:   selectNum,
				//	Text:  selectText,
				//})
				selects = append(selects, selectText)
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
				//selects = append(selects, entity.TopicSelect{
				//	Value: answerId,
				//	Num:   answerId,
				//	Text:  "",
				//})
				selects = append(selects, answerId)
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
		examTopic := xuexitong.YingHuaExamTopic{
			AnswerId: topicMap[num],
			Index:    num,
			Source:   source,
			Question: qentity.Question{
				Type:    turnTypeStr(tag),
				Content: content,
				Options: selects,
			},
			//Content:  content,
			//Type:    tag,
			//Selects: selects,
		}

		// Add the topic to the ExamTopics map
		//exchangeTopics.YingHuaExamTopics[topicMap[num]] = examTopic
		topics = append(topics, examTopic)
	}

	return topics
}

// 转标准类型
func turnTypeStr(origin string) string {
	switch origin {
	case "单选":
		return "单选题"
	case "多选":
		return "多选题"
	case "判断":
		return "判断题"
	case "填空":
		return "填空题"
	case "简答":
		return "简答题"
	}
	return "其他"
}

// Deprecated: 此方法将在未来版本中删除
// 组装AI问题消息
func AIProblemMessage(testPaperTitle string, question qentity.Question) aiq.AIChatMessages {
	topicType := question.Type
	problem := `试卷名称：` + testPaperTitle + `
题目类型：` + topicType + `
题目内容：` + question.Content + "\n"

	//选择题
	if topicType == "单选" {
		for _, v := range question.Options {
			//problem += v.Num + v.Text + "\n"
			problem += v + "\n"
		}
		return aiq.AIChatMessages{Messages: []aiq.Message{
			{
				Role:    "system",
				Content: `接下来你只需要回答选项对应内容即可，不能回答任何选项无关的任何内容，包括解释以及标点符也不需要。`,
			},
			{
				Role:    "system",
				Content: `就算你不知道选什么也随机选输出其选项内容，回答的格式一定要严格为单个数组格式，比如：["选项1"]，注意回复的时候不要带选项字母，你只需回复答案对应格式内容即可，无需回答任何解释！！！`,
			},
			{
				Role: "system",
				Content: `比如：` + `
				试卷名称：考试
				题目类型：单选
				题目内容：新中国是什么时候成立的
				A. 1949年10月5日
				B. 1949年10月1日
				C. 1949年09月1日
				D. 2002年10月1日
				` + `
				那么你应该回答选项B的内容：“["1949年10月1日"]”
				`,
			},
			{
				Role:    "user",
				Content: problem,
			},
		}}
	} else if topicType == "多选" {
		for _, v := range question.Options {
			//problem += v.Num + v.Text + "\n"
			problem += v + "\n"
		}
		return aiq.AIChatMessages{Messages: []aiq.Message{
			{
				Role:    "system",
				Content: `接下来你只需要回答选项对应内容即可，不能回答任何选项无关的任何内容，包括解释以及标点符也不需要。`,
			},
			{
				Role:    "system",
				Content: `就算你不知道选什么也随机选输出其选项内容，回答的格式一定要严格为单个数组格式，比如：["选项1","选项2"]，注意回复的时候不要带选项字母，你只需回复答案对应格式内容即可，无需回答任何解释！！！`,
			},
			{
				Role: "system",
				Content: `比如：` + `
				试卷名称：考试
				题目类型：多选选
				题目内容：马克思关于资本积累的学说是剩余价值理论的重要组成部分。资本积累是
				A. 资本主义扩大再生产的源泉
				B. 资本有机构成呈现不断降低趋势的根本原因
				C. 社会财富占有两极分化的重要原因
				D. 资本主义社会失业现象产生的根源
				` + `
				那么你应该回答选项A、B、D的内容：“["资本主义扩大再生产的源泉","社会财富占有两极分化的重要原因","资本主义社会失业现象产生的根源"]”
				`,
			},
			{
				Role:    "user",
				Content: problem,
			},
		}}
	} else if topicType == "判断" {
		for _, v := range question.Options {
			//problem += v.Num + v.Text + "\n"
			problem += v + "\n"
		}
		return aiq.AIChatMessages{Messages: []aiq.Message{
			{
				Role:    "system",
				Content: `接下来你只需要回答“正确”或者“错误”即可，不能回答任何无关的内容，包括解释以及标点符也不需要。`,
			},
			{
				Role:    "system",
				Content: `就算你不知道选什么也随机选输出其选项内容，回答的格式一定要严格为单个数组格式，比如：["正确"]，注意回复的时候不要带选项字母，你只需回复答案对应格式内容即可，无需回答任何解释！！！`,
			},
			{
				Role: "system",
				Content: `比如：` + `
				试卷名称：考试
				题目类型：判断
				题目内容：新中国是什么时候成立是1949年10月1日吗？
				A. 正确
				B. 错误
				` + `
				那么你应该回答选项A的内容：“["正确"]”
				`,
			},
			{
				Role:    "user",
				Content: problem,
			},
		}}
	} else if topicType == "填空" { //填空题
		return aiq.AIChatMessages{Messages: []aiq.Message{
			{
				Role:    "system",
				Content: `其中，“（answer_数字）”相关字样的地方是你需要填写答案的地方，现在你只需要按顺序回复我对应每个填空项的答案即可，回答的格式一定要严格为单个数组格式，比如["答案1","答案2"]其他不符合格式的内容无需回复。你只需回复答案对应格式内容即可，无需回答任何解释！！！`,
			},
			{
				Role:    "user",
				Content: problem,
			},
		}}
	} else if topicType == "简答" { //简答
		return aiq.AIChatMessages{Messages: []aiq.Message{
			{
				Role:    "system",
				Content: `这是一个简答题，现在你只需要回复我对应简答题答案即可，回答的格式一定要严格为单个数组格式，比如["答案"]，但是注意你只需要把所有答案填写在一个元素项里面就行，别分开，比如你不能["xxx","zzz"]这样写，你只能["xxxzzz"]这样写，其他不符合格式的内容无需回复。你只需回复答案对应格式内容即可，无需回答任何解释！！！`,
			},
			{
				Role:    "user",
				Content: problem,
			},
		}}
	}
	return aiq.AIChatMessages{Messages: []aiq.Message{}}
}
