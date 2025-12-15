package xuexitong

import (
	"bytes"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/yatori-dev/yatori-go-core/api/xuexitong"
	"github.com/yatori-dev/yatori-go-core/que-core/qtype"
)

// 学习通考试结构体
type XXTExam struct {
	Name             string `json:"name"`
	Status           string `json:"status"`
	RemainTime       string `json:"remain_time"`
	RawURL           string `json:"raw_url"`
	Params           map[string]string
	CourseId         string `json:"course_id"`
	UserId           string `json:"user_id"`
	ClazzId          string `json:"clazz_id"`
	Type             string `json:"type"`
	EncTask          string `json:"enc_task"`
	TaskRefId        string `json:"taskrefId"`
	MsgId            string `json:"msgId"`
	CaptchaCaptchaId string
	ExamRelationId   string
	AnswerId         string `json:"answerId"`
	Cpi              string
	Validate         string //过验证码用的
	Paper            xuexitong.XXTExamPaper
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
		fmt.Println(refererUrl)
		return err
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
		fmt.Printf("ID: %-35s Name: %-20s Value: %s\n", f.ID, f.Name, f.Value)
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

	return nil
}

// 拉取考试试卷
func PullExamPaperAction(cache *xuexitong.XueXiTUserCache, exam *XXTExam, isSubmit bool) error {
	pullPaperHtml, err := cache.PullExamPaperHtmlApi(exam.CourseId, exam.ClazzId, exam.ExamRelationId, "0", exam.AnswerId, exam.Cpi, "1", xuexitong.IMEI, exam.Validate, "0", 3, nil)
	if err != nil {
		return err
	}

	qsEntity, err := HtmlPaperTurnEntity(pullPaperHtml)
	exam.Paper = qsEntity
	if err != nil {
		return err
	}
	for i := 0; ; i++ {
		pullQuestion, err1 := cache.PullExamQuestionApi(exam.CourseId, exam.ClazzId, exam.ExamRelationId, exam.AnswerId, exam.Cpi, exam.Paper.EncRemainTime, exam.Paper.Enc, exam.Paper.EncLastUpdateTime, i)
		if err1 != nil {
			return err1
		}
		isLastQuestion := strings.Contains(pullQuestion, `<div class="lastQuestion"> 已经是最后一题了</div>`)
		qsEntity, err = HtmlPaperTurnEntity(pullQuestion)

		exam.Paper = qsEntity
		//fmt.Println(pullPaperHtml)

		answerResult, err2 := SubmitExamAnswerAction(cache, exam, isLastQuestion && isSubmit)
		if err2 != nil {
			return err2
		}
		fmt.Println(answerResult)
		if isLastQuestion { //如果已经是最后一题则直接退出
			break
		}
	}

	return nil
}

// 提交学习通考试答案
func SubmitExamAnswerAction(cache *xuexitong.XueXiTUserCache, exam *XXTExam, isSubmit bool /*是否提交，true为提交，false为暂存*/) (string, error) {
	api, err := cache.SubmitExamAnswerApi(exam.ClazzId, exam.CourseId, exam.Paper.TestPaperId, exam.Paper.TestUserRelationId, exam.Cpi, exam.Paper.RemainTime, exam.Paper.EncRemainTime, exam.Paper.EncLastUpdateTime, exam.ExamRelationId, exam.AnswerId, exam.RemainTime, !isSubmit, exam.Paper.Enc, exam.Paper.EnterPageTime, exam.Paper.XXTQuestion.QuestionId, exam.Paper.Type, exam.Paper.XXTQuestion.TypeName, &exam.Paper)
	if err != nil {
		return "", err
	}
	//fmt.Println(api)
	return api, nil
}

// html转Exam实体
func HtmlPaperTurnEntity(paperHtml string) (xuexitong.XXTExamPaper, error) {
	xxtExamPaper := xuexitong.XXTExamPaper{}
	// 使用 goquery 解析 HTML
	paperDoc, err := goquery.NewDocumentFromReader(bytes.NewReader([]byte(paperHtml)))
	if err != nil {
		log.Fatal(err)
	}
	questionId, exists := paperDoc.Find("#questionId").Attr("value") //题目id
	if exists {
		fmt.Println("question:", questionId)
	}
	questionType, exists := paperDoc.Find(`input[name="` + `type` + questionId + `"]`).Attr("value")
	if exists {
		fmt.Println("questionType:", questionType)
	}
	get := func(id string) string {
		v, _ := paperDoc.Find("#" + id).Attr("value")
		return v
	}
	xxtExamPaper.CourseId = get("courseId")
	xxtExamPaper.TestPaperId = get("testPaperId")
	xxtExamPaper.TestUserRelationId = get("testUserRelationId")
	xxtExamPaper.ClassId = get("classId")
	xxtExamPaper.Type = get("type")
	xxtExamPaper.IsPhone = get("isphone")
	xxtExamPaper.Imei = get("imei")
	xxtExamPaper.SubCount = get("subCount")
	xxtExamPaper.RemainTime = get("remainTime")
	xxtExamPaper.TempSave = get("tempSave")
	xxtExamPaper.TimeOver = get("timeOver")
	xxtExamPaper.EncRemainTime = get("encRemainTime")
	xxtExamPaper.EncLastUpdateTime = get("encLastUpdateTime")
	xxtExamPaper.Cpi = get("cpi")
	xxtExamPaper.Enc = get("enc")
	xxtExamPaper.Source = get("source")
	xxtExamPaper.UserId = get("userId")
	xxtExamPaper.EnterPageTime = get("enterPageTime")
	xxtExamPaper.AnsweredView = get("answeredView")
	xxtExamPaper.ExitdTime = get("exitdtime")
	xxtExamPaper.PaperGroupId = get("paperGroupId")

	switch questionType {
	case "0": //单选题
		turn, err1 := singleTurn(paperDoc)
		if err1 != nil {
			fmt.Println(err1)
		}
		xxtExamPaper.XXTQuestion = turn
		fmt.Println(turn)
	case "1": //多选题
		turn, err1 := multipleTurn(paperDoc)
		if err1 != nil {
			fmt.Println(err1)
		}
		xxtExamPaper.XXTQuestion = turn
		fmt.Println(turn)
	case "2": //填空题
		turn, err1 := fillTurn(paperDoc)
		if err1 != nil {
			fmt.Println(err1)
		}
		xxtExamPaper.XXTQuestion = turn
		fmt.Println(turn)
	case "3": //判断题
		turn, err1 := trueOrFalseTurn(paperDoc)
		if err1 != nil {
			fmt.Println(err1)
		}
		xxtExamPaper.XXTQuestion = turn
		fmt.Println(turn)
	case "4": //简答题
		turn, err1 := shortAnswerTurn(paperDoc)
		if err1 != nil {
			fmt.Println(err1)
		}
		xxtExamPaper.XXTQuestion = turn
		fmt.Println(turn)
	case "6": //论述题
		turn, err1 := essayTurn(paperDoc)
		if err1 != nil {
			fmt.Println(err1)
		}
		xxtExamPaper.XXTQuestion = turn
		fmt.Println(turn)
	}

	return xxtExamPaper, nil
}

