package enaea

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/yatori-dev/yatori-go-core/core"
	agg "github.com/yatori-dev/yatori-go-core/internal/aggregation/enaea"
	api "github.com/yatori-dev/yatori-go-core/internal/api/enaea"
	"github.com/yatori-dev/yatori-go-core/internal/driver/driverutil"
)

func init() {
	core.Register(core.ENAAE, New)
}

type client struct {
	driverutil.Base
	cache    *api.EnaeaUserCache
	loggedIn bool

	courses map[string]agg.EnaeaCourse
	videos  map[string]agg.EnaeaVideo
}

// New constructs an ENAAE client.
func New(cfg core.Config) (core.Client, error) {
	if strings.TrimSpace(cfg.Account.Username) == "" && strings.TrimSpace(cfg.Account.Cookie) == "" {
		return nil, core.Invalid(core.ENAAE, "new", "username or cookie is required")
	}
	cache := &api.EnaeaUserCache{
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
		courses: map[string]agg.EnaeaCourse{},
		videos:  map[string]agg.EnaeaVideo{},
	}, nil
}

func (c *client) Platform() core.Platform { return core.ENAAE }

func (c *client) Capabilities() core.Capability {
	return core.CapLogin | core.CapCookieLogin | core.CapCourses | core.CapTree |
		core.CapStudy | core.CapExams
}

