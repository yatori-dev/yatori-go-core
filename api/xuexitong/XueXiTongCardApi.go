package xuexitong

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/yatori-dev/yatori-go-core/api/entity"
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

func (cache *XueXiTUserCache) VideoDtoFetch(p *entity.PointVideoDto) (bool, error) {
	params := url.Values{}
	params.Set("k", strconv.Itoa(p.FID))
	params.Set("flag", "normal")
	params.Set("_dc", strconv.FormatInt(time.Now().UnixNano()/1e6, 10))
	method := "GET"
	client := &http.Client{}
	resp, err := http.NewRequest(method, fmt.Sprintf("%s/%s?%s", APIChapterCardResource, p.ObjectID, params.Encode()), nil)
	// resp, err := p.Session.Client.Get(fmt.Sprintf("%s/%s?%s", APIChapterCardResource, p.ObjectID, params.Encode()))
	if err != nil {
		return false, err
	}

	resp.Header.Add("Cookie", cache.cookie)

	res, err := client.Do(resp)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return false, fmt.Errorf("failed to fetch video, status code: %d", res.StatusCode)
	}

	var jsonResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&jsonResponse); err != nil {
		return false, err
	}

	p.DToken = jsonResponse["dtoken"].(string)
	p.Duration = int(jsonResponse["duration"].(float64))
	p.Title = jsonResponse["filename"].(string)

	if jsonResponse["status"].(string) == "success" {
		p.Logger.Printf("Fetch successful: %s", p)
		return true, nil
	}

	p.Logger.Println("Fetch failed")
	return false, nil
}

func VideoDtoPlayReport(p *entity.PointVideoDto, playingTime int) (map[string]interface{}, error) {
	clipTime := fmt.Sprintf("0_%d", p.Duration)
	hash := md5.Sum([]byte(fmt.Sprintf("[%s][%s][%s][%d][%s][%d][%s]",
		p.PUID, p.JobID, p.ObjectID, playingTime*1000, "d_yHJ!$pdA~5", p.Duration*1000, clipTime)))
	enc := hex.EncodeToString(hash[:])

	params := url.Values{}
	params.Set("otherInfo", p.OtherInfo)
	params.Set("playingTime", strconv.Itoa(playingTime))
	params.Set("duration", strconv.Itoa(p.Duration))
	params.Set("jobid", p.JobID)
	params.Set("clipTime", clipTime)
	params.Set("clazzId", strconv.Itoa(p.FID))
	params.Set("objectId", p.ObjectID)
	params.Set("userid", p.Session.Acc.PUID)
	params.Set("isdrag", "0")
	params.Set("enc", enc)
	params.Set("rt", fmt.Sprintf("%f", p.RT))
	params.Set("dtype", "Video")
	params.Set("view", "pc")
	params.Set("_t", strconv.FormatInt(time.Now().UnixNano()/1e6, 10))

	reqURL := fmt.Sprintf("%s/%d/%s?%s", APIVideoPlayReport, p.FID, p.DToken, params.Encode())
	resp, err := p.Session.Client.Get(reqURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to report play, status code: %d", resp.StatusCode)
	}

	var jsonResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&jsonResponse); err != nil {
		return nil, err
	}

	if errorMsg, exists := jsonResponse["error"].(string); exists {
		return nil, &APIError{Message: errorMsg}
	}

	p.Logger.Printf("Play report successful: %d/%d", playingTime, p.Duration)
	return jsonResponse, nil
}
