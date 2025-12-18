package xuexitong

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/yatori-dev/yatori-go-core/api/xuexitong"
	"github.com/yatori-dev/yatori-go-core/models/ctype"
	"github.com/yatori-dev/yatori-go-core/que-core/aiq"
	"github.com/yatori-dev/yatori-go-core/que-core/external"
	"github.com/yatori-dev/yatori-go-core/que-core/qtype"
	log2 "github.com/yatori-dev/yatori-go-core/utils/log"
)

// 学习通考试结构体
type XXTWork struct {
	Name              string `json:"name"`
	Status            string `json:"status"`
	RemainTime        string `json:"remain_time"`
	RawURL            string `json:"raw_url"`
	Params            map[string]string
	CourseId          string `json:"course_id"`
	UserId            string `json:"user_id"`
	ClazzId           string `json:"clazz_id"`
	Type              string `json:"type"`
	EncTask           string `json:"enc_task"`
	TaskRefId         string `json:"taskrefId"`
	MsgId             string `json:"msgId"`
	CaptchaCaptchaId  string
	ExamRelationId    string
	AnswerId          string `json:"answerId"`
	Cpi               string
	Validate          string //过验证码用的
	QuestionTotal     int    //题目数量
	Enc               string
	EncRemainTime     string
	EncLastUpdateTime string
}

type XXTWorkQuestion struct {
	xuexitong.XXTWorkQuestionSubmitEntity
}

// 拉取考试列表
func PullWorkListAction(cache *xuexitong.XueXiTUserCache, course XueXiTCourse) ([]XXTWork, error) {
	examList := []XXTWork{}
	examListHtml, err := cache.PullWorkListHtmlApi(course.CourseID, course.Key, fmt.Sprintf("%d", course.Cpi), 3, nil)
	if err != nil {
		return examList, err
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(examListHtml))
	if err != nil {
		log.Fatal(err)
	}

	// 遍历所有 <li>
	doc.Find("ul.nav li").Each(func(i int, li *goquery.Selection) {
		rawURL, _ := li.Attr("data")

		// 解析 URL 参数
		parsed, _ := url.Parse(rawURL)
		params := map[string]string{}
		for k, v := range parsed.Query() {
			if len(v) > 0 {
				params[k] = v[0]
			}
		}

		div := li.Find("div")

		name := strings.TrimSpace(div.Find("p").Text())

		spanList := div.Find("span")
		status := strings.TrimSpace(spanList.Eq(0).Text())

		remain := ""
		if spanList.Length() > 1 {
			remain = strings.TrimSpace(spanList.Eq(1).Text())
		}

		exam := XXTWork{
			Name:       name,
			Status:     status,
			RemainTime: remain,
			RawURL:     rawURL,
			Params:     params,
			TaskRefId:  params["taskrefId"],
			CourseId:   params["courseId"],
			UserId:     params["userId"],
			ClazzId:    params["clazzId"],
			Type:       params["type"],
			EncTask:    params["enc_task"],
			MsgId:      params["msgId"],
		}
		examList = append(examList, exam)
	})

	return examList, nil
}

