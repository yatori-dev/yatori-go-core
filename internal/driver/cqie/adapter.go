package cqie

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/yatori-dev/yatori-go-core/v3/core"
	agg "github.com/yatori-dev/yatori-go-core/v3/internal/aggregation/cqie"
	api "github.com/yatori-dev/yatori-go-core/v3/internal/api/cqie"
	"github.com/yatori-dev/yatori-go-core/v3/internal/driver/driverutil"
)

func init() {
	core.Register(core.CQIE, New)
}

type client struct {
	driverutil.Base
	cache    *api.CqieUserCache
	loggedIn bool

	courses map[string]agg.CqieCourse
	videos  map[string]agg.CqieVideo
}

// New constructs a CQIE client.
func New(cfg core.Config) (core.Client, error) {
	if strings.TrimSpace(cfg.Account.Username) == "" && strings.TrimSpace(cfg.Account.Cookie) == "" {
		return nil, core.Invalid(core.CQIE, "new", "username or token is required")
	}
	cache := &api.CqieUserCache{
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
		courses: map[string]agg.CqieCourse{},
		videos:  map[string]agg.CqieVideo{},
	}, nil
}

func (c *client) Platform() core.Platform { return core.CQIE }

func (c *client) Capabilities() core.Capability {
	return core.CapLogin | core.CapCookieLogin | core.CapCourses | core.CapTree | core.CapStudy
}

func (c *client) Login(ctx context.Context) error {
	unlock := c.Lock()
	defer unlock()
	if strings.TrimSpace(c.Config.Account.Cookie) != "" {
		return c.loginWithTokenLocked(c.Config.Account.Cookie)
	}
	driverutil.EnsureOCR()
	if err := agg.CqieLoginAction(c.cache); err != nil {
		return core.Wrap(core.CQIE, "login", err)
	}
	c.loggedIn = true
	return nil
}

func (c *client) LoginWithCookie(ctx context.Context, cookie string) error {
	unlock := c.Lock()
	defer unlock()
	return c.loginWithTokenLocked(cookie)
}

func (c *client) loginWithTokenLocked(token string) error {
	if err := agg.CqieLoginTokenAction(c.cache, token); err != nil {
		return core.Wrap(core.CQIE, "login_with_token", err)
	}
	c.loggedIn = true
	return nil
}

func (c *client) Session() core.Session {
	return core.Session{
		Platform: core.CQIE,
		Account:  c.cache.Account,
		UserID:   c.cache.GetUserId(),
		Cookie:   c.cache.GetCookie(),
		Token:    c.cache.GetAccess_Token(),
		Extra: map[string]string{
			"token":     c.cache.GetToken(),
			"appId":     c.cache.GetAppId(),
			"studentId": c.cache.GetStudentId(),
			"userName":  c.cache.GetUserName(),
		},
	}
}

