package xuexitong

import (
	"errors"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"testing"

	"github.com/yatori-dev/yatori-go-core/v3/core"
)

type courseTestTransport func(*http.Request) (*http.Response, error)

func (f courseTestTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestAppCourseSession(t *testing.T) {
	courseURL, err := url.Parse("https://mooc1-api.chaoxing.com/mycourse/backclazzdata")
	if err != nil {
		t.Fatal(err)
	}
	passportURL, err := url.Parse("https://passport2.chaoxing.com/login")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, body string
		redirected bool
		rejected   bool
	}{
		{name: "json", body: `{"channelList":[]}`},
		{name: "redirect", redirected: true, body: "<html>login</html>", rejected: true},
		{name: "embedded-login-form", body: `<html><input id="phone"><input id="pwd"></html>`, rejected: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			jar, err := cookiejar.New(nil)
			if err != nil {
				t.Fatal(err)
			}
			jar.SetCookies(courseURL, []*http.Cookie{{Name: "course_session", Value: "current", Path: "/mycourse"}})
			jar.SetCookies(passportURL, []*http.Cookie{{Name: "passport_only", Value: "scoped", Path: "/"}})
			calls := 0
			transport := courseTestTransport(func(r *http.Request) (*http.Response, error) {
				calls++
				if got := r.UserAgent(); got != "Test App UA" {
					t.Errorf("got user agent %q", got)
				}
				response := &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(tc.body)), Request: r}
				if r.URL.Host == courseURL.Host {
					if r.Header.Get("Cookie") != "course_session=current" {
						t.Errorf("wrong cookie on course request: %q", r.Header.Get("Cookie"))
					}
					if tc.redirected {
						response.StatusCode = http.StatusFound
						response.Header.Set("Location", passportURL.String())
					}
				} else if r.URL.Host == passportURL.Host && tc.redirected {
					if r.Header.Get("Cookie") != "passport_only=scoped" {
						t.Errorf("wrong cookie on redirected request: %q", r.Header.Get("Cookie"))
					}
				} else {
					t.Fatalf("unexpected request URL %s", r.URL)
				}
				return response, nil
			})
			cache := XueXiTUserCache{
				cookies:      []*http.Cookie{{Name: "legacy", Value: "unused"}},
				appSession:   &http.Client{Transport: transport, Jar: jar},
				appUserAgent: "Test App UA",
			}
			body, err := cache.CourseListApiContext(t.Context(), 3, nil)
			if tc.rejected {
				if body != "" || !errors.Is(err, core.ErrNotLoggedIn) {
					t.Fatalf("expected a typed session error with no response body, got %q, %v", body, err)
				}
			} else if err != nil || body != tc.body {
				t.Fatalf("expected course JSON, got %q, %v", body, err)
			}
			if (tc.redirected && calls != 2) || (!tc.redirected && calls != 1) {
				t.Fatalf("unexpected request count: %d", calls)
			}
		})
	}
}
