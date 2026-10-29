package qsxt

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/yatori-dev/yatori-go-core/core"
	agg "github.com/yatori-dev/yatori-go-core/internal/aggregation/qingshuxuetang"
	api "github.com/yatori-dev/yatori-go-core/internal/api/qingshuxuetang"
	"github.com/yatori-dev/yatori-go-core/internal/driver/driverutil"
)

func init() {
	core.Register(core.QSXT, New)
}

type client struct {
	driverutil.Base
	cache    *api.QsxtUserCache
	loggedIn bool

	courses  map[string]agg.QsxtCourse
	nodes    map[string]agg.QsxtNode
	works    map[string]*agg.QsxtWork
	startIDs map[string]string
}

// New constructs a QSXT client.
func New(cfg core.Config) (core.Client, error) {
	if strings.TrimSpace(cfg.Account.Username) == "" && strings.TrimSpace(cfg.Account.Cookie) == "" {
		return nil, core.Invalid(core.QSXT, "new", "username or token is required")
	}
	cache := &api.QsxtUserCache{
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
		courses:  map[string]agg.QsxtCourse{},
		nodes:    map[string]agg.QsxtNode{},
		works:    map[string]*agg.QsxtWork{},
		startIDs: map[string]string{},
	}, nil
}

func (c *client) Platform() core.Platform { return core.QSXT }

func (c *client) Capabilities() core.Capability {
	return core.CapLogin | core.CapCookieLogin | core.CapCourses | core.CapTree |
		core.CapStudy | core.CapWorks
}

func (c *client) Login(ctx context.Context) error {
	unlock := c.Lock()
	defer unlock()
	if strings.TrimSpace(c.Config.Account.Cookie) != "" {
		agg.QsxtCookieLoginAction(c.cache)
		c.loggedIn = true
		return nil
	}
	driverutil.EnsureOCR()
	if _, err := agg.QsxtLoginAction(c.cache); err != nil {
		return core.Wrap(core.QSXT, "login", err)
	}
	c.loggedIn = true
	return nil
}

func (c *client) LoginWithCookie(ctx context.Context, cookie string) error {
	unlock := c.Lock()
	defer unlock()
	c.cache.Password = cookie
	agg.QsxtCookieLoginAction(c.cache)
	c.loggedIn = true
	return nil
}

func (c *client) Session() core.Session {
	return core.Session{
		Platform: core.QSXT,
		Account:  c.cache.Account,
		Cookie:   driverutil.CookieHeader(c.cache.Cookies),
		Token:    c.cache.Token,
	}
}

