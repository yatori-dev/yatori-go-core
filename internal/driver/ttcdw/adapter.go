package ttcdw

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/yatori-dev/yatori-go-core/v3/core"
	agg "github.com/yatori-dev/yatori-go-core/v3/internal/aggregation/ttcdw"
	api "github.com/yatori-dev/yatori-go-core/v3/internal/api/ttcdw"
	"github.com/yatori-dev/yatori-go-core/v3/internal/driver/driverutil"
)

func init() {
	core.Register(core.TTCDW, New)
}

type courseRef struct {
	Project   agg.TtcdwProject
	ClassRoom agg.TtcdwClassRoom
	Course    agg.TtcdwCourse
}

type client struct {
	driverutil.Base
	cache    *api.TtcdwUserCache
	loggedIn bool

	courses map[string]courseRef
	videos  map[string]agg.TtcdwVideo
}

// New constructs a TTCDW client.
func New(cfg core.Config) (core.Client, error) {
	if strings.TrimSpace(cfg.Account.Username) == "" && strings.TrimSpace(cfg.Account.Cookie) == "" {
		return nil, core.Invalid(core.TTCDW, "new", "username or cookie is required")
	}
	cache := &api.TtcdwUserCache{
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
		courses: map[string]courseRef{},
		videos:  map[string]agg.TtcdwVideo{},
	}, nil
}

func (c *client) Platform() core.Platform { return core.TTCDW }

func (c *client) Capabilities() core.Capability {
	return core.CapLogin | core.CapCookieLogin | core.CapCourses | core.CapTree | core.CapStudy
}

