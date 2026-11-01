package qsxt

import (
	"testing"

	agg "github.com/yatori-dev/yatori-go-core/v3/internal/aggregation/qingshuxuetang"
)

func TestNodeProgressForDocumentUsesStudySeconds(t *testing.T) {
	node := agg.QsxtNode{
		NodeType:           "html",
		TotalStudyDuration: 175,
	}
	if got := nodeProgress(node); got != 50 {
		t.Fatalf("nodeProgress() = %v, want 50", got)
	}

	node.TotalStudyDuration = 400
	if got := nodeProgress(node); got != 100 {
		t.Fatalf("nodeProgress() = %v, want 100", got)
	}
}

func TestNodeCompletedForDocument(t *testing.T) {
	node := agg.QsxtNode{NodeType: "html", TotalStudyDuration: 349}
	if nodeCompleted(node) {
		t.Fatal("nodeCompleted() = true before threshold")
	}
	node.TotalStudyDuration = 350
	if !nodeCompleted(node) {
		t.Fatal("nodeCompleted() = false at threshold")
	}
}

func TestNodeCompletedForVideoUsesMaxPosition(t *testing.T) {
	node := agg.QsxtNode{NodeType: "video", Duration: 1000, MaxStudyPosition: 999}
	if nodeCompleted(node) {
		t.Fatal("nodeCompleted() = true before duration")
	}
	node.MaxStudyPosition = 1000
	if !nodeCompleted(node) {
		t.Fatal("nodeCompleted() = false at duration")
	}
}
