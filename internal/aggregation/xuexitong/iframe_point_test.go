package xuexitong

import "testing"

func TestIframePointHasJobID(t *testing.T) {
	cases := []struct {
		name string
		data map[string]interface{}
		want bool
	}{
		{name: "ballchart", data: map[string]interface{}{"keyword": "理想信念", "mid": "123"}, want: false},
		{name: "resourceVR", data: map[string]interface{}{"mid": "123", "previewUrl": "https://example.com"}, want: false},
		{name: "string jobid", data: map[string]interface{}{"_jobid": "456"}, want: true},
		{name: "float jobid", data: map[string]interface{}{"jobid": float64(456)}, want: true},
		{name: "zero jobid", data: map[string]interface{}{"_jobid": "0"}, want: false},
		{name: "empty jobid", data: map[string]interface{}{"jobid": ""}, want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			point := IframeAttributes{Data: tc.data, HasData: true}
			if got := iframePointHasJobID(point); got != tc.want {
				t.Fatalf("iframePointHasJobID() = %v, want %v", got, tc.want)
			}
		})
	}
}
