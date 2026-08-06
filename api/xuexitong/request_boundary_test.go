package xuexitong

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestRequestClientsShareDirectTransportAndPolicies(t *testing.T) {
	first := (&XueXiTUserCache{}).requestClient()
	secondCache := &XueXiTUserCache{}
	second := secondCache.requestClient()
	noRedirect := secondCache.noRedirectRequestClient()

	if first != second {
		t.Fatal("direct request clients are not shared")
	}
	if first == noRedirect {
		t.Fatal("normal and no-redirect clients share redirect policy")
	}
	if first.Transport != noRedirect.Transport {
		t.Fatal("normal and no-redirect clients do not share transport")
	}
	if first.Timeout != 30*time.Second || noRedirect.Timeout != 30*time.Second {
		t.Fatalf("client timeouts = %v and %v", first.Timeout, noRedirect.Timeout)
	}
	transport, ok := first.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("transport type = %T", first.Transport)
	}
	if transport.ResponseHeaderTimeout != 15*time.Second || transport.TLSHandshakeTimeout != 10*time.Second || transport.IdleConnTimeout != 90*time.Second {
		t.Fatalf("unexpected transport timeouts: %+v", transport)
	}
	if transport.MaxIdleConnsPerHost != 32 || transport.MaxConnsPerHost != 0 || !transport.ForceAttemptHTTP2 {
		t.Fatalf("unexpected transport pool settings: %+v", transport)
	}
	if err := noRedirect.CheckRedirect(nil, nil); !errors.Is(err, http.ErrUseLastResponse) {
		t.Fatalf("no-redirect policy error = %v", err)
	}
}

func TestRequestClientsCacheProxyTransportByConfiguration(t *testing.T) {
	cache := &XueXiTUserCache{IpProxySW: true, ProxyIP: "127.0.0.1:8080"}
	first := cache.requestClient()
	if second := cache.requestClient(); first != second {
		t.Fatal("proxy client was rebuilt without a configuration change")
	}
	if first == sharedRequestClients.normal {
		t.Fatal("proxy request reused direct client")
	}
	proxyURL, err := first.Transport.(*http.Transport).Proxy(&http.Request{})
	if err != nil || proxyURL.String() != "http://127.0.0.1:8080" {
		t.Fatalf("proxy URL = %v, error = %v", proxyURL, err)
	}

	cache.ProxyIP = "127.0.0.1:9090"
	updated := cache.requestClient()
	if updated == first {
		t.Fatal("proxy client was not rebuilt after configuration change")
	}
	proxyURL, err = updated.Transport.(*http.Transport).Proxy(&http.Request{})
	if err != nil || proxyURL.String() != "http://127.0.0.1:9090" {
		t.Fatalf("updated proxy URL = %v, error = %v", proxyURL, err)
	}
}

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
