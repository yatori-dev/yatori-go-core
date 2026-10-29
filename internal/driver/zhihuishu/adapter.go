package zhihuishu

import (
	"context"
	"time"

	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/core"
	api "github.com/yatori-dev/yatori-go-core/internal/api/zhihuishu"
	"github.com/yatori-dev/yatori-go-core/internal/driver/driverutil"
)

func init() {
	core.Register(core.ZhiHuiShu, New)
}

type client struct {
	driverutil.Base
	cache *api.ZhiDaoUserCache
}

// New constructs a ZhiHuiShu client.
func New(cfg core.Config) (core.Client, error) {
	cache := &api.ZhiDaoUserCache{
		Account:  cfg.Account.Username,
		Password: cfg.Account.Password,
	}
	if enabled, proxy := driverutil.ProxyConfig(cfg); enabled {
		cache.IpProxySW = true
		cache.ProxyIP = proxy
	}
	return &client{Base: driverutil.NewBase(cfg), cache: cache}, nil
}

func (c *client) Platform() core.Platform { return core.ZhiHuiShu }

func (c *client) Capabilities() core.Capability { return core.CapQRCodeLogin }

func (c *client) Login(ctx context.Context) error {
	return core.Invalid(core.ZhiHuiShu, "login", "use QRCodeLogin.QRCode and WaitQRCode for this platform")
}

func (c *client) LoginWithCookie(ctx context.Context, cookie string) error {
	return core.Unsupported(core.ZhiHuiShu, "login_with_cookie")
}

func (c *client) Session() core.Session {
	return core.Session{Platform: core.ZhiHuiShu, Account: c.cache.Account}
}

// QRCode returns the login QR image (base64) and the pending login token.
func (c *client) QRCode(ctx context.Context) (string, string, error) {
	image, token, err := c.cache.ZhidaoQrCode()
	if err != nil {
		return "", "", core.Wrap(core.ZhiHuiShu, "qr_code", err)
	}
	return image, token, nil
}

// WaitQRCode polls the QR login state until it is confirmed.
func (c *client) WaitQRCode(ctx context.Context, token string) (core.Session, error) {
	for {
		select {
		case <-ctx.Done():
			return core.Session{}, ctx.Err()
		default:
		}
		check, err := c.cache.ZhidaoQrCheck(token)
		if err != nil {
			return core.Session{}, core.Wrap(core.ZhiHuiShu, "qr_check", err)
		}
		status := gojsonq.New().JSONString(check).Find("status")
		if status != nil {
			if value, ok := status.(float64); ok && int(value) == 1 {
				oncePassword, _ := gojsonq.New().JSONString(check).Find("oncePassword").(string)
				uuid, _ := gojsonq.New().JSONString(check).Find("uuid").(string)
				return core.Session{
					Platform: core.ZhiHuiShu,
					Account:  c.cache.Account,
					Token:    oncePassword,
					Extra: map[string]string{
						"oncePassword": oncePassword,
						"uuid":         uuid,
					},
				}, nil
			}
		}
		select {
		case <-ctx.Done():
			return core.Session{}, ctx.Err()
		case <-time.After(5 * time.Second):
		}
	}
}

func (c *client) Courses(ctx context.Context) ([]core.Course, error) {
	return nil, core.Unsupported(core.ZhiHuiShu, "courses")
}

func (c *client) Tree(ctx context.Context, course core.Course) ([]core.Node, error) {
	return nil, core.Unsupported(core.ZhiHuiShu, "tree")
}

func (c *client) Study(ctx context.Context, req core.StudyRequest) (*core.StudyResult, error) {
	return nil, core.Unsupported(core.ZhiHuiShu, "study")
}

func (c *client) Works(ctx context.Context, course core.Course) ([]core.Work, error) {
	return nil, core.Unsupported(core.ZhiHuiShu, "works")
}

func (c *client) DoWork(ctx context.Context, req core.WorkRequest) (*core.AnswerResult, error) {
	return nil, core.Unsupported(core.ZhiHuiShu, "work")
}

func (c *client) Exams(ctx context.Context, course core.Course) ([]core.Exam, error) {
	return nil, core.Unsupported(core.ZhiHuiShu, "exams")
}

func (c *client) DoExam(ctx context.Context, req core.ExamRequest) (*core.AnswerResult, error) {
	return nil, core.Unsupported(core.ZhiHuiShu, "exam")
}

func (c *client) SignIn(ctx context.Context, req core.SignInRequest) (*core.SignInResult, error) {
	return nil, core.Unsupported(core.ZhiHuiShu, "sign_in")
}

func (c *client) Call(ctx context.Context, action core.Action, params map[string]any) (any, error) {
	return core.Call(ctx, c, action, params)
}

func (c *client) Close() error { return nil }

var _ core.QRCodeLogin = (*client)(nil)

var _ core.Client = (*client)(nil)
