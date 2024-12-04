package entity

import (
	"errors"
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
	isPassed   bool
	FID        int
	DToken     string
	PlayTime   int
	Duration   int
	JobID      string
	OtherInfo  string
	Title      string
	RT         float64
	Logger     *log.Logger
	PUID       string
	Session    *Session
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
}

// PointDocumentDto 文档查看任务点
type PointDocumentDto struct {
	CardIndex   int
	CourseID    string
	ClassID     string
	KnowledgeID int
	Cpi         string
	ObjectID    string
}
type Session struct {
	Client *http.Client
	Acc    *Account
}

type Account struct {
	PUID string
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
		if property["objectid"] == p.ObjectID {
			var otherInfo string
			p.JobID = property["jobid"].(string)
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
			}
			p.PlayTime = int(attachment["playTime"].(float64)) / 1000
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
