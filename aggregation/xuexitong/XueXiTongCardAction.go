package xuexitong

import "C"
import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/api/entity"
	"github.com/yatori-dev/yatori-go-core/api/xuexitong"
	que_core "github.com/yatori-dev/yatori-go-core/que-core/aiq"
	"github.com/yatori-dev/yatori-go-core/que-core/qtype"
	"github.com/yatori-dev/yatori-go-core/utils"
	log2 "github.com/yatori-dev/yatori-go-core/utils/log"
	"golang.org/x/net/html"
)

// ChapterNotOpened 是未打开章节时的错误类型
type ChapterNotOpened struct{}

func (e ChapterNotOpened) Error() string {
	return "章节未开放"
}

// APIError 是 API 相关错误的一般错误
// 这里之后统一做整合
type APIError struct {
	Message string
}

func (e APIError) Error() string {
	return e.Message
}

func PageMobileChapterCardAction(
	cache *xuexitong.XueXiTUserCache,
	classId, courseId, knowledgeId, cardIndex, cpi int) (interface{}, string, error) {
	cardHtml, err := cache.PageMobileChapterCard(classId, courseId, knowledgeId, cardIndex, cpi)
	var att interface{}

	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch pageMobileChapterCard: %w", err)
	}
	doc, err := html.Parse(bytes.NewReader([]byte(cardHtml)))
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	var scriptContent string
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "script" {
			// Check if the script tag has a type attribute and its value is "text/javascript"
			hasTypeAttr := false
			for _, attr := range n.Attr {
				if attr.Key == "type" && strings.TrimSpace(attr.Val) == "text/javascript" {
					hasTypeAttr = true
					break
				}
			}

			if hasTypeAttr {
				// Check if the script tag has a child node and it's a text node
				if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
					scriptContent = strings.TrimSpace(n.FirstChild.Data)
					return // Exit the traversal as we found what we need
				}
			}
		}

		// Continue traversing the children nodes
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)

	// Define the regex pattern to match window.AttachmentSetting JSON object
	re := regexp.MustCompile(`window\.AttachmentSetting\s*=\s*(\{.*?\});`)
	matches := re.FindStringSubmatch(scriptContent)
	if len(matches) > 1 {
		var attachment interface{}
		if err := json.Unmarshal([]byte(matches[1]), &attachment); err != nil {
			return nil, "", fmt.Errorf("failed to parse JSON: %w", err)
		}
		att = attachment
	} else {
		var blankTips string
		traverse = func(n *html.Node) {
			if n.Type == html.ElementNode && n.Data == "p" {
				for _, attr := range n.Attr {
					if attr.Key == "class" && strings.Contains(attr.Val, "blankTips") {
						blankTips = strings.TrimSpace(n.FirstChild.Data)
						return
					}
				}
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				traverse(c)
			}
		}
		traverse(doc)

		if strings.TrimSpace(blankTips) == "章节未开放！" {
			log.Println("章节未开放")
			return ChapterNotOpened{}, "", nil
		}
		return APIError{Message: blankTips}, "", nil
	}
	// 截取
	reEnc := regexp.MustCompile(`<input type="hidden" id="from" value="[^_]+_[^_]+_[^_]+_([^"]+)"/>`)
	matchesEnc := reEnc.FindStringSubmatch(cardHtml)
	enc := ""
	if len(matchesEnc) > 1 {
		enc = matchesEnc[1]
	}
	//log.Println("Attachment拉取成功")
	return att, enc, nil
}

func VideoDtoFetchAction(cache *xuexitong.XueXiTUserCache, p *entity.PointVideoDto) (bool, error) {
	fetch, err := cache.VideoDtoFetch(p)
	if err != nil {
		log.Println("VideoDtoFetchAction:", err)
		return false, err
	}
	dtoken := gojsonq.New().JSONString(fetch).Find("dtoken").(string)
	duration := gojsonq.New().JSONString(fetch).Find("duration").(float64)

	p.DToken = dtoken
	p.Duration = int(duration)
	titleStr, turnErr := url.QueryUnescape(gojsonq.New().JSONString(fetch).Find("filename").(string))
	//转换
	if turnErr == nil {
		log2.Print(log2.INFO, titleStr, "解码失败")
		p.Title = gojsonq.New().JSONString(fetch).Find("filename").(string)
	} else {
		p.Title = titleStr
	}

	if gojsonq.New().JSONString(fetch).Find("status").(string) == "success" {
		return true, nil
	}
	return false, errors.New("fetch failed")
}

