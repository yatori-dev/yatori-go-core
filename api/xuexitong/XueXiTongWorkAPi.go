package xuexitong

import (
	"bytes"
	"fmt"
	"github.com/yatori-dev/yatori-go-core/api/entity"
	"github.com/yatori-dev/yatori-go-core/models/ctype"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// WorkNewSubmitAnswer 新的提交作业答案的接口
func (cache *XueXiTUserCache) WorkNewSubmitAnswer(courseId string, classId string, knowledgeid string,
	cpi string, jobid string, totalQuestionNum string, answerId string,
	workAnswerId string, api string, fullScore string, oldSchoolId string,
	oldWorkId string, workRelationId string, enc_work string, question entity.Question, isSubmit string /*""为直接交卷，1为暂存*/) (string, error) {

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("pyFlag", isSubmit)
	_ = writer.WriteField("courseId", courseId)
	_ = writer.WriteField("classId", classId)
	_ = writer.WriteField("api", api)
	_ = writer.WriteField("workAnswerId", workAnswerId)
	_ = writer.WriteField("answerId", answerId)
	_ = writer.WriteField("totalQuestionNum", totalQuestionNum)
	_ = writer.WriteField("fullScore", fullScore)
	_ = writer.WriteField("knowledgeid", knowledgeid)
	_ = writer.WriteField("oldSchoolId", oldSchoolId)
	_ = writer.WriteField("oldWorkId", oldWorkId)
	_ = writer.WriteField("jobid", jobid)
	_ = writer.WriteField("workRelationId", workRelationId)
	_ = writer.WriteField("enc", "")
	_ = writer.WriteField("enc_work", enc_work)
	_ = writer.WriteField("userId", cache.UserID)
	_ = writer.WriteField("cpi", cpi)
	_ = writer.WriteField("workTimesEnc", "")
	_ = writer.WriteField("randomOptions", "false")
	_ = writer.WriteField("isAccessibleCustomFid", "0")
	answerwqbid := ""
	for _, ch := range question.Choice {
		if ch.Qid != "" {
			answerwqbid += ch.Qid + ","
		}
		if ch.Type == ctype.SingleChoice {
			answers := ""
			for _, item := range ch.Answers {
				for k, v := range ch.Options {
					if strings.Contains(v, item) {
						answers += k
					}

				}
			}
			_ = writer.WriteField("answer"+ch.Qid, answers)
			_ = writer.WriteField("answertype"+ch.Qid, "0")
		}
		if ch.Type == ctype.MultipleChoice {
			answers := ""
			for _, item := range ch.Answers {
				for k, v := range ch.Options {
					if strings.Contains(v, item) {
						answers += k
					}

				}
			}
			_ = writer.WriteField("answer"+ch.Qid, answers)
			_ = writer.WriteField("answertype"+ch.Qid, "1")
		}
	}
	for _, ch := range question.Judge {
		if ch.Qid != "" {
			answerwqbid += ch.Qid + ","
		}
		answers := ""
		for _, item := range ch.Answers {
			if item == "正确" {
				item = "true"
			}
			if item == "错误" {
				item = "false"
			}
			answers += item

		}
		_ = writer.WriteField("answer"+ch.Qid, answers)
		_ = writer.WriteField("answertype"+ch.Qid, "3")
	}

	for _, ch := range question.Fill {
		if ch.Qid != "" {
			answerwqbid += ch.Qid + ","
		}
		for k, v := range ch.OpFromAnswer {
			re := regexp.MustCompile(`\d+`)
			numbers := re.FindAllString(k, -1)
			//answer := "<p>"+v[0]+"</p>"
			_ = writer.WriteField("answer"+ch.Qid+numbers[0], v[0])
		}
		_ = writer.WriteField("tiankongsize"+ch.Qid, strconv.Itoa(len(ch.OpFromAnswer)))
		_ = writer.WriteField("answertype"+ch.Qid, "2")
	}

	for _, ch := range question.Short {
		if ch.Qid != "" {
			answerwqbid += ch.Qid + ","
		}
		for _, v := range ch.OpFromAnswer {
			_ = writer.WriteField("answer"+ch.Qid, v[0])
		}
		_ = writer.WriteField("answertype"+ch.Qid, "4")
	}

	_ = writer.WriteField("answerwqbid", answerwqbid)
	err := writer.Close()
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	//403825664,403825665,403825666,403825667,403825668,403825669,403825670,403825671,403825672,403825673,403825674,403825684,403825685,403825686,403825675,403825676,403825677,403825678,403825687,403825688,
	//403825664,403825665,403825666,403825667,403825668,403825669,403825670,403825671,403825672,403825673,403825674,403825684,403825685,403825686,403825675,403825676,403825677,403825678,403825687,403825688,
	// 构建 URL
	urlStr := fmt.Sprintf("%s?_classId=%s&courseid=%s&token=%s&totalQuestionNum=%s&ua=pc&formType=post&saveStatus=1&version=1&tempsave=1",
		ApiWorkCommitNew, classId, courseId, enc_work, totalQuestionNum)
	// 构建请求
	req, err := http.NewRequest("POST", urlStr, payload)
	if err != nil {
		return "", err
	}
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "mooc1.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Add("Cookie", cache.cookie)

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return "", err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	return string(body), nil
}
