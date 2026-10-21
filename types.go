package yatori

import "github.com/yatori-dev/yatori-go-core/core"

// Platform identifies a supported upstream learning platform.
type Platform = core.Platform

const (
	XueXiTong = core.XueXiTong
	YingHua   = core.YingHua
	ICVE      = core.ICVE
	CQIE      = core.CQIE
	ENAAE     = core.ENAAE
	WeLearn   = core.WeLearn
	QSXT      = core.QSXT
	TTCDW     = core.TTCDW
	Ketangx   = core.Ketangx
	HaiQiKeJi = core.HaiQiKeJi
	WeiBan    = core.WeiBan
	CELA      = core.CELA
	MOOC      = core.MOOC
	ZhiHuiShu = core.ZhiHuiShu
	GongXue   = core.GongXue
)

// Capability is a bit set describing adapter support.
type Capability = core.Capability

const (
	CapLogin       = core.CapLogin
	CapCookieLogin = core.CapCookieLogin
	CapCourses     = core.CapCourses
	CapTree        = core.CapTree
	CapStudy       = core.CapStudy
	CapWorks       = core.CapWorks
	CapExams       = core.CapExams
	CapSignIn      = core.CapSignIn
	CapQRCodeLogin = core.CapQRCodeLogin
)

// Account carries credentials for one upstream account.
type Account = core.Account

// Session is a snapshot of authenticated state.
type Session = core.Session

// AIType mirrors the legacy ctype.AiType values.
type AIType = core.AIType

const (
	AIChatGLM  = core.AIChatGLM
	AIXingHuo  = core.AIXingHuo
	AITongYi   = core.AITongYi
	AIDouBao   = core.AIDouBao
	AIOpenAI   = core.AIOpenAI
	AIDeepSeek = core.AIDeepSeek
	AIMetaAi   = core.AIMetaAi
	AISilicon  = core.AISilicon
	AIOther    = core.AIOther
)

// AIConfig configures the AI answer backend.
type AIConfig = core.AIConfig

// QuestionBankConfig configures an external answer source.
type QuestionBankConfig = core.QuestionBankConfig

// Logger is the SDK logging contract.
type Logger = core.Logger

// Options controls runtime behaviour shared by all adapters.
type Options = core.Options

// Config is the complete input to New.
type Config = core.Config

// NopLogger returns a logger that discards every message.
func NopLogger() Logger { return core.NopLogger() }

// CourseStatus is the normalized lifecycle state of a course.
type CourseStatus = core.CourseStatus

const (
	CourseStatusUnknown    = core.CourseStatusUnknown
	CourseStatusNotStarted = core.CourseStatusNotStarted
	CourseStatusActive     = core.CourseStatusActive
	CourseStatusFinished   = core.CourseStatusFinished
)

// Course is a platform-neutral course summary.
type Course = core.Course

// NodeKind classifies a node inside a course tree.
type NodeKind = core.NodeKind

const (
	NodeChapter  = core.NodeChapter
	NodeVideo    = core.NodeVideo
	NodeAudio    = core.NodeAudio
	NodeDocument = core.NodeDocument
	NodeLive     = core.NodeLive
	NodeBBS      = core.NodeBBS
	NodeLink     = core.NodeLink
	NodeWork     = core.NodeWork
	NodeExam     = core.NodeExam
	NodeUnknown  = core.NodeUnknown
)

// Node is one study unit.
type Node = core.Node

// Work is a platform-neutral assignment summary.
type Work = core.Work

// Exam is a platform-neutral exam summary.
type Exam = core.Exam

// Answer is one answered question.
type Answer = core.Answer

// AnswerResult is returned by DoWork and DoExam.
type AnswerResult = core.AnswerResult

// StudyMode controls how a study tick is submitted.
type StudyMode = core.StudyMode

const (
	StudyModeNormal = core.StudyModeNormal
	StudyModeFast   = core.StudyModeFast
	StudyModeWeb    = core.StudyModeWeb
)

// StudyAction describes one step of a study session.
type StudyAction = core.StudyAction

const (
	StudyActionStart    = core.StudyActionStart
	StudyActionProgress = core.StudyActionProgress
	StudyActionPause    = core.StudyActionPause
	StudyActionFinish   = core.StudyActionFinish
)

// StudyRequest submits one study tick for a single node.
type StudyRequest = core.StudyRequest

// StudyResult reports the effect of one study tick.
type StudyResult = core.StudyResult

// AnswerSource selects how work and exam questions are answered.
type AnswerSource = core.AnswerSource

const (
	AnswerSourceAI      = core.AnswerSourceAI
	AnswerSourceBank    = core.AnswerSourceBank
	AnswerSourceBuiltin = core.AnswerSourceBuiltin
	AnswerSourceManual  = core.AnswerSourceManual
)

// WorkRequest asks an adapter to complete one assignment.
type WorkRequest = core.WorkRequest

// ExamRequest asks an adapter to complete one exam.
type ExamRequest = core.ExamRequest

// SignInRequest is used by platforms that expose attendance.
type SignInRequest = core.SignInRequest

// SignInResult reports the attendance outcome.
type SignInResult = core.SignInResult

// Client is the single public contract implemented by every adapter.
type Client = core.Client

// QRCodeLogin is implemented by platforms that use an interactive QR login.
type QRCodeLogin = core.QRCodeLogin

// Factory constructs a Client for one platform.
type Factory = core.Factory

// Action names the dynamic operations understood by Client.Call.
type Action = core.Action

const (
	ActionLogin           = core.ActionLogin
	ActionLoginWithCookie = core.ActionLoginWithCookie
	ActionSession         = core.ActionSession
	ActionCourses         = core.ActionCourses
	ActionTree            = core.ActionTree
	ActionStudy           = core.ActionStudy
	ActionWorks           = core.ActionWorks
	ActionDoWork          = core.ActionDoWork
	ActionExams           = core.ActionExams
	ActionDoExam          = core.ActionDoExam
	ActionSignIn          = core.ActionSignIn
)

// Error is the SDK's structured error type.
type Error = core.Error
