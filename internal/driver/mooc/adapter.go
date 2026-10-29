package mooc

import (
	"context"
	"strings"

	"github.com/yatori-dev/yatori-go-core/core"
	api "github.com/yatori-dev/yatori-go-core/internal/api/mooc"
	"github.com/yatori-dev/yatori-go-core/internal/driver/driverutil"
)

func init() {
	core.Register(core.MOOC, New)
}

type client struct {
	driverutil.Base
	cache    *api.MOOCUserCache
	loggedIn bool
}

// New constructs a MOOC client.
func New(cfg core.Config) (core.Client, error) {
	if strings.TrimSpace(cfg.Account.Username) == "" {
		return nil, core.Invalid(core.MOOC, "new", "username is required")
	}
	cache := &api.MOOCUserCache{
		Account:  cfg.Account.Username,
		Password: cfg.Account.Password,
	}
	if enabled, proxy := driverutil.ProxyConfig(cfg); enabled {
		cache.IpProxySW = true
		cache.ProxyIP = proxy
	}
	return &client{Base: driverutil.NewBase(cfg), cache: cache}, nil
}

func (c *client) Platform() core.Platform { return core.MOOC }

func (c *client) Capabilities() core.Capability { return core.CapLogin }

func (c *client) Login(ctx context.Context) error {
	unlock := c.Lock()
	defer unlock()
	c.cache.InitCookiesApi()
	c.cache.GtApi()
	c.cache.PowGetPApi()
	c.cache.LoginApi()
	c.loggedIn = true
	return nil
}

func (c *client) LoginWithCookie(ctx context.Context, cookie string) error {
	return core.Unsupported(core.MOOC, "login_with_cookie")
}

func (c *client) Session() core.Session {
	return core.Session{
		Platform: core.MOOC,
		Account:  c.cache.Account,
		Cookie:   driverutil.CookieHeader(c.cache.GetCookies()),
	}
}

func (c *client) Courses(ctx context.Context) ([]core.Course, error) {
	return nil, core.Unsupported(core.MOOC, "courses")
}

func (c *client) Tree(ctx context.Context, course core.Course) ([]core.Node, error) {
	return nil, core.Unsupported(core.MOOC, "tree")
}

func (c *client) Study(ctx context.Context, req core.StudyRequest) (*core.StudyResult, error) {
	return nil, core.Unsupported(core.MOOC, "study")
}

func (c *client) Works(ctx context.Context, course core.Course) ([]core.Work, error) {
	return nil, core.Unsupported(core.MOOC, "works")
}

func (c *client) DoWork(ctx context.Context, req core.WorkRequest) (*core.AnswerResult, error) {
	return nil, core.Unsupported(core.MOOC, "work")
}

func (c *client) Exams(ctx context.Context, course core.Course) ([]core.Exam, error) {
	return nil, core.Unsupported(core.MOOC, "exams")
}

func (c *client) DoExam(ctx context.Context, req core.ExamRequest) (*core.AnswerResult, error) {
	return nil, core.Unsupported(core.MOOC, "exam")
}

func (c *client) SignIn(ctx context.Context, req core.SignInRequest) (*core.SignInResult, error) {
	return nil, core.Unsupported(core.MOOC, "sign_in")
}

func (c *client) Call(ctx context.Context, action core.Action, params map[string]any) (any, error) {
	return core.Call(ctx, c, action, params)
}

func (c *client) Close() error { return nil }

var _ core.Client = (*client)(nil)
