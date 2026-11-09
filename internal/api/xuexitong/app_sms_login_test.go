package xuexitong

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"testing"
)

type appLoginTestTransport func(*http.Request) (*http.Response, error)

func (f appLoginTestTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestAppSMSLoginContinuation(t *testing.T) {
	for _, tt := range []struct {
		name    string
		status  int
		body    string
		wantErr string
	}{
		{"verified", 200, `{"status":true}`, ""},
		{"normal_type", 200, `{"status":true,"type":0}`, ""},
		{"rejected", 200, `{"status":false}`, "短信验证失败"},
		{"http_error", 503, `{"status":true}`, "短信验证失败"},
		{"html", 200, `<html>login</html>`, "短信验证失败"},
		{"missing_status", 200, `{}`, "短信验证失败"},
		{"change_password", 200, `{"status":true,"type":1}`, "修改密码"},
		{"change_password_string", 200, `{"status":true,"type":"1"}`, "修改密码"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := appSMSLoginContinuation(tt.status, []byte(tt.body))
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) || got != "" {
					t.Fatalf("got %q, %v", got, err)
				}
			} else if err != nil || got != "https://sso.chaoxing.com/apis/login/userLogin4Uname.do?ft=true" {
				t.Fatalf("got %q, %v", got, err)
			}
		})
	}
}

func TestAppLoginContinuesVerifiedSession(t *testing.T) {
	// Use a generated test key so the entire SSO response is synthetic.
	originalKey, err := appFacePublicKeyValue()
	if err != nil {
		t.Fatal(err)
	}
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}
	facePublicKey = &key.PublicKey
	t.Cleanup(func() { facePublicKey = originalKey })
	signed, err := rsa.SignPKCS1v15(rand.Reader, key, 0, []byte(`{"cid":"test-cid","sc":"test-sc"}`))
	if err != nil {
		t.Fatal(err)
	}
	clientID := base64.StdEncoding.EncodeToString(signed)
	for _, tt := range []struct {
		name, reply               string
		challenge, cancel, wantOK bool
	}{
		{"sms_success", `{"status":true}`, true, false, true},
		{"sms_rejected", `{"status":false}`, true, false, false},
		{"password_required", `{"status":true,"type":1}`, true, false, false},
		{"cancelled", "", true, true, false},
		{"ordinary_login", "", false, false, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			jar, err := cookiejar.New(nil)
			if err != nil {
				t.Fatal(err)
			}
			calls := map[string]int{}
			prompts := 0
			cache := &XueXiTUserCache{Name: "test-account", Password: "test-password"}
			cache.LoginCodeProvider = func(got context.Context, challenge LoginCodeChallenge) (string, error) {
				prompts++
				if got != ctx || challenge.Kind != "sms" || challenge.SendSMS == nil {
					t.Fatal("invalid challenge")
				}
				if tt.cancel {
					return "", context.Canceled
				}
				return "000000", nil
			}
			client := &http.Client{Jar: jar, Transport: appLoginTestTransport(func(req *http.Request) (*http.Response, error) {
				calls[req.URL.Path]++
				if req.UserAgent() != "test-app-agent" || req.Context() != ctx {
					t.Fatal("session context changed")
				}
				body := ""
				header := make(http.Header)
				switch req.URL.Path {
				case "/v11/loginregister":
					if calls[req.URL.Path] != 1 {
						t.Fatal("loginregister repeated after verification")
					}
					if tt.challenge {
						body = `{"status":false,"webUrl":"https://passport2.chaoxing.com/v11/tocheckmessage"}`
					} else {
						body = `{"status":true,"url":"https://sso.chaoxing.com/apis/login/userLogin4Uname.do?ft=true"}`
					}
				case "/v11/tocheckmessage":
					body = `<input id="enc2" value="test-enc"><input id="countrycode" value="86">`
				case "/v11/checkmessage":
					if err := req.ParseForm(); err != nil {
						t.Fatal(err)
					}
					if req.Form.Get("vcode") != "000000" || req.Form.Get("countrycode") != "86" {
						t.Fatal("verification form lost")
					}
					body = tt.reply
					header.Add("Set-Cookie", "verified=yes; Domain=chaoxing.com; Path=/; Secure")
					header.Add("Set-Cookie", "passport_only=yes; Path=/; Secure")
				case "/apis/login/userLogin4Uname.do":
					if !tt.wantOK {
						t.Fatal("SSO attempted after failed verification")
					}
					if req.Method != http.MethodPost || req.URL.Query().Get("ft") != "true" {
						t.Fatal("invalid SSO request")
					}
					if err := req.ParseForm(); err != nil {
						t.Fatal(err)
					}
					if req.PostForm.Get("data") == "" || req.PostForm.Get("data") != req.URL.Query().Get("hddInfo") {
						t.Fatal("device payload changed")
					}
					if tt.challenge {
						if cookie, err := req.Cookie("verified"); err != nil || cookie.Value != "yes" {
							t.Fatal("verified cookie missing")
						}
					}
					if _, err := req.Cookie("passport_only"); err == nil {
						t.Fatal("host-only cookie leaked to SSO")
					}
					body = fmt.Sprintf(`{"result":1,"msg":{"clientId":%q,"puid":"test-user"}}`, clientID)
				case "/":
					if req.URL.Host != "mooc1-api.chaoxing.com" {
						t.Fatal("unexpected host")
					}
				default:
					t.Fatalf("unexpected request: %s", req.URL.Path)
				}
				return &http.Response{StatusCode: 200, Header: header, Body: io.NopCloser(strings.NewReader(body)), Request: req}, nil
			})}
			result, err := cache.appLogin(ctx, client, &appDeviceProfile{UserAgent: "test-app-agent"})
			if tt.wantOK {
				if err != nil || result == nil || result.Cxcid != "test-cid" || cache.UserID != "test-user" || cache.FaceSc != "test-sc" || calls["/apis/login/userLogin4Uname.do"] != 1 {
					t.Fatalf("login failed: %+v, %v", result, err)
				}
			} else if err == nil || result != nil || cache.UserID != "" || calls["/apis/login/userLogin4Uname.do"] != 0 {
				t.Fatalf("failed login accepted: %+v, %v", result, err)
			}
			if tt.cancel && !errors.Is(err, context.Canceled) {
				t.Fatalf("cancellation lost: %v", err)
			}
			if (tt.challenge && prompts != 1) || (!tt.challenge && prompts != 0) {
				t.Fatal("unexpected challenge count")
			}
		})
	}
}
