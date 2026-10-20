package aiq

import (
	"strings"
	"testing"
)

func TestParseChatCompletionResponse(t *testing.T) {
	cases := []struct {
		name    string
		body    string
		want    string
		wantErr string
	}{
		{
			name: "openai chat completion",
			body: `{"choices":[{"message":{"role":"assistant","content":"[\"B\"]"},"finish_reason":"stop"}]}`,
			want: `["B"]`,
		},
		{
			name:    "xunfei business error",
			body:    `{"code":10007,"message":"AppIdQpsOverFlow"}`,
			wantErr: "code=10007",
		},
		{
			name:    "deepseek thinking truncated",
			body:    `{"choices":[{"message":{"role":"assistant","content":"","reasoning_content":"思考中"},"finish_reason":"length"}]}`,
			wantErr: "截断",
		},
		{
			name: "metaso answer",
			body: `{"answer":"[\"1949年10月1日\"]"}`,
			want: `["1949年10月1日"]`,
		},
		{
			name:    "metaso error",
			body:    `{"errCode":401,"errMsg":"invalid token"}`,
			wantErr: "errCode=401",
		},
		{
			name: "openai responses output",
			body: `{"status":"completed","output":[{"type":"reasoning","content":[]},{"type":"message","content":[{"type":"output_text","text":"[\"正确\"]"}]}]}`,
			want: `["正确"]`,
		},
		{
			name: "content parts",
			body: `{"choices":[{"message":{"role":"assistant","content":[{"type":"text","text":"[\"A\"]"}]},"finish_reason":"stop"}]}`,
			want: `["A"]`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseChatCompletionResponse([]byte(tc.body))
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("want error containing %q, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("want %q, got %q", tc.want, got)
			}
		})
	}
}

func TestNormalizeJSONAnswerContent(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "markdown and think",
			raw:  "```json\n<think>先分析一下</think>\n[\"A\", \"B\"]\n```",
			want: `["A", "B"]`,
		},
		{
			name: "trailing explanation with brackets",
			raw:  `答案是 ["A"]，解析见 [1]`,
			want: `["A"]`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeJSONAnswerContent(tc.raw)
			if got != tc.want {
				t.Fatalf("want %q, got %q", tc.want, got)
			}
		})
	}
}

func TestValidateJSONAnswer(t *testing.T) {
	cases := []struct {
		name    string
		raw     string
		want    string
		wantErr string
	}{
		{name: "valid", raw: "```json\n[\" A \",\"B\"]\n```", want: `["A","B"]`},
		{name: "null", raw: "null", wantErr: "null"},
		{name: "empty array", raw: "[]", wantErr: "空JSON数组"},
		{name: "invalid", raw: "答案是A", wantErr: "未能正常解析"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := validateJSONAnswer(tc.raw)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("want error containing %q, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("want %q, got %q", tc.want, got)
			}
		})
	}
}

func TestSplitSystemMessages(t *testing.T) {
	instructions, input := splitSystemMessages([]Message{
		{Role: "system", Content: "规则一"},
		{Role: "developer", Content: "规则二"},
		{Role: "user", Content: "题目"},
		{Role: "assistant", Content: "上一次回答"},
	})
	if instructions != "规则一\n\n规则二" {
		t.Fatalf("unexpected instructions: %q", instructions)
	}
	if len(input) != 2 || input[0]["role"] != "user" || input[1]["role"] != "assistant" {
		t.Fatalf("unexpected input: %#v", input)
	}
}
