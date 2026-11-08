package xuexitong

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const appDeviceDefaultDir = "assets/xxt_device"

var (
	appDeviceBaseDir = appDeviceDefaultDir
	appDeviceMu      sync.Mutex
)

// appDeviceProfile 是按账号持久化的设备指纹。
//
// 学习通把 deviceUniqueId/mediaDrmId/android_id 以及 UA 里的设备信息当作
// 设备标识。每个账号在首次登录时随机分配一套设备型号和标识并落盘，
// 后续登录复用；不同账号之间的型号、标识、UA 令牌都不同，避免多个账号
// 被关联成同一台设备。
type appDeviceProfile struct {
	Platform       string `json:"platform"`
	Brand          string `json:"brand"`
	Model          string `json:"model"`
	AppVer         string `json:"app_ver"`
	OSName         string `json:"os_name"`
	CDID           string `json:"cdid"`
	DeviceUniqueID string `json:"deviceUniqueId"`
	MediaDrmID     string `json:"mediaDrmId"`
	OSVer          string `json:"os_ver"`
	OSLang         string `json:"os_lang"`
	AppName        string `json:"app_name"`
	CPUArch        string `json:"cpu_ar"`
	Hardware       string `json:"hardware"`
	Resolution     string `json:"resolution"`
	AndroidID      string `json:"android_id"`
	Signatures     string `json:"signatures"`
	Build          string `json:"build"`
	UserAgent      string `json:"user_agent"`
}

// appDeviceTemplate 是可供账号分配的真实机型。
type appDeviceTemplate struct {
	Brand      string
	Model      string
	OSVer      string
	CPUArch    string
	Hardware   string
	Resolution string
	Build      string
}

var appDeviceTemplates = []appDeviceTemplate{
	{Brand: "Xiaomi", Model: "M2102J2SC", OSVer: "11", CPUArch: "arm64-v8a", Hardware: "qcom", Resolution: "1080x2206", Build: "RKQ1.200826.002"},
	{Brand: "Xiaomi", Model: "2201123C", OSVer: "12", CPUArch: "arm64-v8a", Hardware: "qcom", Resolution: "1080x2400", Build: "SKQ1.211006.001"},
	{Brand: "Redmi", Model: "M2012K11AC", OSVer: "11", CPUArch: "arm64-v8a", Hardware: "qcom", Resolution: "1080x2400", Build: "RKQ1.200826.002"},
	{Brand: "HUAWEI", Model: "ELS-AN00", OSVer: "10", CPUArch: "arm64-v8a", Hardware: "kirin", Resolution: "1080x2340", Build: "HUAWEIELS-AN00"},
	{Brand: "HONOR", Model: "RKY-AN10", OSVer: "12", CPUArch: "arm64-v8a", Hardware: "qcom", Resolution: "1080x2400", Build: "HONORRKY-AN10"},
	{Brand: "OPPO", Model: "PDEM30", OSVer: "12", CPUArch: "arm64-v8a", Hardware: "qcom", Resolution: "1080x2412", Build: "SP1A.210812.016"},
	{Brand: "vivo", Model: "V2148A", OSVer: "12", CPUArch: "arm64-v8a", Hardware: "qcom", Resolution: "1080x2400", Build: "SP1A.210812.016"},
	{Brand: "samsung", Model: "SM-G9910", OSVer: "12", CPUArch: "arm64-v8a", Hardware: "qcom", Resolution: "1080x2400", Build: "SP1A.210812.016"},
	{Brand: "realme", Model: "RMX3366", OSVer: "11", CPUArch: "arm64-v8a", Hardware: "qcom", Resolution: "1080x2400", Build: "RKQ1.201112.002"},
	{Brand: "OnePlus", Model: "LE2110", OSVer: "12", CPUArch: "arm64-v8a", Hardware: "qcom", Resolution: "1080x2400", Build: "SKQ1.211113.001"},
}

// loadOrCreateAppDeviceProfile 读取账号对应的设备信息，不存在或损坏时
// 生成一份并落盘。落盘失败不影响本次登录。
func loadOrCreateAppDeviceProfile(account string) *appDeviceProfile {
	appDeviceMu.Lock()
	defer appDeviceMu.Unlock()

	path := appDevicePath(account)
	if data, err := os.ReadFile(path); err == nil {
		var profile appDeviceProfile
		if json.Unmarshal(data, &profile) == nil {
			profile, changed := normalizeAppDeviceProfile(&profile, account)
			if changed {
				_ = saveAppDeviceProfile(path, profile)
			}
			return profile
		}
	}

	profile := newAppDeviceProfile(account)
	_ = saveAppDeviceProfile(path, profile)
	return profile
}

func newAppDeviceProfile(account string) *appDeviceProfile {
	template := pickAppDeviceTemplate(account)
	profile := &appDeviceProfile{
		Platform:       "android",
		Brand:          template.Brand,
		Model:          template.Model,
		AppVer:         "3.7.0.4",
		OSName:         "android",
		DeviceUniqueID: randomHex(16),
		MediaDrmID:     randomHex(16),
		OSVer:          template.OSVer,
		OSLang:         "zh_CN",
		AppName:        "com.chaoxing.mobile",
		CPUArch:        template.CPUArch,
		Hardware:       template.Hardware,
		Resolution:     template.Resolution,
		AndroidID:      randomHex(8),
		Build:          template.Build,
	}
	profile.UserAgent = buildAppDeviceUserAgent(profile)
	return profile
}

