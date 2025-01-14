package gongxueyun_service

type MoguDing struct {
	ID             int              `json:"ID"`
	UserId         string           `json:"userId"`
	RoleKey        string           `json:"roleKey"`
	Authorization  string           `json:"authorization"`
	PlanID         string           `json:"planId"`
	PlanName       string           `json:"planName"`
	PhoneNumber    string           `json:"phoneNumber"`
	Password       string           `json:"password"`
	Email          string           `json:"email"`
	Sign           SignInfo         `json:"sign"`
	CommParameters commonParameters `json:"commParameters"`
	ReportStruct   report           `json:"report"`
	WeekTime       weekTime         `json:"weekTime"`
	JobInfo        JobInfo          `json:"jobInfo"`
}
type SignInfo struct {
	//	构造签到信息
	Address   string `json:"address"`
	City      string `json:"city"`
	Area      string `json:"area"`
	Country   string `json:"country"`
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
	Province  string `json:"province"`
}
type JobInfo struct {
	JobName     string `json:"jobName"`
	Address     string `json:"address"`
	CompanyName string `json:"companyName"`
}
type report struct {
	CreateTime string `json:"createTime"`
	ReportId   string `json:"reportId"`
	ReportType string `json:"reportType"`
	Flag       int    `json:"flag"`
}
type weekTime struct {
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	Week      string `json:"week"`
	IsDefault int    `json:"isDefault"`
	Flag      int    `json:"flag"`
}
type commonParameters struct {
	token     string
	secretKey string
	xY        string
	captcha   string
	JobId     string
}

// Message represents a single message in the conversation.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// RequestData represents the request payload.
type RequestData struct {
	MaxTokens   int       `json:"max_tokens"`
	TopK        int       `json:"top_k"`
	Temperature float64   `json:"temperature"`
	Messages    []Message `json:"messages"`
	Model       string    `json:"model"`
	Stream      bool      `json:"stream"`
}
