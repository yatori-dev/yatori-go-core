package xuexitong

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	ddddocr "github.com/Changbaiqi/ddddocr-go/utils"
	"github.com/PuerkitoBio/goquery"
	"github.com/thedevsaddam/gojsonq"
	ort "github.com/yalue/onnxruntime_go"
	"github.com/yatori-dev/yatori-go-core/api/entity"
	"github.com/yatori-dev/yatori-go-core/api/xuexitong"
	que_core "github.com/yatori-dev/yatori-go-core/que-core/aiq"
	"github.com/yatori-dev/yatori-go-core/que-core/qtype"
	"github.com/yatori-dev/yatori-go-core/utils"
	log2 "github.com/yatori-dev/yatori-go-core/utils/log"
	"golang.org/x/net/html"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
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
	cardHtml, err := cache.PageMobileChapterCard(classId, courseId, knowledgeId, cardIndex, cpi, 3, nil)
	if err != nil {
		if err.Error() == "触发验证码" {
			log2.Print(log2.DEBUG, utils.RunFuncName(), "触发验证码，正在进行AI智能识别绕过.....")
			for {
				codePath, err1 := cache.XueXiTVerificationCodeApi(5, nil)
				if err1 != nil {
					return nil, "", err1
				}
				if codePath == "" { //如果path为空，那么可能是账号问题
					return nil, "", errors.New("无法正常获取对应网站验证码，请检查对应url是否正常")
				}
				img, _ := utils.ReadImg(codePath) //读取验证码图片
				//codeResult := utils.AutoVerification(img, ort.NewShape(1, 23)) //自动识别
				codeResult := ddddocr.SemiOCRVerification(img, ort.NewShape(1, 23))
				utils.DeleteFile(codePath) //删除验证码文件
				status, err1 := cache.XueXiTPassVerificationCode(codeResult, 5, nil)
				//fmt.Println(codeResult)
				//fmt.Println(status)
				if status {
					break
				}
			}
			cardHtml, err = cache.PageMobileChapterCard(classId, courseId, knowledgeId, cardIndex, cpi, 3, nil) //尝试重新拉取卡片信息
			log2.Print(log2.DEBUG, utils.RunFuncName(), "绕过成功")
		}
	}

	if strings.Contains(cardHtml, `<p class="blankTips">章节未开放</p>`) {
		return nil, "", errors.New("章节未开放")
	}
	//如果遇到人脸,则进行过人脸
	if strings.Contains(cardHtml, `title : "人脸识别"`) {
		ObjectId, err1 := PassFacePhoneAction(cache, fmt.Sprintf("%d", courseId), fmt.Sprintf("%d", classId), fmt.Sprintf("%d", cpi), fmt.Sprintf("%d", knowledgeId), "", "", "")
		//人脸重试机制
		for i := 0; i <= 8; i++ {
			if err1 != nil && strings.Contains(err1.Error(), "用户图片信息出错") {
				time.Sleep(1 * time.Second) //隔一下
				ObjectId, err1 = PassFacePhoneAction(cache, fmt.Sprintf("%d", courseId), fmt.Sprintf("%d", classId), fmt.Sprintf("%d", cpi), fmt.Sprintf("%d", knowledgeId), "", "", "")
			} else {
				break
			}
		}
		if err1 != nil {
			log.Println(ObjectId, err1.Error())
			return nil, "", err1
		}
		//过完人脸重新拉取章节信息
		cardHtml, err = cache.PageMobileChapterCard(classId, courseId, knowledgeId, cardIndex, cpi, 3, nil)
		//如果新版本人脸过不去，则再尝试旧版本人脸
		if strings.Contains(cardHtml, `title : "人脸识别"`) {
			ObjectId, err1 = PassFacePhoneOldAction(cache, fmt.Sprintf("%d", courseId), fmt.Sprintf("%d", classId), fmt.Sprintf("%d", cpi), fmt.Sprintf("%d", knowledgeId), "", "", "")
			time.Sleep(1 * time.Second) //隔一下
			cardHtml, err = cache.PageMobileChapterCard(classId, courseId, knowledgeId, cardIndex, cpi, 3, nil)
		}

		if strings.Contains(cardHtml, `title : "人脸识别"`) {
			return nil, "", errors.New("通过人脸识别失败")
		}

	}

	//探测又进度控制的
	docQuery, err := goquery.NewDocumentFromReader(bytes.NewReader([]byte(cardHtml)))
	topInfo := docQuery.Find("div.tipInfo").Text()
	if strings.Contains(topInfo, "本课已开启学习进度控制") {
		// 定义正则，匹配“>数字/数字<”
		re := regexp.MustCompile(`(\d+)/(\d+)`)
		matches := re.FindAllStringSubmatch(topInfo, -1)
		if len(matches) >= 2 {
			taskNum, _ := strconv.Atoi(matches[0][1])
			taskDenom, _ := strconv.Atoi(matches[0][2])
			timeNum, _ := strconv.Atoi(matches[1][1])
			timeDenom, _ := strconv.Atoi(matches[1][2])
			if taskNum == 0 || timeDenom-timeNum <= 0 {
				return nil, "", fmt.Errorf("本课已开启学习进度控制，今日还可完成%d/%d个视频任务点,已观看视频时长%d/%d分钟", taskNum, taskDenom, timeNum, timeDenom)
			}
		}
	}
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
		(att.(map[string]interface{}))["enc"] = enc
	}
	//log.Println("Attachment拉取成功")
	return att, enc, nil
}

