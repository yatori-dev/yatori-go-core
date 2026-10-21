package yinghua

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/yatori-dev/yatori-go-core/core"
	agg "github.com/yatori-dev/yatori-go-core/internal/aggregation/yinghua"
	api "github.com/yatori-dev/yatori-go-core/internal/api/yinghua"
	"github.com/yatori-dev/yatori-go-core/internal/driver/driverutil"
	"github.com/yatori-dev/yatori-go-core/models/ctype"
)

func init() {
	core.Register(core.YingHua, New)
}

type client struct {
	driverutil.Base
	cache    *api.YingHuaUserCache
	loggedIn bool

	courses map[string]agg.YingHuaCourse
	nodes   map[string]agg.YingHuaNode
	trees   map[string][]agg.YingHuaNode
	works   map[string]*agg.YingHuaWork
	exams   map[string]*agg.YingHuaExam
}

// New constructs a YingHua client.
func New(cfg core.Config) (core.Client, error) {
	if strings.TrimSpace(cfg.Account.URL) == "" {
		return nil, core.Invalid(core.YingHua, "new", "url is required")
	}
	cache := &api.YingHuaUserCache{
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
		courses: map[string]agg.YingHuaCourse{},
		nodes:   map[string]agg.YingHuaNode{},
		trees:   map[string][]agg.YingHuaNode{},
		works:   map[string]*agg.YingHuaWork{},
		exams:   map[string]*agg.YingHuaExam{},
	}, nil
}

func (c *client) Platform() core.Platform { return core.YingHua }

func (c *client) Capabilities() core.Capability {
	return core.CapLogin | core.CapCourses | core.CapTree | core.CapStudy |
		core.CapWorks | core.CapExams
}

func (c *client) Login(ctx context.Context) error {
	unlock := c.Lock()
	defer unlock()
	driverutil.EnsureOCR()
	if err := agg.YingHuaLoginAction(c.cache); err != nil {
		return core.Wrap(core.YingHua, "login", err)
	}
	c.loggedIn = true
	return nil
}

func (c *client) LoginWithCookie(ctx context.Context, cookie string) error {
	return core.Unsupported(core.YingHua, "login_with_cookie")
}

func (c *client) Session() core.Session {
	return core.Session{
		Platform: core.YingHua,
		Account:  c.cache.Account,
		Cookie:   c.cache.GetCookie(),
		Token:    c.cache.GetToken(),
		Extra: map[string]string{
			"sign": c.cache.GetSign(),
			"url":  c.cache.PreUrl,
		},
	}
}

