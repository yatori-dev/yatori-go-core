package xuexitong

import (
	"testing"

	api "github.com/yatori-dev/yatori-go-core/v3/internal/api/xuexitong"
)

func TestApplyVideoAttachmentFieldsDurations(t *testing.T) {
	cases := []struct {
		name         string
		obj          map[string]interface{}
		wantPlayTime int
		wantDuration int
	}{
		{
			name:         "float duration and millisecond play time",
			obj:          map[string]interface{}{"playTime": float64(58000), "attDuration": float64(1821)},
			wantPlayTime: 58,
			wantDuration: 1821,
		},
		{
			name:         "string duration",
			obj:          map[string]interface{}{"playTime": float64(12000), "attDuration": "120"},
			wantPlayTime: 12,
			wantDuration: 120,
		},
		{
			name:         "missing fields keep zero",
			obj:          map[string]interface{}{},
			wantPlayTime: 0,
			wantDuration: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := &api.PointVideoDto{}
			applyVideoAttachmentFields(p, tc.obj)
			if p.PlayTime != tc.wantPlayTime {
				t.Fatalf("PlayTime = %d, want %d", p.PlayTime, tc.wantPlayTime)
			}
			if p.Duration != tc.wantDuration {
				t.Fatalf("Duration = %d, want %d", p.Duration, tc.wantDuration)
			}
		})
	}
}
