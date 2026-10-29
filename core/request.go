package core

import "time"

// StudyMode controls how a study tick is submitted.
type StudyMode string

const (
	// StudyModeNormal submits progress as a normal client would.
	StudyModeNormal StudyMode = "normal"
	// StudyModeFast asks the adapter to use its fast path when available.
	StudyModeFast StudyMode = "fast"
	// StudyModeWeb asks the adapter to use the web endpoint when available.
	StudyModeWeb StudyMode = "web"
)

// StudyAction describes one step of a study session.
type StudyAction string

const (
	StudyActionStart    StudyAction = "start"
	StudyActionProgress StudyAction = "progress"
	StudyActionPause    StudyAction = "pause"
	StudyActionFinish   StudyAction = "finish"
)

// StudyRequest submits one study tick for a single node.
//
// Study is intentionally a primitive: the caller (or a runner) decides
// when to start, tick and finish. This keeps platform specific pacing and
// anti-risk-control logic out of the SDK core while still exposing a
// uniform API.
type StudyRequest struct {
	Course   Course
	Node     Node
	Action   StudyAction
	Mode     StudyMode
	Position time.Duration
	Duration time.Duration
	IsDrag   bool
	Extra    map[string]any
}

// StudyResult reports the effect of one study tick.
type StudyResult struct {
	CourseID  string
	NodeID    string
	Kind      NodeKind
	Submitted bool
	Position  time.Duration
	Progress  float64
	Completed bool
	Message   string
	Raw       any
}

// AnswerSource selects how work and exam questions are answered.
type AnswerSource string

const (
	AnswerSourceAI      AnswerSource = "ai"
	AnswerSourceBank    AnswerSource = "bank"
	AnswerSourceBuiltin AnswerSource = "builtin"
	AnswerSourceManual  AnswerSource = "manual"
)

// WorkRequest asks the adapter to complete one assignment.
type WorkRequest struct {
	Course  Course
	Work    Work
	Source  AnswerSource
	Submit  bool
	Answers []Answer
	Extra   map[string]any
}

// ExamRequest asks the adapter to complete one exam.
type ExamRequest struct {
	Course  Course
	Exam    Exam
	Source  AnswerSource
	Submit  bool
	Answers []Answer
	Extra   map[string]any
}

// SignInRequest is used by platforms that expose an attendance action.
type SignInRequest struct {
	UserID   string
	Location string
	Remark   string
	Extra    map[string]any
}

// SignInResult reports the attendance outcome.
type SignInResult struct {
	Success bool
	Message string
	Raw     any
}
