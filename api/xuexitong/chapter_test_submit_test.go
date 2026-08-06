package xuexitong

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSendChapterWorkSubmissionValidatesResponse(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		wantBody   string
		wantErr    string
	}{
		{
			name:       "success",
			statusCode: http.StatusOK,
			body:       `{"status":true,"msg":"success!"}`,
			wantBody:   `{"status":true,"msg":"success!"}`,
		},
		{
			name:       "http failure",
			statusCode: http.StatusBadGateway,
			body:       `upstream unavailable`,
			wantBody:   `upstream unavailable`,
			wantErr:    "HTTP 502",
		},
		{
			name:       "business rejection",
			statusCode: http.StatusOK,
			body:       `{"status":false,"msg":"作业提交失败！"}`,
			wantBody:   `{"status":false,"msg":"作业提交失败！"}`,
			wantErr:    "作业提交失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = io.WriteString(w, tt.body)
			}))
			defer server.Close()

			req, err := http.NewRequest(http.MethodPost, server.URL, nil)
			if err != nil {
				t.Fatal(err)
			}
			body, err := (&XueXiTUserCache{}).sendChapterWorkSubmission(server.Client(), req)
			if body != tt.wantBody {
				t.Fatalf("sendChapterWorkSubmission() body = %q, want %q", body, tt.wantBody)
			}
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("sendChapterWorkSubmission() error = %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("sendChapterWorkSubmission() error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestSendChapterWorkSubmissionPropagatesBodyReadError(t *testing.T) {
	wantErr := errors.New("read failed")
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(errorReader{err: wantErr}),
		}, nil
	})}
	req, err := http.NewRequest(http.MethodPost, "https://example.test", nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = (&XueXiTUserCache{}).sendChapterWorkSubmission(client, req)
	if !errors.Is(err, wantErr) {
		t.Fatalf("sendChapterWorkSubmission() error = %v, want wrapping %v", err, wantErr)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

type errorReader struct {
	err error
}

func (reader errorReader) Read([]byte) (int, error) {
	return 0, reader.err
}
