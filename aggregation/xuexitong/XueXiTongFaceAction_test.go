package xuexitong

import "testing"

func TestFaceCheckResult(t *testing.T) {
	cases := []struct {
		name     string
		body     string
		wantPass bool
		wantWhy  string
	}{
		{"old literal true", "true", true, ""},
		{"old quoted true", `"true"`, true, ""},
		{"status true with non-pass msg", `{"status":true,"msg":"上传成功"}`, true, ""},
		{"status true empty msg", `{"status":true,"msg":""}`, true, ""},
		{"status false", `{"status":false,"msg":"比对失败"}`, false, "比对失败"},
		{"status null", `{"status":null,"msg":"比对失败"}`, false, "比对失败"},
		{"status empty string", `{"status":"","msg":"比对失败"}`, false, "比对失败"},
		{"result 1", `{"result":1}`, true, ""},
		{"result false", `{"result":false,"msg":"失败"}`, false, "失败"},
		{"success string false", `{"success":"false","msg":"失败"}`, false, "失败"},
		{"msg pass", `{"msg":"识别通过"}`, true, ""},
		{"msg liveness failed", `{"msg":"活体检测不通过"}`, false, "活体检测不通过"},
		{"msg image error", `{"msg":"用户图片信息出错"}`, false, "用户图片信息出错"},
		{"msg illegal", `{"msg":"非法请求"}`, false, "非法请求"},
		{"code zero", `{"code":"0"}`, true, ""},
		{"code non-zero", `{"code":"1","msg":"失败"}`, false, "失败"},
		{"empty body", "", false, "接口返回空响应"},
		{"html login page", "<html><body>请登录</body></html>", false, "非JSON响应: <html><body>请登录</body></html>"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			passed, reason := faceCheckResult(tc.body)
			if passed != tc.wantPass {
				t.Fatalf("faceCheckResult(%q) passed=%v, want %v (reason=%q)", tc.body, passed, tc.wantPass, reason)
			}
			if tc.wantWhy != "" && reason != tc.wantWhy {
				t.Fatalf("faceCheckResult(%q) reason=%q, want %q", tc.body, reason, tc.wantWhy)
			}
		})
	}
}
