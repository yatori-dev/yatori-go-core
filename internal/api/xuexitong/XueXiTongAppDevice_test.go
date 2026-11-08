package xuexitong

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppDeviceProfilePersistsPerAccount(t *testing.T) {
	oldDir := appDeviceBaseDir
	appDeviceBaseDir = t.TempDir()
	t.Cleanup(func() { appDeviceBaseDir = oldDir })

	first := loadOrCreateAppDeviceProfile("13800000000")
	if first.DeviceUniqueID == "" || first.MediaDrmID == "" || first.AndroidID == "" {
		t.Fatalf("生成的设备信息不完整: %+v", first)
	}

	path := appDevicePath("13800000000")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("设备信息未落盘: %v", err)
	}

	second := loadOrCreateAppDeviceProfile("13800000000")
	if second.DeviceUniqueID != first.DeviceUniqueID ||
		second.MediaDrmID != first.MediaDrmID ||
		second.AndroidID != first.AndroidID ||
		second.UserAgent != first.UserAgent {
		t.Fatalf("同一账号复用了不同的设备信息: first=%+v second=%+v", first, second)
	}
	if !strings.Contains(first.UserAgent, first.Model) ||
		!strings.Contains(first.UserAgent, "Android "+first.OSVer) {
		t.Fatalf("UA 与设备信息不一致: %q model=%s os=%s", first.UserAgent, first.Model, first.OSVer)
	}

	other := loadOrCreateAppDeviceProfile("13900000000")
	if other.DeviceUniqueID == first.DeviceUniqueID ||
		other.MediaDrmID == first.MediaDrmID ||
		other.AndroidID == first.AndroidID ||
		other.UserAgent == first.UserAgent {
		t.Fatalf("不同账号应该生成不同的设备信息: first=%+v other=%+v", first, other)
	}
}

func TestBuildAppDeviceInfoUsesPersistentProfile(t *testing.T) {
	oldDir := appDeviceBaseDir
	appDeviceBaseDir = t.TempDir()
	t.Cleanup(func() { appDeviceBaseDir = oldDir })

	profile := loadOrCreateAppDeviceProfile("13800000000")
	first, err := buildAppDeviceInfo(profile)
	if err != nil {
		t.Fatalf("构建 deviceInfo 失败: %v", err)
	}
	if first == "" {
		t.Fatal("deviceInfo 为空")
	}
	if _, err := os.Stat(appDevicePath("13800000000")); err != nil {
		t.Fatalf("deviceInfo 没有按账号落盘: %v", err)
	}
}

func TestSanitizeAppDeviceAccount(t *testing.T) {
	if got := sanitizeAppDeviceAccount("13800000000"); got != "13800000000" {
		t.Fatalf("手机号文件名被改写: %q", got)
	}
	if got := sanitizeAppDeviceAccount("user@example.com"); got != "user@example.com" {
		t.Fatalf("邮箱文件名被改写: %q", got)
	}
	got := sanitizeAppDeviceAccount("../../etc/passwd")
	if strings.ContainsAny(got, `/\`) {
		t.Fatalf("文件名仍包含路径分隔符: %q", got)
	}
	if !strings.HasPrefix(got, ".._.._etc_passwd_") {
		t.Fatalf("特殊账号没有追加哈希: %q", got)
	}
}

func TestAppDeviceProfileRecoversFromCorruptFile(t *testing.T) {
	oldDir := appDeviceBaseDir
	appDeviceBaseDir = t.TempDir()
	t.Cleanup(func() { appDeviceBaseDir = oldDir })

	path := appDevicePath("13800000000")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("创建目录失败: %v", err)
	}
	if err := os.WriteFile(path, []byte("not-json"), 0o600); err != nil {
		t.Fatalf("写入损坏文件失败: %v", err)
	}

	profile := loadOrCreateAppDeviceProfile("13800000000")
	if profile.DeviceUniqueID == "" {
		t.Fatalf("损坏文件后没有重新生成设备信息: %+v", profile)
	}

	again := loadOrCreateAppDeviceProfile("13800000000")
	if again.DeviceUniqueID != profile.DeviceUniqueID {
		t.Fatalf("损坏文件恢复后没有落盘复用: first=%s second=%s", profile.DeviceUniqueID, again.DeviceUniqueID)
	}
}