func WorkPageFromAction(cache *xuexitong.XueXiTUserCache, workPoint *entity.PointWorkDto) ([]entity.WorkInputField, error) {
	questionHtml, err := cache.WorkFetchQuestion(workPoint)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch WorkFetchQuestion: %w", err)
	}

	inputPattern := regexp.MustCompile(`<input\s+[^>]*>`)
	attributePattern := regexp.MustCompile(`\b(name|value|type|id)\s*=\s*["']([^"']+)["']`)

	var inputs []entity.WorkInputField

	// Find all matches of <input> tags in the HTML content.
	inputTags := inputPattern.FindAllStringSubmatch(questionHtml, -1)
	for _, tag := range inputTags {
		var inputField entity.WorkInputField
		attributes := attributePattern.FindAllStringSubmatch(tag[0], -1)

		for _, attr := range attributes {
			if len(attr) == 3 { // Ensure we have a key-value pair
				switch strings.ToLower(attr[1]) {
				case "name":
					inputField.Name = attr[2]
				case "value":
					inputField.Value = attr[2]
				case "type":
					inputField.Type = attr[2]
				case "id":
					inputField.ID = attr[2]
				}
			}
		}

		// Include the input field if it has either a name or an id attribute.
		if inputField.Name != "" || inputField.ID != "" {
			inputs = append(inputs, inputField)
		} else {
			fmt.Printf("Skipping input with no name or id attribute: %s\n", tag[0])
		}
	}

	return inputs, nil
}

// cleanText 函数用于净化提取的文本，去除多余的空白字符
func cleanText(text string) string {
	// 去除首尾空白字符
	text = strings.TrimSpace(text)
	// 替换多个连续换行符为单个空格Add commentMore actions
	text = regexp.MustCompile(`\n+`).ReplaceAllString(text, " ")

	// 替换多个连续空格为单个空格
	return regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
}

// WorkInformInputWorkDTO workDTO赋值
func WorkInformInputWorkDTO(informMap map[string]interface{}, question *entity.Question) {
	if v, ok := informMap["title"]; ok {
		question.Title = v.(string)
	}
	if v, ok := informMap["jobid"]; ok {
		question.JobId = v.(string)
	}
	if v, ok := informMap["cpi"]; ok {
		question.Cpi = v.(string)
	}
	if v, ok := informMap["knowledgeid"]; ok {
		question.Knowledgeid = v.(string)
	}
	if v, ok := informMap["userId"]; ok {
		question.UserId = v.(string)
	}
	if v, ok := informMap["workAnswerId"]; ok {
		question.WorkAnswerId = v.(string)
	}
	if v, ok := informMap["answerId"]; ok {
		question.AnswerId = v.(string)
	}
	if v, ok := informMap["totalQuestionNum"]; ok {
		question.TotalQuestionNum = v.(string)
	}
	if v, ok := informMap["workRelationId"]; ok {
		question.WorkRelationId = v.(string)
	}
	if v, ok := informMap["oldSchoolId"]; ok {
		question.OldSchoolId = v.(string)
	}
	if v, ok := informMap["oldWorkId"]; ok {
		question.OldWorkId = v.(string)
	}
	if v, ok := informMap["enc_work"]; ok {
		question.Enc_work = v.(string)
	}
	if v, ok := informMap["fullScore"]; ok {
		question.FullScore = v.(string)
	}
	if v, ok := informMap["api"]; ok {
		question.Api = v.(string)
	}
	if v, ok := informMap["courseId"]; ok {
		question.CourseId = v.(string)
	}
	if v, ok := informMap["classId"]; ok {
		question.ClassId = v.(string)
	}
	if v, ok := informMap["randomOptions"]; ok {
		question.RandomOptions = v.(string)
	}
}

