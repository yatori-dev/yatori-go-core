package welearn

import "testing"

func TestIsCompletionStatus(t *testing.T) {
	cases := map[string]bool{
		"completed":  true,
		"passed":     true,
		"已完成":        true,
		"incomplete": false,
		"":           false,
	}
	for status, want := range cases {
		if got := isCompletionStatus(status); got != want {
			t.Fatalf("isCompletionStatus(%q) = %v, want %v", status, got, want)
		}
	}
}
