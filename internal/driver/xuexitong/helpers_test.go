package xuexitong

import (
	"strings"
	"testing"
)

func TestValidateReadReport(t *testing.T) {
	cases := []struct {
		name    string
		report  string
		wantErr bool
	}{
		{name: "success", report: `{"status":true,"msg":"ok"}`, wantErr: false},
		{name: "business failure", report: `{"status":false,"msg":"该文档不是任务点"}`, wantErr: true},
		{name: "missing status", report: `<html>login</html>`, wantErr: true},
		{name: "empty body", report: "", wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateReadReport(tc.report)
			if (err != nil) != tc.wantErr {
				t.Fatalf("validateReadReport() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestSummarizeReportLimitsLength(t *testing.T) {
	report := strings.Repeat("x", 500)
	summary := summarizeReport(report)
	if len(summary) != 203 {
		t.Fatalf("summary length = %d, want 203", len(summary))
	}
	if !strings.HasSuffix(summary, "...") {
		t.Fatalf("summary = %q, want ... suffix", summary)
	}
}
