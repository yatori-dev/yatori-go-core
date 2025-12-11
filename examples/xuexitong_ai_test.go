package examples

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

type RespChunk struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Content string `json:"content"`
}

func TestXXtAiAnswer(t *testing.T) {
	xxtAiAnswer()
}

func xxtAiAnswer() {
	url := "https://stat2-ans.chaoxing.com/stat2/bot/talk-v1?cozeEnc=129ca94f26a7802fd8061ef32f129b4c&botId=7438777570621653018&userId=346635955&appId=1192651262850&courseid=258101827&clazzid=134204187&ut=s"

	body := `[{"role":"user","content":"题目类型：单选题
题目内容：
21.[单选题] 下列国民党右派制造的反革命活动（事变），其先后顺序是(       )①“七·一五”反革命政变②“四·一二”反革命政变③中山舰事件④西山会议\nA.①②③④
B.②③①④
C.④①②③
D.④③②①
接下来无论出现任何题目，你都必须只回答题目中某个选项对应的内容，并严格按照以下要求作答：

【回答规则】
1. 最终输出必须严格遵循 JSON 数组格式，例如：[\"选项内容\"]
2. 数组中只能有一个字符串元素。
3. 字符串中不能包含选项前缀，如 A. B. C. D. 等，只能输出选项的纯内容。
4. 不能输出解析、解释步骤、理由、提示语或任何多余文本。
5. 不能输出题目本身、不能输出其他格式，只能输出 JSON 数组。
6. 如果你无法判断正确答案，也必须随机选择一个选项的内容进行输出，不允许回答“我不知道”“无法判断”之类内容。

【格式要求】
- 只能输出 JSON
- 不允许换行，若内容中需要换行必须使用\n
- 不能出现额外的空格、标点或第二层数组

【示例】
题目如下：
试卷名称：考试
题目类型：单选题
题目内容：新中国是什么时候成立的？
A. 1949年10月5日
B. 1949年10月1日
C. 1949年09月1日
D. 2002年10月1日

你必须回答：
[\"1949年10月1日\"]
（注意：不能输出 B 或 B. 等前缀，只能输出选项的内容本身）

请严格按照以上规则回答后续所有题目，只输出 JSON 数组格式内容，不得违反任何一项要求。","baseData":{"conversationId":"7582447454588862490","userId":"346635955","appId":"1192651262850","botId":"7438777570621653018","custom_variables":{"courseName":"大学教育","studentName":"蔡卓睿","weakKnowledgePoint":"{}"},"shortcut_command":{},"sourceInfo":"","sdkFlag":"false","courseid":"258101827","clazzid":"134204187","personid":"475997091"}}]`

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte(body)))
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36 Edg/143.0.0.0")
	req.Header.Add("sec-ch-ua-platform", "\"Windows\"")
	req.Header.Add("sec-ch-ua", "\"Microsoft Edge\";v=\"143\", \"Chromium\";v=\"143\", \"Not A(Brand\";v=\"24\"")
	req.Header.Add("sec-ch-ua-mobile", "?0")
	req.Header.Add("Origin", "https://stat2-ans.chaoxing.com")
	req.Header.Add("Sec-Fetch-Site", "same-origin")
	req.Header.Add("Sec-Fetch-Mode", "cors")
	req.Header.Add("Sec-Fetch-Dest", "empty")
	req.Header.Add("Referer", "https://stat2-ans.chaoxing.com/bot/index?fromWorkbench=true&upload=true&clazzid=134204187&showToolbox=false&bgColorNone=true&app_id=1192651262850&courseid=258101827&cpi=411545273&bot_id=7438777570621653018&ut=s")
	req.Header.Add("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6")
	req.Header.Add("Cookie", "tl=1; fanyamoocs=11401F839C536D9E; fid=10596; _uid=346635955; _d=1765424983114; UID=346635955; vc3=WJVDBtvmkUEaY0HsSmz%2Bleu2m5HukYPTw7PB%2FTZKTZjYnxx2qYIDT%2BmHnhaQkE58YFQ0CD01IGQVsiyoLkcweY91pAnogBmn1p1V0RZaxXcJBfJcSmcGFbOYFAAl%2F3PqTI6jAu7NUqfaMaRzUAy75kmarCFVqmqBcbs3i0IytNo%3D105778862c0e2bf79d8b479d117e895d; uf=b2d2c93beefa90dc495549838143a13b264677447b1a2384b8cd17c4874b05f58a6d64458f94593ccfd0e74dd1db6b4ab7f7d292066fbd0bc49d67c0c30ca5043ad701c8b4cc548c0234d89f51c3dccfb0f1a1db51ab43f5fb98ce0e6210c3884a878d0a9a7b05da6103a97f8cd189bc0ecd8afc9e03f695363fdc8451f0c870a8a2139068a68c9d5d1b47daef7984b1da9735baa04d8d5fce71fc6e59483dd39b16e3a7097306130b46772558e12a315adddf55eab87d53e9fdc681bdf07734; cx_p_token=ce364ca7460a4a413cb03059b7d0a3ed; p_auth_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOiIzNDY2MzU5NTUiLCJsb2dpblRpbWUiOjE3NjU0MjQ5ODMxMTUsImV4cCI6MTc2NjAyOTc4M30.Uq9A5gcOUEB6HkxVP4WfGczFiURBghQTt_0NxamBKPg; xxtenc=f8c84ceb53bc45f40b7d9bfaaa413810; DSSTASH_LOG=C_38-UN_10038-US_346635955-T_1765424983116; source=\"\"; thirdRegist=0; k8s=1765425012.682.1508.646214; jrose=4898A8A76C9F0B11111E9891394B03C3.mooc-statistics2-1746325997-srq6z; route=bca6486eee9aca907e6257b7921729c3")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "stat2-ans.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")

	client := &http.Client{
		Timeout: 0, // 必须 0，否则流式会被提前中断
	}

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
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

	fmt.Println("最终回复内容：", finalAnswer)
}
