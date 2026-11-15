package haiqikeji

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	apihqkj "github.com/yatori-dev/yatori-go-core/v3/internal/api/haiqikeji"
)

func TestSubmitHqkjStudyTimeRetriesBusyHeartbeat(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		if n == 1 {
			_, _ = w.Write([]byte(`{"code":500,"msg":"error","data":"心跳过于频繁，请间隔1秒后再试"}`))
			return
		}
		_, _ = w.Write([]byte(`{"code":200,"msg":"ok","data":"ok"}`))
	}))
	defer server.Close()

	oldDelay := hqkjHeartbeatRetryDelay
	hqkjHeartbeatRetryDelay = 10 * time.Millisecond
	defer func() { hqkjHeartbeatRetryDelay = oldDelay }()

	cache := &apihqkj.HqkjUserCache{PreUrl: server.URL, Token: "test-token"}
	result, err := submitHqkjStudyTime(cache, "session-id", 100)
	if err != nil {
		t.Fatalf("submitHqkjStudyTime: %v", err)
	}
	if !strings.Contains(result, `"code":200`) {
		t.Fatalf("unexpected result: %s", result)
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf("requests=%d want 2", got)
	}
}