// EnterWorkAction 进入作业
func EnterWorkAction(cache *xuexitong.XueXiTUserCache, exam *XXTWork) error {
	//这一步拉取必要的参数，比如滑块验证码参数等,注意这里的refererUrl会在后面的滑块验证码中用到
	enterHtml, refererUrl, err := cache.PullWorkEnterInformHtmlApi(exam.TaskRefId, exam.MsgId, exam.CourseId, exam.UserId, exam.ClazzId, exam.Type, exam.EncTask, 3, nil)
	if err != nil {
		//fmt.Println(refererUrl)
		return err
	}

	re := regexp.MustCompile(`共包含\s*(\d+)\s*道题目`)
	match := re.FindStringSubmatch(enterHtml)

	if len(match) > 1 {
		count, _ := strconv.Atoi(match[1])
		exam.QuestionTotal = count
	}
	//如果是待重做
	if strings.Contains(enterHtml, "待重做") {
		re = regexp.MustCompile(`共\s*(\d+)\s*题`)
		match = re.FindStringSubmatch(enterHtml)

		if len(match) > 1 {
			count, _ := strconv.Atoi(match[1])
			exam.QuestionTotal = count
		}
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(enterHtml))
	if err != nil {
		log.Fatal(err)
	}
	type HiddenField struct {
		ID    string
		Name  string
		Value string
	}

	var fields []HiddenField
	// 选择所有隐藏字段
	doc.Find("input[type='hidden']").Each(func(i int, sel *goquery.Selection) {
		id, _ := sel.Attr("id")
		name, _ := sel.Attr("name")
		value, _ := sel.Attr("value")

		fields = append(fields, HiddenField{
			ID:    id,
			Name:  name,
			Value: value,
		})
	})

	// 输出结果
	for _, f := range fields {
		if f.ID == "captchaCaptchaId" {
			exam.CaptchaCaptchaId = f.Value
		}
		if f.ID == "testPaperId" {
			exam.ExamRelationId = f.Value
		}
		if f.ID == "testUserRelationId" {
			exam.AnswerId = f.Value
		}
		if f.ID == "cpi" {
			exam.Cpi = f.Value
		}
		//fmt.Printf("ID: %-35s Name: %-20s Value: %s\n", f.ID, f.Name, f.Value)
	}
	if exam.CaptchaCaptchaId != "" {
		slider := XueXiTSlider{
			CaptchaId: exam.CaptchaCaptchaId,
			Referer:   refererUrl,
		}
		for {
			validate, passErr := slider.Pass(cache)
			if passErr != nil {
				if strings.Contains(passErr.Error(), `"result":false`) {
					continue
				}
			}
			exam.Validate = validate
			break //如果成功了那么直接退出循环
		}
	}
	cpi, answerId, enc := extractParams(enterHtml)
	exam.Cpi = cpi
	exam.AnswerId = answerId
	exam.Enc = enc
	//pullPaperHtml, err := cache.PullExamPaperHtmlApi(exam.CourseId, exam.ClazzId, exam.ExamRelationId, "0", exam.AnswerId, exam.Cpi, "1", xuexitong.IMEI, exam.Validate, "0", 3, nil)
	pullPaperHtml, err := cache.PullWorkPaperHtmlApi(exam.CourseId, exam.ClazzId, exam.TaskRefId, "0", exam.MsgId, exam.Cpi, exam.AnswerId, exam.Enc, "1", 3, nil)
	if err != nil {
		return err

	}
	qsEntity, err1 := HtmlWorkQuestionTurnEntity(pullPaperHtml)
	if err1 != nil {
		return err1
	}
	exam.Enc = qsEntity.Enc
	exam.EncRemainTime = qsEntity.EncRemainTime
	exam.EncLastUpdateTime = qsEntity.EncLastUpdateTime

	//fmt.Println(pullPaperHtml)
	return nil
}

func extractParams(js string) (cpi, workAnswerId, enc string) {
	//cpiRe := regexp.MustCompile(`cpi=(\d+).*?workAnswerId=(\d+).*?enc=([a-fA-F0-9]+)`)
	cpiRe := regexp.MustCompile(`cpi=(\d+).*?`)
	cpiMatch := cpiRe.FindStringSubmatch(js)
	answerIdRe := regexp.MustCompile(`workAnswerId=(\d+).*?`)
	answerIdMatch := answerIdRe.FindStringSubmatch(js)
	encRe := regexp.MustCompile(`enc=([a-fA-F0-9]+)`)
	encMatch := encRe.FindStringSubmatch(js)
	cpiC := ""
	workAnswerIdC := ""
	encC := ""
	if len(cpiMatch) > 1 {
		cpiC = cpiMatch[1]
	}
	if len(answerIdMatch) > 1 {
		workAnswerIdC = answerIdMatch[1]
	}
	if len(encMatch) > 1 {
		encC = encMatch[1]
	}
	return cpiC, workAnswerIdC, encC
}

// 拉取题目
func (exam *XXTWork) PullWorkQuestionAction(cache *xuexitong.XueXiTUserCache, index int /*第几道题*/) (XXTWorkQuestion, error) {
	pullQuestion, err1 := cache.PullWorkQuestionApi(exam.CourseId, exam.ClazzId, exam.TaskRefId, "0", exam.MsgId, exam.Cpi, exam.AnswerId, exam.Enc, "1", index, 3, nil)
	if err1 != nil {
		return XXTWorkQuestion{}, err1
	}
	//isLastQuestion := strings.Contains(pullQuestion, `<div class="lastQuestion"> 已经是最后一题了</div>`)
	qsEntity, err := HtmlWorkQuestionTurnEntity(pullQuestion)
	//fmt.Println(pullPaperHtml)
	if err != nil {
		return XXTWorkQuestion{}, err
	}
	qsEntity.ExamRelationId = exam.ExamRelationId
	qsEntity.AnswerId = exam.AnswerId
	qsEntity.RemainTimeParam = exam.RemainTime
	qsEntity.Tid = exam.TaskRefId
	return qsEntity, nil
}