func (c *client) Courses(ctx context.Context) ([]core.Course, error) {
	unlock := c.Lock()
	defer unlock()
	if !c.loggedIn {
		return nil, core.NotLoggedIn(core.QSXT, "courses")
	}
	list, err := agg.PullCourseListAction(c.cache)
	if err != nil {
		return nil, core.Wrap(core.QSXT, "courses", err)
	}
	out := make([]core.Course, 0, len(list))
	for _, item := range list {
		id := courseKey(item)
		c.courses[id] = item
		status := core.CourseStatusActive
		if item.StudyStatus == 2 {
			status = core.CourseStatusFinished
		}
		out = append(out, core.Course{
			ID:       id,
			Name:     item.CourseName,
			Cover:    item.CourseCoverImg,
			Progress: courseProgress(item),
			Status:   status,
			Extra: map[string]string{
				"classId":         item.ClassId,
				"projectName":     item.ProjectName,
				"schoolId":        item.SchoolId,
				"schoolName":      item.SchoolName,
				"semesterId":      item.SemesterId,
				"semesterYear":    item.SemesterYear,
				"semesterName":    item.SemesterName,
				"courseId":        item.CourseId,
				"studyStatus":     strconv.Itoa(item.StudyStatus),
				"studyStatusName": item.StudyStatusName,
				"allowLearn":      strconv.FormatBool(item.AllowLearn),
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
		return nil, core.NotLoggedIn(core.QSXT, "tree")
	}
	raw, err := c.resolveCourse(course)
	if err != nil {
		return nil, err
	}
	list, err := agg.PullCourseNodeListAction(c.cache, raw)
	if err != nil {
		return nil, core.Wrap(core.QSXT, "tree", err)
	}
	byChapter := map[string]*core.Node{}
	order := make([]string, 0)
	for _, item := range list {
		c.nodes[item.NodeId] = item
		key := item.BigId
		if key == "" {
			key = "default"
		}
		chapter, ok := byChapter[key]
		if !ok {
			chapter = &core.Node{
				ID:       "chapter:" + key,
				CourseID: course.ID,
				Name:     firstNonEmpty(item.BigName, "章节"),
				Kind:     core.NodeChapter,
				Extra: map[string]string{
					"bigId":       item.BigId,
					"bigName":     item.BigName,
					"coverImgUrl": item.BigCoverImgUrl,
				},
			}
			byChapter[key] = chapter
			order = append(order, key)
		}
		chapter.Children = append(chapter.Children, core.Node{
			ID:        item.NodeId,
			ParentID:  chapter.ID,
			CourseID:  course.ID,
			Name:      item.NodeName,
			Kind:      nodeKind(item.NodeType),
			Duration:  driverutil.Seconds(item.Duration / 1000),
			Progress:  nodeProgress(item),
			Completed: item.MaxStudyPosition > 0 && item.Duration > 0 && item.MaxStudyPosition >= item.Duration,
			Extra: map[string]string{
				"classId":            item.ClassId,
				"schoolId":           item.SchoolId,
				"courseId":           item.CourseId,
				"semesterId":         item.SemesterId,
				"bigId":              item.BigId,
				"nodeId":             item.NodeId,
				"nodeType":           item.NodeType,
				"duration":           strconv.Itoa(item.Duration),
				"totalStudyDuration": strconv.Itoa(item.TotalStudyDuration),
				"maxStudyPosition":   strconv.Itoa(item.MaxStudyPosition),
			},
			Raw: item,
		})
	}
	out := make([]core.Node, 0, len(order))
	for _, key := range order {
		out = append(out, *byChapter[key])
	}
	return out, nil
}

func (c *client) Study(ctx context.Context, req core.StudyRequest) (*core.StudyResult, error) {
	unlock := c.Lock()
	defer unlock()
	if !c.loggedIn {
		return nil, core.NotLoggedIn(core.QSXT, "study")
	}
	node, ok := req.Node.Raw.(agg.QsxtNode)
	if !ok {
		if stored, exists := c.nodes[req.Node.ID]; exists {
			node = stored
		} else {
			return nil, core.Invalid(core.QSXT, "study", "node raw data is missing; call Tree first")
		}
	}
	if node.NodeType == "chapter" {
		return nil, core.Unsupported(core.QSXT, "study/chapter node")
	}
	startID := c.startIDs[node.NodeId]
	if startID == "" || req.Action == core.StudyActionStart {
		var err error
		startID, err = node.StartStudyTimeAction(c.cache)
		if err != nil {
			return nil, core.Wrap(core.QSXT, "study.start", err)
		}
		c.startIDs[node.NodeId] = startID
	}
	if req.Action == core.StudyActionStart {
		return &core.StudyResult{
			CourseID: req.Course.ID,
			NodeID:   req.Node.ID,
			Kind:     req.Node.Kind,
			Progress: nodeProgress(node),
			Message:  startID,
			Raw:      startID,
		}, nil
	}
	isEnd := req.Action == core.StudyActionFinish || req.Mode == core.StudyModeFast
	report, err := node.SubmitStudyTimeAction(c.cache, startID, isEnd)
	if err != nil {
		return nil, core.Wrap(core.QSXT, "study.submit", err)
	}
	if isEnd {
		delete(c.startIDs, node.NodeId)
	}
	return &core.StudyResult{
		CourseID:  req.Course.ID,
		NodeID:    req.Node.ID,
		Kind:      req.Node.Kind,
		Submitted: true,
		Progress:  nodeProgress(node),
		Completed: isEnd,
		Message:   report,
		Raw:       report,
	}, nil
}

func (c *client) Works(ctx context.Context, course core.Course) ([]core.Work, error) {
	unlock := c.Lock()
	defer unlock()
	if !c.loggedIn {
		return nil, core.NotLoggedIn(core.QSXT, "works")
	}
	raw, err := c.resolveCourse(course)
	if err != nil {
		return nil, err
	}
	list, err := agg.PullWorkListAction(c.cache, raw)
	if err != nil {
		return nil, core.Wrap(core.QSXT, "works", err)
	}
	out := make([]core.Work, 0, len(list))
	for i := range list {
		item := &list[i]
		c.works[item.Id] = item
		status := ""
		switch item.AnswerStatus {
		case 2:
			status = "finished"
		case 0:
			status = "saved"
		default:
			status = "pending"
		}
		out = append(out, core.Work{
			ID:       item.Id,
			CourseID: course.ID,
			Name:     item.Title,
			Status:   status,
			Extra: map[string]string{
				"quizId":       item.Id,
				"type":         strconv.Itoa(item.Type),
				"totalTime":    strconv.Itoa(item.TotalTime),
				"answerStatus": strconv.Itoa(item.AnswerStatus),
				"courseId":     item.CourseId,
				"classId":      item.ClassId,
				"schoolId":     item.SchoolId,
				"passScore":    fmt.Sprintf("%.1f", item.PassScore),
				"finalExam":    strconv.FormatBool(item.FinalExam),
			},
			Raw: item,
		})
	}
	return out, nil
}

func (c *client) DoWork(ctx context.Context, req core.WorkRequest) (*core.AnswerResult, error) {
	unlock := c.Lock()
	defer unlock()
	if !c.loggedIn {
		return nil, core.NotLoggedIn(core.QSXT, "work")
	}
	work, ok := req.Work.Raw.(*agg.QsxtWork)
	if !ok || work == nil {
		work = c.works[req.Work.ID]
	}
	if work == nil {
		return nil, core.Invalid(core.QSXT, "work", "work raw data is missing; call Works first")
	}
	report, err := agg.WriteWorkAction(c.cache, *work, req.Submit)
	if err != nil {
		return nil, core.Wrap(core.QSXT, "work", err)
	}
	return &core.AnswerResult{
		TaskID:    req.Work.ID,
		Title:     work.Title,
		Submitted: req.Submit,
		Message:   report,
		Raw:       report,
	}, nil
}

func (c *client) Exams(ctx context.Context, course core.Course) ([]core.Exam, error) {
	return nil, core.Unsupported(core.QSXT, "exams")
}

func (c *client) DoExam(ctx context.Context, req core.ExamRequest) (*core.AnswerResult, error) {
	return nil, core.Unsupported(core.QSXT, "exam")
}

func (c *client) SignIn(ctx context.Context, req core.SignInRequest) (*core.SignInResult, error) {
	return nil, core.Unsupported(core.QSXT, "sign_in")
}

func (c *client) Call(ctx context.Context, action core.Action, params map[string]any) (any, error) {
	return core.Call(ctx, c, action, params)
}

func (c *client) Close() error { return nil }

func (c *client) resolveCourse(course core.Course) (agg.QsxtCourse, error) {
	if stored, ok := c.courses[course.ID]; ok {
		return stored, nil
	}
	if raw, ok := course.Raw.(agg.QsxtCourse); ok {
		return raw, nil
	}
	if course.Extra == nil {
		return agg.QsxtCourse{}, core.Invalid(core.QSXT, "course", "course raw data is missing; call Courses first")
	}
	return agg.QsxtCourse{
		ClassId:        course.Extra["classId"],
		ProjectName:    course.Extra["projectName"],
		SchoolId:       course.Extra["schoolId"],
		SchoolName:     course.Extra["schoolName"],
		SemesterId:     course.Extra["semesterId"],
		SemesterYear:   course.Extra["semesterYear"],
		SemesterName:   course.Extra["semesterName"],
		CourseId:       firstNonEmpty(course.Extra["courseId"], course.ID),
		CourseName:     course.Name,
		CourseCoverImg: course.Cover,
	}, nil
}

func courseKey(course agg.QsxtCourse) string {
	return course.SemesterId + ":" + course.ClassId + ":" + course.CourseId
}

func courseProgress(course agg.QsxtCourse) float64 {
	if course.CoursewareLearnTotalScore > 0 {
		return course.CoursewareLearnGainScore / course.CoursewareLearnTotalScore * 100
	}
	if course.StudyStatus == 2 {
		return 100
	}
	return 0
}

func nodeProgress(node agg.QsxtNode) float64 {
	if node.Duration > 0 && node.MaxStudyPosition > 0 {
		value := float64(node.MaxStudyPosition) / float64(node.Duration) * 100
		if value > 100 {
			return 100
		}
		return value
	}
	if node.Duration > 0 && node.TotalStudyDuration > 0 {
		value := float64(node.TotalStudyDuration) / float64(node.Duration) * 100
		if value > 100 {
			return 100
		}
		return value
	}
	return 0
}

func nodeKind(nodeType string) core.NodeKind {
	switch strings.ToLower(nodeType) {
	case "video":
		return core.NodeVideo
	case "audio":
		return core.NodeAudio
	case "html", "document", "doc":
		return core.NodeDocument
	case "chapter":
		return core.NodeChapter
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
