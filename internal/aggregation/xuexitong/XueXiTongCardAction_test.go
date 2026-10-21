package xuexitong

import (
	"errors"
	"testing"
)

func TestParseAttachmentSettingReturnsBusinessError(t *testing.T) {
	_, _, err := parseAttachmentSetting(`<html><body><p class="blankTips">任务点暂不可用</p></body></html>`)
	var apiErr APIError
	if !errors.As(err, &apiErr) || apiErr.Message != "任务点暂不可用" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseAttachmentSettingReturnsChapterNotOpened(t *testing.T) {
	_, _, err := parseAttachmentSetting(`<html><body><p class="blankTips">章节未开放！</p></body></html>`)
	var blocked ChapterNotOpened
	if !errors.As(err, &blocked) {
		t.Fatalf("expected ChapterNotOpened, got %v", err)
	}
}

func TestParseAttachmentSettingReturnsObjectAndEnc(t *testing.T) {
	html := `<html><body><script type="text/javascript">window.AttachmentSetting = {"attachments":[],"defaults":{"fid":"1"}};</script><input type="hidden" id="from" value="a_b_c_enc-value"/></body></html>`
	attachment, enc, err := parseAttachmentSetting(html)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if enc != "enc-value" || attachment["enc"] != enc {
		t.Fatalf("unexpected enc: %q, attachment: %#v", enc, attachment)
	}
}