// AI写题
func (question *XXTWorkQuestion) WriteQuestionForAIAction(cache *xuexitong.XueXiTUserCache, aiUrl, model string, aiType ctype.AiType, apiKey string) error {
	aiChatMessages := aiq.BuildAiQuestionMessage(question.Question)

	aiAnswer, err := aiq.AggregationAIApi(aiUrl, model, aiType, aiChatMessages, apiKey)
	if err != nil {
		log2.Print(log2.INFO, `[`, cache.Name, `] `, log2.BoldRed, "Ai异常，返回信息：", err.Error())
		//os.Exit(0)
	}

	var answers []string
	err = json.Unmarshal([]byte(aiAnswer), &answers)

	if err != nil {
		answers = []string{"A"}
		//fmt.Println("AI回复解析错误，已采用随机答案:", err, fmt.Sprintf("题目：%v \nAI回复： %v", aiChatMessages, aiAnswer))
		log2.Print(log2.INFO, "AI回复解析错误，已采用随机答案:", err.Error(), fmt.Sprintf("题目：%v \nAI回复： %v", aiChatMessages, aiAnswer))
	}
	question.Question.Answers = answers
	return nil
}

// 外置题库写题
func (question *XXTWorkQuestion) WriteQuestionForExternalAction(exUrl string) {
	//赋值选项
	request, err := external.ApiQueRequest(question.Question, exUrl, 5, nil)
	if err != nil {
		log.Fatal(err)
	}
	//赋值答案
	var answers []string
	if request.Code == 404 {
		answers = []string{"A"}
		log2.Print(log2.INFO, "外置题库未找到答案，已使用默认答案A:", fmt.Sprintf("题目：%v \n外置题库回复： %v", question.Question, request))
		request.Question.Answers = answers
		return
	}
	if request.Question.Answers == nil {
		answers = []string{"A"}
		fmt.Println("回复解析错误:", err, fmt.Sprintf("题目：%v \n外置题库回复： %v", question.Question, request))
		request.Question.Answers = answers
	}
	question.Question.Answers = request.Question.Answers
}

// 学习通内置题库写题
func (question *XXTWorkQuestion) WriteQuestionForXXTAIAction(cache *xuexitong.XueXiTUserCache, classId, courseId, cpi string) error {
	aiChatMessages := aiq.BuildAiQuestionMessage(question.Question)

	informHtml, err := cache.XXTAiInformApi(classId, courseId, cpi, 3, nil)
	if err != nil {
		panic(err)
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(informHtml))
	if err != nil {
		panic(err)
	}
	// 再给你示例获取其它值（你可以按需扩展）
	get := func(id string) string {
		v, _ := doc.Find("#" + id).Attr("value")
		return v
	}
	content := ""
	//去除前后"
	trimQuotes := func(s string) string {
		if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
			return s[1 : len(s)-1]
		}
		return s
	}
	for _, msgEntity := range aiChatMessages.Messages {
		msg, _ := json.Marshal(msgEntity.Content)
		content += trimQuotes(string(msg))
	}
	re := regexp.MustCompile(`"studentName"\s*:\s*"([^"]+)"`)
	match := re.FindStringSubmatch(informHtml)
	studentName := ""
	if len(match) > 1 {
		//fmt.Println("courseName:", match[1])
		studentName = match[1]
	} else {
		fmt.Println("未找到 studentName")
	}
	aiAnswer, err := cache.XXTAiAnswerApi(get("cozeEnc"), get("userId"), get("courseId"), get("clazzId"), get("conversationId"), get("courseName"), studentName, get("personId"), content, 3, nil)
	if err != nil {
		panic(err)
	}
	var answers []string
	err = json.Unmarshal([]byte(aiAnswer), &answers)
	if err != nil {
		answers = []string{"A"}
		//fmt.Println("AI回复解析错误，已采用随机答案:", err, fmt.Sprintf("题目：%v \nAI回复： %v", aiChatMessages, aiAnswer))
		log2.Print(log2.INFO, "AI回复解析错误，已采用随机答案:", err.Error(), fmt.Sprintf("题目：%v \nAI回复： %v", aiChatMessages, aiAnswer))
	}
	question.Question.Answers = answers
	return nil
}