// ParseWorkQuestionAction 用于解析作业题目，包括题目类型和题目文本
// TODO 同Question结构体问题 暂时返回未做 全部题目初始化
func ParseWorkQuestionAction(cache *xuexitong.XueXiTUserCache, workPoint *entity.PointWorkDto) entity.Question {
	var questionEntity entity.Question
	var workQuestion []entity.ChoiceQue
	var judgeQuestion []entity.JudgeQue
	var fillQuestion []entity.FillQue
	var shortQuestion []entity.ShortQue
	var readQuestion []entity.ReadQue
	question, _ := cache.WorkFetchQuestion(workPoint)

	// 使用 goquery 解析 HTML
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader([]byte(question)))
	if err != nil {
		log.Fatal(err)
	}

	//用于拉取并完善workPointDto信息
	informMap, err := utils.ParseWorkInform(doc)
	WorkInformInputWorkDTO(informMap, &questionEntity) //转换

	// 内置，用于从文本中提取题目类型
	var extractQuestionType = func(text string) string {
		start := strings.Index(text, "[")
		end := strings.Index(text, "]")

		if start != -1 && end != -1 && end > start {
			content := text[start+1 : end]
			return strings.TrimSpace(content)
		}
		return ""
	}

	questionSets := utils.ParseQuestionSets(doc)
	//fmt.Println(questionSets[9].HTML)
	for _, qs := range questionSets {
		qdoc, err := goquery.NewDocumentFromReader(bytes.NewReader([]byte(qs.HTML)))
		if err != nil {
			log.Fatal(err)
		}

		// 提取题目类型和题目文本
		quesType := qdoc.Find(".Py-m1-title .quesType").Text()
		quesType = extractQuestionType(quesType)

		// 提取题目文本和图片
		// TODO 这里获取图片怎么都拿不到 不会了ಠ_ಠ
		var quesTextBuilder strings.Builder
		quesText := qdoc.Find(".Py-m1-title").Contents().Each(func(i int, n *goquery.Selection) {
			if n.Get(0).Type == html.ElementNode || n.Get(0).Type == html.TextNode && n.Text() != "" {
				quesTextBuilder.WriteString(n.Text())
			}
		}).End().Text()

		// 清理并打印题目文本
		quesText = cleanText(quesTextBuilder.String())

		switch quesType {
		// 单选
		case qtype.SingleChoice.String():
			options := make(map[string]string)
			choiceQue := entity.ChoiceQue{}
			choiceQue.Type = qtype.SingleChoice
			choiceQue.Qid = qs.ID
			choiceQue.Text = quesText
			// 提取选项
			qdoc.Find(".answerList.singleChoice li").Each(func(i int, s *goquery.Selection) {
				optionLetter := s.Find("em.choose-opt").Text()

				// 查找 <cc> 内的内容
				ccContent := s.Find("cc").Contents().First()
				text := ccContent.Text()

				// 如果没有文本，则尝试获取 img 标签的 src 属性
				if text == "" {
					img, exists := s.Find("cc img").Attr("src")
					if exists {
						text = "Image: " + img
					} else {
						text = "No content available"
					}
				}
				options[optionLetter] = text
			})
			choiceQue.Options = options
			workQuestion = append(workQuestion, choiceQue)
			break
			// 多选
		case qtype.MultipleChoice.String():
			options := make(map[string]string)
			choiceQue := entity.ChoiceQue{}
			choiceQue.Type = qtype.MultipleChoice
			choiceQue.Qid = qs.ID
			choiceQue.Text = quesText
			// 提取选项
			qdoc.Find(".answerList.multiChoice li").Each(func(i int, s *goquery.Selection) {
				optionLetter := s.Find("em.choose-opt").Text()

				// 查找 <cc> 内的内容
				ccContent := s.Find("cc").Contents().First()
				text := ccContent.Text()

				// 如果没有文本，则尝试获取 img 标签的 src 属性
				if text == "" {
					img, exists := s.Find("cc img").Attr("src")
					if exists {
						text = "Image: " + img
					} else {
						text = "No content available"
					}
				}
				options[optionLetter] = text
			})
			choiceQue.Options = options
			workQuestion = append(workQuestion, choiceQue)
			break
		case qtype.TrueOrFalse.String():
			options := make(map[string]string)
			judgeQue := entity.JudgeQue{}
			judgeQue.Type = qtype.TrueOrFalse
			judgeQue.Qid = qs.ID
			judgeQue.Text = quesText
			// 提取选项
			qdoc.Find(".answerList.panduan li").Each(func(i int, s *goquery.Selection) {
				optionLetter := s.Find("em").Text()

				// 查找 <p> 内的内容
				ccContent := s.Find("p").Contents().First()
				text := ccContent.Text()

				options[optionLetter] = text
			})
			judgeQue.Options = options
			judgeQuestion = append(judgeQuestion, judgeQue)
		case qtype.FillInTheBlank.String():
			options := make(map[string][]string)
			fillQue := entity.FillQue{}
			fillQue.Type = qtype.FillInTheBlank
			fillQue.Qid = qs.ID
			fillQue.Text = quesText
			// 提取填空题
			qdoc.Find("ul.blankList2").Each(func(i int, selection *goquery.Selection) {
				optionLetter := selection.Find("p").Text()
				splitOptions := strings.Split(optionLetter, ":")
				validParts := splitOptions[:0]
				for _, part := range splitOptions {
					if part != "" { // 忽略空值
						validParts = append(validParts, part)
					}
				}
				for j, s := range validParts {
					options[s] = []string{fmt.Sprintf("填空%d", j+1)}
				}
			})
			fillQue.OpFromAnswer = options
			fillQuestion = append(fillQuestion, fillQue)
		case qtype.ShortAnswer.String():
			options := make(map[string][]string)
			shortQue := entity.ShortQue{}
			shortQue.Type = qtype.ShortAnswer
			shortQue.Qid = qs.ID
			shortQue.Text = quesText
			// 简答暂时未发现有多个textarea标签出现 不做多答案处理
			options["简答"] = []string{"简答答案"}
			shortQue.OpFromAnswer = options
			shortQuestion = append(shortQuestion, shortQue)
		case qtype.ReadingComprehension.String():
			readQue := entity.ReadQue{}
			readQue.Type = qtype.ReadingComprehension
			readQue.Qid = qs.ID
			readQue.Text = quesText
			var (
				ops, childTypes, childIds []string
				textOp                    []map[string]string
			)
			choice := make(map[string]string)
			// TODO 这里我在做尝试 之前是分批处理的 和到一块无法提取
			qdoc.Find("div.readComprehension").Each(func(i int, s *goquery.Selection) {
				s.Find("ul.answerList").Each(func(i int, ul *goquery.Selection) {
					li := ul.Find("li.ignoreli")
					op := li.Find("span").Text() + li.Find("div.ans-cc").Text()
					ops = append(ops, op)
					ul.Find(`li[data="answer"]`).Each(func(i int, li1 *goquery.Selection) {
						text := li1.Find("em").Text()
						choice[text] = ""
						//fmt.Println(li1.Html())
						li1.Find("p.type14").NextAll().EachWithBreak(func(i int, s *goquery.Selection) bool {
							txt := strings.TrimSpace(s.Text())
							if txt != "" {
								//fmt.Println(txt) // 输出 3
								choice[text] = txt
								return false // 停止遍历
							}
							return true
						})
					})
					textOp = append(textOp, choice)
					//fmt.Println("--------")
					choice = make(map[string]string)
				})
				//fmt.Println("ops：", ops)
				//fmt.Println("textOp：", textOp)
				s.Find("input").Each(func(i int, s *goquery.Selection) {
					name, exName := s.Attr("name")
					value, exValue := s.Attr("value")
					if exName && exValue {
						if name == "readCompreHension-childType" {
							childTypes = append(childTypes, value)
						}
						if name == "readCompreHension-childId" {
							childIds = append(childIds, value)
						}
					}
				})
				//fmt.Println("childTypes：", childTypes)
				//fmt.Println("childIds：", childIds)
				//dataItemID = childIds[0]

				opFrom := func(i, childType int, text, childId string, choice map[string]string) map[string]entity.ReadChoice {
					res := make(map[string]entity.ReadChoice)
					res[text] = entity.ReadChoice{
						Text:       choice,
						ChildType:  childType,
						ChildId:    childId,
						DataItemID: childId,
					}
					return res
				}

				for j, text := range ops {
					iChildType, _ := strconv.Atoi(childTypes[j])
					readQue.OpFormAnswer = append(readQue.OpFormAnswer, opFrom(j, iChildType, text, childIds[j], textOp[j]))
				}
				textOp = []map[string]string{}
			})
			readQuestion = append(readQuestion, readQue)
		}
	}

	questionEntity.Choice = workQuestion
	questionEntity.Judge = judgeQuestion
	questionEntity.Fill = fillQuestion
	questionEntity.Short = shortQuestion
	questionEntity.Read = readQuestion
	return questionEntity
}

