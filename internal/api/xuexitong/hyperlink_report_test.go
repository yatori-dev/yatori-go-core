package xuexitong

import (
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"testing"
)

func TestHyperlinkReportUsesMatchingAttachmentAndAppSession(t *testing.T) {
	target, err := url.Parse("https://mooc1.chaoxing.com/ananas/job/hyperlink")
	if err != nil {
		t.Fatal(err)
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	jar.SetCookies(target, []*http.Cookie{{Name: "session", Value: "mock-app", Path: "/ananas"}})
	cache := XueXiTUserCache{
		appUserAgent: "mock app",
		cookies:      []*http.Cookie{{Name: "legacy", Value: "unused"}},
		appSession: &http.Client{
			Jar: jar,
			Transport: courseTestTransport(func(req *http.Request) (*http.Response, error) {
				if req.URL.Host != target.Host || req.URL.Path != target.Path || req.Method != "GET" {
					t.Fatalf("unexpected report URL or method: %s %s", req.Method, req.URL)
				}
				params := req.URL.Query()
				expected := map[string]string{"jobid": "link/1&2", "knowledgeid": "7", "courseid": "2", "clazzid": "3", "jtoken": "token+&=/x"}
				for key, want := range expected {
					if params.Get(key) != want {
						t.Errorf("%s = %q, want %q", key, params.Get(key), want)
					}
				}
				if params.Has("microTopicId") || params.Get("_dc") == "" {
					t.Error("wrong optional report parameters")
				}
				if req.UserAgent() != "mock app" || req.Header.Get("Cookie") != "session=mock-app" {
					t.Fatalf("App session not reused")
				}
				return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"status":true}`)), Request: req}, nil
			}),
		},
	}
	p := PointHyperlinkDto{JobID: "iframe-job", KnowledgeID: 7, CourseID: "2", ClassID: "3"}
	card := map[string]interface{}{"attachments": []interface{}{
		map[string]interface{}{"jobid": "other", "jtoken": "wrong"},
		map[string]interface{}{"jobid": "link/1&2", "jtoken": "token+&=/x", "property": map[string]interface{}{"_jobid": "iframe-job", "name": "External reading"}},
	}}
	matched, err := p.AttachmentsDetection(card)
	if err != nil || !matched || p.Title != "External reading" {
		t.Fatalf("attachment did not populate external reading: matched=%t, err=%v, title=%q", matched, err, p.Title)
	}
	if body, err := cache.HyperlinkDtoCompleteReport(&p, 0, nil); err != nil || body != `{"status":true}` {
		t.Fatalf("report response %q, %v", body, err)
	}
}

func TestHyperlinkReportDoesNotFollowExternalRedirect(t *testing.T) {
	calls := 0
	cache := XueXiTUserCache{appSession: &http.Client{Transport: courseTestTransport(func(req *http.Request) (*http.Response, error) {
		calls++
		return &http.Response{StatusCode: 302, Header: http.Header{"Location": []string{"https://external.example/read"}}, Body: io.NopCloser(strings.NewReader("")), Request: req}, nil
	})}}
	point := PointHyperlinkDto{JobID: "old-link", KnowledgeID: 7, CourseID: "2", ClassID: "3"}
	if _, err := cache.HyperlinkDtoCompleteReport(&point, 0, nil); err == nil || calls != 1 {
		t.Fatalf("external redirect not rejected: %v, requests: %d", err, calls)
	}
}
