package xuexitong

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/yatori-dev/yatori-go-core/api/entity"
	"github.com/yatori-dev/yatori-go-core/que-core/qtype"
	"github.com/yatori-dev/yatori-go-core/utils"
	"github.com/yatori-dev/yatori-go-core/utils/qutils"

	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
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
	_ = writer.WriteField("randomOptions", "true")
	_ = writer.WriteField("isAccessibleCustomFid", "0")
	answerwqbid := ""
	//选择题
	for _, ch := range question.Choice {
		if ch.Qid != "" {
			answerwqbid += ch.Qid + ","
		}
		if ch.Type == qtype.SingleChoice {
			answers := ""
			candidateSelects := []string{} //待选
			for _, option := range ch.Options {
				candidateSelects = append(candidateSelects, option)
			}
			for _, item := range ch.Answers {
				answers += qutils.SimilarityArraySelect(item, candidateSelects)
			}
			_ = writer.WriteField("answer"+ch.Qid, answers)
			_ = writer.WriteField("answertype"+ch.Qid, "0")
		}
		if ch.Type == qtype.MultipleChoice {
			answers := ""
			candidateSelects := []string{} //待选
			for _, option := range ch.Options {
				candidateSelects = append(candidateSelects, option)
			}
			for _, item := range ch.Answers {
				answers += qutils.SimilarityArraySelect(item, candidateSelects)
			}
			//答案排序
			r := []rune(answers)                                      // 将字符串转换为字符数组
			sort.Slice(r, func(i, j int) bool { return r[i] < r[j] }) // 使用 sort 包进行排序
			answers = string(r)
			_ = writer.WriteField("answer"+ch.Qid, answers)
			_ = writer.WriteField("answertype"+ch.Qid, "1")
		}
	}
	//判断题
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
	//填空题
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
	//简答题
	for _, ch := range question.Short {
		if ch.Qid != "" {
			answerwqbid += ch.Qid + ","
		}
		for _, v := range ch.OpFromAnswer {
			_ = writer.WriteField("answer"+ch.Qid, v[0])
		}
		_ = writer.WriteField("answertype"+ch.Qid, "4")
	}
	//名词解释
	for _, ch := range question.TermExplanation {
		if ch.Qid != "" {
			answerwqbid += ch.Qid + ","
		}
		for _, v := range ch.OpFromAnswer {
			_ = writer.WriteField("answer"+ch.Qid, v[0])
		}
		_ = writer.WriteField("answertype"+ch.Qid, "5")
	}
	//论述题
	for _, ch := range question.Essay {
		if ch.Qid != "" {
			answerwqbid += ch.Qid + ","
		}
		for _, v := range ch.OpFromAnswer {
			_ = writer.WriteField("answer"+ch.Qid, v[0])
		}
		_ = writer.WriteField("answertype"+ch.Qid, "6")
	}
	//连线题
	for _, ch := range question.Matching {
		if ch.Qid != "" {
			answerwqbid += ch.Qid + ","
		}
		//临时构建
		type SelData struct {
			Name    int    `json:"name"`
			Content string `json:"content"`
		}
		listSel := []SelData{}
		for i, answer := range ch.Answers {
			answerSel := qutils.SimilarityArraySelect(strings.Split(answer, "->")[1], ch.Selects)
			listSel = append(listSel, SelData{
				Name:    i + 1,
				Content: answerSel,
			})
			_ = writer.WriteField("dept", answerSel)
		}
		listSelJson, _ := json.Marshal(listSel)
		_ = writer.WriteField("answer"+ch.Qid, url.QueryEscape(string(listSelJson)))
		_ = writer.WriteField("answertype"+ch.Qid, "11")
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
	req.Header.Add("User-Agent", GetUA("mobile"))
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "mooc1.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")

	req.Header.Set("Content-Type", writer.FormDataContentType())
	//req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}

	// 发送请求
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}

	//如果开启了IP代理，那么就直接添加代理
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse(cache.ProxyIP) // 设置代理
		}
	}
	client := &http.Client{
		Transport: tr,
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return "", err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if strings.Contains(string(body), "请输入验证码") {
		return "", errors.New("触发验证码")
	}
	utils.CookiesAddNoRepetition(&cache.cookies, resp.Cookies()) //赋值cookie
	return string(body), nil
}