// 定义题型处理策略函数类型
type problemMessageStrategy func(context string, topic entity.ExamTurn) que_core.AIChatMessages

// 策略映射表：题型 -> 处理函数
var problemStrategies = map[string]problemMessageStrategy{
	"单选题": handleSingleChoice,
	"多选题": handleMultipleChoice,
	"判断题": handleTrueFalse,
	"填空题": handleFillInTheBlank,
	"简答题": handleShortAnswer,
}

func AIProblemMessage(testPaperTitle, text string, topic entity.ExamTurn) que_core.AIChatMessages {

	context := buildProblemContext(testPaperTitle, text, topic)

	// 查找对应的处理策略
	if strategy, exists := problemStrategies[testPaperTitle]; exists {
		return strategy(context, topic)
	}

	// 默认返回空消息
	return que_core.AIChatMessages{Messages: []que_core.Message{}}
}

// buildProblemContext 构建通用的题目上下文
func buildProblemContext(testPaperTitle, text string, topic entity.ExamTurn) (context string) {
	switch testPaperTitle {
	case qtype.SingleChoice.String():
		for c, q := range topic.XueXChoiceQue.Options {
			context += text + "\n"
			context += fmt.Sprintf("\n%v. %v", c, q)
		}
		//for _, v := range topic.Selects {
		//	context += v.Num + v.Text + "\n"
		//}
	case qtype.MultipleChoice.String():
		for c, q := range topic.XueXChoiceQue.Options {
			context += text + "\n"
			context += fmt.Sprintf("\n%v. %v", c, q)
		}
		//for _, v := range topic.Selects {
		//	context += v.Num + v.Text + "\n"
		//}
	case qtype.FillInTheBlank.String():
		for c, q := range topic.XueXFillQue.OpFromAnswer {
			context += text + "\n"
			context += fmt.Sprintf("\n%v. %v", c, q)
		}
		//for _, v := range topic.Selects {
		//	context += v.Num + v.Text + "\n"
		//}
	case qtype.TrueOrFalse.String():
		for c, q := range topic.XueXJudgeQue.Options {
			context += text + "\n"
			context += fmt.Sprintf("\n%v. %v", c, q)
		}
		//for _, v := range topic.Selects {
		//	context += v.Num + v.Text + "\n"
		//}
	case qtype.ShortAnswer.String():
		for c, q := range topic.XueXShortQue.OpFromAnswer {
			context += text + "\n"
			context += fmt.Sprintf("\n%v. %v", c, q)
		}
		//for _, v := range topic.Selects {
		//	context += v.Num + v.Text + "\n"
		//}
	}
	return context
}

