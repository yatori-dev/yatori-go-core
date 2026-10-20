package aiq

// AI 请求公共层
//
// 这里统一处理各平台 OpenAI 兼容接口的 HTTP 请求、错误解析、重试和
// 答案解析。各平台实现只负责拼装自己的请求体，避免每个平台各写一套
// 容易出错的解析逻辑。

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	aiRequestTimeout   = 150 * time.Second
	aiMaxResponseBytes = 8 << 20
	aiDefaultMaxTokens = 4096
)

// aiHTTPClient 复用连接。部分中转/代理证书不规范，沿用原来的跳过校验行为。
var aiHTTPClient = &http.Client{
	Transport: &http.Transport{
		// 沿用旧行为：不走环境变量代理，避免用户本机 HTTP_PROXY 影响 AI 请求。
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		MaxIdleConns:        32,
		MaxIdleConnsPerHost: 8,
		IdleConnTimeout:     90 * time.Second,
	},
	Timeout: aiRequestTimeout,
}

type aiHTTPStatusError struct {
	StatusCode int
	Message    string
	Retryable  bool
}

func (e *aiHTTPStatusError) Error() string {
	return fmt.Sprintf("AI接口返回HTTP %d：%s", e.StatusCode, e.Message)
}

func sleepBeforeRetry(attempt int) {
	if attempt < 0 {
		attempt = 0
	}
	d := time.Duration(attempt+1) * 400 * time.Millisecond
	if d > 3*time.Second {
		d = 3 * time.Second
	}
	time.Sleep(d)
}

// doJSONPost 发送 JSON 请求并返回原始响应。
// retryable 表示该错误是否值得重试（网络错误、429、5xx 等）。
func doJSONPost(endpoint, apiKey string, body map[string]interface{}) ([]byte, int, bool, error) {
	return doJSONPostWithHeaders(endpoint, apiKey, body, nil)
}

func doJSONPostWithHeaders(endpoint, apiKey string, body map[string]interface{}, extraHeaders map[string]string) ([]byte, int, bool, error) {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, 0, false, fmt.Errorf("failed to marshal JSON data: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(jsonData))
	if err != nil {
		return nil, 0, false, fmt.Errorf("failed to create HTTP request: %v", err)
	}
	if strings.TrimSpace(apiKey) != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "yatori-go-core/aiq")
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}

	resp, err := aiHTTPClient.Do(req)
	if err != nil {
		return nil, 0, true, fmt.Errorf("failed to execute HTTP request: %v", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, aiMaxResponseBytes))
	if err != nil {
		return nil, resp.StatusCode, true, fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		retryable := resp.StatusCode == http.StatusRequestTimeout ||
			resp.StatusCode == http.StatusConflict ||
			resp.StatusCode == http.StatusTooManyRequests ||
			resp.StatusCode >= http.StatusInternalServerError
		return data, resp.StatusCode, retryable, &aiHTTPStatusError{
			StatusCode: resp.StatusCode,
			Message:    extractErrorMessage(data),
			Retryable:  retryable,
		}
	}

	return data, resp.StatusCode, false, nil
}

func extractErrorMessage(data []byte) string {
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err == nil {
		if msg := apiErrorMessage(m); msg != "" {
			return msg
		}
	}
	return truncateText(string(data), 600)
}

