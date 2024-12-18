package xuexitong

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/yatori-dev/yatori-go-core/api/entity"
	"github.com/yatori-dev/yatori-go-core/api/xuexitong"
	"github.com/yatori-dev/yatori-go-core/models/ctype"
	"golang.org/x/net/html"
	"log"
	"regexp"
	"strconv"
	"strings"
)

// Card 代表卡片信息
type Card struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	CardOrder   int    `json:"cardorder"`
	KnowledgeID int    `json:"knowledgeid"`
}

// DataItem 代表data数组中的每个项目
type DataItem struct {
	ClickCount     int    `json:"clickcount"`
	Createtime     int64  `json:"createtime"`
	OpenLock       int    `json:"openlock"`
	IndexOrder     int    `json:"indexorder"`
	Name           string `json:"name"`
	LastModifyTime int64  `json:"lastmodifytime"`
	ID             int    `json:"id"`
	Label          string `json:"label"`
	Layer          int    `json:"layer"`
	Card           struct {
		Data []Card `json:"data"`
	} `json:"card"`
	ParentNodeID int    `json:"parentnodeid"`
	Status       string `json:"status"`
}

// APIResponse 代表API返回的完整JSON结构
type APIResponse struct {
	Data []DataItem `json:"data"`
}

// ChapterFetchCardsAction 解析章节节点
// return: Card(节点总数结构), []interface{}(解析出可被刷取的节点结构), error
// 三数据返回
// 节点总数在界面请求中需要他们的index做对应渲染 解析后的需要与后续节点请求刷取中的参数对应

