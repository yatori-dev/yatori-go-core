package core

import (
	"context"
	"testing"
	"time"
)

type fakeClient struct {
	lastStudy StudyRequest
}

func (f *fakeClient) Platform() Platform                            { return MOOC }
func (f *fakeClient) Capabilities() Capability                      { return CapCourses | CapTree | CapStudy }
func (f *fakeClient) Login(context.Context) error                   { return nil }
func (f *fakeClient) LoginWithCookie(context.Context, string) error { return nil }
func (f *fakeClient) Session() Session                              { return Session{Platform: MOOC, Account: "fake"} }
func (f *fakeClient) Courses(context.Context) ([]Course, error) {
	return []Course{{ID: "c1", Name: "course"}}, nil
}
func (f *fakeClient) Tree(_ context.Context, course Course) ([]Node, error) {
	return []Node{{ID: "n1", CourseID: course.ID, Name: "node"}}, nil
}
func (f *fakeClient) Study(_ context.Context, req StudyRequest) (*StudyResult, error) {
	f.lastStudy = req
	return &StudyResult{CourseID: req.Course.ID, NodeID: req.Node.ID, Submitted: true}, nil
}
func (f *fakeClient) Works(context.Context, Course) ([]Work, error) { return nil, ErrUnsupported }
func (f *fakeClient) DoWork(context.Context, WorkRequest) (*AnswerResult, error) {
	return nil, ErrUnsupported
}
func (f *fakeClient) Exams(context.Context, Course) ([]Exam, error) { return nil, ErrUnsupported }
func (f *fakeClient) DoExam(context.Context, ExamRequest) (*AnswerResult, error) {
	return nil, ErrUnsupported
}
func (f *fakeClient) SignIn(context.Context, SignInRequest) (*SignInResult, error) {
	return nil, ErrUnsupported
}
func (f *fakeClient) Call(ctx context.Context, action Action, params map[string]any) (any, error) {
	return Call(ctx, f, action, params)
}
func (f *fakeClient) Close() error { return nil }

func TestCallDispatch(t *testing.T) {
	client := &fakeClient{}
	ctx := context.Background()

	courses, err := Call(ctx, client, ActionCourses, nil)
	if err != nil {
		t.Fatalf("courses: %v", err)
	}
	if len(courses.([]Course)) != 1 || courses.([]Course)[0].ID != "c1" {
		t.Fatalf("unexpected courses: %#v", courses)
	}

	tree, err := Call(ctx, client, ActionTree, map[string]any{"courseId": "c1"})
	if err != nil {
		t.Fatalf("tree: %v", err)
	}
	if len(tree.([]Node)) != 1 || tree.([]Node)[0].ID != "n1" {
		t.Fatalf("unexpected tree: %#v", tree)
	}

	_, err = Call(ctx, client, ActionStudy, map[string]any{
		"courseId": "c1",
		"nodeId":   "n1",
		"position": "30s",
		"mode":     "fast",
	})
	if err != nil {
		t.Fatalf("study: %v", err)
	}
	if client.lastStudy.Course.ID != "c1" || client.lastStudy.Node.ID != "n1" {
		t.Fatalf("unexpected study request: %#v", client.lastStudy)
	}
	if client.lastStudy.Position != 30*time.Second {
		t.Fatalf("position = %v, want 30s", client.lastStudy.Position)
	}
	if client.lastStudy.Mode != StudyModeFast {
		t.Fatalf("mode = %v, want fast", client.lastStudy.Mode)
	}
}

func TestCallRejectsMissingArguments(t *testing.T) {
	_, err := Call(context.Background(), &fakeClient{}, ActionStudy, nil)
	if !IsInvalidArgument(err) {
		t.Fatalf("expected invalid argument, got %v", err)
	}
}