// apiErrorMessage 兼容 OpenAI / 讯飞 / 硅基流动 / 秘塔等常见错误结构。
func apiErrorMessage(m map[string]interface{}) string {
	if e, ok := m["error"]; ok {
		switch v := e.(type) {
		case map[string]interface{}:
			if s, ok := v["message"].(string); ok && strings.TrimSpace(s) != "" {
				return s
			}
		case string:
			if strings.TrimSpace(v) != "" {
				return v
			}
		}
	}

	// 讯飞星火：{"code":10007,"message":"..."}
	if code, ok := numberToInt(m["code"]); ok && code != 0 {
		msg, _ := m["message"].(string)
		if strings.TrimSpace(msg) == "" {
			msg = "unknown error"
		}
		return fmt.Sprintf("code=%d, message=%s", code, msg)
	}

	// 秘塔：{"errCode":401,"errMsg":"..."}
	if code, ok := numberToInt(m["errCode"]); ok && code != 0 {
		msg, _ := m["errMsg"].(string)
		if strings.TrimSpace(msg) == "" {
			msg = "unknown error"
		}
		return fmt.Sprintf("errCode=%d, errMsg=%s", code, msg)
	}

	if success, ok := m["success"].(bool); ok && !success {
		if msg, ok := m["message"].(string); ok && strings.TrimSpace(msg) != "" {
			return msg
		}
	}

	// 只有确认不是正常回复结构时，才把顶层 message 当错误。
	if _, hasChoices := m["choices"]; !hasChoices {
		if _, hasAnswer := m["answer"]; !hasAnswer {
			if _, hasOutput := m["output"]; !hasOutput {
				if msg, ok := m["message"].(string); ok &&
					strings.TrimSpace(msg) != "" &&
					!strings.EqualFold(msg, "success") {
					return msg
				}
			}
		}
	}

	return ""
}

func numberToInt(v interface{}) (int, bool) {
	switch n := v.(type) {
	case float64:
		return int(n), true
	case float32:
		return int(n), true
	case int:
		return n, true
	case int64:
		return int(n), true
	case json.Number:
		i, err := n.Int64()
		if err != nil {
			return 0, false
		}
		return int(i), true
	default:
		return 0, false
	}
}

// contentToText 兼容 content 为字符串或内容块数组的返回结构。
func contentToText(v interface{}) string {
	switch c := v.(type) {
	case nil:
		return ""
	case string:
		return c
	case []interface{}:
		var sb strings.Builder
		for _, part := range c {
			switch p := part.(type) {
			case string:
				sb.WriteString(p)
			case map[string]interface{}:
				if t, ok := p["text"].(string); ok {
					sb.WriteString(t)
					continue
				}
				if t, ok := p["content"].(string); ok {
					sb.WriteString(t)
				}
			}
		}
		return sb.String()
	case map[string]interface{}:
		if t, ok := c["text"].(string); ok {
			return t
		}
		if t, ok := c["content"].(string); ok {
			return t
		}
	}
	return ""
}

// parseChatCompletionResponse 解析标准 chat/completions 返回，同时兼容
// OpenAI Responses、秘塔 answer 等非标准结构。
func parseChatCompletionResponse(data []byte) (string, error) {
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return "", fmt.Errorf("failed to parse JSON response: %v  response body: %s", err, truncateText(string(data), 800))
	}

	if msg := apiErrorMessage(m); msg != "" {
		return "", errors.New(msg)
	}

	if choices, ok := m["choices"].([]interface{}); ok && len(choices) > 0 {
		choice, _ := choices[0].(map[string]interface{})
		if message, ok := choice["message"].(map[string]interface{}); ok {
			if text := strings.TrimSpace(contentToText(message["content"])); text != "" {
				return text, nil
			}

			finish, _ := choice["finish_reason"].(string)
			switch finish {
			case "length":
				return "", errors.New("AI输出因长度限制被截断（finish_reason=length），请增大max_tokens或精简题目内容")
			case "content_filter":
				return "", errors.New("AI输出被内容安全策略拦截（finish_reason=content_filter）")
			}
			if strings.TrimSpace(contentToText(message["reasoning_content"])) != "" {
				return "", errors.New("AI只返回了思考过程，没有返回最终答案（可能达到max_tokens限制），请增大max_tokens")
			}
			return "", errors.New("AI回复内容为空")
		}
		if text := strings.TrimSpace(contentToText(choice["text"])); text != "" {
			return text, nil
		}
	}

	// 秘塔 AI 搜索
	if answer, ok := m["answer"].(string); ok && strings.TrimSpace(answer) != "" {
		return answer, nil
	}

	// OpenAI Responses
	if text, err := parseResponsesOutput(m); err == nil && strings.TrimSpace(text) != "" {
		return text, nil
	}

	return "", fmt.Errorf("AI回复内容未找到，AI返回信息：%s", truncateText(string(data), 800))
}

