package xuexitong

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/api/entity"
	"github.com/yatori-dev/yatori-go-core/api/xuexitong"
	"golang.org/x/net/html"
	"log"
	"regexp"
	"strings"
)

// ChapterNotOpened 是未打开章节时的错误类型
type ChapterNotOpened struct{}

func (e ChapterNotOpened) Error() string {
	return "章节未开放"
}

// APIError 是 API 相关错误的一般错误
// 这里之后统一做整合
type APIError struct {
	Message string
}

func (e APIError) Error() string {
	return e.Message
}

func PageMobileChapterCardAction(
	cache *xuexitong.XueXiTUserCache,
	classId, courseId, knowledgeId, cardIndex, cpi int) (interface{}, error) {
	cardHtml, err := cache.PageMobileChapterCard(classId, courseId, knowledgeId, cardIndex, cpi)

	var att interface{}

	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	doc, err := html.Parse(bytes.NewReader([]byte(cardHtml)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var scriptContent string
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "script" {
			// Check if the script tag has a type attribute and its value is "text/javascript"
			hasTypeAttr := false
			for _, attr := range n.Attr {
				if attr.Key == "type" && strings.TrimSpace(attr.Val) == "text/javascript" {
					hasTypeAttr = true
					break
				}
			}

			if hasTypeAttr {
				// Check if the script tag has a child node and it's a text node
				if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
					scriptContent = strings.TrimSpace(n.FirstChild.Data)
					return // Exit the traversal as we found what we need
				}
			}
		}

		// Continue traversing the children nodes
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)

	// Define the regex pattern to match window.AttachmentSetting JSON object
	re := regexp.MustCompile(`window\.AttachmentSetting\s*=\s*(\{.*?\});`)
	matches := re.FindStringSubmatch(scriptContent)
	if len(matches) > 1 {
		var attachment interface{}
		if err := json.Unmarshal([]byte(matches[1]), &attachment); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}
		att = attachment
	} else {
		var blankTips string
		traverse = func(n *html.Node) {
			if n.Type == html.ElementNode && n.Data == "p" {
				for _, attr := range n.Attr {
					if attr.Key == "class" && strings.Contains(attr.Val, "blankTips") {
						blankTips = strings.TrimSpace(n.FirstChild.Data)
						return
					}
				}
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				traverse(c)
			}
		}
		traverse(doc)

		if strings.TrimSpace(blankTips) == "章节未开放！" {
			log.Println("章节未开放")
			return ChapterNotOpened{}, nil
		}
		return APIError{Message: blankTips}, nil
	}

	log.Println("Attachment拉取成功")

	return att, nil
}

func VideoDtoFetchAction(cache *xuexitong.XueXiTUserCache, p *entity.PointVideoDto) (bool, error) {
	fetch, err := cache.VideoDtoFetch(p)
	if err != nil {
		log.Println("VideoDtoFetchAction:", err)
		return false, err
	}
	dtoken := gojsonq.New().JSONString(fetch).Find("dtoken").(string)
	duration := gojsonq.New().JSONString(fetch).Find("duration").(float64)

	p.DToken = dtoken
	p.Duration = int(duration)
	p.Title = gojsonq.New().JSONString(fetch).Find("filename").(string)

	if gojsonq.New().JSONString(fetch).Find("status").(string) == "success" {
		return true, nil
	}

	log.Println("Fetch failed")
	return false, nil
}
