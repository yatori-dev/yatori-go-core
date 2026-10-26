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

func TestParseStartQuestionsTextTypes(t *testing.T) {
	jsonStr := `{
		"code": 200,
		"data": {
			"workTopics": [
				{"id": 3001, "recordId": 9001, "type": 4, "topic": "简述TCP三次握手"},
				{"id": 3002, "recordId": 9001, "type": 5, "topic": "1+1=____"},
				{"id": 3003, "recordId": 9001, "type": 1, "topic": "没有选项的单选题"}
			]
		}
	}`
	topics := parseStartQuestions(jsonStr)
	if len(topics) != 2 {
		t.Fatalf("expected 2 topics (text types kept, choice without options dropped), got %d", len(topics))
	}
	if topics[0].TopicId != "3001" || topics[0].TypeInt != 4 || topics[0].Question.Type != qtype.ShortAnswer.String() {
		t.Fatalf("unexpected short answer topic: %#v", topics[0])
	}
	if topics[1].TopicId != "3002" || topics[1].TypeInt != 5 || topics[1].Question.Type != qtype.FillInTheBlank.String() {
		t.Fatalf("unexpected fill blank topic: %#v", topics[1])
	}
}

func TestParseStartQuestionsBuiltinText(t *testing.T) {
	jsonStr := `{
		"code": 200,
		"data": {
			"workTopics": [
				{"id": 4001, "recordId": 9001, "type": 4, "topic": "简述题", "referenceAnswer": "参考答案A"},
				{"id": 4002, "recordId": 9001, "type": 5, "topic": "填空____", "option": [
					{"idx": "a1", "answer": "第一空答案"},
					{"idx": "a2", "answer": "第二空答案"}
				]}
			]
		}
	}`
	topics := parseStartQuestions(jsonStr)
	if len(topics) != 2 {
		t.Fatalf("expected 2 topics, got %d", len(topics))
	}
	if got := hqkjBuiltinAnswers(topics[0]); !reflect.DeepEqual(got, []string{"参考答案A"}) {
		t.Fatalf("short answer builtin mismatch: %#v", got)
	}
	if got := hqkjBuiltinAnswers(topics[1]); !reflect.DeepEqual(got, []string{"第一空答案", "第二空答案"}) {
		t.Fatalf("fill blank builtin mismatch: %#v", got)
	}
	// 填空题的 option[] 是参考答案，不能当选项塞给 AI
	if len(topics[1].Question.Options) != 0 {
		t.Fatalf("fill blank options should be empty, got %#v", topics[1].Question.Options)
	}
}

func TestExamAutoSubmitSemantics(t *testing.T) {
	cases := []struct {
		value int
		want  bool
	}{
		{0, false}, // 只保存
		{1, true},  // 正式提交/交卷
		{2, false},
	}
	for _, tc := range cases {
		setting := HqkjAnswerSetting{ExamAutoSubmit: tc.value}
		if got := setting.shouldSubmit(); got != tc.want {
			t.Fatalf("examAutoSubmit=%d: want %v, got %v", tc.value, tc.want, got)
		}
	}
}

func TestParseStartRecordId(t *testing.T) {
	cases := []struct {
		name string
		json string
		want string
	}{
		{"workTopics", `{"data":{"workTopics":[{"recordId":2002}]}}`, "2002"},
		{"examTopics string id", `{"data":{"examTopics":[{"recordId":"3003"}]}}`, "3003"},
		{"float string id", `{"data":{"recordId":"4004.0"}}`, "4004"},
		{"data recordId", `{"data":{"recordId":4004}}`, "4004"},
		{"top recordId string", `{"recordId":"5005"}`, "5005"},
		{"data id", `{"data":{"id":6006}}`, "6006"},
		{"top id", `{"id":7007}`, "7007"},
		{"missing", `{"data":{}}`, ""},
		{"zero", `{"data":{"recordId":0}}`, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := parseStartRecordId(tc.json); got != tc.want {
				t.Fatalf("want %q, got %q", tc.want, got)
			}
		})
	}
}

func TestHqkjBlankAnswers(t *testing.T) {
	got := hqkjBlankAnswers([]string{"第一空", "第二空"})
	want := map[string]string{"a1": "第一空", "a2": "第二空"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("want %#v, got %#v", want, got)
	}
}

func TestMapHqkjType(t *testing.T) {
	cases := map[int]string{
		1: qtype.SingleChoice.String(),
		2: qtype.MultipleChoice.String(),
		3: qtype.TrueOrFalse.String(),
		4: qtype.ShortAnswer.String(),
		5: qtype.FillInTheBlank.String(),
	}
	for code, want := range cases {
		if got := mapHqkjType(code); got != want {
			t.Fatalf("type %d: want %q, got %q", code, want, got)
		}
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
