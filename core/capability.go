package core

import "strings"

// Capability is a bit set describing what a platform adapter implements.
//
// Callers should check Capabilities before invoking optional operations.
// Unsupported operations return an error wrapping ErrUnsupported.
type Capability uint32

const (
	CapLogin Capability = 1 << iota
	CapCookieLogin
	CapCourses
	CapTree
	CapStudy
	CapWorks
	CapExams
	CapSignIn
	CapQRCodeLogin
)

var capabilityNames = []struct {
	flag Capability
	name string
}{
	{CapLogin, "login"},
	{CapCookieLogin, "cookie-login"},
	{CapCourses, "courses"},
	{CapTree, "tree"},
	{CapStudy, "study"},
	{CapWorks, "works"},
	{CapExams, "exams"},
	{CapSignIn, "sign-in"},
	{CapQRCodeLogin, "qr-code-login"},
}

// Has reports whether all bits in want are present.
func (c Capability) Has(want Capability) bool {
	return c&want == want
}

// String returns a comma separated capability list.
func (c Capability) String() string {
	if c == 0 {
		return "none"
	}
	parts := make([]string, 0, len(capabilityNames))
	for _, item := range capabilityNames {
		if c.Has(item.flag) {
			parts = append(parts, item.name)
		}
	}
	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, ",")
}