// 单选题处理策略
func handleSingleChoice(context string, topic entity.ExamTurn) que_core.AIChatMessages {
	problem := buildProblemHeader(topic.XueXChoiceQue.Type.String(), "单选题", context)
	return que_core.AIChatMessages{Messages: []que_core.Message{
		{Role: "user", Content: "接下来你只需要回答选项对应内容即可...格式：[\"选项1\"]"},
		{Role: "user", Content: "就算你不知道选什么也随机选...无需回答任何解释！！！"},
		{Role: "user", Content: exampleSingleChoice()},
		{Role: "user", Content: problem},
	}}
}

// 多选题处理策略
func handleMultipleChoice(context string, topic entity.ExamTurn) que_core.AIChatMessages {
	problem := buildProblemHeader(topic.XueXChoiceQue.Type.String(), "多选题", context)
	return que_core.AIChatMessages{Messages: []que_core.Message{
		{Role: "user", Content: "接下来你只需要回答选项对应内容即可...格式：[\"选项1\",\"选项2\"]"},
		{Role: "user", Content: "就算你不知道选什么也随机选...无需回答任何解释！！！"},
		{Role: "user", Content: exampleMultipleChoice()},
		{Role: "user", Content: problem},
	}}
}

// 判断题处理策略
func handleTrueFalse(context string, topic entity.ExamTurn) que_core.AIChatMessages {
	problem := buildProblemHeader(topic.XueXJudgeQue.Type.String(), "判断题", context)
	return que_core.AIChatMessages{Messages: []que_core.Message{
		{Role: "user", Content: "接下来你只需要回答“正确”或者“错误”即可...格式：[\"正确\"]"},
		{Role: "user", Content: "就算你不知道选什么也随机选...无需回答任何解释！！！"},
		{Role: "user", Content: exampleTrueFalse()},
		{Role: "user", Content: problem},
	}}
}

