package core

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Account carries the credentials for one upstream account.
//
// URL, School and Extra are platform specific. Callers may leave fields
// empty when the platform does not need them.
type Account struct {
	Username string
	Password string
	URL      string
	School   string
	Cookie   string
	Extra    map[string]string
}

// Session is a snapshot of the authenticated state of a Client.
//
// Cookie and Token are populated on a best-effort basis. Platforms that
// keep credentials in private fields may return an empty value here; use
// the platform specific Extra map when a token is required.
type Session struct {
	Platform Platform
	Account  string
	UserID   string
	Cookie   string
	Token    string
	Extra    map[string]string
}

// Clone returns a defensive copy of the session.
func (s Session) Clone() Session {
	out := s
	out.Extra = cloneStringMap(s.Extra)
	return out
}

// AIType mirrors models/ctype.AiType without importing the legacy package.
type AIType string

const (
	AIChatGLM  AIType = "CHATGLM"
	AIXingHuo  AIType = "XINGHUO"
	AITongYi   AIType = "TONGYI"
	AIDouBao   AIType = "DOUBAO"
	AIOpenAI   AIType = "OPENAI"
	AIDeepSeek AIType = "DEEPSEEK"
	AIMetaAi   AIType = "METAAI"
	AISilicon  AIType = "SILICON"
	AIOther    AIType = "OTHER"
)

// AIConfig configures the AI answer backend used by work and exam helpers.
type AIConfig struct {
	Type   AIType
	URL    string
	Model  string
	APIKey string
}

// QuestionBankConfig configures an external question-bank service.
type QuestionBankConfig struct {
	URL     string
	Key     string
	Timeout time.Duration
}

// Logger is the minimal logging contract used by the SDK.
//
// Implementations should be safe for concurrent use. The default logger
// discards all messages.
type Logger interface {
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
}

type nopLogger struct{}

func (nopLogger) Debugf(string, ...any) {}
func (nopLogger) Infof(string, ...any)  {}
func (nopLogger) Warnf(string, ...any)  {}
func (nopLogger) Errorf(string, ...any) {}

// NopLogger returns a Logger that discards every message.
func NopLogger() Logger { return nopLogger{} }

// Options controls runtime behaviour shared by all platform adapters.
type Options struct {
	// ProxyURL is an HTTP proxy such as http://127.0.0.1:7890.
	ProxyURL string
	// Timeout bounds a single high level operation. Adapters that already
	// enforce their own timeouts may treat it as a hint.
	Timeout time.Duration
	// WorkDir is reserved for adapters that need a writable asset/cache
	// directory. It defaults to <user-cache-dir>/yatori. Legacy platform
	// actions still write their bundled assets relative to the process CWD.
	// The SDK never changes the process working directory.
	WorkDir string
	// Logger receives SDK level diagnostics.
	Logger Logger
	// AI configures work/exam answering.
	AI *AIConfig
	// QuestionBank configures an external answer source.
	QuestionBank *QuestionBankConfig
}

// Resolve fills in defaults and returns a copy of the options.
func (o Options) Resolve() Options {
	out := o
	if out.Timeout <= 0 {
		out.Timeout = 30 * time.Second
	}
	if strings.TrimSpace(out.WorkDir) == "" {
		if dir, err := os.UserCacheDir(); err == nil && dir != "" {
			out.WorkDir = filepath.Join(dir, "yatori")
		} else {
			out.WorkDir = filepath.Join(".", "yatori-assets")
		}
	}
	if out.Logger == nil {
		out.Logger = NopLogger()
	}
	return out
}

// Proxy returns the parsed proxy URL, or nil when no proxy is configured.
func (o Options) Proxy() *url.URL {
	raw := strings.TrimSpace(o.ProxyURL)
	if raw == "" {
		return nil
	}
	if !strings.Contains(raw, "://") {
		raw = "http://" + raw
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return nil
	}
	return parsed
}

// Config is the complete input to New.
type Config struct {
	Platform Platform
	Account  Account
	Options  Options
}

// Validate checks the mandatory fields. Credential requirements are left to
// the individual adapters because QR based and cookie based logins differ.
func (c Config) Validate() error {
	if !c.Platform.Valid() {
		return fmt.Errorf("yatori: unsupported platform %q", c.Platform)
	}
	if c.Platform == ZhiHuiShu {
		return nil
	}
	if strings.TrimSpace(c.Account.Username) == "" &&
		strings.TrimSpace(c.Account.Cookie) == "" &&
		len(c.Account.Extra) == 0 {
		return fmt.Errorf("yatori: account or cookie is required for %s", c.Platform)
	}
	return nil
}

func cloneStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
