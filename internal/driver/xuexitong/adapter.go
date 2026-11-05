package xuexitong

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/v3/core"
	agg "github.com/yatori-dev/yatori-go-core/v3/internal/aggregation/xuexitong"
	api "github.com/yatori-dev/yatori-go-core/v3/internal/api/xuexitong"
	"github.com/yatori-dev/yatori-go-core/v3/internal/driver/driverutil"
	"github.com/yatori-dev/yatori-go-core/v3/models/ctype"
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
	if sign := cfg.Options.FaceSignature; sign != nil {
		// 纯本地签名所需的三件套允许放在顶层或 Chapter 里，二者取并集。
		sc := sign.Sc
		keyPEM := sign.KeyPEM
		cxcid := sign.Cxcid
		if ch := sign.Chapter; ch != nil {
			if ch.Sc != "" {
				sc = ch.Sc
			}
			if ch.KeyPEM != "" {
				keyPEM = ch.KeyPEM
			}
			if ch.Cxcid != "" {
				cxcid = ch.Cxcid
			}
		}
		// App 内置私钥已随 SDK 一起发布；调用方只需提供设备 cxcid 与当前 sc。
		if keyPEM == "" && cxcid != "" && sc != "" {
			keyPEM = api.DefaultFaceKeyPEM()
		}
		cache.FaceSc = sc
		if keyPEM != "" {
			cache.FaceKeyPEM = []byte(keyPEM)
		}
		if sign.Cxcid != "" || sign.SignT != "" || sign.SignK != "" {
			cache.FaceSign = &api.FaceSign{
				Cxcid:  sign.Cxcid,
				CxTime: sign.CxTime,
				SignK:  sign.SignK,
				SignT:  sign.SignT,
			}
			if cache.FaceSign.Cxcid == "" {
				cache.FaceSign.Cxcid = cxcid
			}
		}
		if ch := sign.Chapter; ch != nil {
			cache.FaceSignChapter = &api.FaceSign{
				Cxcid:  ch.Cxcid,
				CxTime: ch.CxTime,
				SignK:  ch.SignK,
				SignT:  ch.SignT,
			}
			if cache.FaceSignChapter.Cxcid == "" {
				cache.FaceSignChapter.Cxcid = cxcid
			}
		}
		// 只配置了 Cxcid/Sc（没有显式 Chapter）时，也为章节校验创建纯算入口。
		if cxcid != "" && sc != "" {
			if cache.FaceSign == nil {
				cache.FaceSign = &api.FaceSign{Cxcid: cxcid}
			}
			if cache.FaceSignChapter == nil {
				cache.FaceSignChapter = &api.FaceSign{Cxcid: cxcid}
			}
		}
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
		// 学习通的 myjobsnodesmap 对部分课程会返回空的任务点统计，
		// 此时用子节点完成情况回推章节进度，避免章节一直显示 0%。
		if item.PointTotal == 0 && len(chapterNode.Children) > 0 {
			completed := 0
			for _, child := range chapterNode.Children {
				if child.Completed {
					completed++
				}
			}
			chapterNode.Progress = float64(completed) / float64(len(chapterNode.Children)) * 100
			chapterNode.Completed = completed == len(chapterNode.Children)
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
		id := fmt.Sprintf("video:%d:%d:%s:%s", p.KnowledgeID, p.CardIndex, p.ObjectID, p.JobID)
		completed := p.IsPassed || (p.Duration > 0 && p.PlayTime >= p.Duration)
		out = append(out, core.Node{
			ID:        id,
			CourseID:  courseID,
			Name:      p.Title,
			Kind:      kind,
			Duration:  driverutil.Seconds(p.Duration),
			Progress:  videoProgress(p),
			Completed: completed,
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
			ID:       fmt.Sprintf("work:%d:%d:%s:%s", p.KnowledgeID, p.CardIndex, p.WorkID, p.JobID),
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
		progress := 0.0
		if p.Read {
			progress = 100
		}
		out = append(out, core.Node{
			ID:        fmt.Sprintf("document:%d:%d:%s:%s", p.KnowledgeID, p.CardIndex, p.ObjectID, p.JobID),
			CourseID:  courseID,
			Name:      p.Title,
			Kind:      core.NodeDocument,
			Progress:  progress,
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
			ID:       fmt.Sprintf("link:%d:%d:%s:%s", p.KnowledgeID, p.CardIndex, p.ObjectID, p.JobID),
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
			ID:        fmt.Sprintf("live:%d:%d:%s:%s", p.KnowledgeID, p.CardIndex, p.LiveId, p.JobID),
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
		card, err := fetchMobileCardCached(
			c.cache,
			driverutil.Int(p.ClassID),
			driverutil.Int(p.CourseID),
			p.KnowledgeID,
			p.CardIndex,
			driverutil.Int(p.Cpi),
		)
		if err != nil {
			// 多任务点并发拉卡片时容易触发图片验证码。验证码不是节点本身
			// 的错误，这里返回一个未完成的普通结果，让调用方稍后重试。
			if strings.Contains(err.Error(), "验证码") {
				result.Message = "章节卡片触发验证码，已跳过本次学习"
				return result, nil
			}
			// 人脸锁/章节未开放等错误必须上抛，否则会继续用一个空的
			// DToken 去请求视频状态，最终只看到 203 这种误导性错误。
			return nil, core.Wrap(core.XueXiTong, "study.video.card", err)
		}
		if _, err := p.AttachmentsDetection(card); err != nil {
			c.Logger.Debugf("xuexitong: attachment detection for %s failed: %v", p.ObjectID, err)
		}
	}
	if req.Action == core.StudyActionStart {
		if _, err := agg.VideoDtoFetchAction(c.cache, p); err != nil {
			return nil, core.Wrap(core.XueXiTong, "study.video.start", err)
		}
		result.Position = driverutil.Seconds(p.PlayTime)
		result.Progress = videoProgress(p)
		result.Duration = driverutil.Seconds(p.Duration)
		return result, nil
	}
	if p.Session == nil {
		if _, err := agg.VideoDtoFetchAction(c.cache, p); err != nil {
			return nil, core.Wrap(core.XueXiTong, "study.video.fetch", err)
		}
	}
	result.Duration = driverutil.Seconds(p.Duration)
	position := int(req.Position.Seconds())
	if position <= 0 {
		position = p.PlayTime
	}
	if req.Action == core.StudyActionFinish || req.Mode == core.StudyModeFast {
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
	passed := p.IsPassed
	isPassed := gojsonq.New().JSONString(report).Find("isPassed")
	if isPassed == nil {
		return nil, core.Wrap(core.XueXiTong, "study.video", fmt.Errorf("提交学时返回异常：%s", summarizeReport(report)))
	}
	if flag, ok := isPassed.(bool); ok {
		passed = flag
	}
	// 视频随机人脸抓拍：服务器在 isPassed=false 且已到抓拍时间时，
	// 要求先完成一次人脸采集并把 videoFaceCaptureSuccessEnc 带回。
	if !passed && !p.VideoFaceCaptured && p.VideoFaceCaptureEnc != "" && videoFaceCaptureDue(p, position) {
		if p.CardEnc == "" {
			if card, cardErr := fetchMobileCardCached(c.cache, driverutil.Int(p.ClassID), driverutil.Int(p.CourseID), p.KnowledgeID, p.CardIndex, driverutil.Int(p.Cpi)); cardErr == nil {
				_, _ = p.AttachmentsDetection(card)
			}
		}
		successEnc, faceErr := agg.PassVideoFaceAction(c.cache, p)
		if faceErr != nil {
			c.Logger.Debugf("xuexitong: video face capture for %s failed: %v", p.ObjectID, faceErr)
		} else if successEnc != "" {
			p.VideoFaceCaptureEnc = successEnc
			p.VideoFaceCaptured = true
			if p.Type == ctype.InsertAudio {
				report, err = agg.AudioSubmitStudyTimeAction(c.cache, p, position, mode, isdrag)
			} else {
				report, err = agg.VideoSubmitStudyTimeAction(c.cache, p, position, mode, isdrag)
			}
			if err != nil {
				return nil, core.Wrap(core.XueXiTong, "study.video", err)
			}
			if flag, ok := gojsonq.New().JSONString(report).Find("isPassed").(bool); ok {
				passed = flag
			}
		}
	}
	progress := videoProgress(p)
	if p.Duration > 0 {
		progress = float64(position) / float64(p.Duration) * 100
		if progress > 100 {
			progress = 100
		}
	}
	if req.Action == core.StudyActionFinish || req.Mode == core.StudyModeFast {
		progress = 100
	}
	result.Submitted = true
	result.Position = driverutil.Seconds(position)
	result.Progress = progress
	result.Completed = passed || position >= p.Duration
	result.Message = report
	result.Raw = report
	return result, nil
}

func (c *client) studyDocument(req core.StudyRequest, p *api.PointDocumentDto, result *core.StudyResult) (*core.StudyResult, error) {
	// 文档任务点的 jtoken 在章节卡片 attachment 里，gas/knowledge 的
	// iframe data 不带；缺 jtoken 时 ananas/job/document 会返回
	// “无效的请求参数”。这里按卡片索引补拉一次手机端卡片。
	if p.Jtoken == "" {
		if card, err := fetchMobileCardCached(
			c.cache,
			driverutil.Int(p.ClassID),
			driverutil.Int(p.CourseID),
			p.KnowledgeID,
			p.CardIndex,
			driverutil.Int(p.Cpi),
		); err == nil {
			_, _ = p.AttachmentsDetection(card)
		} else {
			c.Logger.Debugf("xuexitong: document card for %s failed: %v", p.ObjectID, err)
			if strings.Contains(err.Error(), "验证码") {
				result.Message = "章节卡片触发验证码，已跳过本次文档"
				return result, nil
			}
		}
	}
	var (
		report string
		err    error
	)
	switch p.Type {
	case ctype.InsertBook:
		report, err = c.cache.DocumentDtoReadingBookReport(p, 3, nil)
	case ctype.InsertReadV2:
		if req.Mode == core.StudyModeFast {
			report, err = c.cache.ReadV2PointPeReport(p, 3, nil)
		} else {
			report, err = c.cache.ReadV2PointWebReport(p, 3, nil)
		}
	default:
		report, err = c.cache.DocumentDtoReadingReport(p, 3, nil)
	}
	if err != nil {
		return nil, core.Wrap(core.XueXiTong, "study.document", err)
	}
	if err := validateReadReport(report); err != nil {
		// 部分课程会挂一些不属于当前账号任务点的文档（服务端直接返回
		// “无效的请求参数”）。这类节点继续重试也不会成功，按非任务点
		// 跳过，避免整门课被一个无效文档卡住。
		if strings.Contains(report, "无效的请求参数") {
			result.Submitted = false
			result.Progress = 100
			result.Completed = true
			result.Message = "非任务点文档，已跳过"
			result.Raw = report
			return result, nil
		}
		return nil, core.Wrap(core.XueXiTong, "study.document", err)
	}
	result.Submitted = true
	result.Progress = 100
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
	if err := validateReadReport(report); err != nil {
		return nil, core.Wrap(core.XueXiTong, "study.hyperlink", err)
	}
	result.Submitted = true
	result.Progress = 100
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
	if work.QuestionTotal <= 0 {
		return nil, core.Invalid(core.XueXiTong, "work", "作业题目数量为 0，可能是作业未开放或列表数据不完整")
	}
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
	if exam.QuestionTotal <= 0 {
		return nil, core.Invalid(core.XueXiTong, "exam", "考试题目数量为 0，可能是考试未开放或列表数据不完整")
	}
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
