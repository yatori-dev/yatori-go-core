package core

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Action names the dynamic operations understood by Client.Call.
type Action string

const (
	ActionLogin           Action = "login"
	ActionLoginWithCookie Action = "login_with_cookie"
	ActionSession         Action = "session"
	ActionCourses         Action = "courses"
	ActionTree            Action = "tree"
	ActionStudy           Action = "study"
	ActionWorks           Action = "works"
	ActionDoWork          Action = "do_work"
	ActionExams           Action = "exams"
	ActionDoExam          Action = "do_exam"
	ActionSignIn          Action = "sign_in"
)

// ParseAction normalizes an action name.
func ParseAction(raw string) (Action, error) {
	value := Action(strings.ToLower(strings.TrimSpace(raw)))
	switch value {
	case ActionLogin, ActionLoginWithCookie, ActionSession, ActionCourses,
		ActionTree, ActionStudy, ActionWorks, ActionDoWork, ActionExams,
		ActionDoExam, ActionSignIn:
		return value, nil
	}
	return "", fmt.Errorf("yatori: unknown action %q", raw)
}

// Call dispatches a dynamic action to a client. It exists so adapters can
// share one implementation of Client.Call.
func Call(ctx context.Context, client Client, action Action, params map[string]any) (any, error) {
	if client == nil {
		return nil, Invalid("", "call", "nil client")
	}
	if params == nil {
		params = map[string]any{}
	}
	switch action {
	case ActionLogin:
		return nil, client.Login(ctx)
	case ActionLoginWithCookie:
		cookie := paramString(params, "cookie")
		if cookie == "" {
			return nil, Invalid(client.Platform(), string(action), "cookie is required")
		}
		return nil, client.LoginWithCookie(ctx, cookie)
	case ActionSession:
		return client.Session(), nil
	case ActionCourses:
		return client.Courses(ctx)
	case ActionTree:
		course, err := paramCourse(client.Platform(), params)
		if err != nil {
			return nil, err
		}
		return client.Tree(ctx, course)
	case ActionStudy:
		course, err := paramCourse(client.Platform(), params)
		if err != nil {
			return nil, err
		}
		node, err := paramNode(client.Platform(), params, course.ID)
		if err != nil {
			return nil, err
		}
		req := StudyRequest{
			Course:   course,
			Node:     node,
			Action:   StudyAction(paramStringDefault(params, "studyAction", string(StudyActionProgress))),
			Mode:     StudyMode(paramStringDefault(params, "mode", string(StudyModeNormal))),
			Position: paramDuration(params, "position"),
			Duration: paramDuration(params, "duration"),
			IsDrag:   paramBool(params, "isDrag"),
			Extra:    paramMap(params, "extra"),
		}
		return client.Study(ctx, req)
	case ActionWorks:
		course, err := paramCourse(client.Platform(), params)
		if err != nil {
			return nil, err
		}
		return client.Works(ctx, course)
	case ActionDoWork:
		course, err := paramCourse(client.Platform(), params)
		if err != nil {
			return nil, err
		}
		work, err := paramWork(client.Platform(), params, course.ID)
		if err != nil {
			return nil, err
		}
		return client.DoWork(ctx, WorkRequest{
			Course: course,
			Work:   work,
			Source: AnswerSource(paramStringDefault(params, "source", string(AnswerSourceAI))),
			Submit: paramBoolDefault(params, "submit", true),
			Extra:  paramMap(params, "extra"),
		})
	case ActionExams:
		course, err := paramCourse(client.Platform(), params)
		if err != nil {
			return nil, err
		}
		return client.Exams(ctx, course)
	case ActionDoExam:
		course, err := paramCourse(client.Platform(), params)
		if err != nil {
			return nil, err
		}
		exam, err := paramExam(client.Platform(), params, course.ID)
		if err != nil {
			return nil, err
		}
		return client.DoExam(ctx, ExamRequest{
			Course: course,
			Exam:   exam,
			Source: AnswerSource(paramStringDefault(params, "source", string(AnswerSourceAI))),
			Submit: paramBoolDefault(params, "submit", true),
			Extra:  paramMap(params, "extra"),
		})
	case ActionSignIn:
		return client.SignIn(ctx, SignInRequest{
			UserID:   paramString(params, "userId"),
			Location: paramString(params, "location"),
			Remark:   paramString(params, "remark"),
			Extra:    paramMap(params, "extra"),
		})
	default:
		return nil, fmt.Errorf("yatori: unknown action %q", action)
	}
}

