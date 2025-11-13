package aiq

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/yatori-dev/yatori-go-core/que-core/qentity"
	"github.com/yatori-dev/yatori-go-core/que-core/qtype"
)

// 单选题处理策略
func handleSingleChoice(topic qentity.Question) AIChatMessages {
	problem := buildProblemHeader(topic.Type, topic)
	return AIChatMessages{Messages: []Message{
		{Role: "system", Content: `接下来你只需要回答选项对应内容即可...回答格式严格遵循json范式，比如：["选项1内容"]`},
		{Role: "system", Content: "就算你不知道选什么也随机选，答题过程中你无需回答任何解释，只需按照预定json范式输出！！！"},
		{Role: "system", Content: exampleSingleChoice()},
		{Role: "user", Content: problem},
	}}
}

// 构建AI题目消息
func BuildAiQuestionMessage(topic qentity.Question) AIChatMessages {
	switch qtype.Index(topic.Type) {
	case qtype.SingleChoice: //单选题
		return handleSingleChoice(topic)
	case qtype.MultipleChoice: //多选题
		return handleMultipleChoice(topic)
	case qtype.TrueOrFalse: //判断题
		return handleTrueFalse(topic)
	case qtype.FillInTheBlank: //填空题
		return handleFillInTheBlank(topic)
	case qtype.ShortAnswer:
		return handleShortAnswer(topic)
	case qtype.TermExplanation:
		return handleTermExplanationAnswer(topic)
	case qtype.Essay:
		return handleEssayAnswer(topic)
	case qtype.Matching:
		return handleMatchingAnswer(topic)

	}
	return AIChatMessages{}
}

// 将AI的回复转换到Question里面
func ResponseTurnQuestion(question qentity.Question, response string) qentity.Question {
	var answers []string
	json.Unmarshal([]byte(response), &answers)
	question.Answers = answers
	return question
}

// 多选题处理策略
func handleMultipleChoice(topic qentity.Question) AIChatMessages {
	problem := buildProblemHeader("多选题", topic)
	return AIChatMessages{Messages: []Message{
		{Role: "system", Content: "接下来你只需要回答选项对应内容即可...格式：[\"选项1\",\"选项2\"]"},
		{Role: "system", Content: "就算你不知道选什么也随机选...无需回答任何解释！！！"},
		{Role: "system", Content: exampleMultipleChoice()},
		{Role: "user", Content: problem},
	}}
}

// 判断题处理策略
func handleTrueFalse(topic qentity.Question) AIChatMessages {
	problem := buildProblemHeader("判断题", topic)
	return AIChatMessages{Messages: []Message{
		{Role: "system", Content: "接下来你只需要回答“正确”或者“错误”即可...格式：[\"正确\"]"},
		{Role: "system", Content: "就算你不知道选什么也随机选...无需回答任何解释！！！"},
		{Role: "system", Content: exampleTrueFalse()},
		{Role: "user", Content: problem},
	}}
}

// 填空题处理策略
func handleFillInTheBlank(topic qentity.Question) AIChatMessages {
	problem := buildProblemHeader("填空题", topic)
	return AIChatMessages{Messages: []Message{
		{Role: "system", Content: "其中，“（answer_数字）”相关字样的地方是你需要填写答案的地方，回答时请严格遵循json格式：[\"答案1\",\"答案2\"]"},
		{Role: "system", Content: "就算你不知道选什么也随机选...无需回答任何解释！！！"},
		{Role: "system", Content: exampleFillInTheBlank()},
		{Role: "user", Content: problem},
	}}
}

// 简答题处理策略
func handleShortAnswer(topic qentity.Question) AIChatMessages {
	problem := buildProblemHeader("简答题", topic)
	return AIChatMessages{Messages: []Message{
		{Role: "system", Content: `这是一个简答题，回答时请严格遵循json格式，包括换行等特殊符号也要遵循json语法：["答案"]，注意不要拆分答案！！！`},
		//{Role: "system", Content: "就算你不知道选什么也随机选...无需回答任何解释！！！"},
		{Role: "system", Content: exampleShortAnswer()},
		{Role: "user", Content: problem},
	}}
}

// 名词解释处理策略
func handleTermExplanationAnswer(topic qentity.Question) AIChatMessages {
	problem := buildProblemHeader(topic.Type, topic)
	return AIChatMessages{Messages: []Message{
		{Role: "system", Content: `这是一个名词解释题，回答时请严格遵循json格式，包括换行等特殊符号也要遵循json语法：["答案"]，注意不要拆分答案！！！`},
		{Role: "system", Content: exampleTermExplanationAnswer()},
		{Role: "user", Content: problem},
	}}
}

// 论述题处理策略
func handleEssayAnswer(topic qentity.Question) AIChatMessages {
	problem := buildProblemHeader(topic.Type, topic)
	return AIChatMessages{Messages: []Message{
		{Role: "system", Content: `这是一个论述题，回答时请严格遵循json格式，包括换行等特殊符号也要遵循json语法：["答案"]，注意不要拆分答案！！！`},
		{Role: "system", Content: exampleEssayAnswer()},
		{Role: "user", Content: problem},
	}}
}

