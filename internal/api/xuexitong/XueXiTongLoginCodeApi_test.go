package xuexitong

import (
	"context"
	"errors"
	"testing"
)

func TestRequestLoginCodeRetriesRejectedCodes(t *testing.T) {
	cache := &XueXiTUserCache{}
	providerCalls := 0
	cache.LoginCodeProvider = func(_ context.Context, req LoginCodeChallenge) (string, error) {
		providerCalls++
		if req.Attempt != providerCalls {
			t.Fatalf("第 %d 次回调的 Attempt=%d", providerCalls, req.Attempt)
		}
		if providerCalls == 1 && req.LastError != "" {
			t.Fatalf("第一次回调不应带 LastError: %q", req.LastError)
		}
		if providerCalls > 1 && req.LastError == "" {
			t.Fatalf("第 %d 次回调应带 LastError", providerCalls)
		}
		return "code", nil
	}

	verifyCalls := 0
	err := cache.requestLoginCode(context.Background(), func(attempt int, lastErr string) LoginCodeChallenge {
		return LoginCodeChallenge{Kind: "sms", Prompt: "test", Attempt: attempt, LastError: lastErr}
	}, func(value string) error {
		verifyCalls++
		if verifyCalls < maxLoginCodeAttempts {
			return &loginCodeRejectedError{message: "验证码错误"}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("requestLoginCode() error = %v", err)
	}
	if providerCalls != maxLoginCodeAttempts || verifyCalls != maxLoginCodeAttempts {
		t.Fatalf("provider=%d verify=%d, want %d", providerCalls, verifyCalls, maxLoginCodeAttempts)
	}
}

func TestRequestLoginCodeStopsOnNonRejectionError(t *testing.T) {
	cache := &XueXiTUserCache{}
	providerCalls := 0
	cache.LoginCodeProvider = func(_ context.Context, req LoginCodeChallenge) (string, error) {
		providerCalls++
		return "code", nil
	}

	err := cache.requestLoginCode(context.Background(), func(attempt int, lastErr string) LoginCodeChallenge {
		return LoginCodeChallenge{Kind: "sms", Prompt: "test", Attempt: attempt, LastError: lastErr}
	}, func(value string) error {
		return errors.New("network down")
	})
	if err == nil || providerCalls != 1 {
		t.Fatalf("provider=%d, err=%v", providerCalls, err)
	}
}

func TestRequestLoginCodeRequiresProvider(t *testing.T) {
	cache := &XueXiTUserCache{}
	err := cache.requestLoginCode(context.Background(), func(attempt int, lastErr string) LoginCodeChallenge {
		return LoginCodeChallenge{Kind: "sms"}
	}, func(value string) error { return nil })
	if err == nil {
		t.Fatal("没有配置 LoginCodeProvider 时应返回错误")
	}
}
