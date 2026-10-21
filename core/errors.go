package core

import (
	"errors"
	"fmt"
)

var (
	// ErrUnsupported means the platform does not implement the operation.
	ErrUnsupported = errors.New("yatori: operation not supported by platform")
	// ErrNotLoggedIn means an authenticated session is required first.
	ErrNotLoggedIn = errors.New("yatori: not logged in")
	// ErrInvalidArgument means the caller supplied an unusable value.
	ErrInvalidArgument = errors.New("yatori: invalid argument")
	// ErrNotFound means the requested course, node or task was not found.
	ErrNotFound = errors.New("yatori: not found")
)

// Error is the SDK's structured error type.
type Error struct {
	Platform Platform
	Op       string
	Code     string
	Message  string
	Err      error
}

// Error implements the error interface.
func (e *Error) Error() string {
	parts := make([]string, 0, 5)
	if e.Platform != "" {
		parts = append(parts, string(e.Platform))
	}
	if e.Op != "" {
		parts = append(parts, e.Op)
	}
	text := ""
	for _, part := range parts {
		if text != "" {
			text += ": "
		}
		text += part
	}
	if e.Message != "" {
		if text != "" {
			text += ": "
		}
		text += e.Message
	}
	if e.Err != nil {
		if text != "" {
			text += ": "
		}
		text += e.Err.Error()
	}
	if text == "" {
		return "yatori: error"
	}
	if len(parts) == 0 {
		return text
	}
	return "yatori: " + text
}

// Unwrap returns the underlying cause.
func (e *Error) Unwrap() error { return e.Err }

// Wrap annotates err with platform and operation context. It returns nil
// when err is nil.
func Wrap(platform Platform, op string, err error) error {
	if err == nil {
		return nil
	}
	var existing *Error
	if errors.As(err, &existing) {
		return err
	}
	return &Error{Platform: platform, Op: op, Err: err}
}

// Unsupported builds an ErrUnsupported error for the given operation.
func Unsupported(platform Platform, op string) error {
	return &Error{
		Platform: platform,
		Op:       op,
		Code:     "unsupported",
		Err:      ErrUnsupported,
	}
}

// Invalid builds an ErrInvalidArgument error.
func Invalid(platform Platform, op, message string) error {
	return &Error{
		Platform: platform,
		Op:       op,
		Code:     "invalid_argument",
		Message:  message,
		Err:      ErrInvalidArgument,
	}
}

// NotLoggedIn builds an ErrNotLoggedIn error.
func NotLoggedIn(platform Platform, op string) error {
	return &Error{
		Platform: platform,
		Op:       op,
		Code:     "not_logged_in",
		Err:      ErrNotLoggedIn,
	}
}

// NotFound builds an ErrNotFound error.
func NotFound(platform Platform, op, message string) error {
	return &Error{
		Platform: platform,
		Op:       op,
		Code:     "not_found",
		Message:  message,
		Err:      ErrNotFound,
	}
}

// IsUnsupported reports whether err wraps ErrUnsupported.
func IsUnsupported(err error) bool { return errors.Is(err, ErrUnsupported) }

// IsNotLoggedIn reports whether err wraps ErrNotLoggedIn.
func IsNotLoggedIn(err error) bool { return errors.Is(err, ErrNotLoggedIn) }

// IsInvalidArgument reports whether err wraps ErrInvalidArgument.
func IsInvalidArgument(err error) bool { return errors.Is(err, ErrInvalidArgument) }

// Errorf builds a formatted Error with an underlying cause.
func Errorf(platform Platform, op, format string, args ...any) error {
	return &Error{Platform: platform, Op: op, Message: fmt.Sprintf(format, args...)}
}
