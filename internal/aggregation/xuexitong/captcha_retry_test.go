package xuexitong

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestRetryCaptchaStopsAfterSuccess(t *testing.T) {
	attempts := 0
	err := retryCaptcha(context.Background(), "图片验证码", 5, func() (bool, error) {
		attempts++
		return attempts == 3, nil
	})
	if err != nil || attempts != 3 {
		t.Fatalf("attempts=%d, err=%v", attempts, err)
	}
}

func TestRetryCaptchaReturnsLastRejection(t *testing.T) {
	attempts := 0
	err := retryCaptcha(context.Background(), "滑块验证码", 3, func() (bool, error) {
		attempts++
		return false, errors.New("业务拒绝: 风控")
	})
	var retryErr *CaptchaRetryError
	if !errors.As(err, &retryErr) || attempts != 3 || !strings.Contains(err.Error(), "业务拒绝: 风控") {
		t.Fatalf("attempts=%d, err=%v", attempts, err)
	}
}

func TestRetryCaptchaHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	attempts := 0
	err := retryCaptcha(ctx, "图片验证码", 5, func() (bool, error) {
		attempts++
		return false, nil
	})
	if !errors.Is(err, context.Canceled) || attempts != 0 {
		t.Fatalf("attempts=%d, err=%v", attempts, err)
	}
}