func VideoDtoFetchAction(cache *xuexitong.XueXiTUserCache, p *entity.PointVideoDto) (bool, error) {
	fetch, err := cache.VideoDtoFetch(p, 5, nil)
	//500处理
	if err != nil && strings.Contains(err.Error(), "status code: 500") {
		ReLogin(cache) //重新登录
		fetch, err = cache.VideoDtoFetch(p, 5, nil)
	}

	if err != nil {
		log.Println("VideoDtoFetchAction:", err)
		return false, err
	}
	dtoken, ok1 := gojsonq.New().JSONString(fetch).Find("dtoken").(string)
	if ok1 {
		p.DToken = dtoken
	} else {
		log2.Print(log2.INFO, fmt.Sprintf("[%s]", cache.Name), fmt.Sprintf("dtoken获取失败:%s", fetch))
	}
	duration, ok1 := gojsonq.New().JSONString(fetch).Find("duration").(float64)
	if ok1 {
		p.Duration = int(duration)
	} else {
		log2.Print(log2.INFO, fmt.Sprintf("[%s]", cache.Name), fmt.Sprintf("duration获取失败:%s", fetch))
	}

	//titleStr, turnErr := url.QueryUnescape(gojsonq.New().JSONString(fetch).Find("filename").(string))
	////转换
	//if turnErr != nil {
	//	log2.Print(log2.DEBUG, titleStr, "解码失败")
	//	p.Title = gojsonq.New().JSONString(fetch).Find("filename").(string)
	//} else {
	//	p.Title = titleStr
	//}
	stutas, ok := gojsonq.New().JSONString(fetch).Find("status").(string)
	if !ok { //如果转化失败，可能是文件有问题
		return false, nil
	}
	if stutas == "success" {
		return true, nil
	}
	return false, errors.New("fetch failed")
}