// buildAppDeviceUserAgent 根据账号自己的机型生成 UA。
// schild 和 (@Kalimdor) 后面的令牌每个账号都不同，避免 UA 层面被关联。
func buildAppDeviceUserAgent(profile *appDeviceProfile) string {
	model := strings.TrimSpace(profile.Model)
	if model == "" {
		model = "M2102J2SC"
	}
	osVer := strings.TrimSpace(profile.OSVer)
	if osVer == "" {
		osVer = "11"
	}
	build := strings.TrimSpace(profile.Build)
	if build == "" {
		build = "RKQ1.200826.002"
	}
	return fmt.Sprintf(
		"Mozilla/5.0 (Linux; Android %s; %s Build/%s; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/87.0.4280.141 Mobile Safari/537.36 (schild:%s) (device:%s) Language/zh_CN com.chaoxing.mobile/ChaoXingStudy_3_7_0_4_android_phone_10995_342 (@Kalimdor)_%s",
		osVer, model, build, randomHex(16), model, randomHex(16),
	)
}

// pickAppDeviceTemplate 按账号稳定地分配机型：同一账号始终落在同一档，
// 不同账号在型号上尽量分散，但真正的唯一性由随机标识和 UA 令牌保证。
func pickAppDeviceTemplate(account string) appDeviceTemplate {
	sum := sha256.Sum256([]byte(strings.TrimSpace(account)))
	return appDeviceTemplates[int(sum[0])%len(appDeviceTemplates)]
}

// normalizeAppDeviceProfile 补齐旧版本文件里缺失的字段，并报告是否有改动。
// 旧文件没有 model 时，整组机型字段会迁移到该账号对应的模板上。
func normalizeAppDeviceProfile(profile *appDeviceProfile, account string) (*appDeviceProfile, bool) {
	changed := false
	template := pickAppDeviceTemplate(account)
	if profile.Model == "" {
		profile.Brand = template.Brand
		profile.Model = template.Model
		profile.OSVer = template.OSVer
		profile.Hardware = template.Hardware
		profile.Resolution = template.Resolution
		profile.Build = template.Build
		profile.CPUArch = template.CPUArch
		changed = true
	}

	fill := func(target *string, value string) {
		if *target == "" {
			*target = value
			changed = true
		}
	}
	fill(&profile.Platform, "android")
	fill(&profile.Brand, template.Brand)
	fill(&profile.Model, template.Model)
	fill(&profile.AppVer, "3.7.0.4")
	fill(&profile.OSName, "android")
	if profile.DeviceUniqueID == "" {
		profile.DeviceUniqueID = randomHex(16)
		changed = true
	}
	if profile.MediaDrmID == "" {
		profile.MediaDrmID = randomHex(16)
		changed = true
	}
	fill(&profile.OSVer, template.OSVer)
	fill(&profile.OSLang, "zh_CN")
	fill(&profile.AppName, "com.chaoxing.mobile")
	fill(&profile.CPUArch, template.CPUArch)
	fill(&profile.Hardware, template.Hardware)
	fill(&profile.Resolution, template.Resolution)
	if profile.AndroidID == "" {
		profile.AndroidID = randomHex(8)
		changed = true
	}
	fill(&profile.Build, template.Build)
	if profile.UserAgent == "" {
		profile.UserAgent = buildAppDeviceUserAgent(profile)
		changed = true
	}
	return profile, changed
}

func appDevicePath(account string) string {
	return filepath.Join(appDeviceBaseDir, sanitizeAppDeviceAccount(account)+".json")
}

func saveAppDeviceProfile(path string, profile *appDeviceProfile) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return err
	}
	temp := path + ".tmp"
	if err := os.WriteFile(temp, data, 0o600); err != nil {
		return err
	}
	if err := os.Rename(temp, path); err != nil {
		_ = os.Remove(path)
		if retryErr := os.Rename(temp, path); retryErr != nil {
			_ = os.Remove(temp)
			return retryErr
		}
	}
	return nil
}

// sanitizeAppDeviceAccount 保证账号可以安全地作为文件名。
// 常规手机号/邮箱保持原样；包含路径分隔符等特殊字符时追加短哈希，避免碰撞。
func sanitizeAppDeviceAccount(account string) string {
	trimmed := strings.TrimSpace(account)
	if trimmed == "" {
		return "unknown"
	}
	var builder strings.Builder
	for _, r := range trimmed {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9',
			r == '-', r == '_', r == '.', r == '@':
			builder.WriteRune(r)
		default:
			builder.WriteByte('_')
		}
	}
	name := builder.String()
	if name == "" || name == "." || name == ".." {
		name = "account"
	}
	if name != trimmed || len(name) > 80 {
		sum := sha256.Sum256([]byte(trimmed))
		if len(name) > 48 {
			name = name[:48]
		}
		name += "_" + hex.EncodeToString(sum[:8])
	}
	return name
}