// 提交学习通考试答案
func (question *XXTWorkQuestion) SubmitWorkAnswerAction(cache *xuexitong.XueXiTUserCache, isSubmit bool /*是否提交，true为提交，false为暂存*/) (string, error) {
	//api, err := cache.SubmitExamAnswerApi(exam.ClazzId, exam.CourseId, exam.Paper.TestPaperId, exam.Paper.TestUserRelationId, exam.Cpi, exam.Paper.RemainTime, exam.Paper.EncRemainTime, exam.Paper.EncLastUpdateTime, exam.ExamRelationId, exam.AnswerId, exam.RemainTime, !isSubmit, exam.Paper.Enc, exam.Paper.EnterPageTime, exam.Paper.XXTExamQuestion.QuestionId, exam.Paper.Type, exam.Paper.XXTExamQuestion.TypeName, &exam.Paper)
	api, err := cache.SubmitWorkAnswerApi(&question.XXTWorkQuestionSubmitEntity, !isSubmit)
	if err != nil {
		return "", err
	}
	//fmt.Println(api)
	return api, nil
}

// html转Exam实体
func HtmlWorkQuestionTurnEntity(paperHtml string) (XXTWorkQuestion, error) {
	//xxtExamPaper := XXTExamPaper{}
	question := XXTWorkQuestion{}
	// 使用 goquery 解析 HTML
	paperDoc, err := goquery.NewDocumentFromReader(bytes.NewReader([]byte(paperHtml)))
	if err != nil {
		log.Fatal(err)
	}
	questionId, exists := paperDoc.Find("#questionId").Attr("value") //题目id
	if exists {
		//fmt.Println("question:", questionId)
		log2.Print(log2.DEBUG, questionId)
	}
	questionTypeCode, exists := paperDoc.Find(`input[name="` + `type` + questionId + `"]`).Attr("value")
	if exists {
		question.QuestionTypeCode = questionTypeCode
		log2.Print(log2.DEBUG, questionTypeCode)
		//fmt.Println("questionType:", questionTypeCode)
	}
	questionTypeStr, _ := paperDoc.Find(`span.focusSpan`).Attr("aria-label")
	// 去除前置题序，如 "1. "、"12. "
	re := regexp.MustCompile(`^\s*\d+\.\s*`)
	questionTypeStr = re.ReplaceAllString(questionTypeStr, "")

	get := func(id string) string {
		v, _ := paperDoc.Find("#" + id).Attr("value")
		return v
	}
	getForName := func(id string) string {
		v, _ := paperDoc.Find(`input[name="` + id + `"]`).Attr("value")
		return v
	}
	switch questionTypeCode {
	case "0": //单选题
		turn, err1 := singleWorkTurn(paperDoc)
		if err1 != nil {
			fmt.Println(err1)
		}
		question = turn
		//fmt.Println(turn)
	case "1": //多选题
		turn, err1 := multipleWorkTurn(paperDoc)
		if err1 != nil {
			fmt.Println(err1)
		}
		question = turn
		//fmt.Println(turn)
	case "2": //填空题
		turn, err1 := fillWorkTurn(paperDoc)
		if err1 != nil {
			fmt.Println(err1)
		}
		question = turn
		//fmt.Println(turn)
	case "3": //判断题
		turn, err1 := trueOrFalseWorkTurn(paperDoc)
		if err1 != nil {
			fmt.Println(err1)
		}
		question = turn
		//fmt.Println(turn)
	case "4": //简答题
		turn, err1 := shortAnswerWorkTurn(paperDoc)
		if err1 != nil {
			fmt.Println(err1)
		}
		question = turn
		//fmt.Println(turn)
	case "6": //论述题
		turn, err1 := essayWorkTurn(paperDoc)
		if err1 != nil {
			fmt.Println(err1)
		}
		question = turn
		//fmt.Println(turn)
	default:
		fmt.Printf("未知题型%s   %s", questionTypeCode, questionTypeStr)
	}
	question.CourseId = get("courseId")
	question.TestUserRelationId = get("testUserRelationId")
	question.ClassId = get("classId")
	question.Type = get("type")
	question.IsPhone = get("isphone")
	question.Imei = get("imei")
	question.SubCount = get("subCount")
	question.RemainTime = get("remainTime")
	question.TempSave = get("tempSave")
	question.TimeOver = get("timeOver")
	question.EncRemainTime = get("encRemainTime")
	question.EncLastUpdateTime = get("encLastUpdateTime")
	question.Cpi = get("cpi")
	question.Enc = get("enc")
	question.Source = get("source")
	question.Score = getForName("score" + questionId)
	question.UserId = get("userId")
	question.EnterPageTime = get("enterPageTime")
	question.AnsweredView = get("answeredView")
	question.AnsweredView = get("answeredView")
	question.PaperGroupId = get("paperGroupId")
	question.WordId = get("workId")
	question.CurrentTime = get("currentTime")
	question.CurrentCpi = get("currentCpi")
	question.CurrentUploadEnc = get("currentUploadEnc")
	question.MatchEnc = get("matchEnc")
	question.Cfid = get("cfid")
	question.AddTimes = get("addTimes")
	question.LimitWorkSubmitTimes = get("limitWorkSubmitTimes")
	question.QuestionTypeCode = questionTypeCode
	question.QuestionTypeStr = questionTypeStr
	question.EncWork = get("encWork")
	question.Index = get("index")
	return question, nil
}