func (c *client) Courses(ctx context.Context) ([]core.Course, error) {
	unlock := c.Lock()
	defer unlock()
	if !c.loggedIn {
		return nil, core.NotLoggedIn(core.YingHua, "courses")
	}
	list, err := agg.CourseListAction(c.cache)
	if err != nil {
		return nil, core.Wrap(core.YingHua, "courses", err)
	}
	now := time.Now()
	out := make([]core.Course, 0, len(list))
	for _, item := range list {
		c.courses[item.Id] = item
		status := core.CourseStatusActive
		if !item.StartDate.IsZero() && now.Before(item.StartDate) {
			status = core.CourseStatusNotStarted
		}
		if item.Progress >= 1 || (!item.EndDate.IsZero() && now.After(item.EndDate)) {
			status = core.CourseStatusFinished
		}
		out = append(out, core.Course{
			ID:       item.Id,
			Name:     item.Name,
			Progress: percent(item.Progress),
			Status:   status,
			Extra: map[string]string{
				"mode":         strconv.Itoa(item.Mode),
				"startDate":    item.StartDate.Format("2006-01-02"),
				"endDate":      item.EndDate.Format("2006-01-02"),
				"videoCount":   strconv.Itoa(item.VideoCount),
				"videoLearned": strconv.Itoa(item.VideoLearned),
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
		return nil, core.NotLoggedIn(core.YingHua, "tree")
	}
	raw, err := c.resolveCourse(course)
	if err != nil {
		return nil, err
	}
	nodes, err := c.treeLocked(raw)
	if err != nil {
		return nil, err
	}
	out := make([]core.Node, 0, len(nodes))
	for _, node := range nodes {
		out = append(out, core.Node{
			ID:        node.Id,
			CourseID:  course.ID,
			Name:      node.Name,
			Kind:      nodeKind(node),
			Duration:  driverutil.Seconds(node.VideoDuration),
			Progress:  float64(node.Progress),
			Completed: node.Progress >= 100,
			Extra: map[string]string{
				"courseId":      node.CourseId,
				"nodeId":        node.Id,
				"tabVideo":      strconv.FormatBool(node.TabVideo),
				"tabFile":       strconv.FormatBool(node.TabFile),
				"tabVote":       strconv.FormatBool(node.TabVote),
				"tabWork":       strconv.FormatBool(node.TabWork),
				"tabExam":       strconv.FormatBool(node.TabExam),
				"nodeLock":      strconv.Itoa(node.NodeLock),
				"unlockTime":    node.UnlockTime.Format("2006-01-02 15:04"),
				"viewedSeconds": strconv.Itoa(node.ViewedDuration),
				"duration":      strconv.Itoa(node.VideoDuration),
			},
			Raw: node,
		})
	}
	return out, nil
}

func (c *client) treeLocked(course agg.YingHuaCourse) ([]agg.YingHuaNode, error) {
	if nodes, ok := c.trees[course.Id]; ok {
		return nodes, nil
	}
	nodes, err := agg.VideosListAction(c.cache, course)
	if err != nil {
		return nil, core.Wrap(core.YingHua, "tree", err)
	}
	c.trees[course.Id] = nodes
	for i := range nodes {
		c.nodes[nodes[i].Id] = nodes[i]
	}
	return nodes, nil
}

func (c *client) Study(ctx context.Context, req core.StudyRequest) (*core.StudyResult, error) {
	unlock := c.Lock()
	defer unlock()
	if !c.loggedIn {
		return nil, core.NotLoggedIn(core.YingHua, "study")
	}
	node, ok := req.Node.Raw.(agg.YingHuaNode)
	if !ok {
		if stored, exists := c.nodes[req.Node.ID]; exists {
			node = stored
		} else {
			return nil, core.Invalid(core.YingHua, "study", "node raw data is missing; call Tree first")
		}
	}
	position := int(req.Position.Seconds())
	if position <= 0 {
		position = node.ViewedDuration
	}
	if req.Action == core.StudyActionStart {
		position = node.ViewedDuration
	}
	if req.Action == core.StudyActionFinish {
		position = node.VideoDuration
	}
	if node.VideoDuration > 0 && position > node.VideoDuration {
		position = node.VideoDuration
	}
	report, err := api.SubmitStudyTimeApi(*c.cache, node.Id, "0", position, 10, nil)
	if err != nil {
		return nil, core.Wrap(core.YingHua, "study", err)
	}
	progress := float64(node.Progress)
	if node.VideoDuration > 0 {
		progress = float64(position) / float64(node.VideoDuration) * 100
	}
	return &core.StudyResult{
		CourseID:  req.Course.ID,
		NodeID:    req.Node.ID,
		Kind:      nodeKind(node),
		Submitted: true,
		Position:  driverutil.Seconds(position),
		Progress:  progress,
		Completed: progress >= 100,
		Message:   report,
		Raw:       report,
	}, nil
}

func (c *client) Works(ctx context.Context, course core.Course) ([]core.Work, error) {
	unlock := c.Lock()
	defer unlock()
	if !c.loggedIn {
		return nil, core.NotLoggedIn(core.YingHua, "works")
	}
	raw, err := c.resolveCourse(course)
	if err != nil {
		return nil, err
	}
	nodes, err := c.treeLocked(raw)
	if err != nil {
		return nil, err
	}
	out := make([]core.Work, 0)
	for _, node := range nodes {
		if !node.TabWork {
			continue
		}
		list, err := agg.WorkDetailAction(c.cache, node.Id)
		if err != nil {
			c.Logger.Debugf("yinghua: work detail for node %s failed: %v", node.Id, err)
			continue
		}
		for i := range list {
			item := &list[i]
			id := firstNonEmpty(item.WorkId, item.Id)
			c.works[id] = item
			out = append(out, core.Work{
				ID:       id,
				CourseID: course.ID,
				Name:     item.Title,
				Deadline: item.EndTime,
				Extra: map[string]string{
					"nodeId":    item.NodeId,
					"workId":    item.WorkId,
					"courseId":  item.CourseId,
					"startTime": item.StartTime.Format("2006-01-02 15:04:05"),
					"endTime":   item.EndTime.Format("2006-01-02 15:04:05"),
					"allow":     strconv.Itoa(item.Allow),
					"frequency": strconv.Itoa(item.Frequency),
					"score":     fmt.Sprintf("%.0f", item.Score),
				},
				Raw: item,
			})
		}
	}
	return out, nil
}

func (c *client) DoWork(ctx context.Context, req core.WorkRequest) (*core.AnswerResult, error) {
	unlock := c.Lock()
	defer unlock()
	if !c.loggedIn {
		return nil, core.NotLoggedIn(core.YingHua, "work")
	}
	work, err := c.resolveWork(req.Work)
	if err != nil {
		return nil, err
	}
	submit := boolInt(req.Submit)
	result := &core.AnswerResult{TaskID: req.Work.ID, Title: work.Title, Submitted: req.Submit}
	switch req.Source {
	case core.AnswerSourceAI:
		ai := c.Config.Options.AI
		if ai == nil || ai.APIKey == "" {
			return nil, core.Invalid(core.YingHua, "work", "AI config is required for AnswerSourceAI")
		}
		if err := agg.StartWorkAction(c.cache, work, ai.URL, ai.Model, ai.APIKey, ctype.AiType(ai.Type), submit); err != nil {
			return nil, core.Wrap(core.YingHua, "work", err)
		}
	case core.AnswerSourceBank:
		bank := c.Config.Options.QuestionBank
		if bank == nil || bank.URL == "" {
			return nil, core.Invalid(core.YingHua, "work", "question bank config is required for AnswerSourceBank")
		}
		if err := agg.StartWorkForExternalAction(c.cache, bank.URL, work, submit); err != nil {
			return nil, core.Wrap(core.YingHua, "work", err)
		}
	default:
		return nil, core.Unsupported(core.YingHua, "work/source "+string(req.Source))
	}
	if submit == 1 {
		if score, err := agg.WorkedFinallyScoreAction(c.cache, work); err == nil {
			result.Score, _ = strconv.ParseFloat(score, 64)
			result.Message = "score=" + score
		}
	}
	return result, nil
}

func (c *client) Exams(ctx context.Context, course core.Course) ([]core.Exam, error) {
	unlock := c.Lock()
	defer unlock()
	if !c.loggedIn {
		return nil, core.NotLoggedIn(core.YingHua, "exams")
	}
	raw, err := c.resolveCourse(course)
	if err != nil {
		return nil, err
	}
	nodes, err := c.treeLocked(raw)
	if err != nil {
		return nil, err
	}
	out := make([]core.Exam, 0)
	for _, node := range nodes {
		if !node.TabExam {
			continue
		}
		list, err := agg.ExamDetailAction(c.cache, node.Id)
		if err != nil {
			c.Logger.Debugf("yinghua: exam detail for node %s failed: %v", node.Id, err)
			continue
		}
		for i := range list {
			item := &list[i]
			id := firstNonEmpty(item.ExamId, item.Id)
			c.exams[id] = item
			out = append(out, core.Exam{
				ID:       id,
				CourseID: course.ID,
				Name:     item.Title,
				StartAt:  item.StartTime,
				EndAt:    item.EndTime,
				Duration: time.Duration(float64(item.LimitedTime) * float64(time.Minute)),
				Extra: map[string]string{
					"nodeId":   item.NodeId,
					"examId":   item.ExamId,
					"courseId": item.CourseId,
					"score":    fmt.Sprintf("%.0f", item.Score),
				},
				Raw: item,
			})
		}
	}
	return out, nil
}

func (c *client) DoExam(ctx context.Context, req core.ExamRequest) (*core.AnswerResult, error) {
	unlock := c.Lock()
	defer unlock()
	if !c.loggedIn {
		return nil, core.NotLoggedIn(core.YingHua, "exam")
	}
	exam, err := c.resolveExam(req.Exam)
	if err != nil {
		return nil, err
	}
	submit := boolInt(req.Submit)
	result := &core.AnswerResult{TaskID: req.Exam.ID, Title: exam.Title, Submitted: req.Submit}
	switch req.Source {
	case core.AnswerSourceAI:
		ai := c.Config.Options.AI
		if ai == nil || ai.APIKey == "" {
			return nil, core.Invalid(core.YingHua, "exam", "AI config is required for AnswerSourceAI")
		}
		if err := agg.StartExamAction(c.cache, exam, ai.URL, ai.Model, ai.APIKey, ctype.AiType(ai.Type), submit); err != nil {
			return nil, core.Wrap(core.YingHua, "exam", err)
		}
	case core.AnswerSourceBank:
		bank := c.Config.Options.QuestionBank
		if bank == nil || bank.URL == "" {
			return nil, core.Invalid(core.YingHua, "exam", "question bank config is required for AnswerSourceBank")
		}
		if err := agg.StartExamForExternalAction(c.cache, exam, bank.URL, submit); err != nil {
			return nil, core.Wrap(core.YingHua, "exam", err)
		}
	default:
		return nil, core.Unsupported(core.YingHua, "exam/source "+string(req.Source))
	}
	if submit == 1 {
		if score, err := agg.ExamFinallyScoreAction(c.cache, exam); err == nil {
			result.Score, _ = strconv.ParseFloat(score, 64)
			result.Message = "score=" + score
		}
	}
	return result, nil
}

func (c *client) SignIn(ctx context.Context, req core.SignInRequest) (*core.SignInResult, error) {
	return nil, core.Unsupported(core.YingHua, "sign_in")
}

func (c *client) Call(ctx context.Context, action core.Action, params map[string]any) (any, error) {
	return core.Call(ctx, c, action, params)
}

func (c *client) Close() error { return nil }

func (c *client) resolveCourse(course core.Course) (agg.YingHuaCourse, error) {
	if stored, ok := c.courses[course.ID]; ok {
		return stored, nil
	}
	if raw, ok := course.Raw.(agg.YingHuaCourse); ok {
		return raw, nil
	}
	if course.Extra == nil {
		return agg.YingHuaCourse{}, core.Invalid(core.YingHua, "course", "course raw data is missing; call Courses first")
	}
	return agg.YingHuaCourse{Id: course.ID, Name: course.Name, Progress: course.Progress / 100}, nil
}

func (c *client) resolveWork(work core.Work) (agg.YingHuaWork, error) {
	if raw, ok := work.Raw.(*agg.YingHuaWork); ok && raw != nil {
		return *raw, nil
	}
	if stored, ok := c.works[work.ID]; ok {
		return *stored, nil
	}
	return agg.YingHuaWork{}, core.Invalid(core.YingHua, "work", "work raw data is missing; call Works first")
}

func (c *client) resolveExam(exam core.Exam) (agg.YingHuaExam, error) {
	if raw, ok := exam.Raw.(*agg.YingHuaExam); ok && raw != nil {
		return *raw, nil
	}
	if stored, ok := c.exams[exam.ID]; ok {
		return *stored, nil
	}
	return agg.YingHuaExam{}, core.Invalid(core.YingHua, "exam", "exam raw data is missing; call Exams first")
}

func nodeKind(node agg.YingHuaNode) core.NodeKind {
	if node.TabVideo {
		return core.NodeVideo
	}
	if node.TabFile {
		return core.NodeDocument
	}
	return core.NodeUnknown
}

func percent(value float64) float64 {
	if value <= 1 {
		return value * 100
	}
	return value
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
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
