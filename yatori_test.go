package yatori

import (
	"testing"
)

func TestRegisteredPlatforms(t *testing.T) {
	want := AllPlatforms()
	got := Registered()
	if len(got) != len(want) {
		t.Fatalf("registered %d platforms, want %d: %v", len(got), len(want), got)
	}
	set := make(map[Platform]bool, len(got))
	for _, platform := range got {
		set[platform] = true
	}
	for _, platform := range want {
		if !set[platform] {
			t.Errorf("platform %s is not registered", platform)
		}
	}
}

func TestParsePlatformAliases(t *testing.T) {
	cases := map[string]Platform{
		"xuexitong":  XueXiTong,
		"XXT":        XueXiTong,
		"chaoxing":   XueXiTong,
		"haiqikeji":  HaiQiKeJi,
		"ZXCPKS":     HaiQiKeJi,
		"qingshuxue": QSXT,
		"zhihuishu":  ZhiHuiShu,
		"gongxue":    GongXue,
	}
	for input, want := range cases {
		got, err := ParsePlatform(input)
		if err != nil {
			t.Fatalf("ParsePlatform(%q) failed: %v", input, err)
		}
		if got != want {
			t.Errorf("ParsePlatform(%q) = %s, want %s", input, got, want)
		}
	}
	if _, err := ParsePlatform("unknown"); err == nil {
		t.Fatal("ParsePlatform(unknown) should fail")
	}
}

func TestNewAllPlatformsSmoke(t *testing.T) {
	cases := []Config{
		{Platform: XueXiTong, Account: Account{Username: "u", Password: "p"}},
		{Platform: YingHua, Account: Account{Username: "u", Password: "p", URL: "https://example.com"}},
		{Platform: ICVE, Account: Account{Username: "u", Password: "p"}},
		{Platform: CQIE, Account: Account{Username: "u", Password: "p"}},
		{Platform: ENAAE, Account: Account{Username: "u", Password: "p"}},
		{Platform: WeLearn, Account: Account{Username: "u", Password: "p"}},
		{Platform: QSXT, Account: Account{Username: "u", Password: "p"}},
		{Platform: TTCDW, Account: Account{Username: "u", Password: "p"}},
		{Platform: Ketangx, Account: Account{Username: "u", Password: "p"}},
		{Platform: HaiQiKeJi, Account: Account{Username: "u", Password: "p", URL: "https://example.com"}},
		{Platform: WeiBan, Account: Account{Username: "u", Password: "p", School: "school"}},
		{Platform: CELA, Account: Account{Username: "u", Password: "p"}},
		{Platform: MOOC, Account: Account{Username: "u", Password: "p"}},
		{Platform: ZhiHuiShu},
		{Platform: GongXue, Account: Account{Username: "u", Password: "p"}},
	}
	for _, cfg := range cases {
		client, err := New(cfg)
		if err != nil {
			t.Fatalf("New(%s) failed: %v", cfg.Platform, err)
		}
		if client == nil {
			t.Fatalf("New(%s) returned nil client", cfg.Platform)
		}
		if client.Platform() != cfg.Platform {
			t.Errorf("New(%s).Platform() = %s", cfg.Platform, client.Platform())
		}
		if client.Capabilities() == 0 {
			t.Errorf("New(%s) has no capabilities", cfg.Platform)
		}
		_ = client.Close()
	}
}

func TestNewRejectsUnregisteredPlatform(t *testing.T) {
	_, err := New(Config{Platform: Platform("NOPE"), Account: Account{Username: "u"}})
	if err == nil {
		t.Fatal("New with an unknown platform should fail")
	}
}
