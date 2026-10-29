package welearn

import (
	"context"
	"strconv"
	"strings"

	"github.com/yatori-dev/yatori-go-core/v3/core"
	agg "github.com/yatori-dev/yatori-go-core/v3/internal/aggregation/welearn"
	api "github.com/yatori-dev/yatori-go-core/v3/internal/api/welearn"
	"github.com/yatori-dev/yatori-go-core/v3/internal/driver/driverutil"
)

func init() {
	core.Register(core.WeLearn, New)
}

type client struct {
	driverutil.Base
	cache    *api.WeLearnUserCache
	loggedIn bool

	courses  map[string]agg.WeLearnCourse
	chapters map[string]agg.WeLearnChapter
	points   map[string]agg.WeLearnPoint
}

// New constructs a WeLearn client.
func New(cfg core.Config) (core.Client, error) {
	if strings.TrimSpace(cfg.Account.Username) == "" && strings.TrimSpace(cfg.Account.Cookie) == "" {
		return nil, core.Invalid(core.WeLearn, "new", "username or cookie is required")
	}
	cache := &api.WeLearnUserCache{
		Account:  cfg.Account.Username,
		Password: cfg.Account.Password,
	}
	if enabled, proxy := driverutil.ProxyConfig(cfg); enabled {
		cache.IpProxySW = true
		cache.ProxyIP = proxy
	}
	return &client{
		Base:     driverutil.NewBase(cfg),
		cache:    cache,
		courses:  map[string]agg.WeLearnCourse{},
		chapters: map[string]agg.WeLearnChapter{},
		points:   map[string]agg.WeLearnPoint{},
	}, nil
}

func (c *client) Platform() core.Platform { return core.WeLearn }

func (c *client) Capabilities() core.Capability {
	return core.CapLogin | core.CapCookieLogin | core.CapCourses | core.CapTree | core.CapStudy
}

func (c *client) Login(ctx context.Context) error {
	unlock := c.Lock()
	defer unlock()
	if strings.TrimSpace(c.Config.Account.Cookie) != "" {
		agg.WeLearnCookieLoginAction(c.cache, c.Config.Account.Cookie)
		c.loggedIn = true
		return nil
	}
	if err := agg.WeLearnLoginAction(c.cache); err != nil {
		return core.Wrap(core.WeLearn, "login", err)
	}
	c.loggedIn = true
	return nil
}

func (c *client) LoginWithCookie(ctx context.Context, cookie string) error {
	unlock := c.Lock()
	defer unlock()
	agg.WeLearnCookieLoginAction(c.cache, cookie)
	c.loggedIn = true
	return nil
}

func (c *client) Session() core.Session {
	return core.Session{
		Platform: core.WeLearn,
		Account:  c.cache.Account,
		Cookie:   driverutil.CookieHeader(c.cache.Cookies),
	}
}

