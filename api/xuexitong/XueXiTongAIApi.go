package xuexitong

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/yatori-dev/yatori-go-core/utils"
)

var courseLocks sync.Map // map[string]*sync.Mutex

func lockByCourse(courseId string) func() {
	muIface, _ := courseLocks.LoadOrStore(courseId, &sync.Mutex{})
	mu := muIface.(*sync.Mutex)
	mu.Lock()

	// 返回 unlock 函数，方便 defer
	return func() {
		mu.Unlock()
	}
}

// 拉取学习通AI必要参数
func (cache *XueXiTUserCache) XXTAiInformApi(clazzId, courseId, cpi string, retry int, lastErr error) (string, error) {

	if retry < 0 {
		return "", lastErr
	}
	//url := "https://stat2-ans.chaoxing.com/bot/index?fromWorkbench=true&upload=true&clazzid=134204187&showToolbox=false&bgColorNone=true&app_id=1192651262850&courseid=258101827&cpi=411545273&bot_id=7438777570621653018&ut=s"
	urlStr := "https://stat2-ans.chaoxing.com/bot/index?fromWorkbench=true&upload=true&clazzid=" + clazzId + "&showToolbox=false&bgColorNone=true&app_id=1192651262850&courseid=" + courseId + "&cpi=" + cpi + "&bot_id=7438777570621653018&ut=s"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, urlStr, nil)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Add("Cache-Control", "max-age=0")
	req.Header.Add("sec-ch-ua", "\"Microsoft Edge\";v=\"143\", \"Chromium\";v=\"143\", \"Not A(Brand\";v=\"24\"")
	req.Header.Add("sec-ch-ua-mobile", "?0")
	req.Header.Add("sec-ch-ua-platform", "\"Windows\"")
	req.Header.Add("Upgrade-Insecure-Requests", "1")
	req.Header.Add("Sec-Fetch-Site", "none")
	req.Header.Add("Sec-Fetch-Mode", "navigate")
	req.Header.Add("Sec-Fetch-User", "?1")
	req.Header.Add("Sec-Fetch-Dest", "document")
	req.Header.Add("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6")
	req.Header.Add("Host", "stat2-ans.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return cache.XXTAiInformApi(clazzId, courseId, cpi, retry-1, lastErr)
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

// 请求学习通AI获取答复
func (cache *XueXiTUserCache) XXTAiAnswerApi(cozeEnc, userId, courseId, classId, conversationId, courseName, studentName, personId, content string, retry int, lastErr error) (string, error) {
	unlock := lockByCourse(courseId) //对AI加锁
	defer unlock()
	//url := "https://stat2-ans.chaoxing.com/stat2/bot/talk-v1?cozeEnc=129ca94f26a7802fd8061ef32f129b4c&botId=7438777570621653018&userId=346635955&appId=1192651262850&courseid=258101827&clazzid=134204187&ut=s"
	urlStr := "https://stat2-ans.chaoxing.com/stat2/bot/talk-v1?cozeEnc=" + cozeEnc + "&botId=7438777570621653018&userId=" + userId + "&appId=1192651262850&courseid=" + courseId + "&clazzid=" + classId + "&ut=s"

	body := `[{"role":"user","content":"` + content + `","baseData":{"conversationId":"` + conversationId + `","userId":"` + userId + `","appId":"1192651262850","botId":"7438777570621653018","custom_variables":{"courseName":"` + courseName + `","studentName":"` + studentName + `","weakKnowledgePoint":"{}"},"shortcut_command":{},"sourceInfo":"","sdkFlag":"false","courseid":"` + courseId + `","clazzid":"` + classId + `","personid":"` + personId + `"}}]`

	req, _ := http.NewRequest("POST", urlStr, bytes.NewBuffer([]byte(body)))
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("sec-ch-ua-platform", "\"Windows\"")
	req.Header.Add("sec-ch-ua", "\"Microsoft Edge\";v=\"143\", \"Chromium\";v=\"143\", \"Not A(Brand\";v=\"24\"")
	req.Header.Add("sec-ch-ua-mobile", "?0")
	req.Header.Add("Origin", "https://stat2-ans.chaoxing.com")
	req.Header.Add("Sec-Fetch-Site", "same-origin")
	req.Header.Add("Sec-Fetch-Mode", "cors")
	req.Header.Add("Sec-Fetch-Dest", "empty")
	req.Header.Add("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "stat2-ans.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}

	client := &http.Client{
		Timeout: 0, // 必须 0，否则流式会被提前中断
	}

	resp, err := client.Do(req)
	if err != nil {
		return cache.XXTAiAnswerApi(cozeEnc, userId, courseId, classId, conversationId, courseName, studentName, personId, content, retry-1, lastErr)
		//panic(err)
	}
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)

	var finalAnswer string

	for {
		// 按行读取
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("read error:", err)
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		type RespChunk struct {
			ID      string `json:"id"`
			Type    string `json:"type"`
			Content string `json:"content"`
		}
		// 每段 chunk 以 $_$ 分割
		parts := strings.Split(line, "$_$")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" || p == "server-heartbeat" || strings.HasPrefix(p, "server-current-chatid") {
				continue
			}

			// 尝试解析 JSON
			var chunk RespChunk
			err := json.Unmarshal([]byte(p), &chunk)
			if err != nil {
				continue
			}

			// 只拼接 type=coreAnswer 的内容
			if chunk.Type == "coreAnswer" {
				finalAnswer += chunk.Content
			}
		}
	}
	finalAnswer = strings.ReplaceAll(finalAnswer, "&quot;", `"`) //替换非法字符
	finalAnswer = strings.ReplaceAll(finalAnswer, "&nbsp;", ` `) //替换非法字符
	finalAnswer = strings.ReplaceAll(finalAnswer, "&amp;", `&`)  //替换非法字符
	finalAnswer = strings.ReplaceAll(finalAnswer, "&lt;", `<`)   //替换非法字符
	finalAnswer = strings.ReplaceAll(finalAnswer, "&gt;", `>`)   //替换非法字符
	//json格式检查逻辑
	var answers []string
	err = json.Unmarshal([]byte(finalAnswer), &answers)
	if err != nil {
		content += "\n\n你刚才生成的回复未严格遵循json格式，我无法正常解析，请你重新生成！！！"
		finalAnswer, err = cache.XXTAiAnswerApi(cozeEnc, userId, courseId, classId, conversationId, courseName, studentName, personId, content, retry-1, lastErr)
	}

	//fmt.Println("最终回复内容：", finalAnswer)
	return finalAnswer, nil
}
