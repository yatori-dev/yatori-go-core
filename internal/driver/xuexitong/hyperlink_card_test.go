package xuexitong

import (
	"fmt"
	"testing"
	"time"

	"github.com/yatori-dev/yatori-go-core/v3/core"
	api "github.com/yatori-dev/yatori-go-core/v3/internal/api/xuexitong"
)

func TestMapPointsEnrichesHyperlinkFromCard(t *testing.T) {
	cache := &api.XueXiTUserCache{Name: t.Name()}
	key := fmt.Sprintf("%s|%d|%d|%d|%d", cache.Name, 3, 2, 7, 1)
	mobileCardCache.Store(key, mobileCardCacheEntry{
		at: time.Now(),
		card: map[string]interface{}{"attachments": []interface{}{
			map[string]interface{}{
				"jobid":  "iframe-link",
				"jtoken": "synthetic-token",
				"property": map[string]interface{}{
					"jobid": "report-link",
					"title": "External reading",
				},
			},
		}},
	})
	t.Cleanup(func() { mobileCardCache.Delete(key) })

	c := client{cache: cache}
	nodes := c.mapPoints("2", 1, 3, []api.PointDto{{PointHyperlinkDto: api.PointHyperlinkDto{
		JobID: "iframe-link", KnowledgeID: 7, CardIndex: 1, CourseID: "2", ClassID: "3", IsSet: true,
	}}})
	if len(nodes) != 1 || nodes[0].Kind != core.NodeLink {
		t.Fatalf("expected one hyperlink node, got %+v", nodes)
	}
	if nodes[0].Name != "External reading" {
		t.Errorf("got title %q, want attachment title", nodes[0].Name)
	}
	point, ok := nodes[0].Raw.(*api.PointHyperlinkDto)
	if !ok || point.JobID != "report-link" || point.Jtoken != "synthetic-token" {
		t.Fatalf("hyperlink is missing report metadata: %+v", nodes[0].Raw)
	}
}
