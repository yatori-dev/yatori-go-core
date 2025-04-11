package xuexitong

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/yatori-dev/yatori-go-core/entity"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// PageMobileChapterCard 客户端章节任务卡片 原始html数据返回
func (cache *XueXiTUserCache) PageMobileChapterCard(
	classId, courseId, knowledgeId, cardIndex, cpi int) (string, error) {
	method := "GET"

	params := url.Values{}
	params.Add("clazzid", strconv.Itoa(classId))
	params.Add("courseid", strconv.Itoa(courseId))
	params.Add("knowledgeid", strconv.Itoa(knowledgeId))
	params.Add("num", strconv.Itoa(cardIndex))
	params.Add("isPhone", "1")
	params.Add("control", "true")
	params.Add("cpi", strconv.Itoa(cpi))
	client := &http.Client{}
	req, err := http.NewRequest(method, PageMobileChapterCard+"?"+params.Encode(), nil)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("Cookie", cache.cookie)
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "mooc1-api.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	return string(body), nil
}

type APIError struct {
	Message string
}

func (e *APIError) Error() string {
	return e.Message
}

// VideoDtoFetch 视频数据
func (cache *XueXiTUserCache) VideoDtoFetch(p *entity.PointVideoDto) (string, error) {
	params := url.Values{}
	params.Set("k", strconv.Itoa(p.FID))
	params.Set("flag", "normal")
	params.Set("_dc", strconv.FormatInt(time.Now().UnixNano()/1e6, 10))
	method := "GET"
	client := &http.Client{}
	resp, err := http.NewRequest(method, fmt.Sprintf("%s/%s?%s", APIChapterCardResource, p.ObjectID, params.Encode()), nil)
	// resp, err := p.Session.Client.Get(fmt.Sprintf("%s/%s?%s", APIChapterCardResource, p.ObjectID, params.Encode()))
	if err != nil {
		return "", err
	}
	resp.Header.Add("Host", " mooc1-api.chaoxing.com")
	resp.Header.Add("Connection", " keep-alive")
	resp.Header.Add("User-Agent", " Mozilla/5.0 (Linux; Android 12; SM-N9006 Build/V417IR; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/95.0.4638.74 Mobile Safari/537.36 (schild:e9b05c3f9fb49fef2f516e86ac3c4ff1) (device:SM-N9006) Language/zh_CN com.chaoxing.mobile/ChaoXingStudy_3_6.3.7_android_phone_10822_249 (@Kalimdor)_4627cad9c4b6415cba5dc6cac39e6c96")
	resp.Header.Add("X-Requested-With", " XMLHttpRequest")
	resp.Header.Add("Accept", " */*")
	resp.Header.Add("Sec-Fetch-Site", " same-origin")
	resp.Header.Add("Sec-Fetch-Mode", " cors")
	resp.Header.Add("Sec-Fetch-Dest", " empty")
	resp.Header.Add("Referer", " https://mooc1-api.chaoxing.com/ananas/modules/video/index_wap.html?v=372024-1121-1947")
	resp.Header.Add("Accept-Language", " zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7")
	resp.Header.Add("Cookie", cache.cookie)

	res, err := client.Do(resp)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch video, status code: %d", res.StatusCode)
	}
	body, err := ioutil.ReadAll(res.Body)

	return string(body), nil
}

