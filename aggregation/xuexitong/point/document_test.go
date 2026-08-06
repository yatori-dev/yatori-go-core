package point

import (
	"errors"
	"strings"
	"testing"

	"github.com/yatori-dev/yatori-go-core/api/xuexitong"
)

func TestExecuteDocumentRetriesWithSameReporterAfterServerError(t *testing.T) {
	p := &xuexitong.PointDocumentDto{Title: "document"}
	calls := 0
	relogins := 0
	reporter := func() (string, error) {
		calls++
		if calls == 1 {
			return "", errors.New("status code: 500")
		}
		return `{"status":true}`, nil
	}

	report, err := executeDocument(p, reporter, func() { relogins++ })
	if err != nil {
		t.Fatalf("executeDocument() error = %v", err)
	}
	if report != `{"status":true}` {
		t.Fatalf("executeDocument() report = %q", report)
	}
	if calls != 2 || relogins != 1 {
		t.Fatalf("calls = %d, relogins = %d, want 2 and 1", calls, relogins)
	}
}

func TestExecuteDocumentDoesNotReloginForOtherErrors(t *testing.T) {
	relogins := 0
	wantErr := errors.New("request canceled")
	_, err := executeDocument(&xuexitong.PointDocumentDto{}, func() (string, error) {
		return "", wantErr
	}, func() { relogins++ })

	if !errors.Is(err, wantErr) {
		t.Fatalf("executeDocument() error = %v, want %v", err, wantErr)
	}
	if relogins != 0 {
		t.Fatalf("relogins = %d, want 0", relogins)
	}
}

func TestExecuteDocumentRejectsInvalidStatus(t *testing.T) {
	tests := []struct {
		name   string
		report string
		want   string
	}{
		{name: "missing", report: `{"msg":"ok"}`, want: "缺少 status"},
		{name: "null", report: `{"status":null}`, want: "缺少 status"},
		{name: "wrong type", report: `{"status":"true"}`, want: "status 不是布尔值"},
		{name: "invalid json", report: `{`, want: "解析文档提交响应失败"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := executeDocument(&xuexitong.PointDocumentDto{}, func() (string, error) {
				return tt.report, nil
			}, func() {})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("executeDocument() error = %v, want containing %q", err, tt.want)
			}
		})
	}
}
