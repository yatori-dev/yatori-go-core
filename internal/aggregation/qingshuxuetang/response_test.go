package qingshuxuetang

import "testing"

func TestCheckQsxtResponse(t *testing.T) {
	cases := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{name: "success", body: `{"hr":0,"data":"ok"}`, wantErr: false},
		{name: "business failure", body: `{"hr":1,"message":"fail"}`, wantErr: true},
		{name: "missing hr", body: `<html></html>`, wantErr: true},
		{name: "empty", body: ``, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := checkQsxtResponse(tc.body)
			if (err != nil) != tc.wantErr {
				t.Fatalf("checkQsxtResponse() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
