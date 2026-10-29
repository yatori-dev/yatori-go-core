package cela

import (
	"context"
	"fmt"
	"strings"

	ddddocr "github.com/Changbaiqi/ddddocr-go/utils"
	"github.com/thedevsaddam/gojsonq"
	ort "github.com/yalue/onnxruntime_go"
	"github.com/yatori-dev/yatori-go-core/v3/core"
	api "github.com/yatori-dev/yatori-go-core/v3/internal/api/cela"
	"github.com/yatori-dev/yatori-go-core/v3/internal/driver/driverutil"
	"github.com/yatori-dev/yatori-go-core/v3/utils"
)

func init() {
	core.Register(core.CELA, New)
}

type client struct {
	driverutil.Base
	cache    *api.CelaUserCache
	loggedIn bool
}

// New constructs a CELA client.
func New(cfg core.Config) (core.Client, error) {
	if strings.TrimSpace(cfg.Account.Username) == "" {
		return nil, core.Invalid(core.CELA, "new", "username is required")
	}
	cache := &api.CelaUserCache{
		Account:  cfg.Account.Username,
		Password: cfg.Account.Password,
	}
	if enabled, proxy := driverutil.ProxyConfig(cfg); enabled {
		cache.IpProxySW = true
		cache.ProxyIP = proxy
	}
	return &client{Base: driverutil.NewBase(cfg), cache: cache}, nil
}

func (c *client) Platform() core.Platform { return core.CELA }

func (c *client) Capabilities() core.Capability { return core.CapLogin }

func (c *client) Login(ctx context.Context) error {
	unlock := c.Lock()
	defer unlock()
	driverutil.EnsureOCR()
	for attempt := 0; attempt < 5; attempt++ {
		c.cache.InitLoginDataApi()
		path, err := c.cache.GetCaptchaApi()
		if err != nil {
			return core.Wrap(core.CELA, "login.captcha", err)
		}
		img, err := utils.ReadImg(path)
		if err != nil {
			return core.Wrap(core.CELA, "login.captcha.read", err)
		}
		c.cache.Code = ddddocr.SemiOCRVerification(img, ort.NewShape(1, 25))
		if _, err := c.cache.CheckCaptchaApi(); err != nil {
			continue
		}
		loginResult := c.cache.LoginApi()
		data := gojsonq.New().JSONString(loginResult).Find("data")
		if data == nil {
			continue
		}
		text, ok := data.(string)
		if !ok || text == "" {
			continue
		}
		c.cache.GetLoginAfterData(text)
		c.loggedIn = true
		return nil
	}
	return fmt.Errorf("yatori: cela login failed after captcha retries")
}

func (c *client) LoginWithCookie(ctx context.Context, cookie string) error {
	return core.Unsupported(core.CELA, "login_with_cookie")
}

func (c *client) Session() core.Session {
	return core.Session{
		Platform: core.CELA,
		Account:  c.cache.Account,
		Cookie:   driverutil.CookieHeader(c.cache.Cookies),
		Extra: map[string]string{
			"ticket": c.cache.Ticket,
		},
	}
}

func (c *client) Courses(ctx context.Context) ([]core.Course, error) {
	return nil, core.Unsupported(core.CELA, "courses")
}

func (c *client) Tree(ctx context.Context, course core.Course) ([]core.Node, error) {
	return nil, core.Unsupported(core.CELA, "tree")
}

func (c *client) Study(ctx context.Context, req core.StudyRequest) (*core.StudyResult, error) {
	return nil, core.Unsupported(core.CELA, "study")
}

func (c *client) Works(ctx context.Context, course core.Course) ([]core.Work, error) {
	return nil, core.Unsupported(core.CELA, "works")
}

func (c *client) DoWork(ctx context.Context, req core.WorkRequest) (*core.AnswerResult, error) {
	return nil, core.Unsupported(core.CELA, "work")
}

func (c *client) Exams(ctx context.Context, course core.Course) ([]core.Exam, error) {
	return nil, core.Unsupported(core.CELA, "exams")
}

func (c *client) DoExam(ctx context.Context, req core.ExamRequest) (*core.AnswerResult, error) {
	return nil, core.Unsupported(core.CELA, "exam")
}

func (c *client) SignIn(ctx context.Context, req core.SignInRequest) (*core.SignInResult, error) {
	return nil, core.Unsupported(core.CELA, "sign_in")
}

func (c *client) Call(ctx context.Context, action core.Action, params map[string]any) (any, error) {
	return core.Call(ctx, c, action, params)
}

func (c *client) Close() error { return nil }

var _ core.Client = (*client)(nil)
