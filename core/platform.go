package core

import (
	"fmt"
	"strings"
)

// Platform identifies a supported upstream learning platform.
//
// The string values match the accountType values used by the legacy
// configuration files, so existing configuration can be mapped directly.
type Platform string

const (
	XueXiTong Platform = "XUEXITONG"
	YingHua   Platform = "YINGHUA"
	ICVE      Platform = "ICVE"
	CQIE      Platform = "CQIE"
	ENAAE     Platform = "ENAAE"
	WeLearn   Platform = "WELEARN"
	QSXT      Platform = "QSXT"
	TTCDW     Platform = "TTCDW"
	Ketangx   Platform = "KETANGX"
	HaiQiKeJi Platform = "ZXCPKS"
	WeiBan    Platform = "WEIBAN"
	CELA      Platform = "CELA"
	MOOC      Platform = "MOOC"
	ZhiHuiShu Platform = "ZHIHUISHU"
	GongXue   Platform = "GONGXUE"
)

var platformDescriptions = map[Platform]string{
	XueXiTong: "超星学习通",
	YingHua:   "英华学堂",
	ICVE:      "智慧职教/职教云",
	CQIE:      "重庆工程学院",
	ENAAE:     "中国教育干部网络学院",
	WeLearn:   "WeLearn 随行课堂",
	QSXT:      "青书学堂",
	TTCDW:     "TTCDW 团课",
	Ketangx:   "课堂行",
	HaiQiKeJi: "海旗科技",
	WeiBan:    "安全微伴",
	CELA:      "CELA",
	MOOC:      "中国大学 MOOC",
	ZhiHuiShu: "智慧树",
	GongXue:   "工学云",
}

var platformAliases = map[string]Platform{
	"XUEXITONG":      XueXiTong,
	"XXT":            XueXiTong,
	"CHAOXING":       XueXiTong,
	"YINGHUA":        YingHua,
	"ICVE":           ICVE,
	"ZJY":            ICVE,
	"CQIE":           CQIE,
	"ENAAE":          ENAAE,
	"WELEARN":        WeLearn,
	"QSXT":           QSXT,
	"QINGSHUXUE":     QSXT,
	"QINGSHUXUETANG": QSXT,
	"TTCDW":          TTCDW,
	"KETANGX":        Ketangx,
	"ZXCPKS":         HaiQiKeJi,
	"HAIQIKEJI":      HaiQiKeJi,
	"HQKJ":           HaiQiKeJi,
	"WEIBAN":         WeiBan,
	"CELA":           CELA,
	"MOOC":           MOOC,
	"ZHIHUISHU":      ZhiHuiShu,
	"ZHS":            ZhiHuiShu,
	"ZHIDAO":         ZhiHuiShu,
	"GONGXUE":        GongXue,
	"GXY":            GongXue,
}

// String returns the canonical platform identifier.
func (p Platform) String() string {
	return string(p)
}

// Description returns a human readable platform name.
func (p Platform) Description() string {
	if v, ok := platformDescriptions[p]; ok {
		return v
	}
	return string(p)
}

// Valid reports whether p is a supported platform.
func (p Platform) Valid() bool {
	_, ok := platformDescriptions[p]
	return ok
}

// ParsePlatform normalizes a platform identifier. It accepts the canonical
// value as well as the common aliases used by legacy configuration files.
func ParsePlatform(raw string) (Platform, error) {
	value := strings.ToUpper(strings.TrimSpace(raw))
	if value == "" {
		return "", fmt.Errorf("yatori: empty platform")
	}
	if p, ok := platformAliases[value]; ok {
		return p, nil
	}
	return "", fmt.Errorf("yatori: unsupported platform %q", raw)
}

// AllPlatforms returns every platform known to the SDK in a stable order.
func AllPlatforms() []Platform {
	return []Platform{
		XueXiTong,
		YingHua,
		ICVE,
		CQIE,
		ENAAE,
		WeLearn,
		QSXT,
		TTCDW,
		Ketangx,
		HaiQiKeJi,
		WeiBan,
		CELA,
		MOOC,
		ZhiHuiShu,
		GongXue,
	}
}
