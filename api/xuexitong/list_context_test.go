package xuexitong

import (
	"context"
	"errors"
	"testing"
)

func TestReadOnlyCourseRequestsHonorCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cache := &XueXiTUserCache{}
	tests := []struct {
		name string
		call func() error
	}{
		{name: "chapter cards", call: func() error {
			_, err := cache.FetchChapterCordsContext(ctx, []int{1}, 0, 2, 0, nil)
			return err
		}},
		{name: "work list", call: func() error {
			_, err := cache.PullWorkListHtmlApiContext(ctx, "1", "2", "3", 0, nil)
			return err
		}},
		{name: "exam list", call: func() error {
			_, err := cache.PullExamListHtmlApiContext(ctx, "1", "2", "3", 0, nil)
			return err
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.call(); !errors.Is(err, context.Canceled) {
				t.Fatalf("error = %v, want context canceled", err)
			}
		})
	}
}

func TestFetchChapterCordsRejectsInvalidIndex(t *testing.T) {
	cache := &XueXiTUserCache{}
	if _, err := cache.FetchChapterCordsContext(context.Background(), nil, 0, 2, 0, nil); err == nil {
		t.Fatal("expected invalid chapter index error")
	}
}
