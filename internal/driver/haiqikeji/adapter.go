package haiqikeji

import (
	"context"
	"strconv"
	"strings"

	"github.com/yatori-dev/yatori-go-core/config"
	"github.com/yatori-dev/yatori-go-core/core"
	agg "github.com/yatori-dev/yatori-go-core/internal/aggregation/haiqikeji"
	api "github.com/yatori-dev/yatori-go-core/internal/api/haiqikeji"
	"github.com/yatori-dev/yatori-go-core/internal/driver/driverutil"
	"github.com/yatori-dev/yatori-go-core/models/ctype"
)

func init() {
	core.Register(core.HaiQiKeJi, New)
}

type client struct {
	driverutil.Base
	cache    *api.HqkjUserCache
	loggedIn bool

	courses  map[string]agg.HqkjCourse
	nodes    map[string]agg.HqkjNode
	sessions map[string]string
}

// New constructs a HaiQiKeJi client.
func New(cfg core.Config) (core.Client, error) {
	if strings.TrimSpace(cfg.Account.URL) == "" {
		return nil, core.Invalid(core.HaiQiKeJi, "new", "url is required")
	}
	cache := &api.HqkjUserCache{
		PreUrl:   cfg.Account.URL,
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
		courses:  map[string]agg.HqkjCourse{},
		nodes:    map[string]agg.HqkjNode{},
		sessions: map[string]string{},
	}, nil
}

func (c *client) Platform() core.Platform { return core.HaiQiKeJi }

func (c *client) Capabilities() core.Capability {
	return core.CapLogin | core.CapCourses | core.CapTree | core.CapStudy |
		core.CapWorks | core.CapExams
}

func (c *client) Login(ctx context.Context) error {
	unlock := c.Lock()
	defer unlock()
	if err := agg.HqkjLoginAction(c.cache); err != nil {
		return core.Wrap(core.HaiQiKeJi, "login", err)
	}
	c.loggedIn = true
	return nil
}

func (c *client) LoginWithCookie(ctx context.Context, cookie string) error {
	return core.Unsupported(core.HaiQiKeJi, "login_with_cookie")
}

func (c *client) Session() core.Session {
	return core.Session{
		Platform: core.HaiQiKeJi,
		Account:  c.cache.Account,
		UserID:   c.cache.UserId,
		Token:    c.cache.Token,
		Extra: map[string]string{
			"schoolId": c.cache.SchoolId,
			"url":      c.cache.PreUrl,
		},
	}
}

