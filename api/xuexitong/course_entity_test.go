package xuexitong

import (
	"encoding/json"
	"testing"
)

func TestQuestionStringReturnsJSON(t *testing.T) {
	question := Question{Title: "章节测验", CourseId: "course-1"}

	var decoded map[string]interface{}
	if err := json.Unmarshal([]byte(question.String()), &decoded); err != nil {
		t.Fatalf("Question.String() returned invalid JSON: %v", err)
	}
	if decoded["title"] != question.Title {
		t.Fatalf("Question.String() title = %v, want %q", decoded["title"], question.Title)
	}
}
