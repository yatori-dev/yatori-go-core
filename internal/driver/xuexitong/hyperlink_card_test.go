package xuexitong

import (
	"fmt"
	"testing"
	"time"

	"github.com/yatori-dev/yatori-go-core/v3/core"
	api "github.com/yatori-dev/yatori-go-core/v3/internal/api/xuexitong"
)

func TestMapPointsEnrichesHyperlinkFromCard(t *testing.T) {
	cache := &api.XueXiTUserCache{Name: "mock-hyperlink-account"}
	key := fmt.Sprintf("%s|%d|%d|%d|%d", cache.Name, 3, 2, 7, 1)
	mobileCardCache.Store(key, mobileCardCacheEntry{
		at: time.Now(),
		card: map[string]interface{}{"attachments": []interface{}{
			map[string]interface{}{"jobid": "example-link", "jtoken": "mock-token", "property": map[string]interface{}{"title": "External reading"}},
		}},
	})
	defer mobileCardCache.Delete(key)
	c := client{cache: cache}
	nodes := c.mapPoints("2", 1, 3, []api.PointDto{{PointHyperlinkDto: api.PointHyperlinkDto{
		JobID: "example-link", KnowledgeID: 7, CardIndex: 1, CourseID: "2", ClassID: "3", IsSet: true,
	}}})
	if len(nodes) != 1 || nodes[0].Kind != core.NodeLink || nodes[0].Name != "External reading" {
		t.Fatalf("bad mapped hyperlink: %+v", nodes)
	}
	raw, ok := nodes[0].Raw.(*api.PointHyperlinkDto)
	if !ok || raw.Jtoken != "mock-token" {
		t.Fatal("mapped hyperlink is missing the attachment token")
	}
}
