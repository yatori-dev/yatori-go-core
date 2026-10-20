package aiq

// AI答题
//
// 各平台接口已按官方最新文档核对（2026-10）：
//   - 通义千问：OpenAI 兼容模式 /compatible-mode/v1/chat/completions，默认 qwen3.7-plus
//   - 智谱 ChatGLM：/api/paas/v4/chat/completions，默认 glm-5.3-flash（思考默认开启，低强度）
//   - 讯飞星火：spark-api-open HTTP 接口，默认 4.0Ultra（Max 套餐 2026-03-10 已下线）
//   - 火山方舟豆包：/api/v3/chat/completions，默认 doubao-seed-2-1-lite-260915
//   - OpenAI：Responses API /v1/responses，默认 gpt-5.5，失败自动回退 Chat Completions
//   - DeepSeek：/chat/completions，默认 deepseek-flash
//   - 硅基流动：/v1/chat/completions，默认 deepseek-ai/DeepSeek-V4-Flash
//   - 秘塔 AI 搜索：/api/v1/chat/completions，默认 fast

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/yatori-dev/yatori-go-core/models/ctype"
)

const (
	tongYiEndpoint          = "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"
	chatGLMEndpoint         = "https://open.bigmodel.cn/api/paas/v4/chat/completions"
	xingHuoEndpoint         = "https://spark-api-open.xf-yun.com/v1/chat/completions"
	douBaoEndpoint          = "https://ark.cn-beijing.volces.com/api/v3/chat/completions"
	openAIResponsesEndpoint = "https://api.openai.com/v1/responses"
	openAIChatEndpoint      = "https://api.openai.com/v1/chat/completions"
	deepSeekEndpoint        = "https://api.deepseek.com/chat/completions"
	siliconFlowEndpoint     = "https://api.siliconflow.cn/v1/chat/completions"
	metaAIEndpoint          = "https://metaso.cn/api/v1/chat/completions"
)

// AIChatMessages ChatGLMChat struct that holds the chat messages.
type AIChatMessages struct {
	Messages []Message `json:"messages"`
}

// Message struct represents individual messages.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AiSem 限制同时并发的 AI 调用数量（容量=2）
var AiSem = make(chan struct{}, 2)

// AggregationAIApi 聚合所有AI接口，直接通过aiType判断然后返回内容
func AggregationAIApi(url,
	model string,
	aiType ctype.AiType,
	aiChatMessages AIChatMessages,
	apiKey string) (string, error) {

	// 获取信号量（阻塞，直到有空位）
	AiSem <- struct{}{}
	defer func() {
		<-AiSem // 释放信号量
	}()

	switch aiType {
	case ctype.ChatGLM:
		return ChatGLMChatReplyApi(model, apiKey, aiChatMessages, 7, nil)
	case ctype.XingHuo:
		return XingHuoChatReplyApi(model, apiKey, aiChatMessages, 7, nil)
	case ctype.TongYi:
		return TongYiChatReplyApi(model, apiKey, aiChatMessages, 7, nil)
	case ctype.DouBao:
		return DouBaoChatReplyApi(model, apiKey, aiChatMessages, 7, nil)
	case ctype.OpenAi:
		return OpenAiReplyApi(model, apiKey, aiChatMessages, 7, nil)
	case ctype.MetaAi:
		return MetaAIReplyApi(model, apiKey, aiChatMessages, 7, nil)
	case ctype.DeepSeek:
		return DeepSeekChatReplyApi(model, apiKey, aiChatMessages, 7, nil)
	case ctype.Silicon:
		return SiliconFlowReplyApi(model, apiKey, aiChatMessages, 7, nil)
	case ctype.Other:
		return OtherChatReplyApi(url, model, apiKey, aiChatMessages, 7, nil)
	default:
		return "", errors.New("AI Type: " + string(aiType))
	}
}

// AICheck AI可用性检测
func AICheck(url, model, apiKey string, aiType ctype.AiType) error {
	aiChatMessages := AIChatMessages{
		Messages: []Message{
			{
				Role:    "user",
				Content: `请你原模原样输出：["测试成功"]`,
			},
		},
	}

	if aiType == "" {
		return errors.New("AI Type: " + "请先填写AIType参数，详细参考官方文档：https://yatori-dev.github.io/yatori-docs/yatori-go-console/docs.html")
	}
	// OTHER 类型允许本地模型不填 API Key，其余平台必须有 Key。
	if aiType != ctype.Other && strings.TrimSpace(apiKey) == "" {
		return errors.New("无效apiKey，请检查apiKey是否正确填写，详细情参考对应AI平台的API文档或者对应yatori的官方文档：https://yatori-dev.github.io/yatori-docs/yatori-go-console/docs.html")
	}
	_, err := AggregationAIApi(url, model, aiType, aiChatMessages, apiKey)

	return err
}

