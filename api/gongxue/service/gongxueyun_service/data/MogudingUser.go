package data

// MogudingUser 代表一个蘑菇钉用户
type MogudingUser struct {
	// 对接数据库 ID
	ID int `json:"id"`

	// 账号
	Username string `json:"username"`

	// 密码
	Password string `json:"password"`

	UserID   string `json:"userId"`
	Token    string `json:"token"`
	UserType string `json:"userType"`

	PlanID   string `json:"planId"`
	PlanName string `json:"planName"`
	Email    string `json:"email"`

	SubmitType  string `json:"submitType"`
	SubmitTitle string `json:"submitTitle"`

	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`

	WeekTotalFlag int    `json:"weekTotalFlag"`
	Content       string `json:"content"`

	IsVipType int `json:"isVipType"`
}
