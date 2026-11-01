package examples

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/yatori-dev/yatori-go-core/v3"
	"github.com/yatori-dev/yatori-go-core/v3/global"
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
	req.Header.Add("Cookie", os.Getenv("YATORI_COOKIE"))
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

func TestXXT(t *testing.T) {
	//测试账号
	setup()
	user := global.Config.Users[78]
	client, err := yatori.New(yatori.Config{
		Platform: yatori.XueXiTong,
		Account: yatori.Account{
			Username: user.Account,
			Password: user.Password,
			// 英华、海旗等平台需要填写 URL：
			// URL: "https://example.com",
		},
		Options: yatori.Options{
			// ProxyURL: "http://127.0.0.1:7890",
			// AI: &yatori.AIConfig{
			// 	Type:   yatori.AIDeepSeek,
			// 	Model:  "MODEL",
			// 	APIKey: "KEY",
			// },
		},
	})
	if err != nil {
		panic(err)
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.Login(ctx); err != nil {
		panic(err)
	}

	courses, err := client.Courses(ctx)
	if err != nil {
		panic(err)
	}
	for _, course := range courses {
		fmt.Println(course.Name, course.Progress)
		nodes, err := client.Tree(ctx, course)
		if err != nil {
			continue
		}
		for _, node := range yatori.Flatten(nodes) {
			if node.Completed || node.Kind == yatori.NodeChapter {
				continue
			}
			// Study 是“提交一次学习进度”的原语：
			// 调用方自行控制循环、间隔和风控节奏。
			_, _ = client.Study(ctx, yatori.StudyRequest{
				Course:   course,
				Node:     node,
				Action:   yatori.StudyActionProgress,
				Position: 60 * time.Second,
				Mode:     yatori.StudyModeNormal,
			})
		}
	}
}
