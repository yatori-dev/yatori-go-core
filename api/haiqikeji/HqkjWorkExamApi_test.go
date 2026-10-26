package haiqikeji

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestFinishWorkAnswerApiParams(t *testing.T) {
	var gotPath, gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"code":200}`))
	}))
	defer srv.Close()

	cache := &HqkjUserCache{PreUrl: srv.URL, SchoolId: "11", UserId: "22", Token: "token"}
	res, err := cache.FinishWorkAnswerApi("33", "44", "55", 0, nil)
	if err != nil {
		t.Fatalf("FinishWorkAnswerApi error: %v", err)
	}
	if res != `{"code":200}` {
		t.Fatalf("unexpected response: %s", res)
	}
	if gotPath != "/api/user/yee_work_answer_finish" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	q, err := url.ParseQuery(gotQuery)
	if err != nil {
		t.Fatalf("parse query: %v", err)
	}
	want := map[string]string{
		"schoolId":  "11",
		"studentId": "22",
		"userId":    "22",
		"workId":    "44",
		"state":     "2",
		"recordId":  "55",
		"courseId":  "33",
	}
	for k, v := range want {
		if got := q.Get(k); got != v {
			t.Fatalf("query %s=%q, want %q (raw: %s)", k, got, v, gotQuery)
		}
	}
}

func TestSubmitWorkAnswerBlankApiPayload(t *testing.T) {
	var payload map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/user/yee_work_answer_addBlank" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("decode payload: %v", err)
		}
		_, _ = w.Write([]byte(`{"code":200}`))
	}))
	defer srv.Close()

	cache := &HqkjUserCache{PreUrl: srv.URL, SchoolId: "11", UserId: "22", Token: "token"}
	if _, err := cache.SubmitWorkAnswerBlankApi("33", "44", "55", "66", "5", map[string]string{"a1": "第一空"}, 0, nil); err != nil {
		t.Fatalf("SubmitWorkAnswerBlankApi error: %v", err)
	}
	if payload["workId"] != float64(44) || payload["type"] != float64(5) {
		t.Fatalf("unexpected payload: %#v", payload)
	}
	answer, ok := payload["answer"].(map[string]interface{})
	if !ok || answer["a1"] != "第一空" {
		t.Fatalf("unexpected blank answer: %#v", payload["answer"])
	}
}

func TestSubmitExamAnswerTextApiPayload(t *testing.T) {
	var payload map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/user/yee_exam_answer_addText" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("decode payload: %v", err)
		}
		_, _ = w.Write([]byte(`{"code":200}`))
	}))
	defer srv.Close()

	cache := &HqkjUserCache{PreUrl: srv.URL, SchoolId: "11", UserId: "22", Token: "token"}
	if _, err := cache.SubmitExamAnswerTextApi("33", "44", "55", "66", "4", "主观题答案", 0, nil); err != nil {
		t.Fatalf("SubmitExamAnswerTextApi error: %v", err)
	}
	if payload["examId"] != float64(44) || payload["type"] != float64(4) {
		t.Fatalf("unexpected payload: %#v", payload)
	}
	if _, exists := payload["workId"]; exists {
		t.Fatalf("exam payload must not contain workId: %#v", payload)
	}
	if payload["answer"] != "主观题答案" {
		t.Fatalf("unexpected text answer: %#v", payload["answer"])
	}
}

func TestAnswerApiRejectsInvalidRecordId(t *testing.T) {
	cache := &HqkjUserCache{PreUrl: "http://127.0.0.1:1", SchoolId: "1", UserId: "2"}
	if _, err := cache.SubmitWorkAnswerApi("1", "2", "3", "0", "1", []string{"A"}, 0, nil); err == nil {
		t.Fatal("expected error for recordId=0")
	}
	if _, err := cache.FinishExamAnswerApi("1", "2", "", 0, nil); err == nil {
		t.Fatal("expected error for empty recordId")
	}
}
