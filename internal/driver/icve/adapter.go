package icve

import (
	"context"
	"strconv"
	"strings"

	"github.com/yatori-dev/yatori-go-core/core"
	agg "github.com/yatori-dev/yatori-go-core/internal/aggregation/icve"
	api "github.com/yatori-dev/yatori-go-core/internal/api/icve"
	"github.com/yatori-dev/yatori-go-core/internal/driver/driverutil"
	"github.com/yatori-dev/yatori-go-core/utils"
)

func init() {
	core.Register(core.ICVE, New)
}

type client struct {
	driverutil.Base
	cache    *api.IcveUserCache
	loggedIn bool

	courses map[string]agg.IcveCourse
	nodes   map[string]agg.IcveCourseNode
}

// New constructs an ICVE client.
func New(cfg core.Config) (core.Client, error) {
	if strings.TrimSpace(cfg.Account.Username) == "" && strings.TrimSpace(cfg.Account.Cookie) == "" {
		return nil, core.Invalid(core.ICVE, "new", "username or cookie is required")
	}
	cache := &api.IcveUserCache{
		PreUrl:   cfg.Account.URL,
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
		courses: map[string]agg.IcveCourse{},
		nodes:   map[string]agg.IcveCourseNode{},
	}, nil
}

func (c *client) Platform() core.Platform { return core.ICVE }

func (c *client) Capabilities() core.Capability {
	return core.CapLogin | core.CapCookieLogin | core.CapCourses | core.CapTree | core.CapStudy
}

func (c *client) Login(ctx context.Context) error {
	unlock := c.Lock()
	defer unlock()
	if strings.TrimSpace(c.Config.Account.Cookie) != "" {
		return c.loginWithCookieLocked(c.Config.Account.Cookie)
	}
	driverutil.EnsureOCR()
	if err := utils.ExtractAssets("./assets", "tencentCollect.exe", "tencentPowSolve.exe"); err != nil {
		return core.Wrap(core.ICVE, "login.extract_assets", err)
	}
	if err := agg.IcveLoginAction(c.cache); err != nil {
		return core.Wrap(core.ICVE, "login", err)
	}
	c.loggedIn = true
	return nil
}

func (c *client) LoginWithCookie(ctx context.Context, cookie string) error {
	unlock := c.Lock()
	defer unlock()
	return c.loginWithCookieLocked(cookie)
}

func (c *client) loginWithCookieLocked(cookie string) error {
	c.cache.Password = cookie
	if err := agg.IcveCookieLogin(c.cache); err != nil {
		return core.Wrap(core.ICVE, "login_with_cookie", err)
	}
	c.loggedIn = true
	return nil
}

func (c *client) Session() core.Session {
	return core.Session{
		Platform: core.ICVE,
		Account:  c.cache.Account,
		UserID:   c.cache.UserId,
		Cookie:   driverutil.CookieHeader(c.cache.Cookies),
		Token:    c.cache.Token,
		Extra: map[string]string{
			"accessToken":    c.cache.AccessToken,
			"zykAccessToken": c.cache.ZYKAccessToken,
		},
	}
}

