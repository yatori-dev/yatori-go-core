package xuexitong

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultRequestTimeout     = 30 * time.Second
	responseHeaderTimeout     = 15 * time.Second
	tlsHandshakeTimeout       = 10 * time.Second
	idleConnectionTimeout     = 90 * time.Second
	maxIdleConnectionsPerHost = 32
)

var (
	sharedRequestTransport = newSharedTransport(nil)
	sharedRequestClients   = newRequestClients(sharedRequestTransport)
)

type requestClients struct {
	transport  *http.Transport
	normal     *http.Client
	noRedirect *http.Client
}

// RemoteError preserves enough response context to distinguish HTTP failures from business rejections.
type RemoteError struct {
	Operation   string
	StatusCode  int
	ContentType string
	Message     string
	BodySummary string
}

func (e *RemoteError) Error() string {
	parts := []string{e.Operation}
	if e.StatusCode != 0 {
		parts = append(parts, fmt.Sprintf("HTTP %d", e.StatusCode))
	}
	if e.Message != "" {
		parts = append(parts, e.Message)
	}
	if e.BodySummary != "" && e.BodySummary != e.Message {
		parts = append(parts, "响应: "+e.BodySummary)
	}
	return strings.Join(parts, ": ")
}

func newRequestClient(transport http.RoundTripper) *http.Client {
	return &http.Client{Transport: transport, Timeout: defaultRequestTimeout}
}

func newSharedTransport(proxy func(*http.Request) (*url.URL, error)) *http.Transport {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = proxy
	transport.ForceAttemptHTTP2 = true
	transport.ResponseHeaderTimeout = responseHeaderTimeout
	transport.TLSHandshakeTimeout = tlsHandshakeTimeout
	transport.IdleConnTimeout = idleConnectionTimeout
	transport.MaxIdleConnsPerHost = maxIdleConnectionsPerHost
	transport.MaxConnsPerHost = 0
	return transport
}

func newRequestClients(transport *http.Transport) *requestClients {
	return &requestClients{
		transport: transport,
		normal: &http.Client{
			Transport: transport,
			Timeout:   defaultRequestTimeout,
		},
		noRedirect: &http.Client{
			Transport: transport,
			Timeout:   defaultRequestTimeout,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

func (cache *XueXiTUserCache) requestClient() *http.Client {
	return cache.requestClients().normal
}

func (cache *XueXiTUserCache) noRedirectRequestClient() *http.Client {
	return cache.requestClients().noRedirect
}

func (cache *XueXiTUserCache) requestClients() *requestClients {
	if !cache.IpProxySW {
		return sharedRequestClients
	}

	cache.requestClientsMu.Lock()
	defer cache.requestClientsMu.Unlock()
	if cache.cachedRequestClients != nil && cache.cachedProxyIP == cache.ProxyIP {
		return cache.cachedRequestClients
	}
	if cache.cachedRequestClients != nil {
		cache.cachedRequestClients.transport.CloseIdleConnections()
	}
	proxyIP := cache.ProxyIP
	transport := newSharedTransport(func(*http.Request) (*url.URL, error) {
		return url.Parse("http://" + proxyIP)
	})
	cache.cachedProxyIP = proxyIP
	cache.cachedRequestClients = newRequestClients(transport)
	return cache.cachedRequestClients
}

func validateResponse(operation string, response *http.Response, body []byte) error {
	summary := responseSummary(body)
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return &RemoteError{
			Operation:   operation,
			StatusCode:  response.StatusCode,
			ContentType: response.Header.Get("Content-Type"),
			BodySummary: summary,
		}
	}

	message, rejected := rejectedBusinessResponse(body)
	if rejected {
		return &RemoteError{
			Operation:   operation,
			StatusCode:  response.StatusCode,
			ContentType: response.Header.Get("Content-Type"),
			Message:     message,
			BodySummary: summary,
		}
	}
	return nil
}

func rejectedBusinessResponse(body []byte) (string, bool) {
	trimmed := strings.TrimSpace(string(body))
	if start, end := strings.Index(trimmed, "("), strings.LastIndex(trimmed, ")"); start >= 0 && end > start {
		body = []byte(trimmed[start+1 : end])
	}
	var payload map[string]interface{}
	if json.Unmarshal(body, &payload) != nil {
		return "", false
	}

	rejected := false
	for _, key := range []string{"status", "success", "result"} {
		if value, exists := payload[key]; exists {
			if status, ok := value.(bool); ok && !status {
				rejected = true
			}
		}
	}
	if !rejected {
		return "", false
	}

	for _, key := range []string{"msg", "message", "errorMsg", "error"} {
		if message, ok := payload[key].(string); ok && strings.TrimSpace(message) != "" {
			return strings.TrimSpace(message), true
		}
	}
	return "业务请求被拒绝", true
}

func responseSummary(body []byte) string {
	const limit = 512
	summary := strings.Join(strings.Fields(string(body)), " ")
	if len(summary) > limit {
		return summary[:limit] + "..."
	}
	return summary
}

func waitForRetry(ctx context.Context, duration time.Duration) error {
	if duration <= 0 {
		return nil
	}
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
