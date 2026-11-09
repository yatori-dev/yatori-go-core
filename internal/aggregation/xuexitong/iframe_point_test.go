package xuexitong

import "testing"

func TestIframePointJobID(t *testing.T) {
	cases := []struct {
		name string
		data map[string]interface{}
		want string
	}{
		{name: "missing", data: map[string]interface{}{}, want: ""},
		{name: "iframe id", data: map[string]interface{}{"_jobid": " 456 "}, want: "456"},
		{name: "property id", data: map[string]interface{}{"jobid": "789"}, want: "789"},
		{name: "float id", data: map[string]interface{}{"jobid": float64(456)}, want: "456"},
		{name: "integer id", data: map[string]interface{}{"jobid": 456}, want: "456"},
		{name: "zero id", data: map[string]interface{}{"_jobid": "0"}, want: ""},
		{name: "zero fallback", data: map[string]interface{}{"_jobid": "0", "jobid": "789"}, want: "789"},
		{name: "empty fallback", data: map[string]interface{}{"_jobid": "", "jobid": "789"}, want: "789"},
		{name: "iframe precedence", data: map[string]interface{}{"_jobid": "456", "jobid": "789"}, want: "456"},
		{name: "unsupported type", data: map[string]interface{}{"jobid": true}, want: ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := iframePointJobID(IframeAttributes{Data: tc.data, HasData: true}); got != tc.want {
				t.Fatalf("iframePointJobID() = %q, want %q", got, tc.want)
			}
		})
	}
}
