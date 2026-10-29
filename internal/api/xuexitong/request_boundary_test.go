package xuexitong

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestValidateResponsePreservesHTTPFailure(t *testing.T) {
	response := &http.Response{
		StatusCode: http.StatusTooManyRequests,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}
	err := validateResponse("拉取章节状态", response, []byte(`{"msg":"请求过于频繁"}`))
	var remoteErr *RemoteError
	if !errors.As(err, &remoteErr) {
		t.Fatalf("expected RemoteError, got %T", err)
	}
	if remoteErr.StatusCode != http.StatusTooManyRequests || !strings.Contains(err.Error(), "请求过于频繁") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateResponsePreservesBusinessMessage(t *testing.T) {
	response := &http.Response{StatusCode: http.StatusOK, Header: make(http.Header)}
	err := validateResponse("拉取课程", response, []byte(`{"status":false,"msg":"操作频繁，请稍后重试"}`))
	var remoteErr *RemoteError
	if !errors.As(err, &remoteErr) {
		t.Fatalf("expected RemoteError, got %T", err)
	}
	if remoteErr.Message != "操作频繁，请稍后重试" {
		t.Fatalf("unexpected message: %q", remoteErr.Message)
	}
}

func TestValidateResponsePreservesJSONPBusinessMessage(t *testing.T) {
	response := &http.Response{StatusCode: http.StatusOK, Header: make(http.Header)}
	err := validateResponse("提交滑块验证码", response, []byte(`cx_captcha_function({"result":false,"msg":"验证失败"})`))
	var remoteErr *RemoteError
	if !errors.As(err, &remoteErr) || remoteErr.Message != "验证失败" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateResponseAcceptsSuccessfulJSON(t *testing.T) {
	response := &http.Response{StatusCode: http.StatusOK, Header: make(http.Header)}
	if err := validateResponse("拉取课程", response, []byte(`{"status":true,"result":true}`)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