func (c *client) Login(ctx context.Context) error {
	unlock := c.Lock()
	defer unlock()
	if strings.TrimSpace(c.Config.Account.Cookie) != "" {
		return c.loginWithCookieLocked(c.Config.Account.Cookie)
	}
	if _, err := agg.EnaeaLoginAction(c.cache); err != nil {
		return core.Wrap(core.ENAAE, "login", err)
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
	if err := agg.EnaeaCookieLoginAction(c.cache, cookie); err != nil {
		return core.Wrap(core.ENAAE, "login_with_cookie", err)
	}
	c.loggedIn = true
	return nil
}

func (c *client) Session() core.Session {
	return core.Session{
		Platform: core.ENAAE,
		Account:  c.cache.Account,
		Cookie:   c.cache.Cookie,
		Token:    c.cache.Asuss,
		Extra: map[string]string{
			"asuss": c.cache.Asuss,
		},
	}
}

func (c *client) Courses(ctx context.Context) ([]core.Course, error) {
	unlock := c.Lock()
	defer unlock()
	if !c.loggedIn {
		return nil, core.NotLoggedIn(core.ENAAE, "courses")
	}
	projects, err := agg.ProjectListAction(c.cache)
	if err != nil {
		return nil, core.Wrap(core.ENAAE, "projects", err)
	}
	out := make([]core.Course, 0)
	for _, project := range projects {
		list, err := agg.CourseListAction(c.cache, project.CircleId)
		if err != nil {
			c.Logger.Debugf("enaea: courses for circle %s failed: %v", project.CircleId, err)
			continue
		}
		for _, item := range list {
			id := courseKey(item.CircleId, item.CourseId)
			c.courses[id] = item
			status := core.CourseStatusActive
			if item.StudyProgress >= 100 {
				status = core.CourseStatusFinished
			}
			out = append(out, core.Course{
				ID:       id,
				Name:     item.CourseTitle,
				Progress: float64(item.StudyProgress),
				Status:   status,
				Extra: map[string]string{
					"circleId":          item.CircleId,
					"courseId":          item.CourseId,
					"syllabusId":        item.SyllabusId,
					"titleTag":          item.TitleTag,
					"remark":            item.Remark,
					"courseContentType": item.CourseContentType,
					"externalType":      item.CourseExternalType,
					"externalLink":      item.CourseContentLink,
					"projectName":       project.CircleName,
				},
				Raw: item,
			})
		}
	}
	return out, nil
}

func (c *client) Tree(ctx context.Context, course core.Course) ([]core.Node, error) {
	unlock := c.Lock()
	defer unlock()
	if !c.loggedIn {
		return nil, core.NotLoggedIn(core.ENAAE, "tree")
	}
	raw, err := c.resolveCourse(course)
	if err != nil {
		return nil, err
	}
	list, err := agg.VideoListAction(c.cache, &raw)
	if err != nil {
		return nil, core.Wrap(core.ENAAE, "tree", err)
	}
	out := make([]core.Node, 0, len(list))
	for _, item := range list {
		c.videos[item.Id] = item
		out = append(out, core.Node{
			ID:        item.Id,
			CourseID:  course.ID,
			Name:      item.CourseContentStr,
			Kind:      core.NodeVideo,
			Duration:  driverutil.Seconds(item.VideoLength),
			Progress:  float64(item.StudyProgress),
			Completed: item.StudyProgress >= 100,
			Extra: map[string]string{
				"courseId":     item.CourseId,
				"circleId":     item.CircleId,
				"tccId":        item.TccId,
				"fileName":     item.FileName,
				"courseName":   item.CourseName,
				"videoLength":  strconv.Itoa(item.VideoLength),
				"studyPercent": fmt.Sprintf("%.2f", item.StudyProgress),
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
		return nil, core.NotLoggedIn(core.ENAAE, "study")
	}
	video, ok := req.Node.Raw.(agg.EnaeaVideo)
	if !ok {
		if stored, exists := c.videos[req.Node.ID]; exists {
			video = stored
		} else {
			return nil, core.Invalid(core.ENAAE, "study", "node raw data is missing; call Tree first")
		}
	}
	if err := agg.StatisticTicForCCVideAction(c.cache, &video); err != nil {
		return nil, core.Wrap(core.ENAAE, "study.statistic", err)
	}
	c.videos[video.Id] = video
	if req.Action == core.StudyActionStart {
		return &core.StudyResult{
			CourseID: req.Course.ID,
			NodeID:   req.Node.ID,
			Kind:     core.NodeVideo,
			Progress: float64(video.StudyProgress),
			Raw:      video,
		}, nil
	}
	model := int64(0)
	value := time.Now().UnixMilli()
	if req.Mode == core.StudyModeFast {
		model = 1
		minutes := int64(req.Position.Minutes())
		if minutes <= 0 {
			minutes = int64(video.VideoLength / 60)
		}
		if minutes <= 0 {
			minutes = 1
		}
		value = minutes
	}
	if err := agg.SubmitStudyTimeAction(c.cache, &video, value, model); err != nil {
		return nil, core.Wrap(core.ENAAE, "study.submit", err)
	}
	c.videos[video.Id] = video
	return &core.StudyResult{
		CourseID:  req.Course.ID,
		NodeID:    req.Node.ID,
		Kind:      core.NodeVideo,
		Submitted: true,
		Progress:  float64(video.StudyProgress),
		Completed: video.StudyProgress >= 100,
		Raw:       video,
	}, nil
}

func (c *client) Works(ctx context.Context, course core.Course) ([]core.Work, error) {
	return nil, core.Unsupported(core.ENAAE, "works")
}

func (c *client) DoWork(ctx context.Context, req core.WorkRequest) (*core.AnswerResult, error) {
	return nil, core.Unsupported(core.ENAAE, "work")
}

func (c *client) Exams(ctx context.Context, course core.Course) ([]core.Exam, error) {
	unlock := c.Lock()
	defer unlock()
	if !c.loggedIn {
		return nil, core.NotLoggedIn(core.ENAAE, "exams")
	}
	raw, err := c.resolveCourse(course)
	if err != nil {
		return nil, err
	}
	list, err := agg.ExamListAction(c.cache, raw.CircleId)
	if err != nil {
		return nil, core.Wrap(core.ENAAE, "exams", err)
	}
	out := make([]core.Exam, 0, len(list))
	for _, item := range list {
		out = append(out, core.Exam{
			ID:       firstNonEmpty(item.ExamId, item.ExamTitle),
			CourseID: course.ID,
			Name:     item.ExamTitle,
			Status:   item.Remark,
			StartAt:  item.StartTime,
			EndAt:    item.EndTime,
			Extra: map[string]string{
				"examId":                item.ExamId,
				"circleId":              raw.CircleId,
				"syllabusResourceId":    item.SyllabusResourceId,
				"resourceRandomPaperId": item.ResourceRandomPaperId,
				"totalScore":            fmt.Sprintf("%.0f", item.TotalScore),
				"passScore":             fmt.Sprintf("%.0f", item.PassScore),
			},
			Raw: item,
		})
	}
	return out, nil
}

func (c *client) DoExam(ctx context.Context, req core.ExamRequest) (*core.AnswerResult, error) {
	return nil, core.Unsupported(core.ENAAE, "exam")
}

func (c *client) SignIn(ctx context.Context, req core.SignInRequest) (*core.SignInResult, error) {
	return nil, core.Unsupported(core.ENAAE, "sign_in")
}

func (c *client) Call(ctx context.Context, action core.Action, params map[string]any) (any, error) {
	return core.Call(ctx, c, action, params)
}

func (c *client) Close() error { return nil }

func (c *client) resolveCourse(course core.Course) (agg.EnaeaCourse, error) {
	if stored, ok := c.courses[course.ID]; ok {
		return stored, nil
	}
	if raw, ok := course.Raw.(agg.EnaeaCourse); ok {
		return raw, nil
	}
	if course.Extra == nil {
		return agg.EnaeaCourse{}, core.Invalid(core.ENAAE, "course", "course raw data is missing; call Courses first")
	}
	return agg.EnaeaCourse{
		CircleId:           course.Extra["circleId"],
		CourseId:           course.Extra["courseId"],
		SyllabusId:         course.Extra["syllabusId"],
		TitleTag:           course.Extra["titleTag"],
		Remark:             course.Extra["remark"],
		CourseTitle:        course.Name,
		CourseContentType:  course.Extra["courseContentType"],
		CourseExternalType: course.Extra["externalType"],
		CourseContentLink:  course.Extra["externalLink"],
		StudyProgress:      float32(course.Progress),
	}, nil
}

func courseKey(circleID, courseID string) string {
	return circleID + ":" + courseID
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