func (c *client) Courses(ctx context.Context) ([]core.Course, error) {
	unlock := c.Lock()
	defer unlock()
	if !c.loggedIn {
		return nil, core.NotLoggedIn(core.CQIE, "courses")
	}
	list, err := agg.CqiePullCourseListAction(c.cache)
	if err != nil {
		return nil, core.Wrap(core.CQIE, "courses", err)
	}
	out := make([]core.Course, 0, len(list))
	for _, item := range list {
		c.courses[item.Id] = item
		status := core.CourseStatusActive
		if item.SumUnit > 0 && item.HaveUnit >= item.SumUnit {
			status = core.CourseStatusFinished
		}
		out = append(out, core.Course{
			ID:       item.Id,
			Name:     item.CourseName,
			Progress: driverutil.Progress(item.HaveUnit, item.SumUnit),
			Status:   status,
			Extra: map[string]string{
				"courseId":        item.Id,
				"studentCourseId": item.StudentCourseId,
				"coursewareId":    item.CoursewareId,
				"version":         item.Version,
				"sumUnit":         strconv.Itoa(item.SumUnit),
				"haveUnit":        strconv.Itoa(item.HaveUnit),
				"learned":         item.Learned,
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
		return nil, core.NotLoggedIn(core.CQIE, "tree")
	}
	raw, err := c.resolveCourse(course)
	if err != nil {
		return nil, err
	}
	list, err := agg.PullCourseVideoListAndProgress(c.cache, &raw)
	if err != nil {
		list, err = agg.PullCourseVideoListAction(c.cache, &raw)
		if err != nil {
			return nil, core.Wrap(core.CQIE, "tree", err)
		}
	}
	out := make([]core.Node, 0, len(list))
	for i := range list {
		item := list[i]
		c.videos[item.VideoId] = item
		progress := 0.0
		if item.TimeLength > 0 {
			progress = float64(item.StudyTime) / float64(item.TimeLength) * 100
		}
		out = append(out, core.Node{
			ID:        item.VideoId,
			CourseID:  course.ID,
			Name:      item.VideoName,
			Kind:      core.NodeVideo,
			Duration:  driverutil.Seconds(item.TimeLength),
			Progress:  progress,
			Completed: item.TimeLength > 0 && item.StudyTime >= item.TimeLength,
			Extra: map[string]string{
				"videoId":         item.VideoId,
				"courseId":        item.CourseId,
				"unitId":          item.UnitId,
				"studentCourseId": item.StudentCourseId,
				"coursewareId":    item.CoursewareId,
				"version":         item.Version,
				"studyId":         item.StudyId,
				"maxCurrentPos":   strconv.Itoa(item.MaxCurrentPos),
				"studyTime":       strconv.Itoa(item.StudyTime),
				"timeLength":      strconv.Itoa(item.TimeLength),
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
		return nil, core.NotLoggedIn(core.CQIE, "study")
	}
	video, ok := req.Node.Raw.(agg.CqieVideo)
	if !ok {
		if stored, exists := c.videos[req.Node.ID]; exists {
			video = stored
		} else {
			return nil, core.Invalid(core.CQIE, "study", "node raw data is missing; call Tree first")
		}
	}
	if req.Action == core.StudyActionStart || video.StudyId == "" {
		if err := agg.StartStudyVideoAction(c.cache, &video); err != nil {
			return nil, core.Wrap(core.CQIE, "study.start", err)
		}
	}
	position := int(req.Position.Seconds())
	if position <= 0 {
		position = video.StudyTime
	}
	if req.Action == core.StudyActionStart {
		return &core.StudyResult{
			CourseID: req.Course.ID,
			NodeID:   req.Node.ID,
			Kind:     core.NodeVideo,
			Position: driverutil.Seconds(position),
			Progress: progress(video),
			Raw:      video,
		}, nil
	}
	if req.Action == core.StudyActionFinish || position > video.TimeLength {
		position = video.TimeLength
	}
	start := video.StudyTime
	if start < 0 {
		start = 0
	}
	stop := position
	if stop <= start {
		stop = start + 5
		if stop > video.TimeLength {
			stop = video.TimeLength
		}
	}
	maxPos := video.MaxCurrentPos
	if stop > maxPos {
		maxPos = stop
	}
	if err := agg.SubmitStudyTimeAction(c.cache, &video, time.Now(), start, stop, maxPos); err != nil {
		return nil, core.Wrap(core.CQIE, "study.submit", err)
	}
	if err := agg.SaveVideoStudyTimeAction(c.cache, &video, start, stop); err != nil {
		c.Logger.Debugf("cqie: save study point failed: %v", err)
	}
	video.StudyTime = stop
	if maxPos > video.MaxCurrentPos {
		video.MaxCurrentPos = maxPos
	}
	c.videos[video.VideoId] = video
	return &core.StudyResult{
		CourseID:  req.Course.ID,
		NodeID:    req.Node.ID,
		Kind:      core.NodeVideo,
		Submitted: true,
		Position:  driverutil.Seconds(stop),
		Progress:  progress(video),
		Completed: video.TimeLength > 0 && stop >= video.TimeLength,
	}, nil
}

func (c *client) Works(ctx context.Context, course core.Course) ([]core.Work, error) {
	return nil, core.Unsupported(core.CQIE, "works")
}

func (c *client) DoWork(ctx context.Context, req core.WorkRequest) (*core.AnswerResult, error) {
	return nil, core.Unsupported(core.CQIE, "work")
}

func (c *client) Exams(ctx context.Context, course core.Course) ([]core.Exam, error) {
	return nil, core.Unsupported(core.CQIE, "exams")
}

func (c *client) DoExam(ctx context.Context, req core.ExamRequest) (*core.AnswerResult, error) {
	return nil, core.Unsupported(core.CQIE, "exam")
}

func (c *client) SignIn(ctx context.Context, req core.SignInRequest) (*core.SignInResult, error) {
	return nil, core.Unsupported(core.CQIE, "sign_in")
}

func (c *client) Call(ctx context.Context, action core.Action, params map[string]any) (any, error) {
	return core.Call(ctx, c, action, params)
}

func (c *client) Close() error { return nil }

func (c *client) resolveCourse(course core.Course) (agg.CqieCourse, error) {
	if stored, ok := c.courses[course.ID]; ok {
		return stored, nil
	}
	if raw, ok := course.Raw.(agg.CqieCourse); ok {
		return raw, nil
	}
	if course.Extra == nil {
		return agg.CqieCourse{}, core.Invalid(core.CQIE, "course", "course raw data is missing; call Courses first")
	}
	return agg.CqieCourse{
		Id:              firstNonEmpty(course.Extra["courseId"], course.ID),
		CourseName:      course.Name,
		StudentCourseId: course.Extra["studentCourseId"],
		CoursewareId:    course.Extra["coursewareId"],
		Version:         course.Extra["version"],
		Learned:         course.Extra["learned"],
		SumUnit:         driverutil.Int(course.Extra["sumUnit"]),
		HaveUnit:        driverutil.Int(course.Extra["haveUnit"]),
	}, nil
}

func progress(video agg.CqieVideo) float64 {
	if video.TimeLength <= 0 {
		return 0
	}
	value := float64(video.StudyTime) / float64(video.TimeLength) * 100
	if value > 100 {
		return 100
	}
	return value
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
