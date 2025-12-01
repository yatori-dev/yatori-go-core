package xuexitong

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/yatori-dev/yatori-go-core/api/entity"
	"github.com/yatori-dev/yatori-go-core/que-core/qtype"
	"github.com/yatori-dev/yatori-go-core/utils"
	"github.com/yatori-dev/yatori-go-core/utils/qutils"

	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
)

// WorkNewSubmitAnswer 新的提交作业答案的接口
// stuStatus：3代表待批阅
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
	//选择题
	for _, ch := range question.Choice {
		if ch.Qid != "" {
			answerwqbid += ch.Qid + ","
		}
		if ch.Type == qtype.SingleChoice {
			answers := ""
			candidateSelects := []string{} //待选
			resSelect := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N"}
			for _, option := range resSelect {
				if ch.Options[option] == "" {
					break
				}
				candidateSelects = append(candidateSelects, ch.Options[option])
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
			resSelect := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N"}
			for _, option := range resSelect {
				if ch.Options[option] == "" {
					break
				}
				candidateSelects = append(candidateSelects, ch.Options[option])
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
	//for _, ch := range question.Fill {
	//	if ch.Qid != "" {
	//		answerwqbid += ch.Qid + ","
	//	}
	//	for k, v := range ch.OpFromAnswer {
	//		re := regexp.MustCompile(`\d+`)
	//		numbers := re.FindAllString(k, -1)
	//		//answer := "<p>"+v[0]+"</p>"
	//		_ = writer.WriteField("answer"+ch.Qid+numbers[0], v[0])
	//	}
	//	_ = writer.WriteField("tiankongsize"+ch.Qid, strconv.Itoa(len(ch.OpFromAnswer)))
	//	_ = writer.WriteField("answertype"+ch.Qid, "2")
	//}
	//填空题
	for _, ch := range question.Fill {
		if ch.Qid != "" {
			answerwqbid += ch.Qid + ","
		}
		for k, v := range ch.OpFromAnswer {
			_ = writer.WriteField("answer"+ch.Qid+fmt.Sprintf("%d", k+1), v)
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

	//其它
	for _, ch := range question.Other {
		if ch.Qid != "" {
			answerwqbid += ch.Qid + ","
		}
		for _, v := range ch.OpFromAnswer {
			_ = writer.WriteField("answer"+ch.Qid, v[0])
		}
		_ = writer.WriteField("answertype"+ch.Qid, "8")
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

// PullWorkListHtmlApi 拉取邮箱作业列表
func (cache *XueXiTUserCache) PullWorkListHtmlApi(courseId string, classId string, cpi string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	urlStr := "https://mooc1-api.chaoxing.com/work/task-list?courseId=" + courseId + "&classId=" + classId + "&cpi=" + cpi
	method := "GET"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}
	client := &http.Client{
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return errors.New("太多重定向")
			}

			// 复制 Cookie
			if len(via) > 0 {
				for _, c := range via[0].Cookies() {
					req.AddCookie(c)
				}
			}
			return nil // 允许重定向
		},
	}
	req, err := http.NewRequest(method, urlStr, nil)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Linux; Android 8.1.0; MI 5X Build/OPM1.171019.019; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/71.0.3578.99 Mobile Safari/537.36 (schild:ce5175d20950c8ee955fb03246f762da) (device:MI 5X) Language/zh_CN com.chaoxing.mobile/ChaoXingStudy_3_6.7.2_android_phone_10936_311 (@Kalimdor)_76c82452584d47e39ab79aa54ea86554")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	req.Header.Add("Upgrade-Insecure-Requests", "1")
	req.Header.Add("accept-language", "zh_CN")
	req.Header.Add("X-Requested-With", "com.chaoxing.mobile")
	req.Header.Add("Cookie", "k8s=1764505505.713.24258.990048; route=1ab934bb3bbdaaef56ce3b0da45c52ed; fid=123402; _uid=410496399; UID=410496399; xxtenc=bac7e84bc401921a0c9bbbc730d6a802; fidsCount=1; _industry=6; sso_role=3; _tid=362220710; sso_puid=410496399; wfwIncode=xw85779; wfwfid=123402; spaceFid=123402; source=num2; wfwEnc=EFD203BB75DE4CE9DDE7348502FBDF88; _d=1764567030436; vc3=ZMa3ZNbJZxzO27Le2nMWqYNHxFC1CsL%2BMiQVphB%2FkiABCazrlQ3v7CpOvpx9ZtMRW%2BzDfNg5kuXRHrQ6H2C8PHF2pc0G4UsTAH%2BcgnAJEMFnbnWYR3a5TTiZP3jlFjh4Bx0d0wzm1iMyoTndqd5DWJsmQFJjwr0b3U7Ckr0uGxQ%3Def0b614ffd3367c1fc06afe55e074c0c; uf=b2d2c93beefa90dc51b37d0932243830be3902fbfb1fc62ce5e0881cbf78ed4a7cc932f0757146e08b64b4d624315739179900d10480c3f3ea4a1670a3a8352fe9295d8c89b08ad0f44425e20f927c6b97cb2aec7f1e5a9afb98ce0e6210c3884a878d0a9a7b05da6103a97f8cd189bccc5d4a9b0d1b6a7a1939dc78d50c971ae2ead16ef2f50a8d7420ea0e76898b41da9735baa04d8d5fce71fc6e59483dd39b16e3a7097306134bf1828e50f6a9b65f822a7404fce8bfe9fdc681bdf07734; cx_p_token=3a75006e93978cdefe29431a1d503240; p_auth_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOiI0MTA0OTYzOTkiLCJsb2dpblRpbWUiOjE3NjQ1NjcwMzA0MzgsImV4cCI6MTc2NTE3MTgzMH0.8Rnwx7-fAHepel9EwqQKf0h48126-AngAhnZqZ5YzB4; DSSTASH_LOG=C_38-UN_2010-US_410496399-T_1764567030438; sso_t=1764567030436; sso_v=586660f4fc7a775c7920a0ff2d6d4c9b; KI4SO_SERVER_EC=RERFSWdRQWdsckJiQXZ5ZmdkWW10bkdSREo2RU9sU1hpRldOcmRNcVkzcGZhejZxVm1pLzNlbDZi%0AZEc0MVhuWkRtU2djWlRjU2NTVQp1NG13bWtscWxkc3ZDODB1b05kYVB2b29OblVmd253N3dlclJt%0ARDFtdWE0OVd4bXBHSTZoVnhLTCswbDhsSW1RZFpncDlWNGgzNGMzCkFZZUlFck8zajRzTVBLcU1C%0AM0RYZEtodUhpeStxWG1paHdJYTByQ3B1NlBmSlpXSEhUb21wRWs4QkdRcFJWalNhcGVra0NzWCtR%0AWlcKUVNwUHdzV3hJbS9rOURPcEJkbXJNMGlFRE5uQTFqQnhsazFqL2oyNHhqSllPL3Q1STBLbUxW%0ASTBDK0ZjdUlETU1FSFVXTEJtMVE4TQp3NFBOeXMyTWVKRHl5VjlBVmkrRFUzMHRYWGlMekVmMGdl%0ASEVsRWZUa2hrSml1TUhNZlV0MUZpMDFWWUlablFjYllWdlhBKzRyWDVHCkhGN08wOEZaS1p2OXdS%0ASWNKVFozajdMRTJwMFJYTDI4c0pyUDR2NVZncFpPY0xVeDdqUTZuMUt1WFNqeGQ0MHluWC9jQ2ZU%0AeTlNZmUKV3VjPT9hcHBJZD0xJmtleUlkPTE%3D; jrose=208D25304E08B788854AC10FCE760EA7.mooc-p4-3965218474-jf1q6")
	req.Header.Add("Host", "mooc1-api.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")

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

//用于解析试卷加密字符的html----------------------------------------------------------------------------------------------------------
// --------------------- 全局变量 ---------------------

var glyfHashed map[string]uint16
var cmap map[string]rune

// --------------------- 工具函数 ---------------------

func keyFor(data []byte) string {
	h1 := sha1.Sum(data)
	h2 := md5.Sum(data)
	return hex.EncodeToString(h1[:]) + "|" + hex.EncodeToString(h2[:])
}

func loadJSONTables() error {
	gData, _ := os.ReadFile("glyfHashed.json")
	cData, _ := os.ReadFile("cmap.json")
	if err := json.Unmarshal(gData, &glyfHashed); err != nil {
		return err
	}
	if err := json.Unmarshal(cData, &cmap); err != nil {
		return err
	}
	fmt.Println("✅ 字体映射表加载完成:", len(glyfHashed), "个哈希项")
	return nil
}

// --------------------- 解析 TTF ---------------------

type tableRecord struct {
	Offset uint32
	Length uint32
}

type ttfFile struct {
	src    []byte
	tables map[string]tableRecord
}

func parseTTF(b []byte) (*ttfFile, error) {
	r := bytes.NewReader(b)
	var numTables uint16
	if _, err := r.Seek(4, io.SeekStart); err != nil {
		return nil, err
	}
	binary.Read(r, binary.BigEndian, &numTables)
	r.Seek(6, io.SeekCurrent)
	tables := make(map[string]tableRecord)
	for i := 0; i < int(numTables); i++ {
		tag := make([]byte, 4)
		r.Read(tag)
		r.Seek(4, io.SeekCurrent)
		var off, length uint32
		binary.Read(r, binary.BigEndian, &off)
		binary.Read(r, binary.BigEndian, &length)
		tables[string(tag)] = tableRecord{Offset: off, Length: length}
	}
	return &ttfFile{src: b, tables: tables}, nil
}

func (t *ttfFile) table(tag string) ([]byte, error) {
	rec, ok := t.tables[tag]
	if !ok {
		return nil, fmt.Errorf("missing table %s", tag)
	}
	return t.src[rec.Offset : rec.Offset+rec.Length], nil
}

// head: indexToLocFormat @ offset 50
func parseHeadIndexToLocFormat(head []byte) int16 {
	return int16(binary.BigEndian.Uint16(head[50:]))
}

// maxp: numGlyphs @ offset 4
func parseMaxpNumGlyphs(maxp []byte) uint16 {
	return binary.BigEndian.Uint16(maxp[4:6])
}

// loca -> glyph offsets
func parseLoca(loca []byte, long bool, numGlyphs uint16) []uint32 {
	offsets := make([]uint32, numGlyphs+1)
	if long {
		for i := range offsets {
			offsets[i] = binary.BigEndian.Uint32(loca[i*4:])
		}
	} else {
		for i := range offsets {
			offsets[i] = uint32(binary.BigEndian.Uint16(loca[i*2:])) * 2
		}
	}
	return offsets
}

// translate: 计算哈希映射
func translate(font []byte) map[rune]rune {
	mapping := make(map[rune]rune)

	ttf, err := parseTTF(font)
	if err != nil {
		fmt.Println("字体解析错误:", err)
		return mapping
	}

	head, _ := ttf.table("head")
	maxp, _ := ttf.table("maxp")
	loca, _ := ttf.table("loca")
	glyf, _ := ttf.table("glyf")

	locFormat := parseHeadIndexToLocFormat(head)
	numGlyphs := parseMaxpNumGlyphs(maxp)
	offsets := parseLoca(loca, locFormat != 0, numGlyphs)

	for i := 0; i < int(numGlyphs); i++ {
		start, end := offsets[i], offsets[i+1]
		if end <= start || int(end) > len(glyf) {
			continue
		}
		raw := glyf[start:end]
		k := keyFor(raw)

		refGID, ok := glyfHashed[k]
		if !ok {
			continue
		}
		targetRune, ok := cmap[fmt.Sprint(refGID)]
		if !ok {
			continue
		}
		mapping[rune(i)] = targetRune
	}
	return mapping
}

// --------------------- HTML 替换 ---------------------

func decodeHTML(html string) (string, error) {
	re := regexp.MustCompile(`data:(?:application|font)/font-ttf[^,]*,([A-Za-z0-9+/=]+)`)
	match := re.FindStringSubmatch(html)
	if len(match) == 0 {
		return "", errors.New("未检测到字体 base64")
	}
	fontBytes, _ := base64.StdEncoding.DecodeString(match[1])
	fmt.Println("✅ 检测到 Base64 字体")

	mapping := translate(fontBytes)
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(html))

	doc.Find("*").Each(func(_ int, s *goquery.Selection) {
		text := s.Text()
		var out strings.Builder
		for _, r := range text {
			if newRune, ok := mapping[r]; ok {
				out.WriteRune(newRune)
			} else {
				out.WriteRune(r)
			}
		}
		s.SetText(out.String())
	})

	return doc.Text(), nil
}
