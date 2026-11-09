package xuexitong

import "testing"

func TestAppLoginReplyShape(t *testing.T) {
	for _, tc := range []struct {
		payload, want string
	}{
		{`{"status":true,"url":"https://example.test/?session=private","webUrl":"","mes":"person@example.test","private-key":123456}`, "HTTP 200，status=true，非空字段=url/mes"},
		{`{"status":false,"data":null,"result":0,"msg":{}}`, "HTTP 200，status=false，非空字段=result"},
		{`{"status":"private","data":{"url":"private"}}`, "HTTP 200，status=非布尔，非空字段=data"},
		{`<html>private</html>`, "HTTP 200，非 JSON 对象"},
	} {
		if got := appLoginReplyShape(200, []byte(tc.payload)); got != tc.want {
			t.Errorf("got %q, want %q", got, tc.want)
		}
	}
}

func TestAppSMSLoginContinuation(t *testing.T) {
	for _, tc := range []struct {
		name    string
		status  int
		body    string
		wantURL string
		wantErr string
	}{
		{"verified", 200, `{"status":true}`, "https://sso.chaoxing.com/apis/login/userLogin4Uname.do?ft=true", ""},
		{"invalid code", 200, `{"status":false,"mes":"private"}`, "", "短信验证失败（HTTP 200，status=false，非空字段=mes）"},
		{"HTTP failure", 403, `{"status":true}`, "", "短信验证失败（HTTP 403，status=true，非空字段=无）"},
		{"password reset", 200, `{"status":true,"type":"1","token":"private"}`, "", "短信验证已通过，学习通要求先在官方页面修改密码"},
		{"password reset numeric", 200, `{"status":true,"type":1}`, "", "短信验证已通过，学习通要求先在官方页面修改密码"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			loginURL, err := appSMSLoginContinuation(tc.status, []byte(tc.body))
			if loginURL != tc.wantURL {
				t.Fatalf("login URL = %q, want %q", loginURL, tc.wantURL)
			}
			if tc.wantErr == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || err.Error() != tc.wantErr {
				t.Fatalf("error = %v, want %q", err, tc.wantErr)
			}
		})
	}
}