// 连线题处理策略
func handleMatchingAnswer(topic qentity.Question) AIChatMessages {
	problem := buildProblemHeader(topic.Type, topic)
	return AIChatMessages{Messages: []Message{
		{Role: "system", Content: "接下来你只需要以json格式回答选项对应内容即可，比如：[\"xxx->xxx\",\"xxx->xxx\"]"},
		{Role: "system", Content: "就算你不知道选什么也随机按指定要求格式回答...无需回答任何解释！！！"},
		{Role: "system", Content: exampleMatchingAnswer()},
		{Role: "user", Content: problem},
	}}
}

// 构建题目头部信息
func buildProblemHeader(topicType string, topic qentity.Question) string {
	selectStr := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}
	sprintf := fmt.Sprintf(`题目类型：%s
题目内容：
%s\n`, topicType, topic.Content)
	switch qtype.Index(topic.Type) {
	case qtype.SingleChoice: //单选题
		for i, option := range topic.Options {
			sprintf += selectStr[i] + "." + option + "\n"
		}
	case qtype.MultipleChoice: //多选题
		for i, option := range topic.Options {
			sprintf += selectStr[i] + "." + option + "\n"
		}
	case qtype.TrueOrFalse: //判断题
		for i, option := range topic.Options {
			sprintf += selectStr[i] + "." + option + "\n"
		}
	case qtype.FillInTheBlank: //填空题
		for i, option := range topic.Options {
			sprintf += selectStr[i] + "." + option + "\n"
		}
	case qtype.ShortAnswer: //简答题
	case qtype.TermExplanation: //名词解释
	case qtype.Essay: //论述题
	case qtype.Matching:
		sprintf += "组别一：\n"
		for _, option := range topic.Options {
			if strings.HasPrefix(option, "[1]") {
				sprintf += strings.Replace(option, "[1]", "", 1) + "\n"
			}
		}

		sprintf += "组别二：\n"
		for _, option := range topic.Options {
			if strings.HasPrefix(option, "[2]") {
				sprintf += strings.Replace(option, "[2]", "", 1) + "\n"
			}
		}

	}

	return sprintf
}

// 单选题示例
func exampleSingleChoice() string {
	return `比如：
试卷名称：考试
题目类型：单选
题目内容：新中国是什么时候成立的
A. 1949年10月5日
B. 1949年10月1日
C. 1949年09月1日
D. 2002年10月1日

那么你应该回答选项B的内容：["1949年10月1日"]。注意不要携带A，B，C，D等选项前缀。`
}

// 多选题示例
func exampleMultipleChoice() string {
	return `比如：
试卷名称：考试
题目类型：多选题
题目内容：马克思关于资本积累的学说是剩余价值理论的重要组成部分...
A. 资本主义扩大再生产的源泉
B. 资本有机构成呈现不断降低趋势的根本原因
C. 社会财富占有两极分化的重要原因
D. 资本主义社会失业现象产生的根源

那么你应该回答选项A、B、D的内容：["资本主义扩大再生产的源泉","社会财富占有两极分化的重要原因","资本主义社会失业现象产生的根源"]，注意不要携带A，B，C，D等选项前缀。`
}

// 判断题示例
func exampleTrueFalse() string {
	return `比如：
试卷名称：考试
题目类型：判断
题目内容：新中国是什么时候成立是1949年10月1日吗？
A. 正确
B. 错误

那么你应该回答选项A的内容：["正确"]。注意不要携带A，B，C，D等选项前缀。`
}

// 填空题示例
func exampleFillInTheBlank() string {
	return ` 比如：
试卷名称：考试
题目类型：填空
题目内容：新中国成立于（ ）年。
答案：1949

那么你应该回答：["1949"]`
}

func exampleShortAnswer() string {
	return `比如：
试卷名称：考试
题目类型：简答
题目内容：请简述中国和外国的国别 differences
答案：中国和外国的国别 differences

那么你应该回答： ["中国和外国的国别 differences"]`
}

// 名词解释
func exampleTermExplanationAnswer() string {
	return `比如：
试卷名称：考试
题目类型：名词解释
题目内容：绿色设计
答案：绿色设计是指在产品、建筑、工程或系统设计的全过程中，将环境保护和可持续发展理念融入其中的一种设计方法。

那么你应该回答： ["绿色设计是指在产品、建筑、工程或系统设计的全过程中，将环境保护和可持续发展理念融入其中的一种设计方法。"]`
}

// 论述题
func exampleEssayAnswer() string {
	return `比如：
试卷名称：考试
题目类型：论述题
题目内容：试述设计艺术的构成元素
答案：设计艺术的构成元素包括点、线、面、形体、色彩、质感与空间等。它们相互依存、互为补充，通过合理的组织和运用，形成和谐、统一而富有美感的设计作品。

那么你应该回答（回答字数不能少于500字）： ["设计艺术的构成元素包括点、线、面、形体、色彩、质感与空间等。它们相互依存、互为补充，通过合理的组织和运用，形成和谐、统一而富有美感的设计作品。"]`
}

// 连线题
func exampleMatchingAnswer() string {
	return `比如：
试卷名称：考试
题目类型：连线题
题目内容：
5.[连线题] 下列认知心理学家与其所做的经典研究之间的关系：

1、桑代克 ()
2、威特金 ()
3、凯利 ()
4、卡特尔 ()

A、迷箱实验
B、角色建构测验
C、16PF
D、棒框实验

答案：第一空：桑代克->迷箱实验、威特金->棒框实验、凯利->角色建构测验、卡特尔->16PF

那么你应该回答： ["桑代克->迷箱实验","威特金->棒框实验","凯利->角色建构测验","卡特尔->16PF"]`

}
