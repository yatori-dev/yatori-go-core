package xuexitong

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/api/entity"
	"github.com/yatori-dev/yatori-go-core/api/xuexitong"
	"github.com/yatori-dev/yatori-go-core/models/ctype"
	"github.com/yatori-dev/yatori-go-core/utils"
	"golang.org/x/net/html"
	"log"
	"regexp"
	"strings"
	"unicode"
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
	classId, courseId, knowledgeId, cardIndex, cpi int) (interface{}, error) {
	cardHtml, err := cache.PageMobileChapterCard(classId, courseId, knowledgeId, cardIndex, cpi)
	var att interface{}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch pageMobileChapterCard: %w", err)
	}
	doc, err := html.Parse(bytes.NewReader([]byte(cardHtml)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
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
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
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
			return ChapterNotOpened{}, nil
		}
		return APIError{Message: blankTips}, nil
	}
	//log.Println("Attachment拉取成功")
	return att, nil
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
	p.Title = gojsonq.New().JSONString(fetch).Find("filename").(string)

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
	// 将多个连续的空白字符替换为单个空格
	return strings.Join(strings.FieldsFunc(text, func(r rune) bool {
		return unicode.IsSpace(r)
	}), " ")
}

// ParseWorkQuestionAction 用于解析作业题目，包括题目类型和题目文本
func ParseWorkQuestionAction(cache *xuexitong.XueXiTUserCache, workPoint *entity.PointWorkDto) entity.Question {
	var workQuestion []entity.ChoiceQue
	question, _ := cache.WorkFetchQuestion(workPoint)

	// 使用 goquery 解析 HTML
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader([]byte(question)))
	if err != nil {
		log.Fatal(err)
	}

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
	fmt.Println(questionSets[9].HTML)
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
			if n.Get(0).Type == 3 {
				quesTextBuilder.WriteString(n.Text())
			}
		}).End().Text()

		// 清理并打印题目文本
		quesText = cleanText(quesTextBuilder.String())

		switch quesType {
		// 单选
		case ctype.SingleChoice.String():
			options := make(map[string]string)
			choiceQue := entity.ChoiceQue{}
			choiceQue.Type = ctype.SingleChoice
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
		case ctype.MultipleChoice.String():
			options := make(map[string]string)
			choiceQue := entity.ChoiceQue{}
			choiceQue.Type = ctype.MultipleChoice
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
		}
	}
	//for j, q := range workQuestion {
	//	fmt.Printf("Question %d:\nType: %s\nText: %s\noptions: %v\n\n", j+1, q.Type, q.Text, q.options)
	//}
	return entity.Question{Choice: workQuestion}
}

func AIProblemMessage(testPaperTitle string, topic entity.ExamTurn) utils.AIChatMessages {
	topicType := topic.ChoiceQue.Type.String()
	context := topic.ChoiceQue.Text
	for c, q := range topic.ChoiceQue.Options {
		context += fmt.Sprintf("\n%v. %v", c, q)
	}

	problem := `试卷名称：` + testPaperTitle + `
题目类型：` + topicType + `
题目内容：` + context + "\n"

	//选择题
	if topicType == "单选题" {
		for _, v := range topic.Selects {
			problem += v.Num + v.Text + "\n"
		}
		return utils.AIChatMessages{Messages: []utils.Message{
			{
				Role:    "user",
				Content: `接下来你只需要回答选项对应内容即可，不能回答任何选项无关的任何内容，包括解释以及标点符也不需要。`,
			},
			{
				Role:    "user",
				Content: `就算你不知道选什么也随机选输出其选项内容，回答的格式一定要严格为单个数组格式，比如：["选项1"]，注意回复的时候不要带选项字母，你只需回复答案对应格式内容即可，无需回答任何解释！！！`,
			},
			{
				Role: "user",
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
	} else if topicType == "多选题" {
		for _, v := range topic.Selects {
			problem += v.Num + v.Text + "\n"
		}
		return utils.AIChatMessages{Messages: []utils.Message{
			{
				Role:    "user",
				Content: `接下来你只需要回答选项对应内容即可，不能回答任何选项无关的任何内容，包括解释以及标点符也不需要。`,
			},
			{
				Role:    "user",
				Content: `就算你不知道选什么也随机选输出其选项内容，回答的格式一定要严格为单个数组格式，比如：["选项1","选项2"]，注意回复的时候不要带选项字母，你只需回复答案对应格式内容即可，无需回答任何解释！！！`,
			},
			{
				Role: "user",
				Content: `比如：` + `
				试卷名称：考试
				题目类型：多选题
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
	} else if topicType == "判断题" {
		for _, v := range topic.Selects {
			problem += v.Num + v.Text + "\n"
		}
		return utils.AIChatMessages{Messages: []utils.Message{
			{
				Role:    "user",
				Content: `接下来你只需要回答“正确”或者“错误”即可，不能回答任何无关的内容，包括解释以及标点符也不需要。`,
			},
			{
				Role:    "user",
				Content: `就算你不知道选什么也随机选输出其选项内容，回答的格式一定要严格为单个数组格式，比如：["正确"]，注意回复的时候不要带选项字母，你只需回复答案对应格式内容即可，无需回答任何解释！！！`,
			},
			{
				Role: "user",
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
	} else if topicType == "填空题" { //填空题
		return utils.AIChatMessages{Messages: []utils.Message{
			{
				Role:    "user",
				Content: `其中，“（answer_数字）”相关字样的地方是你需要填写答案的地方，现在你只需要按顺序回复我对应每个填空项的答案即可，回答的格式一定要严格为单个数组格式，比如["答案1","答案2"]其他不符合格式的内容无需回复。你只需回复答案对应格式内容即可，无需回答任何解释！！！`,
			},
			{
				Role:    "user",
				Content: problem,
			},
		}}
	} else if topicType == "简答题" { //简答
		return utils.AIChatMessages{Messages: []utils.Message{
			{
				Role:    "user",
				Content: `这是一个简答题，现在你只需要回复我对应简答题答案即可，回答的格式一定要严格为单个数组格式，比如["答案"]，但是注意你只需要把所有答案填写在一个元素项里面就行，别分开，比如你不能["xxx","zzz"]这样写，你只能["xxxzzz"]这样写，其他不符合格式的内容无需回复。你只需回复答案对应格式内容即可，无需回答任何解释！！！`,
			},
			{
				Role:    "user",
				Content: problem,
			},
		}}
	}
	return utils.AIChatMessages{Messages: []utils.Message{}}
}
