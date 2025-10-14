package xuexitong

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	ddddocr "github.com/Changbaiqi/ddddocr-go/utils"
	"github.com/thedevsaddam/gojsonq"
	ort "github.com/yalue/onnxruntime_go"
	"github.com/yatori-dev/yatori-go-core/api/entity"
	"github.com/yatori-dev/yatori-go-core/api/xuexitong"
	"github.com/yatori-dev/yatori-go-core/models/ctype"
	"github.com/yatori-dev/yatori-go-core/utils"
	log2 "github.com/yatori-dev/yatori-go-core/utils/log"
	"golang.org/x/net/html"
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

	cords, err := cache.FetchChapterCords(nodes, index, courseId, 5, nil)

	if err != nil {
		if err.Error() == "status code: 500" {
			log2.Print(log2.DEBUG, "触发请求频繁500，自动重新登录")
			ReLogin(cache)                                                       //越过验证码或者202
			cords, err = cache.FetchChapterCords(nodes, index, courseId, 5, nil) //尝试重新拉取卡片信息
			if err != nil {
				log2.Print(log2.DEBUG, "重新登录后cords拉取错误err值>>", fmt.Sprintf("%s", err.Error()))
			}
			log2.Print(log2.DEBUG, "重新登录后cords值>>", fmt.Sprintf("%+v", cords))
		} else if err.Error() == "触发验证码" {
			log2.Print(log2.DEBUG, utils.RunFuncName(), "触发验证码，正在进行AI智能识别绕过.....")
			for {
				codePath, err1 := cache.XueXiTVerificationCodeApi(5, nil)
				if err1 != nil {
					return nil, nil, err1
				}
				if codePath == "" { //如果path为空，那么可能是账号问题
					return nil, nil, errors.New("无法正常获取对应网站验证码，请检查对应url是否正常")
				}
				img, _ := utils.ReadImg(codePath) //读取验证码图片
				//codeResult := utils.AutoVerification(img, ort.NewShape(1, 23)) //自动识别
				codeResult := ddddocr.SemiOCRVerification(img, ort.NewShape(1, 23))
				utils.DeleteFile(codePath) //删除验证码文件
				status, err1 := cache.XueXiTPassVerificationCode(codeResult, 5, nil)
				//fmt.Println(codeResult)
				//fmt.Println(status)
				if status {
					break
				}
			}
			cords, err = cache.FetchChapterCords(nodes, index, courseId, 5, nil) //尝试重新拉取卡片信息
			log2.Print(log2.DEBUG, utils.RunFuncName(), "绕过成功")
		} else if strings.Contains(err.Error(), "status code: 202") {
			ReLogin(cache)                                                       //重登
			cords, err = cache.FetchChapterCords(nodes, index, courseId, 5, nil) //尝试重新拉取卡片信息
		} else if strings.Contains(err.Error(), "status code: 400") {
			ReLogin(cache)                                                       //重登
			cords, err = cache.FetchChapterCords(nodes, index, courseId, 5, nil) //尝试重新拉取卡片信息
		} else if strings.Contains(err.Error(), "status code: 403") {
			ReLogin(cache)                                                       //重登
			cords, err = cache.FetchChapterCords(nodes, index, courseId, 5, nil) //尝试重新拉取卡片信息
		}
	}

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
					//视屏名称获取
					name := point.Data["name"]
					if name != nil {
						titleStr, turnErr := url.QueryUnescape(name.(string))
						//转换
						if turnErr != nil {
							log2.Print(log2.DEBUG, titleStr, "解码失败")
							pointObj.PointVideoDto.Title = titleStr
						} else {
							pointObj.PointVideoDto.Title = titleStr
						}
					}

					cords2, _ := cache.FetchChapterCords2(strconv.Itoa(classId), strconv.Itoa(courseId), strconv.Itoa(card.KnowledgeID), strconv.Itoa(cpi), 3, nil)
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
			case string(ctype.InsertBook):
				//InserBoot类型直接当文档处理
				jobID, ok3 := point.Data["_jobid"].(string)

				if ok3 && jobID != "" {
					pointObj.PointDocumentDto = entity.PointDocumentDto{
						CardIndex:   cardIndex,
						CourseID:    strconv.Itoa(courseId),
						ClassID:     strconv.Itoa(classId),
						KnowledgeID: card.KnowledgeID,
						Cpi:         strconv.Itoa(cpi),
						ObjectID:    "",
						JobID:       jobID,
						Type:        ctype.InsertBook,
						IsSet:       ok,
					}
				} else {
					log2.Print(log2.DEBUG, "(%d, %d) 任务点 'objectid' 不存在或为空 %+v", cardIndex, pointIndex, point)
					continue
				}
			case string(ctype.Hyperlink):

				jobID, ok3 := point.Data["_jobid"].(string)
				linkType, ok4 := point.Data["linkType"].(float64)

				if ok3 && jobID != "" {
					pointObj.PointHyperlinkDto = entity.PointHyperlinkDto{
						CardIndex:   cardIndex,
						CourseID:    strconv.Itoa(courseId),
						ClassID:     strconv.Itoa(classId),
						KnowledgeID: card.KnowledgeID,
						Cpi:         strconv.Itoa(cpi),
						ObjectID:    "",
						JobID:       jobID,
						Type:        ctype.Hyperlink,
						IsSet:       ok,
					}
					if ok4 {
						pointObj.PointHyperlinkDto.LinkType = int(linkType)
					}
				} else {
					log2.Print(log2.DEBUG, "(%d, %d) 任务点 'objectid' 不存在或为空 %+v", cardIndex, pointIndex, point)
					continue
				}
			case string(ctype.Insertlive):
				module, ok1 := point.Other["module"]
				jobID, ok3 := point.Data["_jobid"].(string)
				title, ok4 := point.Data["title"].(string)
				liveStatus, ok5 := point.Data["liveStatus"].(string)
				streamName, ok6 := point.Data["streamName"].(string)
				liveId, ok7 := point.Data["liveId"].(float64)
				live, ok8 := point.Data["live"].(bool)
				vdoid, ok9 := point.Data["vdoid"].(string)
				if ok3 && jobID != "" {
					pointObj.PointLiveDto = entity.PointLiveDto{
						CardIndex:   cardIndex,
						CourseID:    strconv.Itoa(courseId),
						ClassID:     strconv.Itoa(classId),
						KnowledgeID: card.KnowledgeID,
						Cpi:         strconv.Itoa(cpi),
						JobID:       jobID,
						Type:        ctype.Insertlive,
						IsSet:       ok,
					}
					if ok1 {
						pointObj.PointLiveDto.Module = module
					}

					//这里比较特殊，userid其实是cookie里面的_uid值
					cookies := cache.GetCookies()
					for _, cookie := range cookies {
						if cookie.Name == "_uid" {
							pointObj.PointLiveDto.UserId = cookie.Value
							break
						}
					}
					if ok4 {
						pointObj.PointLiveDto.Title = title
					}
					if ok5 {
						pointObj.PointLiveDto.LiveStatusStr = liveStatus
					}
					if ok6 {
						pointObj.PointLiveDto.StreamName = streamName
					}
					if ok7 {
						pointObj.PointLiveDto.LiveId = strconv.FormatInt(int64(liveId), 10)
					}
					if ok8 {
						pointObj.PointLiveDto.Live = live
					}
					if ok9 {
						pointObj.PointLiveDto.Vdoid = vdoid
					}
				} else {
					log2.Print(log2.DEBUG, "(%d, %d) 任务点 'objectid' 不存在或为空 %+v", cardIndex, pointIndex, point)
					continue
				}
			case string(ctype.Insertbbs):
				module, ok1 := point.Other["module"]
				jobID, ok3 := point.Data["_jobid"].(string)
				title, ok4 := point.Data["title"].(string)
				detail, ok5 := point.Data["detail"].(string)
				mid, ok6 := point.Data["mid"].(string)
				allowViewReply, ok7 := point.Data["allowViewReply"].(float64)
				replytimes, ok8 := point.Data["replytimes"].(string)
				replywordnum, ok9 := point.Data["replywordnum"].(string)
				endtime, ok10 := point.Data["endtime"].(string)
				isJob, ok11 := point.Data["isJob"].(bool)
				if ok3 && jobID != "" {
					pointObj.PointBBsDto = entity.PointBBsDto{
						CardIndex:   cardIndex,
						CourseID:    strconv.Itoa(courseId),
						ClassID:     strconv.Itoa(classId),
						KnowledgeID: card.KnowledgeID,
						Cpi:         strconv.Itoa(cpi),
						JobID:       jobID,
						Type:        ctype.Insertbbs,
						IsSet:       ok,
					}
					if ok1 {
						pointObj.PointBBsDto.Module = module
					}

					//这里比较特殊，userid其实是cookie里面的_uid值
					cookies := cache.GetCookies()
					for _, cookie := range cookies {
						if cookie.Name == "_uid" {
							pointObj.PointLiveDto.UserId = cookie.Value
							break
						}
					}
					if ok4 {
						pointObj.PointBBsDto.Title = title
					}
					if ok5 {
						pointObj.PointBBsDto.Detail = detail
					}
					if ok6 {
						pointObj.PointBBsDto.Mid = mid
					}
					if ok7 {
						pointObj.PointBBsDto.AllowViewReply = int(allowViewReply)
					}
					if ok8 {
						pointObj.PointBBsDto.ReplyTimes = replytimes
					}
					if ok9 {
						pointObj.PointBBsDto.ReplayWordNum = replywordnum
					}
					if ok10 {
						pointObj.PointBBsDto.EndTime = endtime
					}
					if ok11 {
						pointObj.PointBBsDto.IsJob = isJob
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

// 每次进入章节前进行一次调用，防止0任务点无法学习的情况
func EnterChapterForwardCallAction(cache *xuexitong.XueXiTUserCache, courseId, clazzid, chapterId, cpi string) error {
	err := cache.EnterChapterForwardCallApi(courseId, clazzid, chapterId, cpi, 3, nil)
	if err != nil && strings.Contains(err.Error(), "status code: 500") {
		ReLogin(cache) //重登
		err = cache.EnterChapterForwardCallApi(courseId, clazzid, chapterId, cpi, 3, nil)
	}
	if err != nil {
		if err.Error() == "status code: 500" {
			log2.Print(log2.DEBUG, "触发请求频繁500，自动重新登录")
			ReLogin(cache)                                                                    //越过验证码或者202
			err = cache.EnterChapterForwardCallApi(courseId, clazzid, chapterId, cpi, 3, nil) //尝试重新拉取卡片信息
			if err != nil {
				log2.Print(log2.DEBUG, "重新登录后cords拉取错误err值>>", fmt.Sprintf("%s", err.Error()))
			}
		} else if err.Error() == "触发验证码" {
			log2.Print(log2.DEBUG, utils.RunFuncName(), "触发验证码，正在进行AI智能识别绕过.....")
			for {
				codePath, err1 := cache.XueXiTVerificationCodeApi(5, nil)
				if err1 != nil {
					return err1
				}
				if codePath == "" { //如果path为空，那么可能是账号问题
					return errors.New("无法正常获取对应网站验证码，请检查对应url是否正常")
				}
				img, _ := utils.ReadImg(codePath) //读取验证码图片
				codeResult := ddddocr.SemiOCRVerification(img, ort.NewShape(1, 23))
				utils.DeleteFile(codePath) //删除验证码文件
				status, err1 := cache.XueXiTPassVerificationCode(codeResult, 5, nil)
				if status {
					break
				}
			}
			err = cache.EnterChapterForwardCallApi(courseId, clazzid, chapterId, cpi, 3, nil) //尝试重新拉取卡片信息
			log2.Print(log2.DEBUG, utils.RunFuncName(), "绕过成功")
		} else if strings.Contains(err.Error(), "status code: 202") {
			ReLogin(cache)                                                                    //重登
			err = cache.EnterChapterForwardCallApi(courseId, clazzid, chapterId, cpi, 3, nil) //尝试重新拉取卡片信息
		} else if strings.Contains(err.Error(), "status code: 400") {
			ReLogin(cache)                                                                    //重登
			err = cache.EnterChapterForwardCallApi(courseId, clazzid, chapterId, cpi, 3, nil) //尝试重新拉取卡片信息
		} else if strings.Contains(err.Error(), "status code: 403") {
			ReLogin(cache)                                                                    //重登
			err = cache.EnterChapterForwardCallApi(courseId, clazzid, chapterId, cpi, 3, nil) //尝试重新拉取卡片信息
		}
	}

	return err
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
