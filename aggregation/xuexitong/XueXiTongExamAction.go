package xuexitong

import (
	"bytes"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/yatori-dev/yatori-go-core/api/xuexitong"
	"github.com/yatori-dev/yatori-go-core/que-core/qentity"
	"github.com/yatori-dev/yatori-go-core/que-core/qtype"
)

// 学习通考试
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
	Paper            XXTExamPaper
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
	XXTQuestion        []XXTQuestion //题目
}

// 学习通题目
type XXTQuestion struct {
	Id       string      //题目ID
	QType    qtype.QType //题目类型
	question qentity.Question
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
func PullExamPaperAction(cache *xuexitong.XueXiTUserCache, exam *XXTExam) error {
	pullPaperHtml, err := cache.PullExamPaperHtmlApi(exam.CourseId, exam.ClazzId, exam.ExamRelationId, "0", exam.AnswerId, exam.Cpi, "1", xuexitong.IMEI, exam.Validate, "0", 3, nil)
	if err != nil {
		return err
	}
	HtmlPaperTurnEntity(pullPaperHtml)
	fmt.Println(pullPaperHtml)
	return nil
}

// html转Exam实体
func HtmlPaperTurnEntity(paperHtml string) (XXTExamPaper, error) {
	xxtExamPaper := XXTExamPaper{}
	// 使用 goquery 解析 HTML
	paperDoc, err := goquery.NewDocumentFromReader(bytes.NewReader([]byte(paperHtml)))
	if err != nil {
		log.Fatal(err)
	}
	questionId, exists := paperDoc.Attr("data") //题目id
	if exists {
		fmt.Println("question:", questionId)
	}
	questionType, exists := paperDoc.Find(`input[name="` + `type` + questionId + `"]`).Attr("value")
	if exists {
		fmt.Println("questionType:", questionType)
	}

	switch questionType {
	case "0": //单选题
		turn, err1 := singleTurn(paperDoc)
		if err1 != nil {
			fmt.Println(err1)
		}
		fmt.Println(turn)
	}

	return xxtExamPaper, nil
}

func singleTurn(paperDoc *goquery.Document) (XXTQuestion, error) {
	question := XXTQuestion{}
	question.QType = qtype.SingleChoice
	paperDoc.Find("div.questionWrap").Each(func(i int, sel *goquery.Selection) {
		questionId, exists := paperDoc.Attr("data") //题目id
		if exists {
			question.Id = questionId
			fmt.Println("question:", questionId)
		}

		//题目
		title := strings.TrimSpace(sel.Find(`.tit p`).First().Text())
		fmt.Println(title)
		question.question.Content = title
		sel.Find(`.singleChoice`).Each(func(i int, sel *goquery.Selection) {
			letter := strings.TrimSpace(sel.Find(`.No`).Text())
			text := strings.TrimSpace(sel.Find(`.answerInfo cc`).Text())
			fmt.Println(letter, text)
			question.question.Options = append(question.question.Options, letter+text)
		})

	})
	return question, nil
}