func baseChatBody(model string, messages []Message, temperature float64, maxTokens int) map[string]interface{} {
	return map[string]interface{}{
		"model":       model,
		"messages":    messages,
		"temperature": temperature,
		"max_tokens":  maxTokens,
		"stream":      false,
	}
}

// TongYiChatReplyApi 通义千问API
func TongYiChatReplyApi(
	model,
	apiKey string,
	aiChatMessages AIChatMessages,
	retryNum int, /*最大重连次数*/
	lastErr error,
) (string, error) {
	if model == "" {
		model = "qwen3.7-plus"
	}
	return runChatCompletion("通义千问", tongYiEndpoint, apiKey, model, aiChatMessages, retryNum, lastErr, true,
		func(model string, messages []Message) map[string]interface{} {
			body := baseChatBody(model, messages, 0.2, aiDefaultMaxTokens)
			// Qwen3 系列默认混合思考模式，答题场景关闭思考以降低延迟。
			if strings.Contains(strings.ToLower(model), "qwen3") {
				body["enable_thinking"] = false
			}
			return body
		})
}

// ChatGLMChatReplyApi 智谱 ChatGLM API
func ChatGLMChatReplyApi(
	model,
	apiKey string,
	aiChatMessages AIChatMessages,
	retryNum int, /*最大重连次数*/
	lastErr error,
) (string, error) {
	if model == "" {
		model = "glm-5.3-flash"
	}
	return runChatCompletion("智谱GLM", chatGLMEndpoint, apiKey, model, aiChatMessages, retryNum, lastErr, true,
		func(model string, messages []Message) map[string]interface{} {
			body := baseChatBody(model, messages, 0.2, 8192)
			// GLM-4.5/4.6/5.x 支持思考参数；GLM-5.3 只能开启，不支持关闭。
			if isGLMThinkingModel(model) {
				body["thinking"] = map[string]string{"type": "enabled"}
				if isGLM53Model(model) {
					body["reasoning_effort"] = "low"
				}
			}
			return body
		})
}

func isArkThinkingModel(model string) bool {
	m := strings.ToLower(strings.TrimSpace(model))
	if m == "" || strings.HasPrefix(m, "ep-") {
		return false
	}
	// 只对官方标注支持深度思考的模型系列下发 thinking 参数，
	// 避免给 doubao-seed-character-251128 等非思考模型带上无效字段。
	prefixes := []string{
		"doubao-seed-evolving",
		"doubao-seed-2-1-",
		"doubao-seed-2-0-",
		"doubao-seed-1-6-",
		"doubao-seed-character-260628",
		"glm-5-3-",
		"glm-5-2",
		"deepseek-v4-",
		"kimi-",
	}
	for _, p := range prefixes {
		if strings.HasPrefix(m, p) {
			return true
		}
	}
	return false
}

func isGLMThinkingModel(model string) bool {
	m := strings.ToLower(strings.TrimSpace(model))
	return strings.HasPrefix(m, "glm-4.5") ||
		strings.HasPrefix(m, "glm-4.6") ||
		strings.HasPrefix(m, "glm-5")
}

func isGLM53Model(model string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(model)), "glm-5.3")
}

func isDeepSeekV4Model(model string) bool {
	m := strings.ToLower(strings.TrimSpace(model))
	return strings.Contains(m, "v4") || strings.HasPrefix(m, "deepseek-flash")
}

func isSiliconThinkingModel(model string) bool {
	m := strings.ToLower(strings.TrimSpace(model))
	// 仅对混合思考模型下发 enable_thinking，老的 DeepSeek-V3 等模型不带该字段。
	return strings.Contains(m, "deepseek-v4") ||
		strings.Contains(m, "glm-5") ||
		strings.Contains(m, "kimi-k2.7") ||
		strings.Contains(m, "qwen3")
}

// XingHuoChatReplyApi 讯飞星火API
func XingHuoChatReplyApi(model,
	apiKey string,
	aiChatMessages AIChatMessages,
	retryNum int, /*最大重连次数*/
	lastErr error,
) (string, error) {
	if model == "" {
		// Max(generalv3.5) 套餐已于 2026-03-10 下线，默认使用 Ultra。
		model = "4.0Ultra"
	}
	return runChatCompletion("讯飞星火", xingHuoEndpoint, apiKey, model, aiChatMessages, retryNum, lastErr, true,
		func(model string, messages []Message) map[string]interface{} {
			return baseChatBody(model, messages, 0.2, aiDefaultMaxTokens)
		})
}

