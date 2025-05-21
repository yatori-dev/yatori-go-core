package entity

import (
	"errors"
	"github.com/yatori-dev/yatori-go-core/models/ctype"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type IAttachment interface {
	AttachmentsDetection(attachment interface{}) (bool, error)
}

type PointDto struct {
	PointVideoDto
	PointWorkDto
	PointDocumentDto
}

// PointVideoDto 视频任务点
type PointVideoDto struct {
	CardIndex   int
	CourseID    string
	ClassID     string
	KnowledgeID int
	Cpi         string
	ObjectID    string
	// 从SSR视图中获取
	isPassed  bool
	FID       int
	DToken    string
	PlayTime  int
	Duration  int
	JobID     string
	OtherInfo string
	Title     string
	RT        float64
	Logger    *log.Logger
	PUID      string
	Session   *Session

	Type  ctype.CardType
	IsSet bool

	Attachment interface{} //视图获取后的原始map
}

// PointWorkDto 测验任务点
type PointWorkDto struct {
	CardIndex   int
	CourseID    string
	ClassID     string
	KnowledgeID int
	Cpi         string
	WorkID      string
	SchoolID    string
	JobID       string
	PUID        string
	KToken      string
	Enc         string

	Type  ctype.CardType
	IsSet bool
}

// WorkInputField represents an <input> element in the HTML form.
type WorkInputField struct {
	Name  string
	Value string
	Type  string // Optional: to store the type attribute if needed
	ID    string // Store the id attribute if present
}

// PointDocumentDto 文档查看任务点
type PointDocumentDto struct {
	CardIndex   int
	CourseID    string
	ClassID     string
	KnowledgeID int
	Cpi         string

	ObjectID string
	Title    string
	JobID    string
	Jtoken   string

	Type  ctype.CardType
	IsSet bool
}
type Session struct {
	Client *http.Client
	Acc    *Account
}

type Account struct {
	PUID string
}

func ParsePointDto(pointDTOs []PointDto) (videoDTOs []PointVideoDto, workDTOs []PointWorkDto, documentDTOs []PointDocumentDto) {
	//处理返回的任务点对象,这里不要使用else，因为可能会有多个不同类型的任务对象
	for _, card := range pointDTOs {
		if card.PointWorkDto.IsSet == true {
			workDTOs = append(workDTOs, card.PointWorkDto)
		} else if card.PointVideoDto.IsSet == true {
			if card.OtherInfo == "" {
			}
			videoDTOs = append(videoDTOs, card.PointVideoDto)
		} else if card.PointDocumentDto.IsSet == true {
			documentDTOs = append(documentDTOs, card.PointDocumentDto)
		}
	}
	return
}

// AttachmentsDetection 使用接口对每种DTO进行检测再次赋值, 以对应后续的刷取请求
func (p *PointVideoDto) AttachmentsDetection(attachment interface{}) (bool, error) {
	attachmentMap, ok := attachment.(map[string]interface{})
	if !ok {
		return false, errors.New("无法将 Attachment 转换为 map[string]interface{}")
	}
	attachments, ok := attachmentMap["attachments"].([]interface{})
	if !ok {
		return false, errors.New("invalid attachment structure")
	}

	for _, a := range attachments {
		attachment, _ := a.(map[string]interface{})
		property, ok := attachment["property"].(map[string]interface{})
		if !ok {
			return false, errors.New("invalid property structure")
		}
		//防止一个节点多种任务点类型 其他property中没有objectid报panic
		objectid := property["objectid"]
		if objectid == nil {
			continue
		}

		if objectid == p.ObjectID {
			var otherInfo string
			jobID, ok := property["jobid"].(string)
			if !ok {
				jobID2, ok := property["jobid"].(float64)
				if !ok {
					return false, errors.New("invalid jobid structure")
				}
				p.JobID = strconv.FormatFloat(jobID2, 'f', -1, 64)
			} else {
				p.JobID = jobID
			}
			parts := strings.SplitN(attachment["otherInfo"].(string), "&", 2)
			if len(parts) > 0 {
				otherInfo = parts[0]
			}
			p.OtherInfo = otherInfo
			if isPassed, ok := attachment["isPassed"].(bool); ok {
				p.isPassed = isPassed
			} else {
				p.isPassed = false
			}

			// 获取 "rt" 的值
			rt, ok := property["rt"].(float64)
			if !ok {
				// 如果 "rt" 键不存在，则使用默认值 0.9
				rt = 0.9
			} else {
				rt = 0.9
			}
			playTime, ok := attachment["playTime"].(float64)
			if !ok {
				p.PlayTime = 0
			} else {
				p.PlayTime = int(playTime) / 1000
			}

			p.RT = rt
			p.Attachment = attachment
			break
		}
	}
	if p.Attachment == nil {
		p.Logger.Println("Failed to locate resource")
		return false, nil
	}
	defaults, ok := attachmentMap["defaults"].(map[string]interface{})
	if !ok {
		return false, errors.New("invalid defaults structure")
	}
	fid, _ := strconv.Atoi(defaults["fid"].(string))
	p.FID = fid
	p.PUID = defaults["userid"].(string)

	return true, nil
}

func (p *PointWorkDto) AttachmentsDetection(attachment interface{}) (bool, error) {
	attachmentMap, ok := attachment.(map[string]interface{})
	var flag bool
	if !ok {
		return false, errors.New("无法将 Attachment 转换为 map[string]interface{}")
	}
	attachments, ok := attachmentMap["attachments"].([]interface{})
	if !ok {
		return false, errors.New("invalid attachment structure")
	}
	for _, a := range attachments {
		att, _ := a.(map[string]interface{})
		property, ok := att["property"].(map[string]interface{})
		if !ok {
			return false, errors.New("invalid property structure")
		}
		workId := property["workid"]
		if workId == nil {
			continue
		}
		if workId == p.WorkID {
			p.Enc = att["enc"].(string)
			if att["job"] == nil {
				flag = false
			} else {
				flag = att["job"].(bool)
			}

			break
		}
	}
	defaults, ok := attachmentMap["defaults"].(map[string]interface{})
	if !ok {
		return false, errors.New("invalid defaults structure")
	}
	p.KToken = defaults["ktoken"].(string)
	p.PUID = defaults["userid"].(string)
	return flag, nil
}

func (p *PointDocumentDto) AttachmentsDetection(attachment interface{}) (bool, error) {
	attachmentMap, ok := attachment.(map[string]interface{})
	if !ok {
		return false, errors.New("无法将 Attachment 转换为 map[string]interface{}")
	}
	attachments, ok := attachmentMap["attachments"].([]interface{})
	if !ok {
		return false, errors.New("invalid attachment structure")
	}

	for _, a := range attachments {
		att, _ := a.(map[string]interface{})

		// 如果未给出文档类型（垃圾学习通，一点都不规范），那么先进行文档解析尝试。
		if att["type"] == nil {
			property, ok := att["property"].(map[string]interface{})
			if !ok {
				return false, errors.New("invalid property structure")
			}
			objectid := property["objectid"]
			if objectid == p.ObjectID {
				p.Title = property["name"].(string)
				if property["jobid"] == nil {
					p.JobID = ""
				} else {
					p.JobID = property["jobid"].(string)
				}
				p.Jtoken = att["jtoken"].(string)
			}
		} else if att["type"].(string) == "document" {
			property, ok := att["property"].(map[string]interface{})
			if !ok {
				return false, errors.New("invalid property structure")
			}
			objectid := property["objectid"]
			if objectid == p.ObjectID {
				p.Title = property["name"].(string)
				if property["jobid"] == nil {
					p.JobID = ""
				} else {
					p.JobID = property["jobid"].(string)
				}
				p.Jtoken = att["jtoken"].(string)
			}
		} else if att["type"].(string) == "wrodid" { //预留作业的

		}
	}
	return true, nil
}