func WorkPageFromAction(cache *xuexitong.XueXiTUserCache, workPoint *entity.PointWorkDto) ([]entity.WorkInputField, error) {
	questionHtml, err := cache.WorkFetchQuestion(workPoint, 3, nil)
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
func ParseWorkQuestionAction(cache *xuexitong.XueXiTUserCache, workPoint *entity.PointWorkDto) (entity.Question, error) {
	var questionEntity entity.Question
	var workQuestion []entity.ChoiceQue
	var judgeQuestion []entity.JudgeQue
	var fillQuestion []entity.FillQue
	var shortQuestion []entity.ShortQue
	var termQuestion []entity.TermExplanationQue
	var essayQuestion []entity.EssayQue
	var matchingQuestion []entity.MatchingQue
	var otherQuestion []entity.OtherQue

	question, err := cache.WorkFetchQuestion(workPoint, 3, nil)

	if err != nil {
		//若无权限则采用第二种方式
		if strings.Contains(err.Error(), `<p class="blankTips">无效的权限,code=2</p>`) {
			question, err = cache.WorkFetch2Question(workPoint, 3, nil)
		}
	}
	//先检测是否含有加密字体，如果有则先解密

	// 使用 goquery 解析 HTML
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader([]byte(question)))
	if err != nil {
		log.Fatal(err)
	}

	//用于拉取并完善workPointDto信息
	informMap, err := utils.ParseWorkInform(doc)
	WorkInformInputWorkDTO(informMap, &questionEntity) //转换

	//用于判断是否已经截止
	textStatus := doc.Find("div.chapter-content").Text()
	if strings.Contains(textStatus, "已截止，不能作答") {
		return questionEntity, errors.New("已截止，不能作答")
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
	//fmt.Println(questionSets[9].HTML)
	for _, qs := range questionSets {
		qdoc, err := goquery.NewDocumentFromReader(bytes.NewReader([]byte(qs.HTML)))
		if err != nil {
			log.Fatal(err)
		}

		// 提取题目类型和题目文本
		quesType := qdoc.Find(".Py-m1-title .quesType").Text()
		quesType = extractQuestionType(quesType)
		//傻逼学习通踏马题目内容都能放到题目类型里面，只能这么解决了
		if strings.Contains(quesType, "辨析题") {
			//fmt.Println(quesType)
			quesType = "判断题" //将辨析题当判断题进行
		} else if strings.Contains(quesType, "投票题") {
			quesType = "单选题"
		}

		// 提取题目文本和图片
		// TODO 这里获取图片怎么都拿不到 不会了ಠ_ಠ
		var quesTextBuilder strings.Builder //文本题目
		quesText := qdoc.Find(".Py-m1-title .workTextWrap").Contents().Each(func(i int, n *goquery.Selection) {
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
			//用于临时存储选项，后面进行排序，因为学习通可能会打乱答题顺序
			resOptions := make(map[string]string)
			// 提取选项
			qdoc.Find(".answerList.singleChoice li").Each(func(i int, s *goquery.Selection) {
				optionLetter := s.Find("em.choose-opt").Text()
				trueOptionLetter, _ := s.Find("em.choose-opt").Attr("id-param") //如果这个不是空的说明这个才是正确的选项参数
				if trueOptionLetter != "" {
					optionLetter = trueOptionLetter
				}
				// 查找 <cc> 内的内容
				ccContent := s.Find("cc").Contents().First()
				text := ccContent.Text()
				//如果ccFirst没有，则采用Contents.Text()
				if text == "" {
					text = s.Find("cc").Contents().Text()
				}
				// 如果没有文本，则尝试获取 img 标签的 src 属性
				if text == "" {
					img, exists := s.Find("cc img").Attr("src")
					if exists {
						text = "Image: " + img
					} else {
						text = "No content available"
					}
				}
				resOptions[optionLetter] = text
			})
			//对Map的内容按字母选项排序
			resSelect := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N"}
			for _, sel := range resSelect {
				if resOptions[sel] == "" {
					break
				}
				options[sel] = resOptions[sel]
			}
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
			//用于临时存储选项，后面进行排序，因为学习通可能会打乱答题顺序
			resOptions := make(map[string]string)
			// 提取选项
			//qdoc.Find(".answerList li").Each(func(i int, s *goquery.Selection) {
			qdoc.Find(".answerList.multiChoice li").Each(func(i int, s *goquery.Selection) {
				optionLetter := s.Find("em.choose-opt").Text()
				trueOptionLetter, _ := s.Find("em.choose-opt").Attr("id-param") //如果这个不是空的说明这个才是正确的选项参数
				if trueOptionLetter != "" {
					optionLetter = trueOptionLetter
				}
				// 查找 <cc> 内的内容
				ccContent := s.Find("cc").Contents().First()
				//fmt.Println(s.Html())
				//fmt.Println(ccContent.Html())
				//fmt.Println(s.Find("cc").Contents().Text())
				text := ccContent.Text()

				//如果ccFirst没有，则采用Contents.Text()
				if text == "" {
					text = s.Find("cc").Contents().Text()
				}
				// 如果没有文本，则尝试获取 img 标签的 src 属性
				if text == "" {
					img, exists := s.Find("cc img").Attr("src")
					if exists {
						text = "Image: " + img
					} else {
						text = "No content available"
					}
				}
				resOptions[optionLetter] = text
			})
			//对Map的内容按字母选项排序
			resSelect := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N"}
			for _, sel := range resSelect {
				if resOptions[sel] == "" {
					break
				}
				options[sel] = resOptions[sel]
			}
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
				trueoption, _ := s.Attr("val-param")
				if trueoption == "true" {
					optionLetter = "对"
				} else if trueoption == "false" {
					optionLetter = "错"
				}

				// 查找 <p> 内的内容
				ccContent := s.Find("p").Contents().First()
				text := ccContent.Text()

				options[optionLetter] = text
			})
			judgeQue.Options = options
			judgeQuestion = append(judgeQuestion, judgeQue)
		case qtype.FillInTheBlank.String():
			//options := make(map[string][]string)
			options := []string{}
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
				for j, _ := range validParts {
					options = append(options, fmt.Sprintf("%d", j+1))
					//options[s] = []string{fmt.Sprintf("填空%d", j+1)}
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
		case qtype.TermExplanation.String(): //名词解释
			options := make(map[string][]string)
			termExplanationQue := entity.TermExplanationQue{}
			termExplanationQue.Type = qtype.TermExplanation
			termExplanationQue.Qid = qs.ID
			termExplanationQue.Text = quesText
			// 简答暂时未发现有多个textarea标签出现 不做多答案处理
			options["名词解释"] = []string{"名词解释"}
			termExplanationQue.OpFromAnswer = options
			termQuestion = append(termQuestion, termExplanationQue)
		case qtype.Essay.String(): //论述题
			options := make(map[string][]string)
			essayQue := entity.EssayQue{}
			essayQue.Type = qtype.Essay
			essayQue.Qid = qs.ID
			essayQue.Text = quesText
			// 简答暂时未发现有多个textarea标签出现 不做多答案处理
			options["论述"] = []string{"论述"}
			essayQue.OpFromAnswer = options
			essayQuestion = append(essayQuestion, essayQue)
		case qtype.Matching.String():

			matchingQue := entity.MatchingQue{}
			matchingQue.Qid = qs.ID
			matchingQue.Type = qtype.Matching
			matchingQue.Text = quesText
			//连线题截取链接项
			options := []string{}
			selects := []string{}
			//fmt.Println(qdoc.Html())
			qdoc.Find("ul.answerList-line").Each(func(i int, s *goquery.Selection) {
				//fmt.Printf("第 %d 组内容：\n", i+1)
				if i%2 == 0 {
					s.Find("li").Each(func(_ int, li *goquery.Selection) {
						options = append(options, li.Text())
					})
				} else {
					s.Find("li").Each(func(_ int, li *goquery.Selection) {
						selects = append(selects, li.Text())
					})
				}
			})
			matchingQue.Options = options
			matchingQue.Selects = selects
			matchingQuestion = append(matchingQuestion, matchingQue)
		case qtype.QueOther.String():
			options := make(map[string][]string)
			otherQue := entity.OtherQue{}
			otherQue.Type = qtype.QueOther
			otherQue.Qid = qs.ID
			otherQue.Text = quesText
			// 简答暂时未发现有多个textarea标签出现 不做多答案处理
			options["回复"] = []string{""}
			otherQue.OpFromAnswer = options
			otherQuestion = append(otherQuestion, otherQue)
		case qtype.ReadingComprehension.String(): //阅读理解题处理(未完工）
			//截取阅读理解子题目
			doc.Find("ul.answerList").Each(func(i int, s *goquery.Selection) {
				queType := s.Find("li.ignoreli .span").Text() //截取题目类型
				if strings.Contains(queType, "单选题") {
					content := strings.TrimSpace(doc.Find(".ans-cc p").Text())
					fmt.Println(content)
					// 提取选项
					doc.Find(`li[data="answer"]`).Each(func(i int, s *goquery.Selection) {
						option := strings.TrimSpace(s.Find("em").Text())
						selectContent := strings.TrimSpace(s.Find("cc p").Text())
						fmt.Printf("选项 %s: %s\n", option, selectContent)
					})
				}
			})
		case qtype.JournalEntry.String(): //分录题，暂时按照填空题处理
			//options := make(map[string][]string)
			options := []string{}
			fillQue := entity.FillQue{}
			fillQue.Type = qtype.FillInTheBlank
			fillQue.Qid = qs.ID
			fillQue.Text = quesText
			// 提取填空题
			qdoc.Find("ul").Each(func(i int, selection *goquery.Selection) {
				optionLetter := selection.Find("p").Text()
				splitOptions := strings.Split(optionLetter, ":")
				validParts := splitOptions[:0]
				for _, part := range splitOptions {
					if part != "" { // 忽略空值
						validParts = append(validParts, part)
					}
				}
				for j, _ := range validParts {
					options = append(options, fmt.Sprintf("%d", j+1))
					//options[s] = []string{fmt.Sprintf("填空%d", j+1)}
				}
			})
			fillQue.OpFromAnswer = options
			fillQuestion = append(fillQuestion, fillQue)
		default:
			log2.Print(log2.INFO, "[", cache.Name, "] ", "未知题目类型，类型为：", quesType, "默认采用论述题方式")
			options := make(map[string][]string)
			essayQue := entity.EssayQue{}
			essayQue.Type = qtype.Essay
			essayQue.Qid = qs.ID
			essayQue.Text = quesText
			// 简答暂时未发现有多个textarea标签出现 不做多答案处理
			options["论述"] = []string{"论述"}
			essayQue.OpFromAnswer = options
			essayQuestion = append(essayQuestion, essayQue)
		}
	}

	questionEntity.Choice = workQuestion
	questionEntity.Judge = judgeQuestion
	questionEntity.Fill = fillQuestion
	questionEntity.Short = shortQuestion
	questionEntity.TermExplanation = termQuestion
	questionEntity.Essay = essayQuestion
	questionEntity.Matching = matchingQuestion
	questionEntity.Other = otherQuestion
	return questionEntity, nil
}

// Deprecated: 此方法将在未来版本中删除
// 定义题型处理策略函数类型
type problemMessageStrategy func(paperTitle, context string, topic entity.ExamTurn) que_core.AIChatMessages

// Deprecated: 此方法将在未来版本中删除
// 策略映射表：题型 -> 处理函数
var problemStrategies = map[string]problemMessageStrategy{
	"单选题":  handleSingleChoice,
	"多选题":  handleMultipleChoice,
	"判断题":  handleTrueFalse,
	"填空题":  handleFillInTheBlank,
	"简答题":  handleShortAnswer,
	"名词解释": handleTermExplanationAnswer,
	"论述题":  handleEssayAnswer,
	"连线题":  handleMatchingAnswer,
}

// 构建AI问答消息
func AIProblemMessage(paperTitle, typeStr string, topic entity.ExamTurn) que_core.AIChatMessages {

	//context := buildProblemContext(typeStr, topic)
	//
	//// 查找对应的处理策略
	//if strategy, exists := problemStrategies[typeStr]; exists {
	//	return strategy(paperTitle, context, topic)
	//}
	switch typeStr {
	case qtype.SingleChoice.String():
		return que_core.BuildAiQuestionMessage(topic.XueXChoiceQue.TurnStandardQuestion())
	case qtype.MultipleChoice.String():
		return que_core.BuildAiQuestionMessage(topic.XueXChoiceQue.TurnStandardQuestion())
	case qtype.TrueOrFalse.String():
		return que_core.BuildAiQuestionMessage(topic.XueXJudgeQue.TurnStandardQuestion())
	case qtype.FillInTheBlank.String():
		return que_core.BuildAiQuestionMessage(topic.XueXFillQue.TurnStandardQuestion())
	case qtype.ShortAnswer.String():
		return que_core.BuildAiQuestionMessage(topic.XueXShortQue.TurnStandardQuestion())
	case qtype.TermExplanation.String():
		return que_core.BuildAiQuestionMessage(topic.XueXTermExplanationQue.TurnStandardQuestion())
	case qtype.Essay.String():
		return que_core.BuildAiQuestionMessage(topic.XueXEssayQue.TurnStandardQuestion())
	case qtype.Matching.String():
		return que_core.BuildAiQuestionMessage(topic.XueXMatchingQue.TurnStandardQuestion())
	case qtype.QueOther.String():
		resQue := topic.XueXOtherQue.TurnStandardQuestion()
		resQue.Type = qtype.ShortAnswer.String() //按照简答题方式处理
		return que_core.BuildAiQuestionMessage(resQue)
	}

	// 默认返回空消息
	return que_core.AIChatMessages{Messages: []que_core.Message{}}

	//return que_core.BuildAiQuestionMessage(topic.Question)
}

// Deprecated: 此方法将在未来版本中删除
// buildProblemContext 构建通用的题目上下文
func buildProblemContext(problemTypeStr string, topic entity.ExamTurn) (context string) {
	switch problemTypeStr {
	case qtype.SingleChoice.String():
		context += topic.XueXChoiceQue.Text + "\n"
		for c, q := range topic.XueXChoiceQue.Options {
			context += fmt.Sprintf("%v. %v\n", c, q)
		}
	case qtype.MultipleChoice.String():
		context += topic.XueXChoiceQue.Text + "\n"
		for c, q := range topic.XueXChoiceQue.Options {
			context += fmt.Sprintf("%v. %v\n", c, q)
		}
	case qtype.FillInTheBlank.String():
		context += topic.XueXFillQue.Text + "\n"
		for c, q := range topic.XueXFillQue.OpFromAnswer {
			context += fmt.Sprintf("\n%v. %v\n", c, q)
		}
	case qtype.TrueOrFalse.String():
		context += topic.XueXJudgeQue.Text + "\n"
		for c, q := range topic.XueXJudgeQue.Options {
			context += fmt.Sprintf("%v. %v\n", c, q)
		}
	case qtype.ShortAnswer.String():
		context += topic.XueXShortQue.Text + "\n"
		//for c, q := range topic.XueXShortQue.OpFromAnswer {
		//
		//	context += fmt.Sprintf("\n%v. %v", c, q)
		//}
	case qtype.TermExplanation.String(): //名词解释
		context += topic.XueXTermExplanationQue.Text + "\n"
	case qtype.Essay.String(): //论述题
		context += topic.XueXEssayQue.Text + "\n"
	case qtype.Matching.String():
		context += topic.XueXMatchingQue.Text + "\n"
		for _, option := range topic.XueXMatchingQue.Options {
			context += fmt.Sprintf("%s\n", option)
		}
		for _, sel := range topic.XueXMatchingQue.Selects {
			context += fmt.Sprintf("%s\n", sel)
		}
		context += "\n"
	}
	return context
}

// Deprecated: 此方法将在未来版本中删除
// 单选题处理策略
func handleSingleChoice(paperTitle, content string, topic entity.ExamTurn) que_core.AIChatMessages {
	problem := buildProblemHeader(paperTitle, topic.XueXChoiceQue.Type.String(), content)
	return que_core.AIChatMessages{Messages: []que_core.Message{
		{Role: "system", Content: "接下来你只需要以json格式回答选项对应内容即可，比如：[\"选项1\"]"},
		{Role: "system", Content: "就算你不知道选什么也随机选...无需回答任何解释！！！"},
		{Role: "system", Content: exampleSingleChoice()},
		{Role: "user", Content: problem},
	}}
}

// Deprecated: 此方法将在未来版本中删除
// 多选题处理策略
func handleMultipleChoice(paperTitle, context string, topic entity.ExamTurn) que_core.AIChatMessages {
	problem := buildProblemHeader(paperTitle, topic.XueXChoiceQue.Type.String(), context)
	return que_core.AIChatMessages{Messages: []que_core.Message{
		{Role: "system", Content: "接下来你只需要以json格式回答选项对应内容即可，比如：[\"选项1\",\"选项2\"]"},
		{Role: "system", Content: "就算你不知道选什么也随机选...无需回答任何解释！！！"},
		{Role: "system", Content: exampleMultipleChoice()},
		{Role: "user", Content: problem},
	}}
}

// Deprecated: 此方法将在未来版本中删除
// 判断题处理策略
func handleTrueFalse(paperTitle, content string, topic entity.ExamTurn) que_core.AIChatMessages {
	problem := buildProblemHeader(paperTitle, topic.XueXJudgeQue.Type.String(), content)
	return que_core.AIChatMessages{Messages: []que_core.Message{
		{Role: "system", Content: "接下来你只能利用json格式回答“正确”或者“错误”这两个选项，比如：[\"正确\"]，不要回答A、B选项字母！！！"},
		{Role: "system", Content: exampleTrueFalse()},
		{Role: "user", Content: problem},
	}}
}

// Deprecated: 此方法将在未来版本中删除
// 填空题处理策略
func handleFillInTheBlank(paperTitle, content string, topic entity.ExamTurn) que_core.AIChatMessages {
	problem := buildProblemHeader(paperTitle, topic.XueXFillQue.Type.String(), content)
	return que_core.AIChatMessages{Messages: []que_core.Message{
		{Role: "system", Content: "其中，“（answer_数字）”相关字样的地方是你需要填写答案的地方...格式：[\"答案1\",\"答案2\"]"},
		//{Role: "system", Content: "就算你不知道填什么也随机选...无需回答任何解释！！！"},
		{Role: "system", Content: exampleFillInTheBlank()},
		{Role: "user", Content: problem},
	}}
}

// Deprecated: 此方法将在未来版本中删除
// 简答题处理策略
func handleShortAnswer(paperTitle, content string, topic entity.ExamTurn) que_core.AIChatMessages {
	problem := buildProblemHeader(paperTitle, topic.XueXShortQue.Type.String(), content)
	return que_core.AIChatMessages{Messages: []que_core.Message{
		{Role: "system", Content: "这是一个简答题，接下来你只需要以json格式回复答案即可，比如：[\"答案\"]，注意不要拆分答案！！！"},
		{Role: "system", Content: exampleShortAnswer()},
		{Role: "user", Content: problem},
	}}
}

// Deprecated: 此方法将在未来版本中删除
// 名词解释处理策略
func handleTermExplanationAnswer(paperTitle, content string, topic entity.ExamTurn) que_core.AIChatMessages {
	problem := buildProblemHeader(paperTitle, topic.XueXTermExplanationQue.Type.String(), content)
	return que_core.AIChatMessages{Messages: []que_core.Message{
		{Role: "system", Content: "这是一个名词解释题，回答时请严格遵循json格式：[\"答案\"]，注意不要拆分答案！！！"},
		{Role: "system", Content: exampleTermExplanationAnswer()},
		{Role: "user", Content: problem},
	}}
}

// Deprecated: 此方法将在未来版本中删除
// 论述题处理策略
func handleEssayAnswer(paperTitle, content string, topic entity.ExamTurn) que_core.AIChatMessages {
	problem := buildProblemHeader(paperTitle, topic.XueXEssayQue.Type.String(), content)
	return que_core.AIChatMessages{Messages: []que_core.Message{
		{Role: "system", Content: `这是一个论述题，回答时请严格遵循json格式：["答案"]，注意不要拆分答案！！！`},
		{Role: "system", Content: exampleEssayAnswer()},
		{Role: "user", Content: problem},
	}}
}

// Deprecated: 此方法将在未来版本中删除
// 连线题处理策略
func handleMatchingAnswer(paperTitle, context string, topic entity.ExamTurn) que_core.AIChatMessages {
	problem := buildProblemHeader(paperTitle, topic.XueXChoiceQue.Type.String(), context)
	return que_core.AIChatMessages{Messages: []que_core.Message{
		{Role: "system", Content: "接下来你只需要以json格式回答选项对应内容即可，比如：[\"xxx->xxx\",\"xxx->xxx\"]"},
		{Role: "system", Content: "就算你不知道选什么也随机按指定要求格式回答...无需回答任何解释！！！"},
		{Role: "system", Content: exampleMatchingAnswer()},
		{Role: "user", Content: problem},
	}}
}

// Deprecated: 此方法将在未来版本中删除
// 构建题目头部信息
func buildProblemHeader(testPaperTitle, topicType, context string) string {
	return fmt.Sprintf(`试卷名称：%s
题目类型：%s
题目内容：%s`, testPaperTitle, topicType, context)
}

// Deprecated: 此方法将在未来版本中删除
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

那么你应该回答选项B的内容：["1949年10月1日"]`
}

// Deprecated: 此方法将在未来版本中删除
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

那么你应该回答选项A、B、D的内容：["资本主义扩大再生产的源泉","社会财富占有两极分化的重要原因","资本主义社会失业现象产生的根源"]`
}

