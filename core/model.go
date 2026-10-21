package core

import "time"

// CourseStatus is the normalized lifecycle state of a course.
type CourseStatus string

const (
	CourseStatusUnknown    CourseStatus = ""
	CourseStatusNotStarted CourseStatus = "not_started"
	CourseStatusActive     CourseStatus = "active"
	CourseStatusFinished   CourseStatus = "finished"
)

// Course is a platform-neutral course summary.
//
// Extra holds platform identifiers required by Tree and Study. Raw keeps
// the original platform object for callers that need full fidelity. Raw is
// only valid in the process that produced it and must not be serialized.
type Course struct {
	ID       string
	Name     string
	Teacher  string
	Cover    string
	Progress float64
	Status   CourseStatus
	Extra    map[string]string
	Raw      any
}

// ExtraValue returns a platform specific value or an empty string.
func (c Course) ExtraValue(key string) string {
	if c.Extra == nil {
		return ""
	}
	return c.Extra[key]
}

// NodeKind classifies a node inside a course tree.
type NodeKind string

const (
	NodeChapter  NodeKind = "chapter"
	NodeVideo    NodeKind = "video"
	NodeAudio    NodeKind = "audio"
	NodeDocument NodeKind = "document"
	NodeLive     NodeKind = "live"
	NodeBBS      NodeKind = "bbs"
	NodeLink     NodeKind = "link"
	NodeWork     NodeKind = "work"
	NodeExam     NodeKind = "exam"
	NodeUnknown  NodeKind = "unknown"
)

// Node is one study unit. A node either carries children (a chapter or a
// module) or is a leaf that Study can process.
type Node struct {
	ID        string
	ParentID  string
	CourseID  string
	Name      string
	Kind      NodeKind
	Duration  time.Duration
	Progress  float64
	Completed bool
	Children  []Node
	Extra     map[string]string
	Raw       any
}

// ExtraValue returns a platform specific value or an empty string.
func (n Node) ExtraValue(key string) string {
	if n.Extra == nil {
		return ""
	}
	return n.Extra[key]
}

// Walk visits every node depth first. Returning false from fn stops the walk.
func (n Node) Walk(fn func(Node) bool) bool {
	if !fn(n) {
		return false
	}
	for _, child := range n.Children {
		if !child.Walk(fn) {
			return false
		}
	}
	return true
}

// Flatten returns all leaf nodes in depth first order.
func Flatten(nodes []Node) []Node {
	out := make([]Node, 0, len(nodes))
	for _, node := range nodes {
		node.Walk(func(current Node) bool {
			if len(current.Children) == 0 {
				out = append(out, current)
			}
			return true
		})
	}
	return out
}

// Work is a platform-neutral assignment summary.
type Work struct {
	ID            string
	CourseID      string
	Name          string
	Status        string
	Deadline      time.Time
	QuestionCount int
	Extra         map[string]string
	Raw           any
}

// Exam is a platform-neutral exam summary.
type Exam struct {
	ID            string
	CourseID      string
	Name          string
	Status        string
	StartAt       time.Time
	EndAt         time.Time
	Duration      time.Duration
	QuestionCount int
	Extra         map[string]string
	Raw           any
}

// Answer is one answered question.
type Answer struct {
	QuestionID string
	Question   string
	Type       string
	Options    []string
	Value      []string
	Correct    bool
	Raw        any
}

// AnswerResult is returned by DoWork and DoExam.
type AnswerResult struct {
	TaskID    string
	Title     string
	Submitted bool
	Score     float64
	Answers   []Answer
	Message   string
	Raw       any
}
