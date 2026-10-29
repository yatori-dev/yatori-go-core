package xuexitong

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/yatori-dev/yatori-go-core/core"
	agg "github.com/yatori-dev/yatori-go-core/internal/aggregation/xuexitong"
	api "github.com/yatori-dev/yatori-go-core/internal/api/xuexitong"
	"github.com/yatori-dev/yatori-go-core/internal/driver/driverutil"
	"github.com/yatori-dev/yatori-go-core/models/ctype"
)

func init() {
	core.Register(core.XueXiTong, New)
}

type client struct {
	driverutil.Base
	cache    *api.XueXiTUserCache
	loggedIn bool

	courses  map[string]agg.XueXiTCourse
	chapters map[string]agg.ChaptersList
	nodes    map[string]core.Node
	works    map[string]*agg.XXTWork
	exams    map[string]*agg.XXTExam
}

// New constructs a XueXiTong client.
func New(cfg core.Config) (core.Client, error) {
	if strings.TrimSpace(cfg.Account.Username) == "" && strings.TrimSpace(cfg.Account.Cookie) == "" {
		return nil, core.Invalid(core.XueXiTong, "new", "username or cookie is required")
	}
	cache := &api.XueXiTUserCache{
		Name:     cfg.Account.Username,
		Password: cfg.Account.Password,
	}
	if enabled, proxy := driverutil.ProxyConfig(cfg); enabled {
		cache.IpProxySW = true
		cache.ProxyIP = proxy
	}
	return &client{
		Base:     driverutil.NewBase(cfg),
		cache:    cache,
		courses:  map[string]agg.XueXiTCourse{},
		chapters: map[string]agg.ChaptersList{},
		nodes:    map[string]core.Node{},
		works:    map[string]*agg.XXTWork{},
		exams:    map[string]*agg.XXTExam{},
	}, nil
}

func (c *client) Platform() core.Platform { return core.XueXiTong }

func (c *client) Capabilities() core.Capability {
	return core.CapLogin | core.CapCookieLogin | core.CapCourses | core.CapTree |
		core.CapStudy | core.CapWorks | core.CapExams
}