// Deprecated: 此方法将在未来版本中删除
// 判断题示例
func exampleTrueFalse() string {
	return `比如：
试卷名称：考试
题目类型：判断
题目内容：新中国是什么时候成立是1949年10月1日吗？
A. 正确
B. 错误

那么你应该回答选项A的内容：["正确"]`
}

// Deprecated: 此方法将在未来版本中删除
// 填空题示例
func exampleFillInTheBlank() string {
	return ` 比如：
试卷名称：考试
题目类型：填空
题目内容：新中国成立于（ ）年。
答案：1949

那么你应该回答："["1949"]"`
}

// Deprecated: 此方法将在未来版本中删除
// 简答题
func exampleShortAnswer() string {
	return `比如：
试卷名称：考试
题目类型：简答
题目内容：请简述中国和外国的国别 differences
答案：中国和外国的国别 differences

那么你应该回答： ["中国和外国的国别 differences"]`
}

// Deprecated: 此方法将在未来版本中删除
// 名词解释
func exampleTermExplanationAnswer() string {
	return `比如：
试卷名称：考试
题目类型：名词解释
题目内容：绿色设计
答案：绿色设计是指在产品、建筑、工程或系统设计的全过程中，将环境保护和可持续发展理念融入其中的一种设计方法。

那么你应该回答： ["绿色设计是指在产品、建筑、工程或系统设计的全过程中，将环境保护和可持续发展理念融入其中的一种设计方法。"]`
}