// 填空题处理策略
func handleFillInTheBlank(context string, topic entity.ExamTurn) que_core.AIChatMessages {
	problem := buildProblemHeader(topic.XueXFillQue.Type.String(), "填空题", context)
	return que_core.AIChatMessages{Messages: []que_core.Message{
		{Role: "user", Content: "其中，“（answer_数字）”相关字样的地方是你需要填写答案的地方...格式：[\"答案1\",\"答案2\"]"},
		{Role: "user", Content: "就算你不知道选什么也随机选...无需回答任何解释！！！"},
		{Role: "user", Content: exampleFillInTheBlank()},
		{Role: "user", Content: problem},
	}}
}

// 简答题处理策略
func handleShortAnswer(context string, topic entity.ExamTurn) que_core.AIChatMessages {
	problem := buildProblemHeader(topic.XueXShortQue.Type.String(), "简答题", context)
	return que_core.AIChatMessages{Messages: []que_core.Message{
		{Role: "user", Content: "这是一个简答题...格式：[\"答案\"]，注意不要拆分答案！！！"},
		{Role: "user", Content: "就算你不知道选什么也随机选...无需回答任何解释！！！"},
		{Role: "user", Content: exampleShortAnswer()},
		{Role: "user", Content: problem},
	}}
}

// 构建题目头部信息
func buildProblemHeader(testPaperTitle, topicType, context string) string {
	return fmt.Sprintf(`试卷名称：%s
题目类型：%s
题目内容：%s`, testPaperTitle, topicType, context)
}

// 单选题示例
func exampleSingleChoice() string {
	return `比如：
试卷名称：考试
题目类型：单选
题目内容：新中国是什么时候成立的
A. 1949年10月5日
B. 1949年10月1日
C. 1949年09月1日
D. 2002年10月1日

那么你应该回答选项B的内容："["1949年10月1日"]"`
}

// 多选题示例
func exampleMultipleChoice() string {
	return `比如：
试卷名称：考试
题目类型：多选题
题目内容：马克思关于资本积累的学说是剩余价值理论的重要组成部分...
A. 资本主义扩大再生产的源泉
B. 资本有机构成呈现不断降低趋势的根本原因
C. 社会财富占有两极分化的重要原因
D. 资本主义社会失业现象产生的根源

那么你应该回答选项A、B、D的内容："["资本主义扩大再生产的源泉","社会财富占有两极分化的重要原因","资本主义社会失业现象产生的根源"]"`
}

// 判断题示例
func exampleTrueFalse() string {
	return `比如：
试卷名称：考试
题目类型：判断
题目内容：新中国是什么时候成立是1949年10月1日吗？
A. 正确
B. 错误

那么你应该回答选项A的内容："["正确"]"`
}

// 填空题示例
func exampleFillInTheBlank() string {
	return ` 比如：
试卷名称：考试
题目类型：填空
题目内容：新中国成立于（ ）年。
答案：1949

那么你应该回答："["1949"]"`
}

func exampleShortAnswer() string {
	return `比如：
试卷名称：考试
题目类型：简答
题目内容：请简述中国和外国的国别 differences
答案：中国和外国的国别 differences

那么你应该回答： "["中国和外国的国别 differences"]"`
}

