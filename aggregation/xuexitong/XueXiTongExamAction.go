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
	"github.com/yatori-dev/yatori-go-core/que-core/aiq"
	"github.com/yatori-dev/yatori-go-core/que-core/qtype"
	log2 "github.com/yatori-dev/yatori-go-core/utils/log"
)

// 学习通考试结构体
type XXTExam struct {
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

// 考试试卷信息
type XXTExamPaper struct {
	CourseId           string
	TestPaperId        string
	TestUserRelationId string
	ClassId            string
	Type               string
	IsPhone            string
	Imei               string
	SubCount           string
	RemainTime         string
	TempSave           string
	TimeOver           string
	EncRemainTime      string
	EncLastUpdateTime  string
	Cpi                string
	Enc                string
	Source             string
	UserId             string
	EnterPageTime      string
	AnsweredView       string
	ExitdTime          string
	PaperGroupId       string
	TypeName           string
	QuestionTotal      int //一共几道题
	ExamRelationId     string
	AnswerId           string
	//XXTExamQuestion    xuexitong.XXTExamQuestion //题目
}

type XXTExamQuestion struct {
	xuexitong.XXTExamQuestionSubmitEntity
}

// 拉取考试列表
func PullExamListAction(cache *xuexitong.XueXiTUserCache, course XueXiTCourse) ([]XXTExam, error) {
	examList := []XXTExam{}
	examListHtml, err := cache.PullExamListHtmlApi(course.CourseID, course.Key, fmt.Sprintf("%d", course.Cpi), 3, nil)
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

		exam := XXTExam{
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

// EnterExamAction 进入考试
func EnterExamAction(cache *xuexitong.XueXiTUserCache, exam *XXTExam) error {
	//这一步拉取必要的参数，比如滑块验证码参数等,注意这里的refererUrl会在后面的滑块验证码中用到
	enterHtml, refererUrl, err := cache.PullExamEnterInformHtmlApi(exam.TaskRefId, exam.MsgId, exam.CourseId, exam.UserId, exam.ClazzId, exam.Type, exam.EncTask, 3, nil)
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
	pullPaperHtml, err := cache.PullExamPaperHtmlApi(exam.CourseId, exam.ClazzId, exam.ExamRelationId, "0", exam.AnswerId, exam.Cpi, "1", xuexitong.IMEI, exam.Validate, "0", 3, nil)
	if err != nil {
		return err

	}
	qsEntity, err1 := HtmlQuestionTurnEntity(pullPaperHtml)
	if err1 != nil {
		return err1
	}
	exam.Enc = qsEntity.Enc
	exam.EncRemainTime = qsEntity.EncRemainTime
	exam.EncLastUpdateTime = qsEntity.EncLastUpdateTime

	//fmt.Println(pullPaperHtml)
	return nil
}

// 拉取题目
func (exam *XXTExam) PullExamQuestionAction(cache *xuexitong.XueXiTUserCache, index int /*第几道题*/) (XXTExamQuestion, error) {
	pullQuestion, err1 := cache.PullExamQuestionApi(exam.CourseId, exam.ClazzId, exam.ExamRelationId, exam.AnswerId, exam.Cpi, exam.EncRemainTime, exam.Enc, exam.EncLastUpdateTime, index)
	if err1 != nil {
		return XXTExamQuestion{}, err1
	}
	//isLastQuestion := strings.Contains(pullQuestion, `<div class="lastQuestion"> 已经是最后一题了</div>`)
	qsEntity, err := HtmlQuestionTurnEntity(pullQuestion)
	//fmt.Println(pullPaperHtml)
	if err != nil {
		return XXTExamQuestion{}, err
	}
	qsEntity.ExamRelationId = exam.ExamRelationId
	qsEntity.AnswerId = exam.AnswerId
	qsEntity.RemainTimeParam = exam.RemainTime
	qsEntity.Tid = exam.TaskRefId
	return qsEntity, nil
}

// AI写题
func (question *XXTExamQuestion) WriteQuestionForAIAction() {
}

// 外置题库写题
func (question *XXTExamQuestion) WriteQuestionForExternalAction() {

}

// 学习通内置题库写题
func (question *XXTExamQuestion) WriteQuestionForXXTAIAction(cache *xuexitong.XueXiTUserCache, classId, courseId, cpi string) {
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
}

// 提交学习通考试答案
func (question *XXTExamQuestion) SubmitExamAnswerAction(cache *xuexitong.XueXiTUserCache, isSubmit bool /*是否提交，true为提交，false为暂存*/) (string, error) {
	//api, err := cache.SubmitExamAnswerApi(exam.ClazzId, exam.CourseId, exam.Paper.TestPaperId, exam.Paper.TestUserRelationId, exam.Cpi, exam.Paper.RemainTime, exam.Paper.EncRemainTime, exam.Paper.EncLastUpdateTime, exam.ExamRelationId, exam.AnswerId, exam.RemainTime, !isSubmit, exam.Paper.Enc, exam.Paper.EnterPageTime, exam.Paper.XXTExamQuestion.QuestionId, exam.Paper.Type, exam.Paper.XXTExamQuestion.TypeName, &exam.Paper)
	api, err := cache.SubmitExamAnswerApi(&question.XXTExamQuestionSubmitEntity, !isSubmit)
	if err != nil {
		return "", err
	}
	//fmt.Println(api)
	return api, nil
}

// html转Exam实体
func HtmlQuestionTurnEntity(paperHtml string) (XXTExamQuestion, error) {
	//xxtExamPaper := XXTExamPaper{}
	question := XXTExamQuestion{}
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
	questionTypeStr, exists := paperDoc.Find(`input[name="` + `typeName` + questionId + `"]`).Attr("value")
	if exists {
		question.QuestionTypeStr = questionTypeStr
		log2.Print(log2.DEBUG, questionTypeStr)
		//fmt.Println("questionType:", questionTypeStr)
	}

	get := func(id string) string {
		v, _ := paperDoc.Find("#" + id).Attr("value")
		return v
	}

	switch questionTypeCode {
	case "0": //单选题
		turn, err1 := singleTurn(paperDoc)
		if err1 != nil {
			fmt.Println(err1)
		}
		question = turn
		//fmt.Println(turn)
	case "1": //多选题
		turn, err1 := multipleTurn(paperDoc)
		if err1 != nil {
			fmt.Println(err1)
		}
		question = turn
		//fmt.Println(turn)
	case "2": //填空题
		turn, err1 := fillTurn(paperDoc)
		if err1 != nil {
			fmt.Println(err1)
		}
		question = turn
		//fmt.Println(turn)
	case "3": //判断题
		turn, err1 := trueOrFalseTurn(paperDoc)
		if err1 != nil {
			fmt.Println(err1)
		}
		question = turn
		//fmt.Println(turn)
	case "4": //简答题
		turn, err1 := shortAnswerTurn(paperDoc)
		if err1 != nil {
			fmt.Println(err1)
		}
		question = turn
		//fmt.Println(turn)
	case "6": //论述题
		turn, err1 := essayTurn(paperDoc)
		if err1 != nil {
			fmt.Println(err1)
		}
		question = turn
		//fmt.Println(turn)
	}
	question.CourseId = get("courseId")
	question.TestPaperId = get("testPaperId")
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
	question.UserId = get("userId")
	question.EnterPageTime = get("enterPageTime")
	question.AnsweredView = get("answeredView")
	question.ExitdTime = get("exitdtime")
	question.PaperGroupId = get("paperGroupId")
	question.QuestionTypeCode = questionTypeCode
	question.QuestionTypeStr = questionTypeStr
	return question, nil
}

// 单选题转换
func singleTurn(paperDoc *goquery.Document) (XXTExamQuestion, error) {
	question := XXTExamQuestion{}
	question.QType = qtype.SingleChoice
	question.Question.Type = question.QType.String()
	paperDoc.Find("div.questionWrap").Each(func(i int, sel *goquery.Selection) {
		questionId, exists := sel.Attr("data") //题目id
		if exists {
			question.QuestionId = questionId
			//fmt.Println("question:", questionId)
		}

		//题目
		title := strings.TrimSpace(sel.Find(`.tit p`).First().Text())
		//fmt.Println(title)
		question.Question.Content = title
		sel.Find(`.singleChoice`).Each(func(i int, sel *goquery.Selection) {
			letter := strings.TrimSpace(sel.Find(`.No`).Text())
			text := strings.TrimSpace(sel.Find(`.answerInfo cc`).Text())
			//fmt.Println(letter, text)
			question.Question.Options = append(question.Question.Options, letter+text)
		})

	})
	return question, nil
}

// 多选题转换
func multipleTurn(paperDoc *goquery.Document) (XXTExamQuestion, error) {
	question := XXTExamQuestion{}
	question.QType = qtype.MultipleChoice
	question.Question.Type = question.QType.String()
	paperDoc.Find("div.questionWrap").Each(func(i int, sel *goquery.Selection) {
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
		title := strings.TrimSpace(sel.Find(`.tit p`).First().Text())
		//fmt.Println(title)
		question.Question.Content = title
		sel.Find(`.mulChoice`).Each(func(i int, sel *goquery.Selection) {
			letter := strings.TrimSpace(sel.Find(`.No`).Text())
			text := strings.TrimSpace(sel.Find(`.answerInfo cc`).Text())
			//fmt.Println(letter, text)
			question.Question.Options = append(question.Question.Options, letter+text)
		})

	})
	return question, nil
}

// 填空题转换
func fillTurn(paperDoc *goquery.Document) (XXTExamQuestion, error) {
	question := XXTExamQuestion{}
	question.QType = qtype.FillInTheBlank
	question.Question.Type = question.QType.String()
	paperDoc.Find("div.questionWrap").Each(func(i int, sel *goquery.Selection) {
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
		title := strings.TrimSpace(sel.Find(`.tit p`).First().Text())
		//fmt.Println(title)
		question.Question.Content = title
		sel.Find(`.completionList`).Each(func(i int, sel *goquery.Selection) {
			letter := strings.TrimSpace(sel.Find(`.grayTit`).Text())
			//fmt.Println(letter)
			question.Question.Options = append(question.Question.Options, letter)
		})

	})
	return question, nil
}

// 判断题
func trueOrFalseTurn(paperDoc *goquery.Document) (XXTExamQuestion, error) {
	question := XXTExamQuestion{}
	question.QType = qtype.TrueOrFalse
	question.Question.Type = question.QType.String()
	paperDoc.Find("div.questionWrap").Each(func(i int, sel *goquery.Selection) {
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
		title := strings.TrimSpace(sel.Find(`.tit p`).First().Text())
		//fmt.Println(title)
		question.Question.Content = title
		sel.Find(`.answerList`).Each(func(i int, sel *goquery.Selection) {
			letter := strings.TrimSpace(sel.Find(`.No`).Text())
			text := strings.TrimSpace(sel.Find(`.answerInfo`).Text())
			//fmt.Println(letter, text)
			question.Question.Options = append(question.Question.Options, letter+text)
		})

	})
	return question, nil
}

// 简答题
func shortAnswerTurn(paperDoc *goquery.Document) (XXTExamQuestion, error) {
	question := XXTExamQuestion{}
	question.QType = qtype.ShortAnswer
	question.Question.Type = question.QType.String()
	paperDoc.Find("div.questionWrap").Each(func(i int, sel *goquery.Selection) {
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
		title := strings.TrimSpace(sel.Find(`.tit p`).First().Text())
		//fmt.Println(title)
		question.Question.Content = title
		sel.Find(`.completionList`).Each(func(i int, sel *goquery.Selection) {
			letter := strings.TrimSpace(sel.Find(`.grayTit`).Text())
			//fmt.Println(letter)
			question.Question.Options = append(question.Question.Options, letter)
		})

	})
	return question, nil
}

// 论述题
func essayTurn(paperDoc *goquery.Document) (XXTExamQuestion, error) {
	question := XXTExamQuestion{}
	question.QType = qtype.Essay
	question.Question.Type = question.QType.String()
	paperDoc.Find("div.questionWrap").Each(func(i int, sel *goquery.Selection) {
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
		title := strings.TrimSpace(sel.Find(`.tit p`).First().Text())
		//fmt.Println(title)
		question.Question.Content = title
		sel.Find(`.completionList`).Each(func(i int, sel *goquery.Selection) {
			letter := strings.TrimSpace(sel.Find(`.grayTit`).Text())
			//fmt.Println(letter)
			question.Question.Options = append(question.Question.Options, letter)
		})

	})
	return question, nil
}