// parseResponsesOutput 解析 OpenAI Responses API 的 output 结构。
// 原生 HTTP 返回没有 SDK 里的 output_text 便捷字段，需要遍历 output[].content[]。
func parseResponsesOutput(m map[string]interface{}) (string, error) {
	if text, ok := m["output_text"].(string); ok && strings.TrimSpace(text) != "" {
		return text, nil
	}

	output, ok := m["output"].([]interface{})
	if !ok || len(output) == 0 {
		if status, _ := m["status"].(string); status == "incomplete" || status == "failed" {
			return "", fmt.Errorf("OpenAI Responses 状态为 %s，未生成完整答案", status)
		}
		return "", errors.New("OpenAI Responses 中未找到 output 字段")
	}

	var sb strings.Builder
	for _, item := range output {
		im, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		// reasoning 等非 message 项不参与最终答案。
		if typ, _ := im["type"].(string); typ != "" && typ != "message" {
			continue
		}
		if text := contentToText(im["content"]); text != "" {
			sb.WriteString(text)
		}
	}

	if sb.Len() == 0 {
		if status, _ := m["status"].(string); status == "incomplete" || status == "failed" {
			return "", fmt.Errorf("OpenAI Responses 状态为 %s（可能达到 max_output_tokens 限制），未生成完整答案", status)
		}
		return "", errors.New("OpenAI Responses 中没有输出文本")
	}
	return sb.String(), nil
}

// normalizeJSONAnswerContent 尝试从模型返回中提取标准 JSON 数组。
// 兼容 ```json 包裹、<think>...</think> 前置思考、前后解释文本等。
func normalizeJSONAnswerContent(raw string) string {
	s := strings.TrimSpace(strings.TrimPrefix(raw, "\ufeff"))
	if s == "" {
		return s
	}

	// 去掉 markdown 代码块标记
	s = strings.ReplaceAll(s, "```json", "")
	s = strings.ReplaceAll(s, "```JSON", "")
	s = strings.ReplaceAll(s, "```", "")
	s = strings.TrimSpace(s)

	// 去掉 <think ...>...</think>
	for {
		lower := strings.ToLower(s)
		start := strings.Index(lower, "<think")
		if start == -1 {
			break
		}
		openEndRel := strings.Index(lower[start:], ">")
		if openEndRel == -1 {
			s = strings.TrimSpace(s[:start])
			break
		}
		openEnd := start + openEndRel
		closeRel := strings.Index(lower[openEnd:], "</think>")
		if closeRel == -1 {
			s = strings.TrimSpace(s[:start] + " " + s[openEnd+1:])
			break
		}
		closeEnd := openEnd + closeRel + len("</think>")
		s = strings.TrimSpace(s[:start] + " " + s[closeEnd:])
	}

	// 再次清理代码块标记
	s = strings.ReplaceAll(s, "```json", "")
	s = strings.ReplaceAll(s, "```JSON", "")
	s = strings.ReplaceAll(s, "```", "")
	s = strings.TrimSpace(s)

	// 提取首个 JSON 数组
	return extractFirstJSONArray(s)
}

// extractFirstJSONArray 用 json.Decoder 只读取第一个 JSON 值，
// 避免答案后面还有解释文本时把多余内容一起截进来。
func extractFirstJSONArray(s string) string {
	l := strings.Index(s, "[")
	if l == -1 {
		return strings.TrimSpace(s)
	}
	dec := json.NewDecoder(strings.NewReader(s[l:]))
	var raw json.RawMessage
	if err := dec.Decode(&raw); err == nil {
		return strings.TrimSpace(string(raw))
	}
	// 兜底：不是合法 JSON 时仍按首尾括号截取，交给后续校验报错。
	r := strings.LastIndex(s, "]")
	if r > l {
		return strings.TrimSpace(s[l : r+1])
	}
	return strings.TrimSpace(s)
}