// 单选题转换
func singleTurn(paperDoc *goquery.Document) (xuexitong.XXTQuestion, error) {
	question := xuexitong.XXTQuestion{}
	question.QType = qtype.SingleChoice
	paperDoc.Find("div.questionWrap").Each(func(i int, sel *goquery.Selection) {
		questionId, exists := sel.Attr("data") //题目id
		if exists {
			question.QuestionId = questionId
			fmt.Println("question:", questionId)
		}

		//题目
		title := strings.TrimSpace(sel.Find(`.tit p`).First().Text())
		fmt.Println(title)
		question.Question.Content = title
		sel.Find(`.singleChoice`).Each(func(i int, sel *goquery.Selection) {
			letter := strings.TrimSpace(sel.Find(`.No`).Text())
			text := strings.TrimSpace(sel.Find(`.answerInfo cc`).Text())
			fmt.Println(letter, text)
			question.Question.Options = append(question.Question.Options, letter+text)
		})

	})
	return question, nil
}

// 多选题转换
func multipleTurn(paperDoc *goquery.Document) (xuexitong.XXTQuestion, error) {
	question := xuexitong.XXTQuestion{}
	question.QType = qtype.MultipleChoice
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
		fmt.Println(title)
		question.Question.Content = title
		sel.Find(`.mulChoice`).Each(func(i int, sel *goquery.Selection) {
			letter := strings.TrimSpace(sel.Find(`.No`).Text())
			text := strings.TrimSpace(sel.Find(`.answerInfo cc`).Text())
			fmt.Println(letter, text)
			question.Question.Options = append(question.Question.Options, letter+text)
		})

	})
	return question, nil
}

// 填空题转换
func fillTurn(paperDoc *goquery.Document) (xuexitong.XXTQuestion, error) {
	question := xuexitong.XXTQuestion{}
	question.QType = qtype.FillInTheBlank
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
		fmt.Println(title)
		question.Question.Content = title
		sel.Find(`.completionList`).Each(func(i int, sel *goquery.Selection) {
			letter := strings.TrimSpace(sel.Find(`.grayTit`).Text())
			fmt.Println(letter)
			question.Question.Options = append(question.Question.Options, letter)
		})

	})
	return question, nil
}

// 判断题
func trueOrFalseTurn(paperDoc *goquery.Document) (xuexitong.XXTQuestion, error) {
	question := xuexitong.XXTQuestion{}
	question.QType = qtype.TrueOrFalse
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
		fmt.Println(title)
		question.Question.Content = title
		sel.Find(`.answerList`).Each(func(i int, sel *goquery.Selection) {
			letter := strings.TrimSpace(sel.Find(`.No`).Text())
			text := strings.TrimSpace(sel.Find(`.answerInfo`).Text())
			fmt.Println(letter, text)
			question.Question.Options = append(question.Question.Options, letter+text)
		})

	})
	return question, nil
}

// 简答题
func shortAnswerTurn(paperDoc *goquery.Document) (xuexitong.XXTQuestion, error) {
	question := xuexitong.XXTQuestion{}
	question.QType = qtype.ShortAnswer
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
		fmt.Println(title)
		question.Question.Content = title
		sel.Find(`.completionList`).Each(func(i int, sel *goquery.Selection) {
			letter := strings.TrimSpace(sel.Find(`.grayTit`).Text())
			fmt.Println(letter)
			question.Question.Options = append(question.Question.Options, letter)
		})

	})
	return question, nil
}

// 论述题
func essayTurn(paperDoc *goquery.Document) (xuexitong.XXTQuestion, error) {
	question := xuexitong.XXTQuestion{}
	question.QType = qtype.Essay
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
		fmt.Println(title)
		question.Question.Content = title
		sel.Find(`.completionList`).Each(func(i int, sel *goquery.Selection) {
			letter := strings.TrimSpace(sel.Find(`.grayTit`).Text())
			fmt.Println(letter)
			question.Question.Options = append(question.Question.Options, letter)
		})

	})
	return question, nil
}