// DouBaoChatReplyApi 豆包API（火山方舟）
func DouBaoChatReplyApi(model,
	apiKey string,
	aiChatMessages AIChatMessages,
	retryNum int, /*最大重连次数*/
	lastErr error,
) (string, error) {
	if model == "" {
		model = "doubao-seed-2-1-lite-260915"
	}
	return runChatCompletion("火山方舟豆包", douBaoEndpoint, apiKey, model, aiChatMessages, retryNum, lastErr, true,
		func(model string, messages []Message) map[string]interface{} {
			body := baseChatBody(model, messages, 0.2, aiDefaultMaxTokens)
			// Seed 系列默认开启深度思考，答题场景关闭。
			// 老的自定义 ep-xxx 接入点可能绑定不支持思考的旧模型，此时不要带该参数。
			if isArkThinkingModel(model) {
				body["thinking"] = map[string]string{"type": "disabled"}
			}
			return body
		})
}

// OpenAiReplyApi OpenAI API
// 优先使用官方推荐的 Responses API，遇到不支持该端点的中转/旧网关时
// 自动回退到 Chat Completions。
func OpenAiReplyApi(model,
	apiKey string,
	aiChatMessages AIChatMessages,
	retryNum int, /*最大重连次数*/
	lastErr error,
) (string, error) {
	if retryNum < 0 {
		return "", lastErr
	}
	if strings.TrimSpace(apiKey) == "" {
		return "", errors.New("无效apiKey，请检查apiKey是否正确填写")
	}
	if model == "" {
		model = "gpt-5.5"
	}

	messages := append([]Message(nil), aiChatMessages.Messages...)
	if len(messages) == 0 {
		return "", errors.New("AI请求消息为空")
	}

	useChatCompletions := false
	var last error = lastErr
	jsonFailures := 0

	for attempt := 0; attempt <= retryNum; attempt++ {
		var (
			content   string
			retryable bool
			err       error
		)

		if useChatCompletions {
			content, retryable, err = postOpenAIChatCompletions(model, apiKey, messages)
		} else {
			content, retryable, err = postOpenAIResponses(model, apiKey, messages)
			if err != nil {
				var statusErr *aiHTTPStatusError
				if errors.As(err, &statusErr) &&
					(statusErr.StatusCode == http.StatusNotFound || statusErr.StatusCode == http.StatusMethodNotAllowed) {
					// 中转站/旧网关不支持 Responses API，立即回退 Chat Completions。
					useChatCompletions = true
					content, retryable, err = postOpenAIChatCompletions(model, apiKey, messages)
				}
			}
		}

		if err != nil {
			last = fmt.Errorf("OpenAI API调用失败：%w", err)
			if !retryable {
				return "", last
			}
			sleepBeforeRetry(attempt)
			continue
		}

		candidate, verr := validateJSONAnswer(content)
		if verr == nil {
			return candidate, nil
		}

		jsonFailures++
		last = fmt.Errorf("OpenAI 生成的回复未能正常解析成json：%w", verr)
		if jsonFailures >= 2 {
			return "", last
		}
		messages = append(messages,
			Message{Role: "assistant", Content: content},
			Message{Role: "user", Content: "你刚才生成的回复未严格遵循json格式，我未能正常解析，请你重新生成。"},
		)
		sleepBeforeRetry(attempt)
	}

	if last == nil {
		last = errors.New("OpenAI API调用失败")
	}
	return "", last
}

func postOpenAIResponses(model, apiKey string, messages []Message) (string, bool, error) {
	instructions, input := splitSystemMessages(messages)
	body := map[string]interface{}{
		"model":             model,
		"input":             input,
		"max_output_tokens": 8192,
	}
	if instructions != "" {
		body["instructions"] = instructions
	}

	data, _, retryable, err := doJSONPost(openAIResponsesEndpoint, apiKey, body)
	if err != nil {
		return "", retryable, err
	}
	content, perr := parseChatCompletionResponse(data)
	if perr != nil {
		return "", isRetryableProviderError(perr.Error()), perr
	}
	return content, false, nil
}

func postOpenAIChatCompletions(model, apiKey string, messages []Message) (string, bool, error) {
	body := map[string]interface{}{
		"model":    model,
		"messages": messages,
		"stream":   false,
	}
	return postChatCompletion(openAIChatEndpoint, apiKey, body)
}

