package ketangx

import (
	"context"
	"strconv"
	"strings"

	"github.com/yatori-dev/yatori-go-core/v3/core"
	agg "github.com/yatori-dev/yatori-go-core/v3/internal/aggregation/ketangx"
	api "github.com/yatori-dev/yatori-go-core/v3/internal/api/ketangx"
	"github.com/yatori-dev/yatori-go-core/v3/internal/driver/driverutil"
)

func init() {
	core.Register(core.Ketangx, New)
}

type client struct {
	driverutil.Base
	cache    *api.KetangxUserCache
	loggedIn bool

	courses map[string]agg.KetangxCourse
	nodes   map[string]agg.KetangxNode
}

// New constructs a Ketangx client.
func New(cfg core.Config) (core.Client, error) {
	if strings.TrimSpace(cfg.Account.Username) == "" {
		return nil, core.Invalid(core.Ketangx, "new", "username is required")
	}
	cache := &api.KetangxUserCache{
		Account:  cfg.Account.Username,
		Password: cfg.Account.Password,
	}
	if enabled, proxy := driverutil.ProxyConfig(cfg); enabled {
		cache.IpProxySW = true
		cache.ProxyIP = proxy
	}
	return &client{
		Base:    driverutil.NewBase(cfg),
		cache:   cache,
		courses: map[string]agg.KetangxCourse{},
		nodes:   map[string]agg.KetangxNode{},
	}, nil
}

func (c *client) Platform() core.Platform { return core.Ketangx }

func (c *client) Capabilities() core.Capability {
	return core.CapLogin | core.CapCourses | core.CapTree | core.CapStudy
}

func (c *client) Login(ctx context.Context) error {
	unlock := c.Lock()
	defer unlock()
	if err := agg.LoginAction(c.cache); err != nil {
		return core.Wrap(core.Ketangx, "login", err)
	}
	c.loggedIn = true
	return nil
}

func (c *client) LoginWithCookie(ctx context.Context, cookie string) error {
	return core.Unsupported(core.Ketangx, "login_with_cookie")
}

func (c *client) Session() core.Session {
	return core.Session{
		Platform: core.Ketangx,
		Account:  c.cache.Account,
		UserID:   c.cache.UserId,
		Cookie:   driverutil.CookieHeader(c.cache.Cookies),
		Extra: map[string]string{
			"userName": c.cache.UserName,
			"userUnit": c.cache.UserUnit,
			"id":       c.cache.Id,
		},
	}
}

func (c *client) Courses(ctx context.Context) ([]core.Course, error) {
	unlock := c.Lock()
	defer unlock()
	if !c.loggedIn {
		return nil, core.NotLoggedIn(core.Ketangx, "courses")
	}
	list := agg.PullCourseListAction(c.cache)
	out := make([]core.Course, 0, len(list))
	for _, item := range list {
		c.courses[item.ActivityId] = item
		status := core.CourseStatusActive
		if item.Progress >= 100 {
			status = core.CourseStatusFinished
		}
		out = append(out, core.Course{
			ID:       item.ActivityId,
			Name:     item.Title,
			Progress: item.Progress,
			Status:   status,
			Extra: map[string]string{
				"activityId": item.ActivityId,
				"title":      item.Title,
			},
			Raw: item,
		})
	}
	return out, nil
}

func (c *client) Tree(ctx context.Context, course core.Course) ([]core.Node, error) {
	unlock := c.Lock()
	defer unlock()
	if !c.loggedIn {
		return nil, core.NotLoggedIn(core.Ketangx, "tree")
	}
	raw, err := c.resolveCourse(course)
	if err != nil {
		return nil, err
	}
	list := agg.PullNodeListAction(c.cache, &raw)
	out := make([]core.Node, 0, len(list))
	for _, item := range list {
		c.nodes[item.SectId] = item
		out = append(out, core.Node{
			ID:        item.SectId,
			CourseID:  course.ID,
			Name:      item.Title,
			Kind:      nodeKind(item.Type),
			Completed: item.IsComplete,
			Extra: map[string]string{
				"activityId": course.ID,
				"sectId":     item.SectId,
				"type":       item.Type,
				"enterNum":   item.EnterNum,
				"isComplete": strconv.FormatBool(item.IsComplete),
			},
			Raw: item,
		})
	}
	return out, nil
}

func (c *client) Study(ctx context.Context, req core.StudyRequest) (*core.StudyResult, error) {
	unlock := c.Lock()
	defer unlock()
	if !c.loggedIn {
		return nil, core.NotLoggedIn(core.Ketangx, "study")
	}
	node, ok := req.Node.Raw.(agg.KetangxNode)
	if !ok {
		if stored, exists := c.nodes[req.Node.ID]; exists {
			node = stored
		} else {
			return nil, core.Invalid(core.Ketangx, "study", "node raw data is missing; call Tree first")
		}
	}
	report, err := agg.CompleteVideoAction(c.cache, &node)
	if err != nil {
		return nil, core.Wrap(core.Ketangx, "study", err)
	}
	return &core.StudyResult{
		CourseID:  req.Course.ID,
		NodeID:    req.Node.ID,
		Kind:      nodeKind(node.Type),
		Submitted: true,
		Progress:  100,
		Completed: true,
		Message:   report,
		Raw:       report,
	}, nil
}

func (c *client) Works(ctx context.Context, course core.Course) ([]core.Work, error) {
	return nil, core.Unsupported(core.Ketangx, "works")
}

func (c *client) DoWork(ctx context.Context, req core.WorkRequest) (*core.AnswerResult, error) {
	return nil, core.Unsupported(core.Ketangx, "work")
}

func (c *client) Exams(ctx context.Context, course core.Course) ([]core.Exam, error) {
	return nil, core.Unsupported(core.Ketangx, "exams")
}

func (c *client) DoExam(ctx context.Context, req core.ExamRequest) (*core.AnswerResult, error) {
	return nil, core.Unsupported(core.Ketangx, "exam")
}

func (c *client) SignIn(ctx context.Context, req core.SignInRequest) (*core.SignInResult, error) {
	return nil, core.Unsupported(core.Ketangx, "sign_in")
}

func (c *client) Call(ctx context.Context, action core.Action, params map[string]any) (any, error) {
	return core.Call(ctx, c, action, params)
}

func (c *client) Close() error { return nil }

func (c *client) resolveCourse(course core.Course) (agg.KetangxCourse, error) {
	if stored, ok := c.courses[course.ID]; ok {
		return stored, nil
	}
	if raw, ok := course.Raw.(agg.KetangxCourse); ok {
		return raw, nil
	}
	if course.Extra == nil {
		return agg.KetangxCourse{}, core.Invalid(core.Ketangx, "course", "course raw data is missing; call Courses first")
	}
	return agg.KetangxCourse{
		Title:      course.Name,
		Progress:   course.Progress,
		ActivityId: course.ID,
	}, nil
}

func nodeKind(nodeType string) core.NodeKind {
	switch nodeType {
	case "视频":
		return core.NodeVideo
	case "文档":
		return core.NodeDocument
	default:
		return core.NodeUnknown
	}
}

var _ core.Client = (*client)(nil)
