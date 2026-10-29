package gongxue

import (
	"context"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/yatori-dev/yatori-go-core/core"
	agg "github.com/yatori-dev/yatori-go-core/internal/aggregation/gongxue"
	gongxueConfig "github.com/yatori-dev/yatori-go-core/internal/api/gongxue/config"
	gongxueGlobal "github.com/yatori-dev/yatori-go-core/internal/api/gongxue/global"
	"github.com/yatori-dev/yatori-go-core/internal/driver/driverutil"
)

func init() {
	core.Register(core.GongXue, New)
}

type client struct {
	driverutil.Base
	account core.Account
}

// New constructs a GongXue client.
func New(cfg core.Config) (core.Client, error) {
	if strings.TrimSpace(cfg.Account.Username) == "" {
		return nil, core.Invalid(core.GongXue, "new", "username is required")
	}
	if gongxueGlobal.Log == nil {
		gongxueGlobal.Log = logrus.New()
	}
	if gongxueGlobal.Config == nil {
		gongxueGlobal.Config = &gongxueConfig.Config{
			Account: gongxueConfig.Account{Off: true, Promptly: true, PromptlyType: "START"},
		}
	}
	return &client{Base: driverutil.NewBase(cfg), account: cfg.Account}, nil
}

func (c *client) Platform() core.Platform { return core.GongXue }

func (c *client) Capabilities() core.Capability { return core.CapSignIn }

func (c *client) Login(ctx context.Context) error {
	// GongXue has no reusable session; the sign-in call authenticates itself.
	return nil
}

func (c *client) LoginWithCookie(ctx context.Context, cookie string) error {
	return core.Unsupported(core.GongXue, "login_with_cookie")
}

func (c *client) Session() core.Session {
	return core.Session{Platform: core.GongXue, Account: c.account.Username}
}

func (c *client) Courses(ctx context.Context) ([]core.Course, error) {
	return nil, core.Unsupported(core.GongXue, "courses")
}

func (c *client) Tree(ctx context.Context, course core.Course) ([]core.Node, error) {
	return nil, core.Unsupported(core.GongXue, "tree")
}

func (c *client) Study(ctx context.Context, req core.StudyRequest) (*core.StudyResult, error) {
	return nil, core.Unsupported(core.GongXue, "study")
}

func (c *client) Works(ctx context.Context, course core.Course) ([]core.Work, error) {
	return nil, core.Unsupported(core.GongXue, "works")
}

func (c *client) DoWork(ctx context.Context, req core.WorkRequest) (*core.AnswerResult, error) {
	return nil, core.Unsupported(core.GongXue, "work")
}

func (c *client) Exams(ctx context.Context, course core.Course) ([]core.Exam, error) {
	return nil, core.Unsupported(core.GongXue, "exams")
}

func (c *client) DoExam(ctx context.Context, req core.ExamRequest) (*core.AnswerResult, error) {
	return nil, core.Unsupported(core.GongXue, "exam")
}

func (c *client) SignIn(ctx context.Context, req core.SignInRequest) (*core.SignInResult, error) {
	extra := req.Extra
	if extra == nil {
		extra = map[string]any{}
	}
	sign := agg.SignUser{
		Username:  c.account.Username,
		Password:  c.account.Password,
		Country:   stringValue(extra, "country"),
		Province:  stringValue(extra, "province"),
		City:      stringValue(extra, "city"),
		Area:      stringValue(extra, "area"),
		Longitude: stringValue(extra, "longitude"),
		Latitude:  stringValue(extra, "latitude"),
		Address:   firstNonEmpty(stringValue(extra, "address"), req.Location),
		PlayType:  firstNonEmpty(stringValue(extra, "playType"), "START"),
	}
	// The legacy action logs failures and does not return them. Success here
	// means the sign-in flow was executed, not that the server accepted it.
	agg.GongXueSignAction(sign)
	return &core.SignInResult{Success: true, Message: "sign-in submitted"}, nil
}

func (c *client) Call(ctx context.Context, action core.Action, params map[string]any) (any, error) {
	return core.Call(ctx, c, action, params)
}

func (c *client) Close() error { return nil }

func stringValue(values map[string]any, key string) string {
	value, ok := values[key]
	if !ok || value == nil {
		return ""
	}
	if text, ok := value.(string); ok {
		return text
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

var _ core.Client = (*client)(nil)