func ChapterFetchCardsAction(
	cache *xuexitong.XueXiTUserCache,
	chapters *ChaptersList,
	nodes []int,
	index, courseId, classId, cpi int) ([]Card, []entity.PointDto, error) {
	var apiResp APIResponse

	var pointObj entity.PointDto

	cords, err := cache.FetchChapterCords(nodes, index, courseId)
	if err != nil {
		return []Card{}, nil, err
	}
	if err := json.NewDecoder(bytes.NewBuffer([]byte(cords))).Decode(&apiResp); err != nil {
		return []Card{}, nil, err
	}
	if len(apiResp.Data) == 0 {
		log.Printf("获取章节任务节点卡片失败 [%s:%s(Id.%d)]",
			chapters.Knowledge[index].Label, chapters.Knowledge[index].Name, chapters.Knowledge[index].ID)
		return []Card{}, nil, err
	}

	dataItem := apiResp.Data[0]
	cards := dataItem.Card.Data
	log.Printf("获取章节任务节点卡片成功 共 %d 个 [%s:%s(Id.%d)]",
		len(cards),
		chapters.Knowledge[index].Label, chapters.Knowledge[index].Name, chapters.Knowledge[index].ID)

	pointObjs := make([]entity.PointDto, 0)
	for cardIndex, card := range cards {
		if card.Description == "" {
			log.Printf("(%d) 卡片 iframe 不存在 %+v", cardIndex, card)
			continue
		}
		points, err := parseIframeData(card.Description)
		if err != nil {
			log.Printf("解析卡片失败 %v", err)
			continue
		}
		log.Printf("(%d) 解析卡片成功 共 %d 个任务点", cardIndex, len(points))

		for pointIndex, point := range points {
			pointType, ok := point.Other["module"]
			if !ok {
				log.Printf("(%d, %d) 任务点 type 不存在 %+v", cardIndex, pointIndex, point)
				continue
			}

			if !point.HasData {
				log.Printf("(%d, %d) 任务点 data 为空或不存在 %+v", cardIndex, pointIndex, point)
				continue
			}

			// 这里data的有些参数可能还会出现参数不存在的问题 导致interface{} is nil, not from string
			// 在console正式发布后需要用户的实际反馈修改
			switch pointType {
			case string(ctype.Video):
				if objectID, ok := point.Data["objectid"].(string); ok && objectID != "" {
					pointObj.PointVideoDto = entity.PointVideoDto{
						CardIndex:   cardIndex,
						CourseID:    strconv.Itoa(courseId),
						ClassID:     strconv.Itoa(classId),
						KnowledgeID: card.KnowledgeID,
						Cpi:         strconv.Itoa(cpi),
						ObjectID:    objectID,
						Type:        ctype.Video,
						IsSet:       ok,
					}
				} else {
					log.Printf("(%d, %d) 任务点 'objectid' 不存在或为空 %+v", cardIndex, pointIndex, point)
					continue
				}
			case string(ctype.Work):

				workID, ok1 := point.Data["workid"].(string)
				// 此ID可能有时候不存在 暂不知有何作用先不做强制处理
				schoolID, _ := point.Data["schoolid"].(string)
				jobID, ok3 := point.Data["_jobid"].(string)

				if schoolID == "" {
					schoolID = "0"
				}

				if ok1 && workID != "" && ok3 && jobID != "" {
					pointObj.PointWorkDto = entity.PointWorkDto{
						CardIndex:   cardIndex,
						CourseID:    strconv.Itoa(courseId),
						ClassID:     strconv.Itoa(classId),
						KnowledgeID: card.KnowledgeID,
						Cpi:         strconv.Itoa(cpi),
						WorkID:      workID,
						SchoolID:    schoolID,
						JobID:       jobID,
						Type:        ctype.Work,
						IsSet:       ok,
					}
				} else {
					log.Printf("(%d, %d) 任务点 'workid', 'schoolid' 或 '_jobid' 不存在或为空 %+v", cardIndex, pointIndex, point)
					continue
				}
			case string(ctype.Insertdoc):
				// 同为文档类型，暂未做区分
				fallthrough
			case string(ctype.Document):

				jobID, ok3 := point.Data["_jobid"].(string)

				if objectID, ok := point.Data["objectid"].(string); ok && objectID != "" && ok3 && jobID != "" {
					pointObj.PointDocumentDto = entity.PointDocumentDto{
						CardIndex:   cardIndex,
						CourseID:    strconv.Itoa(courseId),
						ClassID:     strconv.Itoa(classId),
						KnowledgeID: card.KnowledgeID,
						Cpi:         strconv.Itoa(cpi),
						ObjectID:    objectID,
						JobID:       jobID,
						Type:        ctype.Document,
						IsSet:       ok,
					}
				} else {
					log.Printf("(%d, %d) 任务点 'objectid' 不存在或为空 %+v", cardIndex, pointIndex, point)
					continue
				}
			default:
				log.Printf("未知的任务点类型: %s\n", pointType)
				log.Printf("%+v", point)
				continue
			}

			pointObjs = append(pointObjs, pointObj)
		}
	}

	log.Printf("章节 可刷取任务节点解析成功 共 %d 个 [%s:%s(Id.%d)]",
		len(pointObjs), chapters.Knowledge[index].Label, chapters.Knowledge[index].Name, chapters.Knowledge[index].ID)
	return cards, pointObjs, nil
}

// IframeAttributes iframe 的属性
type IframeAttributes struct {
	Data    map[string]interface{} `json:"data"`
	Other   map[string]string
	HasData bool // 表示data属性是否存在且非空
}

func parseIframeData(htmlString string) ([]IframeAttributes, error) {
	// 解析HTML内容
	node, err := html.Parse(strings.NewReader(htmlString))
	if err != nil {
		return nil, err
	}

	var iframes []IframeAttributes
	var traverse func(n *html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "iframe" {
			attrs := IframeAttributes{
				Other: make(map[string]string),
			}
			hasData := false
			for _, attr := range n.Attr {
				if attr.Key == "data" && strings.TrimSpace(attr.Val) != "" {
					hasData = true
					// 清理data字符串：移除多余的空格和转义引号
					cleanedData := strings.ReplaceAll(attr.Val, "&quot;", "\"")
					cleanedData = regexp.MustCompile(`\s+`).ReplaceAllString(cleanedData, "")

					// 尝试将清理后的字符串解析为JSON对象
					if err := json.Unmarshal([]byte(cleanedData), &attrs.Data); err != nil {
						fmt.Printf("Failed to decode JSON: %v\n", err)
					}
				} else {
					attrs.Other[attr.Key] = attr.Val
				}
			}
			attrs.HasData = hasData
			iframes = append(iframes, attrs)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(node)
	return iframes, nil
}
