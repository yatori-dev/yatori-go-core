package weiban

import "testing"

func TestValidateWeiBanLogin(t *testing.T) {
	cases := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{name: "success code", body: `{"code":200,"data":{"token":"x"}}`, wantErr: false},
		{name: "success flag", body: `{"success":true}`, wantErr: false},
		{name: "business failure", body: `{"success":false,"msg":"验证码错误"}`, wantErr: true},
		{name: "error code", body: `{"code":500,"msg":"密码错误"}`, wantErr: true},
		{name: "empty body", body: ``, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateWeiBanLogin(tc.body)
			if (err != nil) != tc.wantErr {
				t.Fatalf("validateWeiBanLogin() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