func (c *client) Login(ctx context.Context) error {
	unlock := c.Lock()
	defer unlock()
	if strings.TrimSpace(c.Config.Account.Cookie) != "" {
		return c.loginWithCookieLocked(c.Config.Account.Cookie)
	}
	if err := agg.TTCDWLoginAction(c.cache); err != nil {
		return core.Wrap(core.TTCDW, "login", err)
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
	if err := agg.TTCDWCookieLoginAction(c.cache); err != nil {
		return core.Wrap(core.TTCDW, "login_with_cookie", err)
	}
	c.loggedIn = true
	return nil
}

func (c *client) Session() core.Session {
	return core.Session{
		Platform: core.TTCDW,
		Account:  c.cache.Account,
		Cookie:   driverutil.CookieHeader(c.cache.Cookies),
	}
}

func (c *client) Courses(ctx context.Context) ([]core.Course, error) {
	unlock := c.Lock()
	defer unlock()
	if !c.loggedIn {
		return nil, core.NotLoggedIn(core.TTCDW, "courses")
	}
	projects, err := agg.PullProjectAction(c.cache)
	if err != nil {
		return nil, core.Wrap(core.TTCDW, "projects", err)
	}
	out := make([]core.Course, 0)
	for _, project := range projects {
		classRooms, err := agg.PullClassRoomAction(c.cache, project)
		if err != nil {
			c.Logger.Debugf("ttcdw: classrooms for project %s failed: %v", project.CourseProjectId, err)
			continue
		}
		for _, classRoom := range classRooms {
			courses, err := agg.PullCourseAction(c.cache, classRoom)
			if err != nil {
				c.Logger.Debugf("ttcdw: courses for item %s failed: %v", classRoom.ItemId, err)
				continue
			}
			for _, course := range courses {
				id := courseKey(project, classRoom, course)
				ref := courseRef{Project: project, ClassRoom: classRoom, Course: course}
				c.courses[id] = ref
				status := core.CourseStatusActive
				if course.Progress >= 100 {
					status = core.CourseStatusFinished
				}
				out = append(out, core.Course{
					ID:       id,
					Name:     course.Name,
					Progress: float64(course.Progress),
					Status:   status,
					Extra: map[string]string{
						"projectId":    project.CourseProjectId,
						"projectName":  project.Name,
						"classId":      project.ClassId,
						"orgId":        project.OrgId,
						"itemId":       classRoom.ItemId,
						"segmentId":    classRoom.SegmentId,
						"classRoom":    classRoom.Name,
						"classRoomTag": classRoom.Title,
						"courseId":     course.CourseId,
						"originalId":   course.OriginalId,
						"shortId":      course.ShortCourseId,
						"duration":     strconv.Itoa(course.Duration),
					},
					Raw: ref,
				})
			}
		}
	}
	return out, nil
}

func (c *client) Tree(ctx context.Context, course core.Course) ([]core.Node, error) {
	unlock := c.Lock()
	defer unlock()
	if !c.loggedIn {
		return nil, core.NotLoggedIn(core.TTCDW, "tree")
	}
	ref, err := c.resolveCourse(course)
	if err != nil {
		return nil, err
	}
	list, err := agg.PullVideoListAction(c.cache, ref.Project, ref.ClassRoom, ref.Course)
	if err != nil {
		return nil, core.Wrap(core.TTCDW, "tree", err)
	}
	out := make([]core.Node, 0, len(list))
	for _, item := range list {
		c.videos[item.VideoId] = item
		out = append(out, core.Node{
			ID:        item.VideoId,
			CourseID:  course.ID,
			Name:      item.Name,
			Kind:      core.NodeVideo,
			Duration:  driverutil.Seconds(item.Duration),
			Progress:  float64(item.Progress),
			Completed: item.Progress >= 100,
			Extra: map[string]string{
				"videoId":       item.VideoId,
				"parentId":      item.ParentId,
				"tag":           item.Tag,
				"duration":      strconv.Itoa(item.Duration),
				"studyProgress": strconv.Itoa(item.StudyProgress),
				"videoType":     strconv.Itoa(item.VideoType),
				"courseware":    strconv.Itoa(item.CourseWareType),
				"projectId":     ref.Project.CourseProjectId,
				"orgId":         ref.Project.OrgId,
				"itemId":        ref.ClassRoom.ItemId,
				"segmentId":     ref.ClassRoom.SegmentId,
				"courseId":      ref.Course.CourseId,
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
		return nil, core.NotLoggedIn(core.TTCDW, "study")
	}
	video, ok := req.Node.Raw.(agg.TtcdwVideo)
	if !ok {
		if stored, exists := c.videos[req.Node.ID]; exists {
			video = stored
		} else {
			return nil, core.Invalid(core.TTCDW, "study", "node raw data is missing; call Tree first")
		}
	}
	ref, err := c.resolveCourse(req.Course)
	if err != nil {
		return nil, err
	}
	position := int(req.Position.Seconds())
	if position <= 0 {
		position = video.StudyProgress
	}
	if req.Action == core.StudyActionFinish || position > video.Duration {
		position = video.Duration
	}
	isFinish := req.Action == core.StudyActionFinish
	report, err := c.cache.StudyTimeSubmitApi(
		ref.Project.OrgId,
		ref.Course.CourseId,
		ref.ClassRoom.ItemId,
		video.VideoId,
		position,
		ref.ClassRoom.SegmentId,
		isFinish,
		"1", "1", "599", "", "0", "",
		3, nil,
	)
	if err != nil {
		return nil, core.Wrap(core.TTCDW, "study", err)
	}
	progress := float64(video.Progress)
	if video.Duration > 0 {
		progress = float64(position) / float64(video.Duration) * 100
	}
	return &core.StudyResult{
		CourseID:  req.Course.ID,
		NodeID:    req.Node.ID,
		Kind:      core.NodeVideo,
		Submitted: true,
		Position:  driverutil.Seconds(position),
		Progress:  progress,
		Completed: isFinish || progress >= 100,
		Message:   report,
		Raw:       report,
	}, nil
}

func (c *client) Works(ctx context.Context, course core.Course) ([]core.Work, error) {
	return nil, core.Unsupported(core.TTCDW, "works")
}

func (c *client) DoWork(ctx context.Context, req core.WorkRequest) (*core.AnswerResult, error) {
	return nil, core.Unsupported(core.TTCDW, "work")
}

func (c *client) Exams(ctx context.Context, course core.Course) ([]core.Exam, error) {
	return nil, core.Unsupported(core.TTCDW, "exams")
}

func (c *client) DoExam(ctx context.Context, req core.ExamRequest) (*core.AnswerResult, error) {
	return nil, core.Unsupported(core.TTCDW, "exam")
}

func (c *client) SignIn(ctx context.Context, req core.SignInRequest) (*core.SignInResult, error) {
	return nil, core.Unsupported(core.TTCDW, "sign_in")
}

func (c *client) Call(ctx context.Context, action core.Action, params map[string]any) (any, error) {
	return core.Call(ctx, c, action, params)
}

func (c *client) Close() error { return nil }

func (c *client) resolveCourse(course core.Course) (courseRef, error) {
	if stored, ok := c.courses[course.ID]; ok {
		return stored, nil
	}
	if raw, ok := course.Raw.(courseRef); ok {
		return raw, nil
	}
	if course.Extra == nil {
		return courseRef{}, core.Invalid(core.TTCDW, "course", "course raw data is missing; call Courses first")
	}
	return courseRef{
		Project: agg.TtcdwProject{
			CourseProjectId: course.Extra["projectId"],
			ClassId:         course.Extra["classId"],
			Name:            course.Extra["projectName"],
			OrgId:           course.Extra["orgId"],
		},
		ClassRoom: agg.TtcdwClassRoom{
			Name:      course.Extra["classRoom"],
			Title:     course.Extra["classRoomTag"],
			ItemId:    course.Extra["itemId"],
			SegmentId: course.Extra["segmentId"],
		},
		Course: agg.TtcdwCourse{
			CourseId:      firstNonEmpty(course.Extra["courseId"], course.ID),
			Name:          course.Name,
			Progress:      float32(course.Progress),
			Duration:      driverutil.Int(course.Extra["duration"]),
			OriginalId:    course.Extra["originalId"],
			ShortCourseId: course.Extra["shortId"],
		},
	}, nil
}

func courseKey(project agg.TtcdwProject, classRoom agg.TtcdwClassRoom, course agg.TtcdwCourse) string {
	return fmt.Sprintf("%s:%s:%s:%s", project.CourseProjectId, classRoom.SegmentId, classRoom.ItemId, course.CourseId)
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
