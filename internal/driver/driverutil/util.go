// Package driverutil contains small helpers shared by platform adapters.
package driverutil

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	ddddocr "github.com/Changbaiqi/ddddocr-go/utils"

	"github.com/yatori-dev/yatori-go-core/v3/core"
)

// Base holds the state common to every adapter.
type Base struct {
	Config core.Config
	Logger core.Logger
	mu     sync.Mutex
}

// NewBase builds a Base from a resolved config.
func NewBase(cfg core.Config) Base {
	return Base{Config: cfg, Logger: cfg.Options.Logger}
}

// Lock serializes access to the underlying legacy session.
func (b *Base) Lock() func() {
	b.mu.Lock()
	return b.mu.Unlock
}

// Options returns the resolved options.
func (b *Base) Options() core.Options { return b.Config.Options }

// Platform returns the configured platform.
func (b *Base) Platform() core.Platform { return b.Config.Platform }

// Account returns the configured account.
func (b *Base) Account() core.Account { return b.Config.Account }

// ProxyConfig returns the legacy proxy switch and address for cfg.
func ProxyConfig(cfg core.Config) (bool, string) {
	proxy := cfg.Options.Proxy()
	if proxy == nil {
		return false, ""
	}
	return true, proxy.Host
}

// CookieHeader renders cookies as a Cookie request header value.
func CookieHeader(cookies []*http.Cookie) string {
	if len(cookies) == 0 {
		return ""
	}
	parts := make([]string, 0, len(cookies))
	for _, cookie := range cookies {
		if cookie == nil || cookie.Name == "" {
			continue
		}
		parts = append(parts, cookie.Name+"="+cookie.Value)
	}
	return strings.Join(parts, "; ")
}

// Seconds converts an integer second count to a duration.
func Seconds(value int) time.Duration {
	if value <= 0 {
		return 0
	}
	return time.Duration(value) * time.Second
}

// Progress converts a completed/total pair to a 0..100 percentage.
func Progress(done, total int) float64 {
	if total <= 0 {
		return 0
	}
	value := float64(done) / float64(total) * 100
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}

// ExtraString returns a course extra value by key.
func ExtraString(course core.Course, key string) string {
	return course.ExtraValue(key)
}

// NodeExtraString returns a node extra value by key.
func NodeExtraString(node core.Node, key string) string {
	return node.ExtraValue(key)
}

// Int converts a string to an int, returning 0 on failure.
func Int(value string) int {
	out, _ := strconv.Atoi(strings.TrimSpace(value))
	return out
}

// Int64 converts a string to an int64, returning 0 on failure.
func Int64(value string) int64 {
	out, _ := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	return out
}

// Bool converts common string representations to a bool.
func Bool(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}

// SetString writes key/value into a map, creating it when nil.
func SetString(target *map[string]string, key, value string) {
	if *target == nil {
		*target = map[string]string{}
	}
	(*target)[key] = value
}

// SetAny writes key/value into a map, creating it when nil.
func SetAny(target *map[string]any, key string, value any) {
	if *target == nil {
		*target = map[string]any{}
	}
	(*target)[key] = value
}

var ocrOnce sync.Once

// EnsureOCR initializes the bundled OCR engine at most once.
func EnsureOCR() {
	ocrOnce.Do(func() { ddddocr.DDDDOcrCoreInit() })
}
