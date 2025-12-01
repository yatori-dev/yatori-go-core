package xuexitong

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/yatori-dev/yatori-go-core/api/xuexitong"
)

// 学习通考试
type XXTExam struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	RemainTime string `json:"remain_time"`
	RawURL     string `json:"raw_url"`
	Params     map[string]string
	CourseId   string `json:"course_id"`
	UserId     string `json:"user_id"`
	ClazzId    string `json:"clazz_id"`
	Type       string `json:"type"`
	EncTask    string `json:"enc_task"`
	TaskRefId  string `json:"taskrefId"`
	msgId      string `json:"msgId"`
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
			msgId:      params["msgId"],
		}
		examList = append(examList, exam)
	})

	return examList, nil
}

// EnterExamAction 进入考试
func EnterExamAction(cache *xuexitong.XueXiTUserCache, exam *XXTExam) error {
	//这一步拉取必要的参数，比如滑块验证码参数等,注意这里的refererUrl会在后面的滑块验证码中用到
	enterHtml, refererUrl, err := cache.PullExamEnterInformHtmlApi(exam.TaskRefId, exam.msgId, exam.CourseId, exam.UserId, exam.ClazzId, exam.Type, exam.EncTask, 3, nil)
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

	var captchaCaptchaId HiddenField //寄存验证码ID参数
	// 输出结果
	for _, f := range fields {
		if f.ID == "captchaCaptchaId" {
			captchaCaptchaId = f
		}
		fmt.Printf("ID: %-35s Name: %-20s Value: %s\n", f.ID, f.Name, f.Value)
	}
	if captchaCaptchaId.Value != "" {
		slider := XueXiTSlider{
			CaptchaId: captchaCaptchaId.Value,
			Referer:   refererUrl,
		}
		slider.Pass(cache)
	}

	return nil
}