func (c *client) Courses(ctx context.Context) ([]core.Course, error) {
	unlock := c.Lock()
	defer unlock()
	if !c.loggedIn {
		return nil, core.NotLoggedIn(core.ICVE, "courses")
	}
	list, err := agg.PullZYKCourseAction(c.cache)
	if err != nil {
		return nil, core.Wrap(core.ICVE, "courses", err)
	}
	out := make([]core.Course, 0, len(list))
	for _, item := range list {
		id := firstNonEmpty(item.CourseId, item.Id)
		c.courses[id] = item
		status := core.CourseStatusActive
		if item.Status == "3" {
			status = core.CourseStatusFinished
		}
		out = append(out, core.Course{
			ID:      id,
			Name:    item.CourseName,
			Teacher: item.TeacherName,
			Status:  status,
			Extra: map[string]string{
				"id":           item.Id,
				"courseId":     item.CourseId,
				"courseInfoId": item.CourseInfoId,
				"courseType":   item.CourseType,
				"schoolName":   item.SchoolName,
				"weekStr":      item.WeekStr,
				"status":       item.Status,
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
		return nil, core.NotLoggedIn(core.ICVE, "tree")
	}
	raw, err := c.resolveCourse(course)
	if err != nil {
		return nil, err
	}
	list, err := agg.PullZYKCourseNodeAction(c.cache, raw)
	if err != nil {
		return nil, core.Wrap(core.ICVE, "tree", err)
	}
	out := make([]core.Node, 0, len(list))
	for _, item := range list {
		c.nodes[item.Id] = item
		out = append(out, core.Node{
			ID:        item.Id,
			CourseID:  course.ID,
			ParentID:  item.ParentId,
			Name:      item.Name,
			Kind:      nodeKind(item.FileType),
			Duration:  driverutil.Seconds(int(item.TotalNum)),
			Progress:  item.Speed,
			Completed: item.IsLook || item.Speed >= 100,
			Extra: map[string]string{
				"courseId":     item.CourseId,
				"courseInfoId": item.CourseInfoId,
				"parentId":     item.ParentId,
				"fileType":     item.FileType,
				"fileUrl":      item.FileUrl,
				"isLook":       strconv.FormatBool(item.IsLook),
				"speed":        strconv.FormatFloat(item.Speed, 'f', -1, 64),
				"totalNum":     strconv.FormatFloat(item.TotalNum, 'f', -1, 64),
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
		return nil, core.NotLoggedIn(core.ICVE, "study")
	}
	node, ok := req.Node.Raw.(agg.IcveCourseNode)
	if !ok {
		if stored, exists := c.nodes[req.Node.ID]; exists {
			node = stored
		} else {
			return nil, core.Invalid(core.ICVE, "study", "node raw data is missing; call Tree first")
		}
	}
	report, err := agg.SubmitZYKStudyTimeAction(c.cache, node)
	if err != nil {
		return nil, core.Wrap(core.ICVE, "study", err)
	}
	return &core.StudyResult{
		CourseID:  req.Course.ID,
		NodeID:    req.Node.ID,
		Kind:      nodeKind(node.FileType),
		Submitted: true,
		Progress:  node.Speed,
		Completed: true,
		Message:   report,
		Raw:       report,
	}, nil
}

func (c *client) Works(ctx context.Context, course core.Course) ([]core.Work, error) {
	return nil, core.Unsupported(core.ICVE, "works")
}

func (c *client) DoWork(ctx context.Context, req core.WorkRequest) (*core.AnswerResult, error) {
	return nil, core.Unsupported(core.ICVE, "work")
}

func (c *client) Exams(ctx context.Context, course core.Course) ([]core.Exam, error) {
	return nil, core.Unsupported(core.ICVE, "exams")
}

func (c *client) DoExam(ctx context.Context, req core.ExamRequest) (*core.AnswerResult, error) {
	return nil, core.Unsupported(core.ICVE, "exam")
}

func (c *client) SignIn(ctx context.Context, req core.SignInRequest) (*core.SignInResult, error) {
	return nil, core.Unsupported(core.ICVE, "sign_in")
}

func (c *client) Call(ctx context.Context, action core.Action, params map[string]any) (any, error) {
	return core.Call(ctx, c, action, params)
}

func (c *client) Close() error { return nil }

func (c *client) resolveCourse(course core.Course) (agg.IcveCourse, error) {
	if stored, ok := c.courses[course.ID]; ok {
		return stored, nil
	}
	if raw, ok := course.Raw.(agg.IcveCourse); ok {
		return raw, nil
	}
	if course.Extra == nil {
		return agg.IcveCourse{}, core.Invalid(core.ICVE, "course", "course raw data is missing; call Courses first")
	}
	return agg.IcveCourse{
		Id:           firstNonEmpty(course.Extra["id"], course.ID),
		CourseId:     firstNonEmpty(course.Extra["courseId"], course.ID),
		CourseName:   course.Name,
		SchoolName:   course.Extra["schoolName"],
		CourseType:   course.Extra["courseType"],
		CourseInfoId: course.Extra["courseInfoId"],
		WeekStr:      course.Extra["weekStr"],
		Status:       course.Extra["status"],
	}, nil
}

func nodeKind(fileType string) core.NodeKind {
	switch strings.ToLower(fileType) {
	case "mp4":
		return core.NodeVideo
	case "mp3":
		return core.NodeAudio
	case "pdf", "doc", "docx", "ppt", "pptx", "zip":
		return core.NodeDocument
	case "测验":
		return core.NodeExam
	default:
		return core.NodeUnknown
	}
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