func (c *client) Login(ctx context.Context) error {
	unlock := c.Lock()
	defer unlock()
	if strings.TrimSpace(c.Config.Account.Cookie) != "" {
		return c.loginWithCookieLocked(c.Config.Account.Cookie)
	}
	driverutil.EnsureOCR()
	if err := agg.XueXiTLoginAction(c.cache); err != nil {
		return core.Wrap(core.XueXiTong, "login", err)
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
	if err := agg.XueXiTCookieLoginAction(c.cache); err != nil {
		return core.Wrap(core.XueXiTong, "login_with_cookie", err)
	}
	c.loggedIn = true
	return nil
}

func (c *client) Session() core.Session {
	cookie := c.cache.GetCookie()
	if cookie == "" {
		cookie = driverutil.CookieHeader(c.cache.GetCookies())
	}
	return core.Session{
		Platform: core.XueXiTong,
		Account:  c.cache.Name,
		UserID:   c.cache.UserID,
		Cookie:   cookie,
		Extra: map[string]string{
			"userID": c.cache.UserID,
		},
	}
}

func (c *client) Courses(ctx context.Context) ([]core.Course, error) {
	unlock := c.Lock()
	defer unlock()
	driverutil.EnsureOCR()
	if !c.loggedIn {
		return nil, core.NotLoggedIn(core.XueXiTong, "courses")
	}
	list, err := agg.XueXiTPullCourseActionContext(ctx, c.cache)
	if err != nil {
		return nil, core.Wrap(core.XueXiTong, "courses", err)
	}
	out := make([]core.Course, 0, len(list))
	for _, item := range list {
		c.courses[item.CourseID] = item
		status := core.CourseStatusActive
		if !item.IsStart {
			status = core.CourseStatusNotStarted
		}
		if item.State == 1 || item.JobRate >= 100 {
			status = core.CourseStatusFinished
		}
		out = append(out, core.Course{
			ID:       item.CourseID,
			Name:     item.CourseName,
			Teacher:  item.CourseTeacher,
			Cover:    item.CourseImage,
			Progress: item.JobRate,
			Status:   status,
			Extra: map[string]string{
				"cpi":          strconv.Itoa(item.Cpi),
				"key":          item.Key,
				"classId":      item.Key,
				"chatId":       item.ChatID,
				"courseDataId": strconv.Itoa(item.CourseDataID),
				"contentId":    strconv.Itoa(item.ContentID),
				"isStart":      strconv.FormatBool(item.IsStart),
				"state":        strconv.Itoa(item.State),
			},
			Raw: item,
		})
	}
	return out, nil
}

func (c *client) Tree(ctx context.Context, course core.Course) ([]core.Node, error) {
	unlock := c.Lock()
	defer unlock()
	driverutil.EnsureOCR()
	if !c.loggedIn {
		return nil, core.NotLoggedIn(core.XueXiTong, "tree")
	}
	raw, err := c.resolveCourse(course)
	if err != nil {
		return nil, err
	}
	cpi := raw.Cpi
	key := driverutil.Int(raw.Key)
	courseID := driverutil.Int(raw.CourseID)
	userID := driverutil.Int(c.cache.UserID)

	chapters, _, err := agg.PullCourseChapterActionContext(ctx, c.cache, cpi, key)
	if err != nil {
		return nil, core.Wrap(core.XueXiTong, "tree.chapters", err)
	}
	c.chapters[course.ID] = chapters

	ids := make([]int, 0, len(chapters.Knowledge))
	for _, item := range chapters.Knowledge {
		ids = append(ids, item.ID)
	}
	if updated, err := agg.ChapterFetchPointActionContext(ctx, c.cache, ids, &chapters, key, userID, cpi, courseID); err == nil {
		chapters = updated
		c.chapters[course.ID] = chapters
	} else {
		c.Logger.Debugf("xuexitong: chapter point status failed: %v", err)
	}

	out := make([]core.Node, 0, len(chapters.Knowledge))
	for index, item := range chapters.Knowledge {
		chapterNode := core.Node{
			ID:        fmt.Sprintf("chapter:%d", item.ID),
			CourseID:  course.ID,
			Name:      item.Name,
			Kind:      core.NodeChapter,
			Progress:  progressPercent(item.PointFinished, item.PointTotal),
			Completed: item.PointTotal > 0 && item.PointFinished >= item.PointTotal,
			Extra: map[string]string{
				"knowledgeId":  strconv.Itoa(item.ID),
				"label":        item.Label,
				"layer":        strconv.Itoa(item.Layer),
				"indexOrder":   strconv.Itoa(item.IndexOrder),
				"pointTotal":   strconv.Itoa(item.PointTotal),
				"pointFinish":  strconv.Itoa(item.PointFinished),
				"parentNodeId": strconv.Itoa(item.ParentNodeID),
			},
		}
		_, points, err := agg.ChapterFetchCardsAction(c.cache, &chapters, ids, index, courseID, key, cpi)
		if err != nil {
			c.Logger.Debugf("xuexitong: chapter %d cards failed: %v", item.ID, err)
		} else {
			chapterNode.Children = c.mapPoints(course.ID, cpi, key, points)
		}
		for _, child := range chapterNode.Children {
			c.nodes[child.ID] = child
		}
		c.nodes[chapterNode.ID] = chapterNode
		out = append(out, chapterNode)
	}
	return out, nil
}

func (c *client) mapPoints(courseID string, cpi, classID int, points []api.PointDto) []core.Node {
	videos, works, documents, hyperlinks, lives, bbs := api.ParsePointDto(points)
	out := make([]core.Node, 0, len(videos)+len(works)+len(documents)+len(hyperlinks)+len(lives)+len(bbs))

	for i := range videos {
		p := &videos[i]
		kind := core.NodeVideo
		if p.Type == ctype.InsertAudio {
			kind = core.NodeAudio
		}
		id := fmt.Sprintf("video:%d:%d:%s", p.KnowledgeID, p.CardIndex, firstNonEmpty(p.ObjectID, p.JobID))
		out = append(out, core.Node{
			ID:        id,
			CourseID:  courseID,
			Name:      p.Title,
			Kind:      kind,
			Duration:  driverutil.Seconds(p.Duration),
			Progress:  videoProgress(p),
			Completed: p.IsPassed,
			Extra: map[string]string{
				"knowledgeId": strconv.Itoa(p.KnowledgeID),
				"cardIndex":   strconv.Itoa(p.CardIndex),
				"cpi":         p.Cpi,
				"classId":     p.ClassID,
				"courseId":    p.CourseID,
				"objectId":    p.ObjectID,
				"jobId":       p.JobID,
				"title":       p.Title,
				"duration":    strconv.Itoa(p.Duration),
				"playTime":    strconv.Itoa(p.PlayTime),
				"isJob":       strconv.FormatBool(p.IsJob),
			},
			Raw: p,
		})
	}
	for i := range works {
		p := &works[i]
		out = append(out, core.Node{
			ID:       fmt.Sprintf("work:%d:%d:%s", p.KnowledgeID, p.CardIndex, p.WorkID),
			CourseID: courseID,
			Name:     "work",
			Kind:     core.NodeWork,
			Extra: map[string]string{
				"knowledgeId": strconv.Itoa(p.KnowledgeID),
				"cardIndex":   strconv.Itoa(p.CardIndex),
				"cpi":         p.Cpi,
				"classId":     p.ClassID,
				"courseId":    p.CourseID,
				"workId":      p.WorkID,
				"jobId":       p.JobID,
				"isJob":       strconv.FormatBool(p.IsJob),
			},
			Raw: p,
		})
	}
	for i := range documents {
		p := &documents[i]
		out = append(out, core.Node{
			ID:        fmt.Sprintf("document:%d:%d:%s", p.KnowledgeID, p.CardIndex, firstNonEmpty(p.ObjectID, p.JobID)),
			CourseID:  courseID,
			Name:      p.Title,
			Kind:      core.NodeDocument,
			Completed: p.Read,
			Extra: map[string]string{
				"knowledgeId": strconv.Itoa(p.KnowledgeID),
				"cardIndex":   strconv.Itoa(p.CardIndex),
				"cpi":         p.Cpi,
				"classId":     p.ClassID,
				"courseId":    p.CourseID,
				"objectId":    p.ObjectID,
				"jobId":       p.JobID,
				"title":       p.Title,
				"isJob":       strconv.FormatBool(p.IsJob),
			},
			Raw: p,
		})
	}
	for i := range hyperlinks {
		p := &hyperlinks[i]
		out = append(out, core.Node{
			ID:       fmt.Sprintf("link:%d:%d:%s", p.KnowledgeID, p.CardIndex, firstNonEmpty(p.ObjectID, p.JobID)),
			CourseID: courseID,
			Name:     p.Title,
			Kind:     core.NodeLink,
			Extra: map[string]string{
				"knowledgeId": strconv.Itoa(p.KnowledgeID),
				"cardIndex":   strconv.Itoa(p.CardIndex),
				"cpi":         p.Cpi,
				"classId":     p.ClassID,
				"courseId":    p.CourseID,
				"objectId":    p.ObjectID,
				"jobId":       p.JobID,
				"title":       p.Title,
			},
			Raw: p,
		})
	}
	for i := range lives {
		p := &lives[i]
		out = append(out, core.Node{
			ID:        fmt.Sprintf("live:%d:%d:%s", p.KnowledgeID, p.CardIndex, firstNonEmpty(p.LiveId, p.JobID)),
			CourseID:  courseID,
			Name:      p.Title,
			Kind:      core.NodeLive,
			Duration:  driverutil.Seconds(p.VideoDuration),
			Progress:  p.VideoCompletePercent,
			Completed: p.VideoCompletePercent >= 90,
			Extra: map[string]string{
				"knowledgeId": strconv.Itoa(p.KnowledgeID),
				"cardIndex":   strconv.Itoa(p.CardIndex),
				"cpi":         p.Cpi,
				"classId":     p.ClassID,
				"courseId":    p.CourseID,
				"liveId":      p.LiveId,
				"jobId":       p.JobID,
				"title":       p.Title,
				"isJob":       strconv.FormatBool(p.IsJob),
			},
			Raw: p,
		})
	}
	for i := range bbs {
		p := &bbs[i]
		out = append(out, core.Node{
			ID:       fmt.Sprintf("bbs:%d:%d:%s", p.KnowledgeID, p.CardIndex, firstNonEmpty(p.JobID, strconv.Itoa(p.CardIndex))),
			CourseID: courseID,
			Name:     p.Title,
			Kind:     core.NodeBBS,
			Extra: map[string]string{
				"knowledgeId": strconv.Itoa(p.KnowledgeID),
				"cardIndex":   strconv.Itoa(p.CardIndex),
				"cpi":         p.Cpi,
				"classId":     p.ClassID,
				"courseId":    p.CourseID,
				"jobId":       p.JobID,
				"title":       p.Title,
				"isJob":       strconv.FormatBool(p.IsJob),
			},
			Raw: p,
		})
	}
	return out
}

func (c *client) Study(ctx context.Context, req core.StudyRequest) (*core.StudyResult, error) {
	unlock := c.Lock()
	defer unlock()
	if !c.loggedIn {
		return nil, core.NotLoggedIn(core.XueXiTong, "study")
	}
	raw := req.Node.Raw
	if raw == nil {
		if stored, ok := c.nodes[req.Node.ID]; ok {
			raw = stored.Raw
			if req.Node.Kind == "" {
				req.Node.Kind = stored.Kind
			}
		}
	}
	result := &core.StudyResult{
		CourseID: req.Course.ID,
		NodeID:   req.Node.ID,
		Kind:     req.Node.Kind,
	}
	switch value := raw.(type) {
	case *api.PointVideoDto:
		return c.studyVideo(req, value, result)
	case *api.PointDocumentDto:
		return c.studyDocument(req, value, result)
	case *api.PointHyperlinkDto:
		return c.studyHyperlink(req, value, result)
	case *api.PointWorkDto:
		return nil, core.Unsupported(core.XueXiTong, "study/use DoWork for work nodes")
	case *api.PointLiveDto:
		return nil, core.Unsupported(core.XueXiTong, "study/live")
	case *api.PointBBsDto:
		return nil, core.Unsupported(core.XueXiTong, "study/bbs")
	default:
		return nil, core.Invalid(core.XueXiTong, "study", "node raw data is missing; call Tree first")
	}
}

func (c *client) studyVideo(req core.StudyRequest, p *api.PointVideoDto, result *core.StudyResult) (*core.StudyResult, error) {
	// The point DTO from the chapter list contains identifiers only. The
	// mobile card carries the attachment metadata (otherInfo/rt/face enc)
	// required by the submit endpoint, so fetch it once before submitting.
	if p.DToken == "" {
		card, _, err := agg.PageMobileChapterCardAction(
			c.cache,
			driverutil.Int(p.ClassID),
			driverutil.Int(p.CourseID),
			p.KnowledgeID,
			p.CardIndex,
			driverutil.Int(p.Cpi),
		)
		if err != nil {
			c.Logger.Debugf("xuexitong: mobile card for %s failed: %v", p.ObjectID, err)
		} else if _, err := p.AttachmentsDetection(card); err != nil {
			c.Logger.Debugf("xuexitong: attachment detection for %s failed: %v", p.ObjectID, err)
		}
	}
	if req.Action == core.StudyActionStart {
		if _, err := agg.VideoDtoFetchAction(c.cache, p); err != nil {
			return nil, core.Wrap(core.XueXiTong, "study.video.start", err)
		}
		result.Position = driverutil.Seconds(p.PlayTime)
		result.Progress = videoProgress(p)
		return result, nil
	}
	if p.Session == nil {
		if _, err := agg.VideoDtoFetchAction(c.cache, p); err != nil {
			return nil, core.Wrap(core.XueXiTong, "study.video.fetch", err)
		}
	}
	position := int(req.Position.Seconds())
	if position <= 0 {
		position = p.PlayTime
	}
	if req.Action == core.StudyActionFinish {
		position = p.Duration
	}
	if position > p.Duration {
		position = p.Duration
	}
	mode := 1
	if req.Mode == core.StudyModeWeb {
		mode = 0
	}
	isdrag := 0
	if req.Action == core.StudyActionStart {
		isdrag = 3
	}
	if req.IsDrag {
		isdrag = 2
	}
	var (
		report string
		err    error
	)
	if p.Type == ctype.InsertAudio {
		report, err = agg.AudioSubmitStudyTimeAction(c.cache, p, position, mode, isdrag)
	} else {
		report, err = agg.VideoSubmitStudyTimeAction(c.cache, p, position, mode, isdrag)
	}
	if err != nil {
		return nil, core.Wrap(core.XueXiTong, "study.video", err)
	}
	result.Submitted = true
	result.Position = driverutil.Seconds(position)
	result.Progress = videoProgress(p)
	result.Completed = p.IsPassed || position >= p.Duration
	result.Message = report
	result.Raw = report
	return result, nil
}

func (c *client) studyDocument(req core.StudyRequest, p *api.PointDocumentDto, result *core.StudyResult) (*core.StudyResult, error) {
	var (
		report string
		err    error
	)
	if req.Mode == core.StudyModeFast {
		report, err = c.cache.ReadV2PointPeReport(p, 3, nil)
	} else {
		report, err = c.cache.ReadV2PointWebReport(p, 3, nil)
	}
	if err != nil {
		return nil, core.Wrap(core.XueXiTong, "study.document", err)
	}
	result.Submitted = true
	result.Completed = true
	result.Message = report
	result.Raw = report
	return result, nil
}

func (c *client) studyHyperlink(req core.StudyRequest, p *api.PointHyperlinkDto, result *core.StudyResult) (*core.StudyResult, error) {
	report, err := c.cache.HyperlinkDtoCompleteReport(p, 3, nil)
	if err != nil {
		return nil, core.Wrap(core.XueXiTong, "study.hyperlink", err)
	}
	result.Submitted = true
	result.Completed = true
	result.Message = report
	result.Raw = report
	return result, nil
}

func (c *client) Works(ctx context.Context, course core.Course) ([]core.Work, error) {
	unlock := c.Lock()
	defer unlock()
	if !c.loggedIn {
		return nil, core.NotLoggedIn(core.XueXiTong, "works")
	}
	raw, err := c.resolveCourse(course)
	if err != nil {
		return nil, err
	}
	list, err := agg.PullWorkListAction(c.cache, raw)
	if err != nil {
		return nil, core.Wrap(core.XueXiTong, "works", err)
	}
	out := make([]core.Work, 0, len(list))
	for i := range list {
		item := &list[i]
		id := firstNonEmpty(item.TaskRefId, item.AnswerId, item.RawURL, item.Name)
		c.works[id] = item
		out = append(out, core.Work{
			ID:            id,
			CourseID:      course.ID,
			Name:          item.Name,
			Status:        item.Status,
			QuestionCount: item.QuestionTotal,
			Extra: map[string]string{
				"courseId":    item.CourseId,
				"classId":     item.ClazzId,
				"cpi":         item.Cpi,
				"answerId":    item.AnswerId,
				"taskRefId":   item.TaskRefId,
				"msgId":       item.MsgId,
				"encTask":     item.EncTask,
				"type":        item.Type,
				"rawUrl":      item.RawURL,
				"questionNum": strconv.Itoa(item.QuestionTotal),
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
		return nil, core.NotLoggedIn(core.XueXiTong, "work")
	}
	raw, ok := req.Work.Raw.(*agg.XXTWork)
	if !ok || raw == nil {
		raw = c.works[req.Work.ID]
	}
	if raw == nil {
		return nil, core.Invalid(core.XueXiTong, "work", "work raw data is missing; call Works first")
	}
	work := *raw
	if err := agg.EnterWorkAction(c.cache, &work); err != nil {
		return nil, core.Wrap(core.XueXiTong, "work.enter", err)
	}
	result := &core.AnswerResult{TaskID: req.Work.ID, Title: work.Name}
	for i := 0; i < work.QuestionTotal; i++ {
		question, err := work.PullWorkQuestionAction(c.cache, i)
		if err != nil {
			return nil, core.Wrap(core.XueXiTong, fmt.Sprintf("work.question.%d", i), err)
		}
		if err := c.answerQuestion(&question, req.Source); err != nil {
			return nil, err
		}
		isSubmit := req.Submit && i == work.QuestionTotal-1
		report, err := question.SubmitWorkAnswerAction(c.cache, isSubmit)
		if err != nil {
			return nil, core.Wrap(core.XueXiTong, fmt.Sprintf("work.submit.%d", i), err)
		}
		result.Message = report
		result.Raw = report
		result.Answers = append(result.Answers, answerFromQuestion(question.Question))
	}
	result.Submitted = req.Submit
	return result, nil
}

func (c *client) Exams(ctx context.Context, course core.Course) ([]core.Exam, error) {
	unlock := c.Lock()
	defer unlock()
	if !c.loggedIn {
		return nil, core.NotLoggedIn(core.XueXiTong, "exams")
	}
	raw, err := c.resolveCourse(course)
	if err != nil {
		return nil, err
	}
	list, err := agg.PullExamListAction(c.cache, raw)
	if err != nil {
		return nil, core.Wrap(core.XueXiTong, "exams", err)
	}
	out := make([]core.Exam, 0, len(list))
	for i := range list {
		item := &list[i]
		id := firstNonEmpty(item.TaskRefId, item.AnswerId, item.RawURL, item.Name)
		c.exams[id] = item
		out = append(out, core.Exam{
			ID:            id,
			CourseID:      course.ID,
			Name:          item.Name,
			Status:        item.Status,
			QuestionCount: item.QuestionTotal,
			Extra: map[string]string{
				"courseId":    item.CourseId,
				"classId":     item.ClazzId,
				"cpi":         item.Cpi,
				"answerId":    item.AnswerId,
				"taskRefId":   item.TaskRefId,
				"msgId":       item.MsgId,
				"encTask":     item.EncTask,
				"type":        item.Type,
				"rawUrl":      item.RawURL,
				"questionNum": strconv.Itoa(item.QuestionTotal),
			},
			Raw: item,
		})
	}
	return out, nil
}

func (c *client) DoExam(ctx context.Context, req core.ExamRequest) (*core.AnswerResult, error) {
	unlock := c.Lock()
	defer unlock()
	if !c.loggedIn {
		return nil, core.NotLoggedIn(core.XueXiTong, "exam")
	}
	raw, ok := req.Exam.Raw.(*agg.XXTExam)
	if !ok || raw == nil {
		raw = c.exams[req.Exam.ID]
	}
	if raw == nil {
		return nil, core.Invalid(core.XueXiTong, "exam", "exam raw data is missing; call Exams first")
	}
	exam := *raw
	if err := agg.EnterExamAction(c.cache, &exam); err != nil {
		return nil, core.Wrap(core.XueXiTong, "exam.enter", err)
	}
	result := &core.AnswerResult{TaskID: req.Exam.ID, Title: exam.Name}
	for i := 0; i < exam.QuestionTotal; i++ {
		question, err := exam.PullExamQuestionAction(c.cache, i)
		if err != nil {
			return nil, core.Wrap(core.XueXiTong, fmt.Sprintf("exam.question.%d", i), err)
		}
		if err := c.answerQuestion(&question, req.Source); err != nil {
			return nil, err
		}
		isSubmit := req.Submit && i == exam.QuestionTotal-1
		report, err := question.SubmitExamAnswerAction(c.cache, isSubmit)
		if err != nil {
			return nil, core.Wrap(core.XueXiTong, fmt.Sprintf("exam.submit.%d", i), err)
		}
		result.Message = report
		result.Raw = report
		result.Answers = append(result.Answers, answerFromQuestion(question.Question))
	}
	result.Submitted = req.Submit
	return result, nil
}

func (c *client) SignIn(ctx context.Context, req core.SignInRequest) (*core.SignInResult, error) {
	return nil, core.Unsupported(core.XueXiTong, "sign_in")
}

func (c *client) Call(ctx context.Context, action core.Action, params map[string]any) (any, error) {
	return core.Call(ctx, c, action, params)
}

func (c *client) Close() error { return nil }

func (c *client) resolveCourse(course core.Course) (agg.XueXiTCourse, error) {
	if stored, ok := c.courses[course.ID]; ok {
		return stored, nil
	}
	if raw, ok := course.Raw.(agg.XueXiTCourse); ok {
		return raw, nil
	}
	if raw, ok := course.Raw.(*agg.XueXiTCourse); ok && raw != nil {
		return *raw, nil
	}
	if course.Extra == nil {
		return agg.XueXiTCourse{}, core.Invalid(core.XueXiTong, "course", "course raw data is missing; call Courses first")
	}
	return agg.XueXiTCourse{
		Cpi:           driverutil.Int(course.Extra["cpi"]),
		Key:           course.Extra["key"],
		CourseID:      course.ID,
		ChatID:        course.Extra["chatId"],
		CourseTeacher: course.Teacher,
		CourseName:    course.Name,
		CourseImage:   course.Cover,
		CourseDataID:  driverutil.Int(course.Extra["courseDataId"]),
		ContentID:     driverutil.Int(course.Extra["contentId"]),
		IsStart:       driverutil.Bool(course.Extra["isStart"]),
		State:         driverutil.Int(course.Extra["state"]),
	}, nil
}

var _ core.Client = (*client)(nil)
