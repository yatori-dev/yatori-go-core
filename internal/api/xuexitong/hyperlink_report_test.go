package xuexitong

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type hyperlinkTestTransport func(*http.Request) (*http.Response, error)

func (f hyperlinkTestTransport) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestHyperlinkAttachmentDetection(t *testing.T) {
	for _, tt := range []struct {
		name               string
		attachment         interface{}
		wantMatch, wantErr bool
	}{
		{"matched_alias", map[string]interface{}{"attachments": []interface{}{
			map[string]interface{}{"jobid": "other", "jtoken": "wrong"},
			map[string]interface{}{"jobid": "link/1&2", "jtoken": "token+&=/x", "property": map[string]interface{}{"_jobid": "iframe-job", "name": "External reading"}},
		}}, true, false},
		{"unmatched", map[string]interface{}{"attachments": []interface{}{map[string]interface{}{"jobid": "other", "jtoken": "wrong"}}}, false, false},
		{"malformed_item", map[string]interface{}{"attachments": []interface{}{42, nil}}, false, false},
		{"bad_root", "error", false, true},
		{"missing_attachments", map[string]interface{}{}, false, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p := PointHyperlinkDto{JobID: "iframe-job", Title: "Original", Jtoken: "original-token"}
			before := p
			matched, err := p.AttachmentsDetection(tt.attachment)
			if matched != tt.wantMatch || (err != nil) != tt.wantErr {
				t.Fatalf("got %t, %v", matched, err)
			}
			if matched {
				if p.JobID != "link/1&2" || p.Jtoken != "token+&=/x" || p.Title != "External reading" {
					t.Fatalf("wrong attachment: %+v", p)
				}
			} else if p != before {
				t.Fatal("unmatched attachment changed the point")
			}
		})
	}
}

func TestHyperlinkAttachmentNilReceiver(t *testing.T) {
	var point *PointHyperlinkDto
	if matched, err := point.AttachmentsDetection(map[string]interface{}{"attachments": []interface{}{}}); matched || err == nil {
		t.Fatalf("nil receiver: matched=%t, err=%v", matched, err)
	}
}

func TestHyperlinkReport(t *testing.T) {
	transportErr := errors.New("test transport failure")
	for _, tt := range []struct {
		name, body string
		status     int
		err        error
	}{
		{"accepted", `{"status":true}`, 200, nil},
		{"rejected", `{"status":false,"msg":"invalid parameters"}`, 200, nil},
		{"http_error", "", 503, nil},
		{"transport_error", "", 0, transportErr},
	} {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			client := &http.Client{Transport: hyperlinkTestTransport(func(req *http.Request) (*http.Response, error) {
				calls++
				if req.URL.Scheme != "https" || req.URL.Host != "mooc1.chaoxing.com" || req.URL.Path != "/ananas/job/hyperlink" || req.Method != "GET" {
					t.Fatalf("wrong request: %s %s", req.Method, req.URL)
				}
				for key, want := range map[string]string{"jobid": "link/1&2", "knowledgeid": "7", "courseid": "2", "clazzid": "3", "jtoken": "token+&=/x", "checkMicroTopic": "true", "microTopicId": "undefined"} {
					if got := req.URL.Query().Get(key); got != want {
						t.Errorf("%s = %q, want %q", key, got, want)
					}
				}
				if req.URL.Query().Get("_dc") == "" {
					t.Error("missing timestamp")
				}
				if tt.err != nil {
					return nil, tt.err
				}
				return &http.Response{StatusCode: tt.status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(tt.body)), Request: req}, nil
			})}
			point := &PointHyperlinkDto{JobID: "link/1&2", KnowledgeID: 7, CourseID: "2", ClassID: "3", Jtoken: "token+&=/x"}
			cache := &XueXiTUserCache{}
			body, err := cache.hyperlinkDtoCompleteReport(point, 0, nil, client)
			if calls != 1 {
				t.Fatalf("requests = %d", calls)
			}
			if tt.err != nil {
				if !errors.Is(err, tt.err) {
					t.Fatalf("transport error lost: %v", err)
				}
			} else if tt.status != 200 {
				if err == nil {
					t.Fatal("HTTP failure accepted")
				}
			} else if err != nil || body != tt.body {
				t.Fatalf("receipt changed: %q, %v", body, err)
			}
		})
	}
}

func TestHyperlinkReportMissingIdentifiers(t *testing.T) {
	client := &http.Client{Transport: hyperlinkTestTransport(func(*http.Request) (*http.Response, error) {
		t.Fatal("invalid point caused a request")
		return nil, nil
	})}
	for _, point := range []*PointHyperlinkDto{nil, {}, {JobID: "link"}} {
		if _, err := (&XueXiTUserCache{}).hyperlinkDtoCompleteReport(point, 0, nil, client); err == nil {
			t.Fatal("missing identifiers accepted")
		}
	}
}