// 单选题转换
func singleWorkTurn(paperDoc *goquery.Document) (XXTWorkQuestion, error) {
	question := XXTWorkQuestion{}
	question.QType = qtype.SingleChoice
	question.Question.Type = question.QType.String()
	paperDoc.Find("div.singleQuesId").Each(func(i int, sel *goquery.Selection) {
		questionId, exists := sel.Attr("data") //题目id
		if exists {
			question.QuestionId = questionId
			//fmt.Println("question:", questionId)
		}

		//题目
		//title := strings.TrimSpace(sel.Find(`.workWrap p`).First().Text())
		title := ""
		//截取题目
		sel.Find(`.workWrap`).Each(func(i int, sel *goquery.Selection) {
			title = extractQuestion(sel)
		})
		//fmt.Println(title)
		question.Question.Content = title
		resOptions := make(map[string]string)
		sel.Find(`div.centerSpan`).Each(func(i int, sel *goquery.Selection) {
			//letter := strings.TrimSpace(sel.Find(`.No`).Text())
			letter, _ := sel.Attr("id")
			text := strings.TrimSpace(sel.Find(`p`).Text())
			resOptions[letter] = letter + text
			//fmt.Println(letter, text)
		})
		resSelect := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N"}
		for _, slt := range resSelect {
			if resOptions[slt] == "" {
				break
			}
			question.Question.Options = append(question.Question.Options, resOptions[slt])
		}
	})
	return question, nil
}

// 多选题转换
func multipleWorkTurn(paperDoc *goquery.Document) (XXTWorkQuestion, error) {
	question := XXTWorkQuestion{}
	question.QType = qtype.MultipleChoice
	question.Question.Type = question.QType.String()
	paperDoc.Find("div.singleQuesId").Each(func(i int, sel *goquery.Selection) {
		questionId, exists := sel.Attr("data") //题目id
		if exists {
			question.QuestionId = questionId
			//fmt.Println("question:", questionId)
		}

		//题目
		//title := strings.TrimSpace(sel.Find(`.workWrap p`).First().Text())
		title := ""
		//截取题目
		sel.Find(`.workWrap`).Each(func(i int, sel *goquery.Selection) {
			title = extractQuestion(sel)
		})
		//fmt.Println(title)
		question.Question.Content = title
		resOptions := make(map[string]string)
		sel.Find(`div.centerSpan`).Each(func(i int, sel *goquery.Selection) {
			//letter := strings.TrimSpace(sel.Find(`.No`).Text())
			letter, _ := sel.Attr("id")
			text := strings.TrimSpace(sel.Find(`p`).Text())
			resOptions[letter] = letter + text
			//fmt.Println(letter, text)
		})
		resSelect := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N"}
		for _, slt := range resSelect {
			if resOptions[slt] == "" {
				break
			}
			question.Question.Options = append(question.Question.Options, resOptions[slt])
		}
	})

	return question, nil
}

