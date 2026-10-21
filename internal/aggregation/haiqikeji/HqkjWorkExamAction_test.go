package haiqikeji

import (
	"reflect"
	"testing"

	"github.com/yatori-dev/yatori-go-core/que-core/qentity"
	"github.com/yatori-dev/yatori-go-core/que-core/qtype"
)

func TestParseStartQuestionsBuiltinAnswer(t *testing.T) {
	jsonStr := `{
		"code": 200,
		"data": {
			"workTopics": [
				{
					"id": 1001,
					"recordId": 2002,
					"type": 1,
					"topic": "<p>1 + 1 = ?</p>",
					"option": [
						{"idx": "A", "answer": "1", "scale": 0},
						{"idx": "B", "answer": "2", "scale": 100}
					]
				},
				{
					"id": 1002,
					"recordId": 2002,
					"type": 2,
					"topic": "以下哪些是偶数？",
					"option": [
						{"idx": "A", "answer": "2", "scale": 100},
						{"idx": "B", "answer": "3", "scale": 0},
						{"idx": "C", "answer": "4", "scale": 100}
					]
				}
			]
		}
	}`

	topics := parseStartQuestions(jsonStr)
	if len(topics) != 2 {
		t.Fatalf("expected 2 topics, got %d", len(topics))
	}
	if got := hqkjBuiltinAnswers(topics[0]); !reflect.DeepEqual(got, []string{"B"}) {
		t.Fatalf("single choice answer mismatch: %#v", got)
	}
	if got := hqkjBuiltinAnswers(topics[1]); !reflect.DeepEqual(got, []string{"A", "C"}) {
		t.Fatalf("multiple choice answer mismatch: %#v", got)
	}
}

func TestParseQuizListStatus(t *testing.T) {
	jsonStr := `{
		"code": 200,
		"data": {
			"examInfo": [
				{"examId": 1008815, "title": "期末考试", "examStatus": 1, "frequency": 1, "score": 100}
			],
			"workInfo": [
				{"workId": 2001, "title": "作业一", "workStatus": 2, "frequency": 1}
			]
		}
	}`

	exams := parseQuizList(jsonStr, "examInfo", "examId")
	if len(exams) != 1 {
		t.Fatalf("expected 1 exam, got %d", len(exams))
	}
	if exams[0].Status != 1 || exams[0].Frequency != 1 || exams[0].AchievedScore != 100 {
		t.Fatalf("unexpected exam fields: %#v", exams[0])
	}
	if exams[0].completed() {
		t.Fatalf("in-progress exam should not be completed")
	}

	works := parseQuizList(jsonStr, "workInfo", "workId")
	if len(works) != 1 {
		t.Fatalf("expected 1 work, got %d", len(works))
	}
	if works[0].Status != 2 || !works[0].completed() {
		t.Fatalf("unexpected work fields: %#v", works[0])
	}
}

func TestHqkjQuizCompleted(t *testing.T) {
	cases := []struct {
		name string
		quiz hqkjQuiz
		want bool
	}{
		{"exam not started", hqkjQuiz{Status: 0, Frequency: 0}, false},
		{"exam in progress", hqkjQuiz{Status: 1, Frequency: 1}, false},
		{"exam finished", hqkjQuiz{Status: 2, Frequency: 1}, true},
		{"legacy work frequency", hqkjQuiz{Status: 0, Frequency: 1}, true},
		{"work in progress", hqkjQuiz{Status: 1, Frequency: 1}, false},
		{"work finished", hqkjQuiz{Status: 2, Frequency: 1}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.quiz.completed(); got != tc.want {
				t.Fatalf("want %v, got %v", tc.want, got)
			}
		})
	}
}

func TestHqkjBuiltinAnswers(t *testing.T) {
	cases := []struct {
		name  string
		topic hqkjTopic
		want  []string
	}{
		{
			name: "single choice keeps first",
			topic: hqkjTopic{
				Question:   qentity.Question{Type: qtype.SingleChoice.String()},
				CorrectIdx: []string{"B", "C"},
			},
			want: []string{"B"},
		},
		{
			name: "true false keeps first",
			topic: hqkjTopic{
				Question:   qentity.Question{Type: qtype.TrueOrFalse.String()},
				CorrectIdx: []string{"A"},
			},
			want: []string{"A"},
		},
		{
			name: "multiple choice deduplicates",
			topic: hqkjTopic{
				Question:   qentity.Question{Type: qtype.MultipleChoice.String()},
				CorrectIdx: []string{"A", "A", "", "C"},
			},
			want: []string{"A", "C"},
		},
		{
			name: "no builtin answer",
			topic: hqkjTopic{
				Question: qentity.Question{Type: qtype.FillInTheBlank.String()},
			},
			want: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := hqkjBuiltinAnswers(tc.topic)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("want %#v, got %#v", tc.want, got)
			}
		})
	}
}