// validateJSONAnswer 校验并重新序列化答案数组。
// 返回的字符串保证是合法 JSON 数组，调用方可以直接 Unmarshal。
func validateJSONAnswer(raw string) (string, error) {
	candidate := normalizeJSONAnswerContent(raw)
	var answers []string
	if err := json.Unmarshal([]byte(candidate), &answers); err != nil {
		return "", fmt.Errorf("生成的回复未能正常解析成json: %v", err)
	}
	if answers == nil {
		return "", errors.New("生成的回复不是有效的JSON数组（结果为null）")
	}
	if len(answers) == 0 {
		return "", errors.New("生成的回复是空JSON数组")
	}
	for i := range answers {
		answers[i] = strings.TrimSpace(answers[i])
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(answers); err != nil {
		return "", fmt.Errorf("failed to encode answer json: %v", err)
	}
	return strings.TrimSpace(buf.String()), nil
}

func truncateText(s string, limit int) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "empty response body"
	}
	r := []rune(s)
	if len(r) <= limit {
		return s
	}
	return string(r[:limit]) + "..."
}

// splitSystemMessages 把 system/developer 消息合并成 Responses API 的
// instructions，其余消息作为 input。
func splitSystemMessages(messages []Message) (string, []map[string]string) {
	var instructions []string
	input := make([]map[string]string, 0, len(messages))
	for _, m := range messages {
		if strings.EqualFold(m.Role, "system") || strings.EqualFold(m.Role, "developer") {
			if strings.TrimSpace(m.Content) != "" {
				instructions = append(instructions, m.Content)
			}
			continue
		}
		input = append(input, map[string]string{
			"role":    m.Role,
			"content": m.Content,
		})
	}
	if len(input) == 0 {
		input = append(input, map[string]string{"role": "user", "content": "请开始作答。"})
	}
	return strings.Join(instructions, "\n\n"), input
}

// isRetryableProviderError 判断 HTTP 200 但业务错误码表示可重试的情况。
func isRetryableProviderError(msg string) bool {
	lower := strings.ToLower(msg)
	markers := []string{
		"appidqpsoverflow", "10007", "11203", "rate limit", "ratelimit",
		"too many requests", "overloaded", "busy", "temporarily", "timeout",
		"system resource", "insufficient_system_resource",
	}
	for _, marker := range markers {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

type chatBodyBuilder func(model string, messages []Message) map[string]interface{}

// runChatCompletion 统一处理：请求、重试、JSON 校验、格式纠错重答。
func runChatCompletion(
	providerName,
	endpoint,
	apiKey,
	model string,
	chat AIChatMessages,
	retryNum int,
	lastErr error,
	requireAPIKey bool,
	build chatBodyBuilder,
) (string, error) {
	if retryNum < 0 {
		return "", lastErr
	}
	if requireAPIKey && strings.TrimSpace(apiKey) == "" {
		return "", errors.New("无效apiKey，请检查apiKey是否正确填写，详细参考对应AI平台的API文档")
	}
	messages := append([]Message(nil), chat.Messages...)
	if len(messages) == 0 {
		return "", errors.New("AI请求消息为空")
	}

	var last error = lastErr
	jsonFailures := 0

	for attempt := 0; attempt <= retryNum; attempt++ {
		content, retryable, err := postChatCompletion(endpoint, apiKey, build(model, messages))
		if err != nil {
			last = fmt.Errorf("%s API调用失败：%w", providerName, err)
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
		last = fmt.Errorf("%s 生成的回复未能正常解析成json：%w", providerName, verr)
		if jsonFailures >= 2 {
			return "", last
		}

		// 让模型看到自己的错误回复并重新生成。
		messages = append(messages,
			Message{Role: "assistant", Content: content},
			Message{Role: "user", Content: "你刚才生成的回复未严格遵循json格式，我未能正常解析，请你重新生成。"},
		)
		sleepBeforeRetry(attempt)
	}

	if last == nil {
		last = errors.New("AI调用失败")
	}
	return "", last
}

func postChatCompletion(endpoint, apiKey string, body map[string]interface{}) (string, bool, error) {
	data, _, retryable, err := doJSONPost(endpoint, apiKey, body)
	if err != nil {
		return "", retryable, err
	}
	content, perr := parseChatCompletionResponse(data)
	if perr != nil {
		return "", isRetryableProviderError(perr.Error()), perr
	}
	return content, false, nil
}
