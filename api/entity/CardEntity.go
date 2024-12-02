package entity

import (
	"log"
	"net/http"
)

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
	FID        int
	DToken     string
	Duration   int
	JobID      string
	OtherInfo  string
	Title      string
	RT         float64
	Logger     *log.Logger
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