//func AIProblemMessage(testPaperTitle string, topic entity.ExamTurn) utils.AIChatMessages {
//	topicType := topic.ChoiceQue.Type.String()
//	context := topic.ChoiceQue.Text
//	for c, q := range topic.ChoiceQue.Options {
//		context += fmt.Sprintf("\n%v. %v", c, q)
//	}
//
//	problem := `试卷名称：` + testPaperTitle + `
//题目类型：` + topicType + `
//题目内容：` + context + "\n"
//
//	//选择题
//	if topicType == "单选题" {
//		for _, v := range topic.Selects {
//			problem += v.Num + v.Text + "\n"
//		}
//		return utils.AIChatMessages{Messages: []utils.Message{
//			{
//				Role:    "user",
//				Content: `接下来你只需要回答选项对应内容即可，不能回答任何选项无关的任何内容，包括解释以及标点符也不需要。`,
//			},
//			{
//				Role:    "user",
//				Content: `就算你不知道选什么也随机选输出其选项内容，回答的格式一定要严格为单个数组格式，比如：["选项1"]，注意回复的时候不要带选项字母，你只需回复答案对应格式内容即可，无需回答任何解释！！！`,
//			},
//			{
//				Role: "user",
//				Content: `比如：` + `
//				试卷名称：考试
//				题目类型：单选
//				题目内容：新中国是什么时候成立的
//				A. 1949年10月5日
//				B. 1949年10月1日
//				C. 1949年09月1日
//				D. 2002年10月1日
//				` + `
//				那么你应该回答选项B的内容：“["1949年10月1日"]”
//				`,
//			},
//			{
//				Role:    "user",
//				Content: problem,
//			},
//		}}
//	} else if topicType == "多选题" {
//		for _, v := range topic.Selects {
//			problem += v.Num + v.Text + "\n"
//		}
//		return utils.AIChatMessages{Messages: []utils.Message{
//			{
//				Role:    "user",
//				Content: `接下来你只需要回答选项对应内容即可，不能回答任何选项无关的任何内容，包括解释以及标点符也不需要。`,
//			},
//			{
//				Role:    "user",
//				Content: `就算你不知道选什么也随机选输出其选项内容，回答的格式一定要严格为单个数组格式，比如：["选项1","选项2"]，注意回复的时候不要带选项字母，你只需回复答案对应格式内容即可，无需回答任何解释！！！`,
//			},
//			{
//				Role: "user",
//				Content: `比如：` + `
//				试卷名称：考试
//				题目类型：多选题
//				题目内容：马克思关于资本积累的学说是剩余价值理论的重要组成部分。资本积累是
//				A. 资本主义扩大再生产的源泉
//				B. 资本有机构成呈现不断降低趋势的根本原因
//				C. 社会财富占有两极分化的重要原因
//				D. 资本主义社会失业现象产生的根源
//				` + `
//				那么你应该回答选项A、B、D的内容：“["资本主义扩大再生产的源泉","社会财富占有两极分化的重要原因","资本主义社会失业现象产生的根源"]”
//				`,
//			},
//			{
//				Role:    "user",
//				Content: problem,
//			},
//		}}
//	} else if topicType == "判断题" {
//		for _, v := range topic.Selects {
//			problem += v.Num + v.Text + "\n"
//		}
//		return utils.AIChatMessages{Messages: []utils.Message{
//			{
//				Role:    "user",
//				Content: `接下来你只需要回答“正确”或者“错误”即可，不能回答任何无关的内容，包括解释以及标点符也不需要。`,
//			},
//			{
//				Role:    "user",
//				Content: `就算你不知道选什么也随机选输出其选项内容，回答的格式一定要严格为单个数组格式，比如：["正确"]，注意回复的时候不要带选项字母，你只需回复答案对应格式内容即可，无需回答任何解释！！！`,
//			},
//			{
//				Role: "user",
//				Content: `比如：` + `
//				试卷名称：考试
//				题目类型：判断
//				题目内容：新中国是什么时候成立是1949年10月1日吗？
//				A. 正确
//				B. 错误
//				` + `
//				那么你应该回答选项A的内容：“["正确"]”
//				`,
//			},
//			{
//				Role:    "user",
//				Content: problem,
//			},
//		}}
//	} else if topicType == "填空题" { //填空题
//		return utils.AIChatMessages{Messages: []utils.Message{
//			{
//				Role:    "user",
//				Content: `其中，“（answer_数字）”相关字样的地方是你需要填写答案的地方，现在你只需要按顺序回复我对应每个填空项的答案即可，回答的格式一定要严格为单个数组格式，比如["答案1","答案2"]其他不符合格式的内容无需回复。你只需回复答案对应格式内容即可，无需回答任何解释！！！`,
//			},
//			{
//				Role:    "user",
//				Content: problem,
//			},
//		}}
//	} else if topicType == "简答题" { //简答
//		return utils.AIChatMessages{Messages: []utils.Message{
//			{
//				Role:    "user",
//				Content: `这是一个简答题，现在你只需要回复我对应简答题答案即可，回答的格式一定要严格为单个数组格式，比如["答案"]，但是注意你只需要把所有答案填写在一个元素项里面就行，别分开，比如你不能["xxx","zzz"]这样写，你只能["xxxzzz"]这样写，其他不符合格式的内容无需回复。你只需回复答案对应格式内容即可，无需回答任何解释！！！`,
//			},
//			{
//				Role:    "user",
//				Content: problem,
//			},
//		}}
//	}
//	return utils.AIChatMessages{Messages: []utils.Message{}}
//}
