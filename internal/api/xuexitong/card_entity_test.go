package xuexitong

import (
	"strings"
	"testing"
)

func TestVideoAttachmentDetectionReportsRootType(t *testing.T) {
	_, err := (&PointVideoDto{}).AttachmentsDetection("business error")
	if err == nil || !strings.Contains(err.Error(), "根节点类型异常: string") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVideoAttachmentDetectionRecoversMalformedNestedData(t *testing.T) {
	attachment := map[string]interface{}{
		"attachments": []interface{}{
			map[string]interface{}{
				"property":  map[string]interface{}{"objectid": "video-1", "jobid": "job-1"},
				"otherInfo": 42,
			},
		},
		"defaults": map[string]interface{}{"fid": "1", "userid": "2"},
	}
	point := &PointVideoDto{ObjectID: "video-1", JobID: "job-1"}
	_, err := point.AttachmentsDetection(attachment)
	if err == nil || !strings.Contains(err.Error(), "otherInfo 类型异常") {
		t.Fatalf("unexpected error: %v", err)
	}
}
