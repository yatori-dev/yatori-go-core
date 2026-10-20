package aiq

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRunChatCompletionRetryAndJSONCorrection(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		switch calls {
		case 1:
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":{"message":"rate limit"}}`))
		case 2:
			_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"答案是A"},"finish_reason":"stop"}]}`))
		default:
			var body struct {
				Messages []Message `json:"messages"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Errorf("decode request: %v", err)
			}
			if len(body.Messages) != 3 {
				t.Errorf("expected 3 messages after correction, got %d", len(body.Messages))
			}
			if body.Messages[1].Role != "assistant" || body.Messages[2].Role != "user" {
				t.Errorf("unexpected correction roles: %#v", body.Messages)
			}
			_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"[\"A\"]"},"finish_reason":"stop"}]}`))
		}
	}))
	defer srv.Close()

	got, err := runChatCompletion(
		"test", srv.URL, "key", "model",
		AIChatMessages{Messages: []Message{{Role: "user", Content: "题目"}}},
		2, nil, true,
		func(model string, messages []Message) map[string]interface{} {
			return baseChatBody(model, messages, 0.2, 128)
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != `["A"]` {
		t.Fatalf("unexpected answer: %q", got)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestRunChatCompletionDoesNotRetryAuthError(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"message":"invalid api key"}}`))
	}))
	defer srv.Close()

	_, err := runChatCompletion(
		"test", srv.URL, "bad-key", "model",
		AIChatMessages{Messages: []Message{{Role: "user", Content: "题目"}}},
		5, nil, true,
		func(model string, messages []Message) map[string]interface{} {
			return baseChatBody(model, messages, 0.2, 128)
		},
	)
	if err == nil || !strings.Contains(err.Error(), "401") {
		t.Fatalf("expected 401 error, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("auth error should not be retried, got %d calls", calls)
	}
}