// Deprecated: 此方法将在未来版本中删除
// 论述题
func exampleEssayAnswer() string {
	return `比如：
试卷名称：考试
题目类型：论述题
题目内容：试述设计艺术的构成元素
答案：设计艺术的构成元素包括点、线、面、形体、色彩、质感与空间等。它们相互依存、互为补充，通过合理的组织和运用，形成和谐、统一而富有美感的设计作品。

那么你应该回答（回答字数不能少于500字）： ["设计艺术的构成元素包括点、线、面、形体、色彩、质感与空间等。它们相互依存、互为补充，通过合理的组织和运用，形成和谐、统一而富有美感的设计作品。"]`
}

// Deprecated: 此方法将在未来版本中删除
// 连线题
func exampleMatchingAnswer() string {
	return `比如：
试卷名称：考试
题目类型：连线题
题目内容：
5.[连线题] 下列认知心理学家与其所做的经典研究之间的关系：

1、桑代克 ()
2、威特金 ()
3、凯利 ()
4、卡特尔 ()

A、迷箱实验
B、角色建构测验
C、16PF
D、棒框实验

答案：第一空：桑代克->迷箱实验、威特金->棒框实验、凯利->角色建构测验、卡特尔->16PF

那么你应该回答： ["桑代克->迷箱实验","威特金->棒框实验","凯利->角色建构测验","卡特尔->16PF"]`

}
