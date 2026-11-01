package yinghua

import (
	"testing"
	"time"
)

func TestTimeWindowStatus(t *testing.T) {
	now := time.Now()
	if got := timeWindowStatus(now.Add(time.Hour), now.Add(2*time.Hour)); got != "not_started" {
		t.Fatalf("timeWindowStatus() = %q, want not_started", got)
	}
	if got := timeWindowStatus(now.Add(-time.Hour), now.Add(time.Hour)); got != "active" {
		t.Fatalf("timeWindowStatus() = %q, want active", got)
	}
	if got := timeWindowStatus(now.Add(-2*time.Hour), now.Add(-time.Hour)); got != "finished" {
		t.Fatalf("timeWindowStatus() = %q, want finished", got)
	}
}

func TestPercent(t *testing.T) {
	if got := percent(0.5); got != 50 {
		t.Fatalf("percent(0.5) = %v, want 50", got)
	}
	if got := percent(100); got != 100 {
		t.Fatalf("percent(100) = %v, want 100", got)
	}
}
