package yatori

import "github.com/yatori-dev/yatori-go-core/v3/core"

var (
	// ErrUnsupported means the platform does not implement the operation.
	ErrUnsupported = core.ErrUnsupported
	// ErrNotLoggedIn means an authenticated session is required first.
	ErrNotLoggedIn = core.ErrNotLoggedIn
	// ErrInvalidArgument means the caller supplied an unusable value.
	ErrInvalidArgument = core.ErrInvalidArgument
	// ErrNotFound means the requested course, node or task was not found.
	ErrNotFound = core.ErrNotFound
)

// IsUnsupported reports whether err wraps ErrUnsupported.
func IsUnsupported(err error) bool { return core.IsUnsupported(err) }

// IsNotLoggedIn reports whether err wraps ErrNotLoggedIn.
func IsNotLoggedIn(err error) bool { return core.IsNotLoggedIn(err) }

// IsInvalidArgument reports whether err wraps ErrInvalidArgument.
func IsInvalidArgument(err error) bool { return core.IsInvalidArgument(err) }
