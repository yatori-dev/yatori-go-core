package yinghua

import "testing"

func TestYinghuaStudyID(t *testing.T) {
	cases := map[string]string{
		`{"result":{"data":{"studyId":12345}}}`:   "12345",
		`{"result":{"data":{"studyId":"67890"}}}`: "67890",
		`{"result":{"data":{}}}`:                  "",
		`not-json`:                                "",
	}
	for report, want := range cases {
		if got := yinghuaStudyID(report); got != want {
			t.Fatalf("yinghuaStudyID(%q)=%q want %q", report, got, want)
		}
	}
}
