package xuexitong

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestXueXiTongTLSVerificationEnabled(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}

	fset := token.NewFileSet()
	for _, filename := range files {
		file, err := parser.ParseFile(fset, filename, nil, 0)
		if err != nil {
			t.Fatalf("解析 %s 失败: %v", filename, err)
		}
		ast.Inspect(file, func(node ast.Node) bool {
			literal, ok := node.(*ast.CompositeLit)
			if !ok {
				return true
			}
			selector, ok := literal.Type.(*ast.SelectorExpr)
			if !ok || selector.Sel.Name != "Config" {
				return true
			}
			ident, ok := selector.X.(*ast.Ident)
			if !ok || ident.Name != "tls" {
				return true
			}
			for _, element := range literal.Elts {
				field, ok := element.(*ast.KeyValueExpr)
				if !ok {
					continue
				}
				name, ok := field.Key.(*ast.Ident)
				if !ok || name.Name != "InsecureSkipVerify" {
					continue
				}
				value, ok := field.Value.(*ast.Ident)
				if ok && value.Name == "true" {
					t.Errorf("%s 禁用了 TLS 证书校验", filename)
				}
			}
			return true
		})
	}
}

func TestPullSliderImgApiDoesNotForwardCookies(t *testing.T) {
	pngBody := newTestPNG(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Cookie"); got != "" {
			t.Errorf("滑块图片请求转发了 Cookie: %q", got)
		}
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(pngBody)
	}))
	defer server.Close()

	cache := &XueXiTUserCache{}
	cache.SetCookies([]*http.Cookie{{Name: "session", Value: "sensitive"}})
	if _, err := cache.PullSliderImgApi(server.URL, 0, nil); err != nil {
		t.Fatalf("拉取未重定向的图片失败: %v", err)
	}
}

func TestPullSliderImgApiRejectsRedirects(t *testing.T) {
	redirectTargetHits := 0
	target := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		redirectTargetHits++
	}))
	defer target.Close()

	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusFound)
	}))
	defer source.Close()

	if _, err := (&XueXiTUserCache{}).PullSliderImgApi(source.URL, 0, nil); err == nil {
		t.Fatal("滑块图片请求接受了重定向")
	}
	if redirectTargetHits != 0 {
		t.Fatalf("滑块图片请求跟随了重定向，目标收到 %d 次请求", redirectTargetHits)
	}
}

func newTestPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.White)
	var body bytes.Buffer
	if err := png.Encode(&body, img); err != nil {
		t.Fatal(err)
	}
	return body.Bytes()
}
