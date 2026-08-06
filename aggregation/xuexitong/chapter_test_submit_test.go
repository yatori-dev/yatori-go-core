package xuexitong

import (
	"context"
	"errors"
	"net"
	"strings"
	"testing"

	xuexitongapi "github.com/yatori-dev/yatori-go-core/api/xuexitong"
)

func TestWorkNewSubmitAnswerActionPropagatesRequestError(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	proxyAddress := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}

	cache := &xuexitongapi.XueXiTUserCache{
		IpProxySW: true,
		ProxyIP:   proxyAddress,
	}
	_, err = WorkNewSubmitAnswerAction(cache, xuexitongapi.Question{}, true)
	if err == nil {
		t.Fatal("WorkNewSubmitAnswerAction() error = nil, want request error")
	}
	if !strings.Contains(err.Error(), "提交章节测试任务点请求失败") {
		t.Fatalf("WorkNewSubmitAnswerAction() error = %v, want submission context", err)
	}
}

func TestWorkNewSubmitAnswerActionContextResubmitsAfterCaptcha(t *testing.T) {
	submissions := 0
	answer, err := workNewSubmitAnswerAction(context.Background(), &xuexitongapi.XueXiTUserCache{},
		func(context.Context) (string, error) {
			submissions++
			if submissions == 1 {
				return "captcha", errors.New("触发验证码")
			}
			return "success", nil
		},
		func(context.Context, *xuexitongapi.XueXiTUserCache) error { return nil },
	)
	if err != nil {
		t.Fatalf("workNewSubmitAnswerAction() error = %v", err)
	}
	if answer != "success" || submissions != 2 {
		t.Fatalf("workNewSubmitAnswerAction() = %q after %d submissions", answer, submissions)
	}
}

func TestWorkNewSubmitAnswerActionContextPropagatesCaptchaFailure(t *testing.T) {
	wantErr := errors.New("captcha rejected")
	_, err := workNewSubmitAnswerAction(context.Background(), &xuexitongapi.XueXiTUserCache{},
		func(context.Context) (string, error) { return "captcha", errors.New("触发验证码") },
		func(context.Context, *xuexitongapi.XueXiTUserCache) error { return wantErr },
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("workNewSubmitAnswerAction() error = %v, want %v", err, wantErr)
	}
}

func TestWorkNewSubmitAnswerActionContextHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := workNewSubmitAnswerAction(ctx, &xuexitongapi.XueXiTUserCache{},
		func(context.Context) (string, error) { return "captcha", errors.New("触发验证码") },
		passImageCaptcha,
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("workNewSubmitAnswerAction() error = %v, want context canceled", err)
	}
}
