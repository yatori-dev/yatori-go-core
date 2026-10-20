package aiq

import "testing"

func TestModelCapabilityDetection(t *testing.T) {
	cases := []struct {
		name string
		got  bool
		want bool
	}{
		{"ark seed 2.1", isArkThinkingModel("doubao-seed-2-1-lite-260915"), true},
		{"ark endpoint id", isArkThinkingModel("ep-20261020123456-abcde"), false},
		{"ark non-thinking character", isArkThinkingModel("doubao-seed-character-251128"), false},
		{"ark deepseek v4", isArkThinkingModel("deepseek-v4-1-flash-261017"), true},

		{"silicon deepseek v3", isSiliconThinkingModel("deepseek-ai/DeepSeek-V3"), false},
		{"silicon deepseek v4", isSiliconThinkingModel("deepseek-ai/DeepSeek-V4-Flash"), true},
		{"silicon qwen3", isSiliconThinkingModel("Qwen/Qwen3-8B"), true},

		{"deepseek legacy chat", isDeepSeekV4Model("deepseek-chat"), false},
		{"deepseek flash", isDeepSeekV4Model("deepseek-flash"), true},
		{"deepseek v4 pro", isDeepSeekV4Model("deepseek-v4-pro"), true},

		{"glm 5.3 flash", isGLMThinkingModel("glm-5.3-flash"), true},
		{"glm 4 flash", isGLMThinkingModel("glm-4-flash"), false},
		{"glm 5.3 effort", isGLM53Model("glm-5.3-flash"), true},
		{"glm 5.2 effort", isGLM53Model("glm-5.2"), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("want %v, got %v", tc.want, tc.got)
			}
		})
	}
}
