package weiban

import (
	"context"
	"strings"

	"github.com/yatori-dev/yatori-go-core/core"
	agg "github.com/yatori-dev/yatori-go-core/internal/aggregation/weiban"
	api "github.com/yatori-dev/yatori-go-core/internal/api/weiban"
	"github.com/yatori-dev/yatori-go-core/internal/driver/driverutil"
)

func init() {
	core.Register(core.WeiBan, New)
}

type client struct {
	driverutil.Base
	cache    *api.WeiBanCache
	loggedIn bool
}

// New constructs a WeiBan client.
func New(cfg core.Config) (core.Client, error) {
	school := strings.TrimSpace(cfg.Account.School)
	if school == "" {
		school = strings.TrimSpace(cfg.Account.Extra["school"])
	}
	if school == "" {
		return nil, core.Invalid(core.WeiBan, "new", "school is required")
	}
	cache := &api.WeiBanCache{
		School:   school,
		Account:  cfg.Account.Username,
		Password: cfg.Account.Password,
	}
	if enabled, proxy := driverutil.ProxyConfig(cfg); enabled {
		cache.IpProxySW = true
		cache.ProxyIP = proxy
	}
	return &client{Base: driverutil.NewBase(cfg), cache: cache}, nil
}

func (c *client) Platform() core.Platform { return core.WeiBan }

func (c *client) Capabilities() core.Capability { return core.CapLogin }

func (c *client) Login(ctx context.Context) error {
	unlock := c.Lock()
	defer unlock()
	driverutil.EnsureOCR()
	if _, err := agg.WeiBanLoginAction(c.cache); err != nil {
		return core.Wrap(core.WeiBan, "login", err)
	}
	c.loggedIn = true
	return nil
}

func (c *client) LoginWithCookie(ctx context.Context, cookie string) error {
	return core.Unsupported(core.WeiBan, "login_with_cookie")
}

func (c *client) Session() core.Session {
	return core.Session{
		Platform: core.WeiBan,
		Account:  c.cache.Account,
		Cookie:   driverutil.CookieHeader(c.cache.Cookies),
		Extra: map[string]string{
			"school":     c.cache.School,
			"tenantCode": c.cache.TenantCode,
		},
	}
}

func (c *client) Courses(ctx context.Context) ([]core.Course, error) {
	return nil, core.Unsupported(core.WeiBan, "courses")
}

func (c *client) Tree(ctx context.Context, course core.Course) ([]core.Node, error) {
	return nil, core.Unsupported(core.WeiBan, "tree")
}

func (c *client) Study(ctx context.Context, req core.StudyRequest) (*core.StudyResult, error) {
	return nil, core.Unsupported(core.WeiBan, "study")
}

func (c *client) Works(ctx context.Context, course core.Course) ([]core.Work, error) {
	return nil, core.Unsupported(core.WeiBan, "works")
}

func (c *client) DoWork(ctx context.Context, req core.WorkRequest) (*core.AnswerResult, error) {
	return nil, core.Unsupported(core.WeiBan, "work")
}

func (c *client) Exams(ctx context.Context, course core.Course) ([]core.Exam, error) {
	return nil, core.Unsupported(core.WeiBan, "exams")
}

func (c *client) DoExam(ctx context.Context, req core.ExamRequest) (*core.AnswerResult, error) {
	return nil, core.Unsupported(core.WeiBan, "exam")
}

func (c *client) SignIn(ctx context.Context, req core.SignInRequest) (*core.SignInResult, error) {
	return nil, core.Unsupported(core.WeiBan, "sign_in")
}

func (c *client) Call(ctx context.Context, action core.Action, params map[string]any) (any, error) {
	return core.Call(ctx, c, action, params)
}

func (c *client) Close() error { return nil }

var _ core.Client = (*client)(nil)