func paramString(params map[string]any, key string) string {
	value, ok := params[key]
	if !ok || value == nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprint(v)
	}
}

func paramStringDefault(params map[string]any, key, fallback string) string {
	if value := paramString(params, key); value != "" {
		return value
	}
	return fallback
}

func paramBool(params map[string]any, key string) bool {
	value, ok := params[key]
	if !ok || value == nil {
		return false
	}
	switch v := value.(type) {
	case bool:
		return v
	case string:
		return strings.EqualFold(v, "true") || v == "1" || strings.EqualFold(v, "yes")
	case int:
		return v != 0
	case int64:
		return v != 0
	case float64:
		return v != 0
	default:
		return false
	}
}

func paramBoolDefault(params map[string]any, key string, fallback bool) bool {
	if _, ok := params[key]; !ok {
		return fallback
	}
	return paramBool(params, key)
}

func paramDuration(params map[string]any, key string) time.Duration {
	value, ok := params[key]
	if !ok || value == nil {
		return 0
	}
	switch v := value.(type) {
	case time.Duration:
		return v
	case int:
		return time.Duration(v) * time.Second
	case int64:
		return time.Duration(v) * time.Second
	case float64:
		return time.Duration(v * float64(time.Second))
	case string:
		if parsed, err := time.ParseDuration(v); err == nil {
			return parsed
		}
	}
	return 0
}

func paramMap(params map[string]any, key string) map[string]any {
	value, ok := params[key]
	if !ok || value == nil {
		return nil
	}
	if out, ok := value.(map[string]any); ok {
		return out
	}
	return nil
}

func paramCourse(platform Platform, params map[string]any) (Course, error) {
	if value, ok := params["course"]; ok && value != nil {
		switch v := value.(type) {
		case Course:
			return v, nil
		case *Course:
			if v != nil {
				return *v, nil
			}
		}
	}
	id := paramString(params, "courseId")
	if id == "" {
		return Course{}, Invalid(platform, "course", "course or courseId is required")
	}
	return Course{ID: id}, nil
}

func paramNode(platform Platform, params map[string]any, courseID string) (Node, error) {
	if value, ok := params["node"]; ok && value != nil {
		switch v := value.(type) {
		case Node:
			if v.CourseID == "" {
				v.CourseID = courseID
			}
			return v, nil
		case *Node:
			if v != nil {
				out := *v
				if out.CourseID == "" {
					out.CourseID = courseID
				}
				return out, nil
			}
		}
	}
	id := paramString(params, "nodeId")
	if id == "" {
		return Node{}, Invalid(platform, "node", "node or nodeId is required")
	}
	return Node{ID: id, CourseID: courseID, Name: paramString(params, "nodeName")}, nil
}

func paramWork(platform Platform, params map[string]any, courseID string) (Work, error) {
	if value, ok := params["work"]; ok && value != nil {
		switch v := value.(type) {
		case Work:
			return v, nil
		case *Work:
			if v != nil {
				return *v, nil
			}
		}
	}
	id := paramString(params, "workId")
	if id == "" {
		return Work{}, Invalid(platform, "work", "work or workId is required")
	}
	return Work{ID: id, CourseID: courseID, Name: paramString(params, "workName")}, nil
}

func paramExam(platform Platform, params map[string]any, courseID string) (Exam, error) {
	if value, ok := params["exam"]; ok && value != nil {
		switch v := value.(type) {
		case Exam:
			return v, nil
		case *Exam:
			if v != nil {
				return *v, nil
			}
		}
	}
	id := paramString(params, "examId")
	if id == "" {
		return Exam{}, Invalid(platform, "exam", "exam or examId is required")
	}
	return Exam{ID: id, CourseID: courseID, Name: paramString(params, "examName")}, nil
}