func (cache *XueXiTUserCache) VideoSubmitStudyTime(p *entity.PointVideoDto, playingTime int, isdrag int /*提交模式，0代表正常视屏播放提交，2代表暂停播放状态，3代表着点击开始播放状态*/, retry int, lastErr error) (string, error) {
	clipTime := fmt.Sprintf("0_%d", p.Duration)
	hash := md5.Sum([]byte(fmt.Sprintf("[%s][%s][%s][%s][%d][%s][%d][%s]",
		p.ClassID, cache.UserID, p.JobID, p.ObjectID, playingTime*1000, "d_yHJ!$pdA~5", p.Duration*1000, clipTime)))
	enc := hex.EncodeToString(hash[:])
	//
	url := "https://mooc1.chaoxing.com/mooc-ans/multimedia/log/a/" + p.Cpi + "/" + p.DToken + "?clazzId=" + p.ClassID + "&playingTime=" + strconv.Itoa(playingTime) + "&duration=" + strconv.Itoa(p.Duration) + "&clipTime=" + clipTime + "&objectId=" + p.ObjectID + "&otherInfo=" + p.OtherInfo + "&jobid=" + p.JobID + "&userid=" + cache.UserID + "&isdrag=" + strconv.Itoa(isdrag) + "&view=pc&enc=" + enc + "&rt=0.9&dtype=Video&_t=" + strconv.FormatInt(time.Now().UnixMilli(), 10)

	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "mooc1.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")
	//req.Header.Add("Cookie", "k8s=1740601804.712.17514.777640; jrose=702DC91E24ECD4078821D470F213E5C6.mooc-1385974980-0c6l3; route=0a65fa708818ad1416475328b69707fd")
	req.Header.Add("Cookie", cache.cookie)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	//fmt.Println(string(body))
	return string(body), nil
}
func (cache *XueXiTUserCache) VideoDtoPlayReport(p *entity.PointVideoDto, playingTime int, isdrag int /*提交模式，0代表正常视屏播放提交，2代表暂停播放状态，3代表着点击开始播放状态*/, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	clipTime := fmt.Sprintf("0_%d", p.Duration)
	hash := md5.Sum([]byte(fmt.Sprintf("[%s][%s][%s][%s][%d][%s][%d][%s]",
		p.ClassID, cache.UserID, p.JobID, p.ObjectID, playingTime*1000, "d_yHJ!$pdA~5", p.Duration*1000, clipTime)))
	enc := hex.EncodeToString(hash[:])
	//fmt.Println(enc)
	client := &http.Client{}
	params := url.Values{}
	params.Set("otherInfo", p.OtherInfo)
	params.Set("playingTime", strconv.Itoa(playingTime))
	params.Set("duration", strconv.Itoa(p.Duration))
	params.Set("jobid", p.JobID)
	params.Set("clipTime", clipTime)
	params.Set("clazzId", p.ClassID)
	params.Set("objectId", p.ObjectID)
	params.Set("userid", cache.UserID)
	params.Set("isdrag", strconv.Itoa(isdrag)) //0为正常播放，2为点击暂停播放状态，3为点击开始播放
	params.Set("enc", enc)
	params.Set("rt", fmt.Sprintf("%f", p.RT))
	//params.Set("retry", "0.9")
	params.Set("dtype", "Video")
	params.Set("view", "pc")
	params.Set("rt", "0.9")
	params.Set("_t", strconv.FormatInt(time.Now().UnixMilli(), 10))

	// 自定义编码函数以保留 & 和 =
	encodedParams := encodeWithSafeChars(params)
	method := "GET"

	resp, err := http.NewRequest(method, fmt.Sprintf("%s/%s/%s?%s", APIVideoPlayReport, p.Cpi, p.DToken, encodedParams), nil)
	if err != nil {
		return "", err
	}

	resp.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36")
	resp.Header.Add("Accept", "*/*")
	resp.Header.Add("Host", "mooc1.chaoxing.com")
	resp.Header.Add("Connection", "keep-alive")
	resp.Header.Add("Cookie", cache.cookie)
	resp.Header.Add("Referer", "https://mooc1.chaoxing.com/ananas/modules/video/index.html?v=2023-1110-1610")
	resp.Header.Add("Content-Type", " application/json")
	res, err := client.Do(resp)
	if err != nil {
		time.Sleep(150 * time.Millisecond)
		return cache.VideoDtoPlayReport(p, playingTime, isdrag, retry-1, err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch video, status code: %d", res.StatusCode)
	}
	body, err := ioutil.ReadAll(res.Body)
	return string(body), nil
}

// encodeWithSafeChars 自定义编码函数，保留 & 和 =
func encodeWithSafeChars(values url.Values) string {
	var result []string
	for key, list := range values {
		for _, value := range list {
			// 手动编码键和值，但不编码 & 和 =
			encodedKey := url.QueryEscape(key)
			encodedValue := url.QueryEscape(value)
			// 替换 %3D (等号) 和 %26 (与号) 回原字符
			encodedKey = replaceSpecialChars(encodedKey)
			encodedValue = replaceSpecialChars(encodedValue)
			result = append(result, encodedKey+"="+encodedValue)
		}
	}
	return strings.Join(result, "&")
}

// replaceSpecialChars 将 %3D 和 %26 替换回等号和与号
func replaceSpecialChars(s string) string {
	return strings.NewReplacer("%3D", "=", "%26", "&").Replace(s)
}

// WorkFetchQuestion 获取作业题目
func (cache *XueXiTUserCache) WorkFetchQuestion(p *entity.PointWorkDto) (string, error) {
	method := "GET"

	params := url.Values{}
	params.Add("courseid", p.CourseID)
	var SorW func(string, string) string
	SorW = func(s string, w string) string {
		if s != "0" {
			return fmt.Sprintf("%s-%s", s, w)
		}
		return w
	}
	params.Add("workid", SorW(p.SchoolID, p.WorkID))
	params.Add("jobid", p.JobID)
	params.Add("needRedirect", "true")
	params.Add("knowledgeid", strconv.Itoa(p.KnowledgeID))
	params.Add("userid", p.PUID)
	params.Add("ut", "s")
	params.Add("clazzId", p.ClassID)
	params.Add("cpi", p.Cpi)
	params.Add("ktoken", p.KToken)
	params.Add("enc", p.Enc)
	client := &http.Client{}
	req, err := http.NewRequest(method, PageMobileWork+"?"+params.Encode(), nil)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "mooc1-api.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Cookie", cache.cookie)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	return string(body), nil
}

func (cache *XueXiTUserCache) WorkCommit(p *entity.PointWorkDto, fields []entity.WorkInputField) (string, error) {
	method := "POST"

	//TODO 此处需要对答案进行分析后提交 具体body模板 在 examples 中
	payload := strings.NewReader("")

	client := &http.Client{}
	req, err := http.NewRequest(method, ApiWorkCommit, payload)

	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	req.Header.Add("Cookie", cache.cookie)
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "mooc1-api.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	return string(body), nil
}

func (cache *XueXiTUserCache) DocumentDtoReadingReport(p *entity.PointDocumentDto) (string, error) {
	method := "GET"

	client := &http.Client{}
	params := url.Values{}

	params.Add("jobid", p.JobID)
	params.Add("knowledgeid", strconv.Itoa(p.KnowledgeID))
	params.Add("courseid", p.CourseID)
	params.Add("clazzid", p.ClassID)
	params.Add("jtoken", p.Jtoken)
	params.Add("_dc", strconv.FormatInt(time.Now().UnixMilli(), 10))

	resp, err := http.NewRequest(method, ApiDocumentReadingReport+"?"+params.Encode(), nil)
	if err != nil {
		return "", err
	}

	resp.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")
	resp.Header.Add("Accept", "*/*")
	resp.Header.Add("Host", "mooc1.chaoxing.com")
	resp.Header.Add("Connection", "keep-alive")
	resp.Header.Add("Cookie", cache.cookie)
	resp.Header.Add("Content-Type", " application/json")

	res, err := client.Do(resp)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch video, status code: %d", res.StatusCode)
	}

	body, err := ioutil.ReadAll(res.Body)
	return string(body), nil
}
