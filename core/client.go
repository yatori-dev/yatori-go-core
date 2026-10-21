package core

import "context"

// Client is the single public contract implemented by every platform
// adapter. Operations that a platform cannot support return an error
// wrapping ErrUnsupported; use Capabilities to check up front.
//
// A Client owns one upstream session. Implementations serialize access to
// their underlying state, so a single Client should not be used to run many
// courses in parallel; create one Client per account instead.
type Client interface {
	// Platform returns the platform this client talks to.
	Platform() Platform
	// Capabilities returns the implemented capability bits.
	Capabilities() Capability

	// Login authenticates with the configured account.
	Login(ctx context.Context) error
	// LoginWithCookie imports an existing browser session. Platforms that
	// do not support cookie import return ErrUnsupported.
	LoginWithCookie(ctx context.Context, cookie string) error
	// Session returns a snapshot of the authenticated state.
	Session() Session

	// Courses lists the courses visible to the account.
	Courses(ctx context.Context) ([]Course, error)
	// Tree returns the chapter/node tree of a course.
	Tree(ctx context.Context, course Course) ([]Node, error)

	// Study submits one study tick for a node.
	Study(ctx context.Context, req StudyRequest) (*StudyResult, error)

	// Works lists assignments of a course.
	Works(ctx context.Context, course Course) ([]Work, error)
	// DoWork completes one assignment.
	DoWork(ctx context.Context, req WorkRequest) (*AnswerResult, error)

	// Exams lists exams of a course.
	Exams(ctx context.Context, course Course) ([]Exam, error)
	// DoExam completes one exam.
	DoExam(ctx context.Context, req ExamRequest) (*AnswerResult, error)

	// SignIn performs an attendance action. Only a few platforms expose it.
	SignIn(ctx context.Context, req SignInRequest) (*SignInResult, error)

	// Call dispatches a dynamic action. It is a thin, weakly typed layer
	// over the methods above and is intended for configuration driven
	// callers. Typed code should call the methods directly.
	Call(ctx context.Context, action Action, params map[string]any) (any, error)

	// Close releases resources held by the client.
	Close() error
}

// QRCodeLogin is implemented by platforms that authenticate through an
// interactive QR code scan. Callers can type assert a Client to this
// interface when CapQRCodeLogin is advertised.
type QRCodeLogin interface {
	// QRCode returns the QR image encoded as base64 and the token that
	// identifies the pending login.
	QRCode(ctx context.Context) (imageBase64 string, token string, err error)
	// WaitQRCode blocks until the QR code is scanned and confirmed.
	WaitQRCode(ctx context.Context, token string) (Session, error)
}

// Supports reports whether c implements every capability in want.
func Supports(c Client, want Capability) bool {
	if c == nil {
		return false
	}
	return c.Capabilities().Has(want)
}