// 填空题转换
func fillWorkTurn(paperDoc *goquery.Document) (XXTWorkQuestion, error) {
	question := XXTWorkQuestion{}
	question.QType = qtype.FillInTheBlank
	question.Question.Type = question.QType.String()
	paperDoc.Find("div.singleQuesId").Each(func(i int, sel *goquery.Selection) {
		questionId, exists := sel.Attr("data") //题目id
		if exists {
			question.QuestionId = questionId
			//fmt.Println("question:", questionId)
		}
		typeName, exists := paperDoc.Find(`input[name="` + `typeName` + questionId + `"]`).Attr("value")
		if exists {
			question.TypeName = typeName
		}

		//题目
		//title := strings.TrimSpace(sel.Find(`.workWrap p`).First().Text())
		title := ""
		//截取题目
		sel.Find(`.workWrap`).Each(func(i int, sel *goquery.Selection) {
			title = extractQuestion(sel)
		})
		//fmt.Println(title)
		question.Question.Content = title
		sel.Find(`.blankItemName`).Each(func(i int, sel *goquery.Selection) {
			letter, _ := sel.Attr(`aria-label`)
			//fmt.Println(letter)
			question.Question.Options = append(question.Question.Options, letter)
		})

	})
	return question, nil
}

// 判断题
func trueOrFalseWorkTurn(paperDoc *goquery.Document) (XXTWorkQuestion, error) {
	question := XXTWorkQuestion{}
	question.QType = qtype.TrueOrFalse
	question.Question.Type = question.QType.String()
	paperDoc.Find("div.singleQuesId").Each(func(i int, sel *goquery.Selection) {
		questionId, exists := sel.Attr("data") //题目id
		if exists {
			question.QuestionId = questionId
			//fmt.Println("question:", questionId)
		}
		typeName, exists := paperDoc.Find(`input[name="` + `typeName` + questionId + `"]`).Attr("value")
		if exists {
			question.TypeName = typeName
		}

		//题目
		//title := strings.TrimSpace(sel.Find(`.workWrap p`).First().Text())
		title := ""
		//截取题目
		sel.Find(`.workWrap`).Each(func(i int, sel *goquery.Selection) {
			title = extractQuestion(sel)
		})
		//fmt.Println(title)
		question.Question.Content = title
		sel.Find(`.Answer`).Each(func(i int, sel *goquery.Selection) {
			letter := sel.Find("span.check").Text()
			text := sel.Find("p.fl").Text()

			//fmt.Println(letter, text)
			question.Question.Options = append(question.Question.Options, letter+text)
		})

	})
	return question, nil
}

// 简答题
func shortAnswerWorkTurn(paperDoc *goquery.Document) (XXTWorkQuestion, error) {
	question := XXTWorkQuestion{}
	question.QType = qtype.ShortAnswer
	question.Question.Type = question.QType.String()
	paperDoc.Find("div.singleQuesId").Each(func(i int, sel *goquery.Selection) {
		questionId, exists := sel.Attr("data") //题目id
		if exists {
			question.QuestionId = questionId
			//fmt.Println("question:", questionId)
		}
		typeName, exists := paperDoc.Find(`input[name="` + `typeName` + questionId + `"]`).Attr("value")
		if exists {
			question.TypeName = typeName
		}

		//题目
		//title := strings.TrimSpace(sel.Find(`.workWrap p`).First().Text())
		title := ""
		//截取题目
		sel.Find(`.workWrap`).Each(func(i int, sel *goquery.Selection) {
			title = extractQuestion(sel)
		})
		//fmt.Println(title)
		question.Question.Content = title

	})
	return question, nil
}

// 论述题
func essayWorkTurn(paperDoc *goquery.Document) (XXTWorkQuestion, error) {
	question := XXTWorkQuestion{}
	question.QType = qtype.Essay
	question.Question.Type = question.QType.String()
	paperDoc.Find("div.singleQuesId").Each(func(i int, sel *goquery.Selection) {
		questionId, exists := sel.Attr("data") //题目id
		if exists {
			question.QuestionId = questionId
			//fmt.Println("question:", questionId)
		}
		typeName, exists := paperDoc.Find(`input[name="` + `typeName` + questionId + `"]`).Attr("value")
		if exists {
			question.TypeName = typeName
		}

		//题目
		//title := strings.TrimSpace(sel.Find(`.workWrap p`).First().Text())
		title := ""
		//截取题目
		sel.Find(`.workWrap`).Each(func(i int, sel *goquery.Selection) {
			title = extractQuestion(sel)
		})
		//fmt.Println(title)
		question.Question.Content = title

	})
	return question, nil
}
