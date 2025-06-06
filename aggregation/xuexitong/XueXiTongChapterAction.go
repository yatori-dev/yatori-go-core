package xuexitong

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/api/entity"
	"github.com/yatori-dev/yatori-go-core/api/xuexitong"
	"github.com/yatori-dev/yatori-go-core/models/ctype"
	log2 "github.com/yatori-dev/yatori-go-core/utils/log"
	"golang.org/x/net/html"
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

	cords, err := cache.FetchChapterCords(nodes, index, courseId)

	if err != nil {
		return []Card{}, nil, err
	}
	if err := json.NewDecoder(bytes.NewBuffer([]byte(cords))).Decode(&apiResp); err != nil {
		return []Card{}, nil, err
	}
	if len(apiResp.Data) == 0 {
		log2.Print(log2.DEBUG, "获取章节任务节点卡片失败 [", chapters.Knowledge[index].Label, ":", chapters.Knowledge[index].Name, "(Id.", fmt.Sprintf("%d", chapters.Knowledge[index].ID), ")]")
		return []Card{}, nil, err
	}

	dataItem := apiResp.Data[0]
	cards := dataItem.Card.Data
	//log.Printf("获取章节任务节点卡片成功 共 %d 个 [%s:%s(Id.%d)]",
	//	len(cards),
	//	chapters.Knowledge[index].Label, chapters.Knowledge[index].Name, chapters.Knowledge[index].ID)

	pointObjs := make([]entity.PointDto, 0)
	for cardIndex, card := range cards {
		if card.Description == "" {
			log2.Print(log2.DEBUG, "(", fmt.Sprintf("%d", cardIndex), ") 卡片 iframe 不存在 ", fmt.Sprintf("%+v", card))
			continue
		}
		points, err := parseIframeData(card.Description)
		if err != nil {
			log2.Print(log2.DEBUG, "解析卡片失败 %v", err)
			continue
		}
		log2.Print(log2.DEBUG, fmt.Sprintf("%d", cardIndex), "解析卡片成功 共 ", fmt.Sprintf("%d", len(points)), "个任务点")

		for pointIndex, point := range points {
			var pointObj entity.PointDto //不要乱移动这玩意位置，OK？
			pointType, ok := point.Other["module"]
			if !ok {
				log2.Print(log2.DEBUG, "(", fmt.Sprintf("%d", cardIndex), ", ", fmt.Sprintf("%d", pointIndex), ") 任务点 type 不存在 %+v", fmt.Sprintf("%+v", point))
				continue
			}

			if !point.HasData {
				log2.Print(log2.DEBUG, "(%d, %d) 任务点 data 为空或不存在 %+v", cardIndex, pointIndex, point)
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
					cords2, _ := cache.FetchChapterCords2(strconv.Itoa(classId), strconv.Itoa(courseId), strconv.Itoa(card.KnowledgeID), strconv.Itoa(cpi))
					find := gojsonq.New().JSONString(cords2).Find("attachments")
					if find != nil {
						list := gojsonq.New().JSONString(cords2).Find("attachments")
						if item, ok := list.([]interface{}); ok {
							for _, item1 := range item {
								if obj, ok := item1.(map[string]interface{}); ok {
									if obj["otherInfo"] != nil {
										if len(obj["otherInfo"].(string)) > 80 {
											pointObj.PointVideoDto.OtherInfo = obj["otherInfo"].(string)
											if obj["jobid"] != nil {
												pointObj.PointVideoDto.JobID = obj["jobid"].(string)
											}

										}
									}
									if obj["playTime"] != nil {
										pointObj.PointVideoDto.PlayTime = int(obj["playTime"].(float64)) / 1000
									}
								}

							}
						}
					}
					//if pointObj.PointVideoDto.OtherInfo == "" {
					//	cords2, _ := cache.FetchChapterCords2(strconv.Itoa(classId), strconv.Itoa(courseId), strconv.Itoa(card.KnowledgeID), strconv.Itoa(cardIndex), strconv.Itoa(cpi))
					//	//fmt.Println(cords2)
					//	sprintf := fmt.Sprintf(`nodeId_[\d]*-cpi_[\d]*-rt_d-ds_[^&]*`)
					//	compile := regexp.MustCompile(sprintf)
					//	find := compile.FindAllStringSubmatch(cords2, -1)
					//	for _, v := range find {
					//		pointObj.PointVideoDto.OtherInfo = v[0]
					//	}
					//}
				} else {
					log2.Print(log2.DEBUG, "(%d, %d) 任务点 'objectid' 不存在或为空 %+v", cardIndex, pointIndex, point)
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
					log2.Print(log2.DEBUG, "(%d, %d) 任务点 'workid', 'schoolid' 或 '_jobid' 不存在或为空 %+v", cardIndex, pointIndex, point)
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
					log2.Print(log2.DEBUG, "(%d, %d) 任务点 'objectid' 不存在或为空 %+v", cardIndex, pointIndex, point)
					continue
				}
			default:
				log2.Print(log2.DEBUG, "未知的任务点类型: %s\n", pointType)
				log2.Print(log2.DEBUG, "%+v", point)
				continue
			}

			pointObjs = append(pointObjs, pointObj)
		}
	}

	//log.Printf("章节 可刷取任务节点解析成功 共 %d 个 [%s:%s(Id.%d)]",
	//	len(pointObjs), chapters.Knowledge[index].Label, chapters.Knowledge[index].Name, chapters.Knowledge[index].ID)
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