func (cache *XueXiTUserCache) WorkNewSubmitAnswerNew(courseId string, classId string, knowledgeid string,
	cpi string, jobid string, totalQuestionNum string, answerId string,
	workAnswerId string, api string, fullScore string, oldSchoolId string,
	oldWorkId string, workRelationId string, enc_work string, question entity.Question, isSubmit string /*""为直接交卷，1为暂存*/) (string, error) {
	//urlStr := "https://mooc1.chaoxing.com/mooc-ans/work/addStudentWorkNew?_classId=125961101&courseid=254298146&token=8e617b972e919e3835df299cd7ffe75f&totalQuestionNum=fa7b891349fa608eae0b751926d519c9&ua=pc&formType=post&saveStatus=1&version=1&tempsave=1"
	urlStr := fmt.Sprintf("%s?_classId=%s&courseid=%s&token=%s&totalQuestionNum=%s&ua=pc&formType=post&saveStatus=1&version=1&tempsave=1",
		ApiWorkCommitNew, classId, courseId, enc_work, totalQuestionNum)
	method := "POST"
	payloadStr := ""
	payload := strings.NewReader("pyFlag=1&courseId=254298146&classId=125961101&api=1&workAnswerId=55322670&answerId=55322670&totalQuestionNum=fa7b891349fa608eae0b751926d519c9&fullScore=100.0&knowledgeid=1009754303&oldSchoolId=&oldWorkId=11828eb3b3ca4ca39b67b257871b9deb&jobid=work-11828eb3b3ca4ca39b67b257871b9deb&workRelationId=44988171&enc=&enc_work=8e617b972e919e3835df299cd7ffe75f&userId=247050353&cpi=275995984&workTimesEnc=&randomOptions=false&isAccessibleCustomFid=0&answer213645519=ABD&answertype213645519=1&answer213645518=ABD&answertype213645518=1&answer213645521=true&answertype213645521=3&answer213645520=true&answertype213645520=3&dept=A&dept=D&dept=B&dept=C&answer213645522=%5B%7B%22name%22%3A1%2C%22content%22%3A%22A%22%7D%2C%7B%22name%22%3A2%2C%22content%22%3A%22D%22%7D%2C%7B%22name%22%3A3%2C%22content%22%3A%22B%22%7D%2C%7B%22name%22%3A4%2C%22content%22%3A%22C%22%7D%5D&answertype213645522=11&answerwqbid=213645519%2C213645518%2C213645521%2C213645520%2C213645522%2C")

	payloadStr += "pyFlag=" + isSubmit
	payloadStr += "&courseId=" + courseId
	payloadStr += "&classId=" + classId
	payloadStr += "&api=" + api
	payloadStr += "&workAnswerId=" + workAnswerId
	payloadStr += "&answerId=" + answerId
	payloadStr += "&totalQuestionNum=" + totalQuestionNum
	payloadStr += "&fullScore" + fullScore
	payloadStr += "&knowledgeid=" + knowledgeid
	payloadStr += "&oldSchoolId=" + oldSchoolId
	payloadStr += "oldWorkId=" + oldWorkId
	payloadStr += "&jobid=" + jobid
	payloadStr += "&workRelationId=" + workRelationId
	payloadStr += "&enc=" + ""
	payloadStr += "&enc_work=" + enc_work
	payloadStr += "&userId=" + cache.UserID
	payloadStr += "&cpi=" + cpi
	payloadStr += "&workTimesEnc=" + ""
	payloadStr += "&randomOptions=" + "false"
	payloadStr += "&isAccessibleCustomFid=" + "0"

	answerwqbid := ""
	//选择题
	for _, ch := range question.Choice {
		if ch.Qid != "" {
			answerwqbid += ch.Qid + ","
		}
		if ch.Type == qtype.SingleChoice {
			answers := ""
			candidateSelects := []string{} //待选
			for _, option := range ch.Options {
				candidateSelects = append(candidateSelects, option)
			}
			for _, item := range ch.Answers {
				answers += qutils.SimilarityArraySelect(item, candidateSelects)
			}
			payloadStr += "&answer" + ch.Qid + "=" + answers
			payloadStr += "&answertype" + ch.Qid + "=" + "0"
		}
		if ch.Type == qtype.MultipleChoice {
			answers := ""
			candidateSelects := []string{} //待选
			for _, option := range ch.Options {
				candidateSelects = append(candidateSelects, option)
			}
			for _, item := range ch.Answers {
				answers += qutils.SimilarityArraySelect(item, candidateSelects)
			}
			//答案排序
			r := []rune(answers)                                      // 将字符串转换为字符数组
			sort.Slice(r, func(i, j int) bool { return r[i] < r[j] }) // 使用 sort 包进行排序
			answers = string(r)
			payloadStr += "&answer" + ch.Qid + "=" + answers
			payloadStr += "&answertype" + ch.Qid + "=" + "1"
		}
	}
	//判断题
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
		payloadStr += "&answer" + ch.Qid + "=" + answers
		payloadStr += "&answertype" + ch.Qid + "=" + "3"
	}
	//填空题
	for _, ch := range question.Fill {
		if ch.Qid != "" {
			answerwqbid += ch.Qid + ","
		}
		for k, v := range ch.OpFromAnswer {
			re := regexp.MustCompile(`\d+`)
			numbers := re.FindAllString(k, -1)
			//answer := "<p>"+v[0]+"</p>"

			payloadStr += "&answer" + ch.Qid + numbers[0] + "=" + url.QueryEscape(v[0])
		}
		payloadStr += "&tiankongsize" + ch.Qid + "=" + strconv.Itoa(len(ch.OpFromAnswer))
		payloadStr += "&answertype" + ch.Qid + "=" + "2"
	}
	//简答题
	for _, ch := range question.Short {
		if ch.Qid != "" {
			answerwqbid += ch.Qid + ","
		}
		for _, v := range ch.OpFromAnswer {
			payloadStr += "&answer" + ch.Qid + "=" + url.QueryEscape(v[0])
		}
		payloadStr += "&answertype" + ch.Qid + "=" + "4"
	}
	//名词解释
	for _, ch := range question.TermExplanation {
		if ch.Qid != "" {
			answerwqbid += ch.Qid + ","
		}
		for _, v := range ch.OpFromAnswer {
			payloadStr += "&answer" + ch.Qid + "=" + url.QueryEscape(v[0])
		}
		payloadStr += "&answertype" + ch.Qid + "=" + "5"
	}
	//论述题
	for _, ch := range question.Essay {
		if ch.Qid != "" {
			answerwqbid += ch.Qid + ","
		}
		for _, v := range ch.OpFromAnswer {
			payloadStr += "&answer" + ch.Qid + "=" + url.QueryEscape(v[0])
		}
		payloadStr += "&answertype" + ch.Qid + "=" + "6"
	}
	//连线题
	for _, ch := range question.Matching {
		if ch.Qid != "" {
			answerwqbid += ch.Qid + ","
		}
		//临时构建
		type SelData struct {
			Name    int    `json:"name"`
			Content string `json:"content"`
		}
		listSel := []SelData{}
		for i, answer := range ch.Answers {
			answerSel := qutils.SimilarityArraySelect(strings.Split(answer, "->")[1], ch.Selects)
			listSel = append(listSel, SelData{
				Name:    i + 1,
				Content: answerSel,
			})
			payloadStr += "&dept=" + answerSel
		}
		listSelJson, _ := json.Marshal(listSel)
		payloadStr += "&answer" + ch.Qid + "=" + url.QueryEscape(string(listSelJson))
		payloadStr += "&answertype" + ch.Qid + "=" + "11"
	}
	//_ = writer.WriteField("answerwqbid", answerwqbid)
	payloadStr += "&answerwqbid=" + url.QueryEscape(answerwqbid)
	//payload := strings.NewReader(payloadStr)
	client := &http.Client{}
	req, err := http.NewRequest(method, urlStr, payload)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Add("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Origin", "https://mooc1.chaoxing.com")
	req.Header.Add("Pragma", "no-cache")
	//req.Header.Add("Referer", "https://mooc1.chaoxing.com/mooc-ans/work/doHomeWorkNew?courseId=254298146&workAnswerId=55322670&workId=44988171&api=1&knowledgeid=1009754303&classId=125961101&oldWorkId=11828eb3b3ca4ca39b67b257871b9deb&jobid=work-11828eb3b3ca4ca39b67b257871b9deb&type=&isphone=false&submit=false&enc=46aeb504461526ded1565ec2adc54bbb&cpi=275995984&mooc2=1&skipHeader=true&originJobId=work-11828eb3b3ca4ca39b67b257871b9deb&fromType=")
	req.Header.Add("Sec-Fetch-Dest", "empty")
	req.Header.Add("Sec-Fetch-Mode", "cors")
	req.Header.Add("Sec-Fetch-Site", "same-origin")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Add("sec-ch-ua", "\"Microsoft Edge\";v=\"141\", \"Not?A_Brand\";v=\"8\", \"Chromium\";v=\"141\"")
	req.Header.Add("sec-ch-ua-mobile", "?0")
	req.Header.Add("sec-ch-ua-platform", "\"Windows\"")
	req.Header.Add("Sec-Fetch-User", "?1")
	req.Header.Add("Upgrade-Insecure-Requests", "1")
	//req.Header.Add("Cookie", "k8s=1761899340.126.18975.419486; fid=3829; source=\"\"; _uid=247050353; _d=1761900076418; UID=247050353; vc3=Hxg9jhItwvuc8ebEnEJzqA70bhaurTMtz3p4XdtxBTSae:\\Yatori-Dev\\yatori-go-core\\models\\ctype\\AiType.govBaj32YQAyrovVnPIuc62BtgOokwm1mgHyvJvN1zgGTwh26BW0%2Bnpn7erBrfJZrFufnZHzRN9ltsNcIW%2Fy%2Fdakly9rP74jOjEEjpOmDB%2B9jVWXtJkvvsE7KS0v%2F4lAU%3D34463274ff9c08f0b8c17ea663b9fdc6; uf=94ffe74515793f367c02eecd2e65af28b42afde16556a41dc962c8f4bbdfd55d8a51b4a415fc6f1f977e73a5069ed94181a6c9ddee30899fd807a544f7930b6aed1e6c11a143bb563b0339d97cdac4ba48c1802635dab65a713028f1ec42bf71b1188854805578cc325fb1782ddea829c6cc1ea97b2b4d97fc8f35766b3e5bfbea823dbb68193e1d0454dcc70487dcf34df7ff280fcb29d10d8a4c92b12beb4bcfa7963e27723ad17a6ae89eec95c32aae5af46d05a99736e7fafd565af53bf2; cx_p_token=561776b9708a75757fc967538ce8ab65; p_auth_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOiIyNDcwNTAzNTMiLCJsb2dpblRpbWUiOjE3NjE5MDAwNzY0MjAsImV4cCI6MTc2MjUwNDg3Nn0.krOXjjpWE-EOFICZhqp-T4i3nCyyBuaBoOAiXXy1XY0; xxtenc=0f1d1ea6064c7e19b5ecdd4a7240c42b; DSSTASH_LOG=C_38-UN_2866-US_247050353-T_1761900076420; thirdRegist=0; _dd412348668=1761900077489; fanyamoocs=BC953FCDEEB4409B8351C072BAF36DBE; route=f537d772be8122bff9ae56a564b98ff6; tl=1; _dd247050353=1761900817019; jrose=1CEC9B901F078384F64A09DFE8F2410A.mooc-4014427331-7j6nq")
	//req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")
	req.Header.Add("User-Agent", GetUA("mobile"))
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("Host", "mooc1.chaoxing.com")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	//fmt.Println(string(body))
	return string(body), nil
}