func (c *client) Courses(ctx context.Context) ([]core.Course, error) {
	unlock := c.Lock()
	defer unlock()
	if !c.loggedIn {
		return nil, core.NotLoggedIn(core.HaiQiKeJi, "courses")
	}
	list, err := agg.HqkjCourseListAction(c.cache)
	if err != nil {
		return nil, core.Wrap(core.HaiQiKeJi, "courses", err)
	}
	out := make([]core.Course, 0, len(list))
	for _, item := range list {
		c.courses[item.Id] = item
		out = append(out, core.Course{
			ID:      item.Id,
			Name:    item.Name,
			Teacher: item.LecturerName,
			Cover:   item.Cover,
			Extra: map[string]string{
				"courseId":   item.Id,
				"collegeId":  item.CollegeId,
				"periodName": item.PeriodName,
				"lecturer":   item.LecturerName,
				"credit":     strconv.FormatFloat(item.Credit, 'f', -1, 64),
				"code":       item.Code,
				"clusterId":  strconv.Itoa(item.ClusterId),
				"schoolId":   strconv.Itoa(item.SchoolId),
				"startDate":  item.StartDate.Format("2006-01-02"),
				"endDate":    item.EndDate.Format("2006-01-02"),
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
		return nil, core.NotLoggedIn(core.HaiQiKeJi, "tree")
	}
	raw, err := c.resolveCourse(course)
	if err != nil {
		return nil, err
	}
	list, err := agg.HqkjNodeListAction(c.cache, raw)
	if err != nil {
		return nil, core.Wrap(core.HaiQiKeJi, "tree", err)
	}
	out := make([]core.Node, 0, len(list))
	for _, item := range list {
		c.nodes[item.Id] = item
		out = append(out, core.Node{
			ID:       item.Id,
			CourseID: course.ID,
			Name:     item.Name,
			Kind:     nodeKind(item),
			Duration: driverutil.Seconds(item.VideoDuration),
			Extra: map[string]string{
				"nodeId":        item.Id,
				"courseId":      item.CourseId,
				"chapterId":     item.ChapterId,
				"videoDuration": strconv.Itoa(item.VideoDuration),
				"tabVideo":      strconv.Itoa(item.TabVideo),
				"tabFile":       strconv.Itoa(item.TabFile),
				"tabWork":       strconv.Itoa(item.TabWork),
				"tabExam":       strconv.Itoa(item.TabExam),
				"videoMode":     strconv.Itoa(item.VideoMode),
				"lock":          strconv.Itoa(item.Lock),
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
		return nil, core.NotLoggedIn(core.HaiQiKeJi, "study")
	}
	node, ok := req.Node.Raw.(agg.HqkjNode)
	if !ok {
		if stored, exists := c.nodes[req.Node.ID]; exists {
			node = stored
		} else {
			return nil, core.Invalid(core.HaiQiKeJi, "study", "node raw data is missing; call Tree first")
		}
	}
	if req.Mode == core.StudyModeFast || req.Action == core.StudyActionFinish {
		report, err := agg.HqkjSubmitFastStudyTimeAction(c.cache, node)
		if err != nil {
			return nil, core.Wrap(core.HaiQiKeJi, "study.fast", err)
		}
		return &core.StudyResult{
			CourseID:  req.Course.ID,
			NodeID:    req.Node.ID,
			Kind:      nodeKind(node),
			Submitted: true,
			Progress:  100,
			Completed: true,
			Message:   report,
			Raw:       report,
		}, nil
	}
	sessionID := c.sessions[node.Id]
	if sessionID == "" || req.Action == core.StudyActionStart {
		var err error
		sessionID, err = agg.HqkjStartStudyAction(c.cache, node)
		if err != nil {
			return nil, core.Wrap(core.HaiQiKeJi, "study.start", err)
		}
		c.sessions[node.Id] = sessionID
	}
	if req.Action == core.StudyActionStart {
		return &core.StudyResult{
			CourseID: req.Course.ID,
			NodeID:   req.Node.ID,
			Kind:     nodeKind(node),
			Message:  sessionID,
			Raw:      sessionID,
		}, nil
	}
	progress := 0
	if node.VideoDuration > 0 {
		progress = int(float64(req.Position.Seconds()) / float64(node.VideoDuration) * 100)
	} else {
		progress = int(req.Position.Seconds())
	}
	if req.Action == core.StudyActionFinish {
		progress = 100
	}
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}
	report, err := agg.HqkjSubmitStudyTimeAction(c.cache, node, sessionID, progress)
	if err != nil {
		return nil, core.Wrap(core.HaiQiKeJi, "study.submit", err)
	}
	if req.Action == core.StudyActionFinish {
		if endReport, err := agg.HqkjEndStudyAction(c.cache, sessionID); err == nil {
			report = endReport
		}
		delete(c.sessions, node.Id)
	}
	return &core.StudyResult{
		CourseID:  req.Course.ID,
		NodeID:    req.Node.ID,
		Kind:      nodeKind(node),
		Submitted: true,
		Progress:  float64(progress),
		Completed: req.Action == core.StudyActionFinish,
		Message:   report,
		Raw:       report,
	}, nil
}

func (c *client) Works(ctx context.Context, course core.Course) ([]core.Work, error) {
	unlock := c.Lock()
	defer unlock()
	if !c.loggedIn {
		return nil, core.NotLoggedIn(core.HaiQiKeJi, "works")
	}
	raw, err := c.resolveCourse(course)
	if err != nil {
		return nil, err
	}
	list, err := c.nodeListLocked(raw)
	if err != nil {
		return nil, err
	}
	out := make([]core.Work, 0)
	for _, node := range list {
		if node.TabWork <= 0 {
			continue
		}
		out = append(out, core.Work{
			ID:       node.Id,
			CourseID: course.ID,
			Name:     node.Name,
			Extra: map[string]string{
				"nodeId":   node.Id,
				"courseId": node.CourseId,
				"kind":     "work",
			},
			Raw: node,
		})
	}
	return out, nil
}

func (c *client) DoWork(ctx context.Context, req core.WorkRequest) (*core.AnswerResult, error) {
	unlock := c.Lock()
	defer unlock()
	if !c.loggedIn {
		return nil, core.NotLoggedIn(core.HaiQiKeJi, "work")
	}
	raw, err := c.resolveCourse(req.Course)
	if err != nil {
		return nil, err
	}
	list, err := c.nodeListLocked(raw)
	if err != nil {
		return nil, err
	}
	results, err := c.runAnswers(raw, list, req.Source, "work", req.Submit)
	if err != nil {
		return nil, err
	}
	return &core.AnswerResult{
		TaskID:    req.Work.ID,
		Title:     req.Work.Name,
		Submitted: req.Submit,
		Answers:   nil,
		Message:   summarize(results),
		Raw:       results,
	}, nil
}

func (c *client) Exams(ctx context.Context, course core.Course) ([]core.Exam, error) {
	unlock := c.Lock()
	defer unlock()
	if !c.loggedIn {
		return nil, core.NotLoggedIn(core.HaiQiKeJi, "exams")
	}
	raw, err := c.resolveCourse(course)
	if err != nil {
		return nil, err
	}
	list, err := c.nodeListLocked(raw)
	if err != nil {
		return nil, err
	}
	out := make([]core.Exam, 0)
	for _, node := range list {
		if node.TabExam <= 0 {
			continue
		}
		out = append(out, core.Exam{
			ID:       node.Id,
			CourseID: course.ID,
			Name:     node.Name,
			Extra: map[string]string{
				"nodeId":   node.Id,
				"courseId": node.CourseId,
				"kind":     "exam",
			},
			Raw: node,
		})
	}
	return out, nil
}

func (c *client) DoExam(ctx context.Context, req core.ExamRequest) (*core.AnswerResult, error) {
	unlock := c.Lock()
	defer unlock()
	if !c.loggedIn {
		return nil, core.NotLoggedIn(core.HaiQiKeJi, "exam")
	}
	raw, err := c.resolveCourse(req.Course)
	if err != nil {
		return nil, err
	}
	list, err := c.nodeListLocked(raw)
	if err != nil {
		return nil, err
	}
	results, err := c.runAnswers(raw, list, req.Source, "exam", req.Submit)
	if err != nil {
		return nil, err
	}
	return &core.AnswerResult{
		TaskID:    req.Exam.ID,
		Title:     req.Exam.Name,
		Submitted: req.Submit,
		Message:   summarize(results),
		Raw:       results,
	}, nil
}

func (c *client) SignIn(ctx context.Context, req core.SignInRequest) (*core.SignInResult, error) {
	return nil, core.Unsupported(core.HaiQiKeJi, "sign_in")
}

func (c *client) Call(ctx context.Context, action core.Action, params map[string]any) (any, error) {
	return core.Call(ctx, c, action, params)
}

func (c *client) Close() error { return nil }

func (c *client) nodeListLocked(course agg.HqkjCourse) ([]agg.HqkjNode, error) {
	list, err := agg.HqkjNodeListAction(c.cache, course)
	if err != nil {
		return nil, core.Wrap(core.HaiQiKeJi, "tree", err)
	}
	for _, node := range list {
		c.nodes[node.Id] = node
	}
	return list, nil
}

func (c *client) runAnswers(course agg.HqkjCourse, list []agg.HqkjNode, source core.AnswerSource, kind string, submit bool) ([]agg.HqkjQuizAnswerResult, error) {
	setting := agg.HqkjAnswerSetting{
		AccountType:    "海旗科技",
		Account:        c.cache.Account,
		AutoExam:       1,
		ExamAutoSubmit: boolInt(submit),
	}
	if ai := c.Config.Options.AI; ai != nil {
		setting.AiSetting = config.AiSetting{
			AiType: ctype.AiType(ai.Type),
			AiUrl:  ai.URL,
			Model:  ai.Model,
			APIKEY: ai.APIKey,
		}
	}
	switch source {
	case core.AnswerSourceAI:
		if setting.AiSetting.APIKEY == "" {
			return nil, core.Invalid(core.HaiQiKeJi, "answer", "AI config is required for AnswerSourceAI")
		}
		agg.HqkjAutoAnswerAction(c.cache, course, list, setting)
		return nil, nil
	case core.AnswerSourceBuiltin:
		return agg.HqkjAutoFillBuiltinAnswerAction(c.cache, course, list, setting), nil
	default:
		return nil, core.Unsupported(core.HaiQiKeJi, "answer/source "+string(source))
	}
}

func (c *client) resolveCourse(course core.Course) (agg.HqkjCourse, error) {
	if stored, ok := c.courses[course.ID]; ok {
		return stored, nil
	}
	if raw, ok := course.Raw.(agg.HqkjCourse); ok {
		return raw, nil
	}
	if course.Extra == nil {
		return agg.HqkjCourse{}, core.Invalid(core.HaiQiKeJi, "course", "course raw data is missing; call Courses first")
	}
	return agg.HqkjCourse{
		Id:           firstNonEmpty(course.Extra["courseId"], course.ID),
		Name:         course.Name,
		Cover:        course.Cover,
		CollegeId:    course.Extra["collegeId"],
		PeriodName:   course.Extra["periodName"],
		LecturerName: course.Extra["lecturer"],
		Credit:       parseFloat(course.Extra["credit"]),
		Code:         course.Extra["code"],
	}, nil
}

func nodeKind(node agg.HqkjNode) core.NodeKind {
	switch {
	case node.TabVideo > 0 || node.VideoDuration > 0:
		return core.NodeVideo
	case node.TabFile > 0:
		return core.NodeDocument
	case node.TabWork > 0:
		return core.NodeWork
	case node.TabExam > 0:
		return core.NodeExam
	default:
		return core.NodeUnknown
	}
}

func summarize(results []agg.HqkjQuizAnswerResult) string {
	if len(results) == 0 {
		return ""
	}
	total, submitted, failed := 0, 0, 0
	for _, item := range results {
		total++
		submitted += item.Submitted
		failed += item.Failed
	}
	return "tasks=" + strconv.Itoa(total) + ", submitted=" + strconv.Itoa(submitted) + ", failed=" + strconv.Itoa(failed)
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func parseFloat(value string) float64 {
	out, _ := strconv.ParseFloat(value, 64)
	return out
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