// DeepSeekChatReplyApi DeepSeek API
func DeepSeekChatReplyApi(model,
	apiKey string,
	aiChatMessages AIChatMessages,
	retryNum int, /*最大重连次数*/
	lastErr error,
) (string, error) {
	if model == "" {
		model = "deepseek-flash" // 官方当前默认模型（DeepSeek-V4.1-Flash）
	}
	return runChatCompletion("DeepSeek", deepSeekEndpoint, apiKey, model, aiChatMessages, retryNum, lastErr, true,
		func(model string, messages []Message) map[string]interface{} {
			body := baseChatBody(model, messages, 0.2, aiDefaultMaxTokens)
			// V4 默认开启思考，答题场景显式关闭。旧模型名不带该参数，避免 400。
			if isDeepSeekV4Model(model) {
				body["thinking"] = map[string]string{"type": "disabled"}
			}
			return body
		})
}

// SiliconFlowReplyApi 硅基流动 API
func SiliconFlowReplyApi(model,
	apiKey string,
	aiChatMessages AIChatMessages,
	retryNum int, /*最大重连次数*/
	lastErr error,
) (string, error) {
	if model == "" {
		model = "deepseek-ai/DeepSeek-V4-Flash"
	}
	return runChatCompletion("硅基流动", siliconFlowEndpoint, apiKey, model, aiChatMessages, retryNum, lastErr, true,
		func(model string, messages []Message) map[string]interface{} {
			body := baseChatBody(model, messages, 0.2, aiDefaultMaxTokens)
			if isSiliconThinkingModel(model) {
				body["enable_thinking"] = false
			}
			return body
		})
}

// MetaAIReplyApi 秘塔AI搜索
func MetaAIReplyApi(model, apiKey string, aiChatMessages AIChatMessages, retryNum int, lastErr error) (string, error) {
	if retryNum < 0 {
		return "", lastErr
	}
	if strings.TrimSpace(apiKey) == "" {
		return "", errors.New("无效apiKey，请检查apiKey是否正确填写")
	}
	if model == "" {
		model = "fast"
	}

	buildString := ""
	for _, message := range aiChatMessages.Messages {
		buildString += message.Content + "\n"
	}
	if strings.TrimSpace(buildString) == "" {
		return "", errors.New("AI请求消息为空")
	}

	var last error = lastErr
	jsonFailures := 0

	for attempt := 0; attempt <= retryNum; attempt++ {
		entity := map[string]interface{}{
			"q":      buildString,
			"model":  model,
			"format": "simple",
			"scope":  "document",
		}

		// 秘塔接口对非浏览器 UA 较敏感，保留原 Apifox UA 避免被网关拦截。
		data, _, retryable, err := doJSONPostWithHeaders(metaAIEndpoint, apiKey, entity, map[string]string{
			"User-Agent": "Apifox/1.0.0 (https://apifox.com)",
			"Accept":     "*/*",
		})
		if err != nil {
			last = fmt.Errorf("秘塔AI API调用失败：%w", err)
			if !retryable {
				return "", last
			}
			sleepBeforeRetry(attempt)
			continue
		}

		content, perr := parseChatCompletionResponse(data)
		if perr != nil {
			last = fmt.Errorf("秘塔AI API调用失败：%w", perr)
			if isRetryableProviderError(perr.Error()) {
				sleepBeforeRetry(attempt)
				continue
			}
			return "", last
		}

		candidate, verr := validateJSONAnswer(content)
		if verr == nil {
			return candidate, nil
		}

		jsonFailures++
		last = fmt.Errorf("秘塔AI 生成的回复未能正常解析成json：%w", verr)
		if jsonFailures >= 2 {
			return "", last
		}
		buildString += "\n\n你刚才生成的回复未严格遵循json格式，我未能正常解析，请你只输出JSON数组重新回答。"
		sleepBeforeRetry(attempt)
	}

	if last == nil {
		last = errors.New("秘塔AI调用失败")
	}
	return "", last
}

// OtherChatReplyApi 其他支持CHATGPT API格式的AI模型接入
func OtherChatReplyApi(url,
	model,
	apiKey string,
	aiChatMessages AIChatMessages,
	retryNum int, /*最大重连次数*/
	lastErr error,
) (string, error) {
	if retryNum < 0 {
		return "", lastErr
	}
	if strings.TrimSpace(url) == "" {
		return "", errors.New("使用OTHER类型时必须填写aiUrl")
	}

	return runChatCompletion("自定义AI", url, apiKey, model, aiChatMessages, retryNum, lastErr, false,
		func(model string, messages []Message) map[string]interface{} {
			// 兼容部分非 OpenAI 标准兼容接口（如 MiniMax chatcompletion_v2），
			// 这类接口不接受 system role，这里退化为 user。
			normalizedMessages := make([]Message, 0, len(messages))
			for _, m := range messages {
				if strings.EqualFold(m.Role, "system") {
					m.Role = "user"
				}
				normalizedMessages = append(normalizedMessages, m)
			}
			return map[string]interface{}{
				"model":       model,
				"temperature": 0.2,
				"messages":    normalizedMessages,
			}
		})
}