func (c *client) Courses(ctx context.Context) ([]core.Course, error) {
	unlock := c.Lock()
	defer unlock()
	if !c.loggedIn {
		return nil, core.NotLoggedIn(core.WeLearn, "courses")
	}
	list, err := agg.WeLearnPullCourseListAction(c.cache)
	if err != nil {
		return nil, core.Wrap(core.WeLearn, "courses", err)
	}
	out := make([]core.Course, 0, len(list))
	for _, item := range list {
		c.courses[item.Cid] = item
		status := core.CourseStatusActive
		if item.Per >= 100 {
			status = core.CourseStatusFinished
		}
		out = append(out, core.Course{
			ID:       item.Cid,
			Name:     item.Name,
			Cover:    item.Img,
			Progress: float64(item.Per),
			Status:   status,
			Extra: map[string]string{
				"scid":      item.Scid,
				"cid":       item.Cid,
				"uid":       item.Uid,
				"classId":   item.ClassId,
				"type":      item.Type,
				"taskCount": item.TaskCount,
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
		return nil, core.NotLoggedIn(core.WeLearn, "tree")
	}
	raw, err := c.resolveCourse(course)
	if err != nil {
		return nil, err
	}
	chapters, err := agg.WeLearnPullCourseChapterAction(c.cache, raw)
	if err != nil {
		return nil, core.Wrap(core.WeLearn, "tree.chapters", err)
	}
	out := make([]core.Node, 0, len(chapters))
	for _, chapter := range chapters {
		c.chapters[chapter.Id] = chapter
		chapterNode := core.Node{
			ID:       "chapter:" + chapter.Id,
			CourseID: course.ID,
			Name:     firstNonEmpty(chapter.Name, chapter.Unitname),
			Kind:     core.NodeChapter,
			Extra: map[string]string{
				"chapterId": chapter.Id,
				"unitIdx":   strconv.Itoa(chapter.UnitIdx),
				"visible":   strconv.FormatBool(chapter.Visible),
			},
		}
		points, err := agg.WeLearnPullChapterPointAction(c.cache, raw, chapter)
		if err != nil {
			c.Logger.Debugf("welearn: points for chapter %s failed: %v", chapter.Id, err)
		}
		for _, point := range points {
			c.points[point.Id] = point
			chapterNode.Children = append(chapterNode.Children, core.Node{
				ID:        point.Id,
				ParentID:  chapterNode.ID,
				CourseID:  course.ID,
				Name:      firstNonEmpty(point.Name, point.Location),
				Kind:      core.NodeUnknown,
				Completed: isComplete(point),
				Extra: map[string]string{
					"chapterId":     chapter.Id,
					"unitIdx":       strconv.Itoa(chapter.UnitIdx),
					"location":      point.Location,
					"isComplete":    point.IsComplete,
					"crate":         point.Crate,
					"isVisible":     strconv.FormatBool(point.IsVisible),
					"isLimited":     strconv.FormatBool(point.IsLimited),
					"learnCount":    strconv.Itoa(point.LearnCount),
					"displayAccess": point.DisplayAccessStyle,
				},
				Raw: point,
			})
		}
		c.nodesForChapter(chapterNode)
		out = append(out, chapterNode)
	}
	return out, nil
}

func (c *client) nodesForChapter(node core.Node) {
	for _, child := range node.Children {
		c.points[child.ID] = child.Raw.(agg.WeLearnPoint)
	}
}

func (c *client) Study(ctx context.Context, req core.StudyRequest) (*core.StudyResult, error) {
	unlock := c.Lock()
	defer unlock()
	if !c.loggedIn {
		return nil, core.NotLoggedIn(core.WeLearn, "study")
	}
	point, ok := req.Node.Raw.(agg.WeLearnPoint)
	if !ok {
		if stored, exists := c.points[req.Node.ID]; exists {
			point = stored
		} else {
			return nil, core.Invalid(core.WeLearn, "study", "node raw data is missing; call Tree first")
		}
	}
	course, err := c.resolveCourse(req.Course)
	if err != nil {
		return nil, err
	}
	if req.Action == core.StudyActionStart {
		if _, err := c.cache.StartStudyApi(course.Cid, point.Id, course.Uid, "100", course.ClassId, false, 3, nil); err != nil {
			return nil, core.Wrap(core.WeLearn, "study.start", err)
		}
		return &core.StudyResult{
			CourseID: req.Course.ID,
			NodeID:   req.Node.ID,
			Kind:     req.Node.Kind,
			Raw:      point,
		}, nil
	}
	if req.Action == core.StudyActionFinish || req.Mode == core.StudyModeFast {
		if err := agg.WeLearnCompletePointAction(c.cache, course, point); err != nil {
			return nil, core.Wrap(core.WeLearn, "study.complete", err)
		}
		return &core.StudyResult{
			CourseID:  req.Course.ID,
			NodeID:    req.Node.ID,
			Kind:      req.Node.Kind,
			Submitted: true,
			Progress:  100,
			Completed: true,
		}, nil
	}
	report, progress, _, _, _, err := agg.WeLearnSubmitStudyTimeAction(c.cache, course, point)
	if err != nil {
		return nil, core.Wrap(core.WeLearn, "study", err)
	}
	return &core.StudyResult{
		CourseID:  req.Course.ID,
		NodeID:    req.Node.ID,
		Kind:      req.Node.Kind,
		Submitted: true,
		Progress:  float64(progress),
		Completed: progress >= 100,
		Message:   report,
		Raw:       report,
	}, nil
}

func (c *client) Works(ctx context.Context, course core.Course) ([]core.Work, error) {
	return nil, core.Unsupported(core.WeLearn, "works")
}

func (c *client) DoWork(ctx context.Context, req core.WorkRequest) (*core.AnswerResult, error) {
	return nil, core.Unsupported(core.WeLearn, "work")
}

func (c *client) Exams(ctx context.Context, course core.Course) ([]core.Exam, error) {
	return nil, core.Unsupported(core.WeLearn, "exams")
}

func (c *client) DoExam(ctx context.Context, req core.ExamRequest) (*core.AnswerResult, error) {
	return nil, core.Unsupported(core.WeLearn, "exam")
}

func (c *client) SignIn(ctx context.Context, req core.SignInRequest) (*core.SignInResult, error) {
	return nil, core.Unsupported(core.WeLearn, "sign_in")
}

func (c *client) Call(ctx context.Context, action core.Action, params map[string]any) (any, error) {
	return core.Call(ctx, c, action, params)
}

func (c *client) Close() error { return nil }

func (c *client) resolveCourse(course core.Course) (agg.WeLearnCourse, error) {
	if stored, ok := c.courses[course.ID]; ok {
		return stored, nil
	}
	if raw, ok := course.Raw.(agg.WeLearnCourse); ok {
		return raw, nil
	}
	if course.Extra == nil {
		return agg.WeLearnCourse{}, core.Invalid(core.WeLearn, "course", "course raw data is missing; call Courses first")
	}
	return agg.WeLearnCourse{
		Scid:      course.Extra["scid"],
		Cid:       firstNonEmpty(course.Extra["cid"], course.ID),
		Name:      course.Name,
		Img:       course.Cover,
		Per:       float32(course.Progress),
		Type:      course.Extra["type"],
		TaskCount: course.Extra["taskCount"],
		Uid:       course.Extra["uid"],
		ClassId:   course.Extra["classId"],
	}, nil
}

func isComplete(point agg.WeLearnPoint) bool {
	return point.IsComplete == "completed" || point.IsComplete == "已完成"
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
