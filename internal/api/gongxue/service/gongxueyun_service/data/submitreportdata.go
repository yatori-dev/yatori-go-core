package data

type SubmitData struct {
	Address          interface{}   `json:"address,omitempty"`
	ApplyId          interface{}   `json:"applyId,omitempty"`
	ApplyName        interface{}   `json:"applyName,omitempty"`
	AttachmentList   interface{}   `json:"attachmentList,omitempty"`
	CommentNum       interface{}   `json:"commentNum,omitempty"`
	CommentContent   interface{}   `json:"commentContent,omitempty"`
	Content          string        `json:"content,omitempty"`
	CreateBy         interface{}   `json:"createBy,omitempty"`
	CreateTime       interface{}   `json:"createTime,omitempty"`
	DepName          interface{}   `json:"depName,omitempty"`
	Reject           interface{}   `json:"reject,omitempty"`
	EndTime          interface{}   `json:"endTime,omitempty"`
	HeadImg          interface{}   `json:"headImg,omitempty"`
	Yearmonth        interface{}   `json:"yearmonth,omitempty"`
	ImageList        interface{}   `json:"imageList,omitempty"`
	IsFine           interface{}   `json:"isFine,omitempty"`
	Latitude         interface{}   `json:"latitude,omitempty"`
	GpmsSchoolYear   interface{}   `json:"gpmsSchoolYear,omitempty"`
	Longitude        interface{}   `json:"longitude,omitempty"`
	PlanId           string        `json:"planId,omitempty"`
	PlanName         interface{}   `json:"planName,omitempty"`
	ReportId         interface{}   `json:"reportId,omitempty"`
	ReportType       string        `json:"reportType,omitempty"`
	ReportTime       string        `json:"reportTime,omitempty"`
	IsOnTime         interface{}   `json:"isOnTime,omitempty"`
	SchoolId         interface{}   `json:"schoolId,omitempty"`
	StartTime        interface{}   `json:"startTime,omitempty"`
	State            interface{}   `json:"state,omitempty"`
	StudentId        interface{}   `json:"studentId,omitempty"`
	StudentNumber    interface{}   `json:"studentNumber,omitempty"`
	SupportNum       interface{}   `json:"supportNum,omitempty"`
	Title            string        `json:"title,omitempty"`
	Url              interface{}   `json:"url,omitempty"`
	Username         interface{}   `json:"username,omitempty"`
	Weeks            interface{}   `json:"weeks,omitempty"`
	VideoUrl         interface{}   `json:"videoUrl,omitempty"`
	VideoTitle       interface{}   `json:"videoTitle,omitempty"`
	Attachments      string        `json:"attachments,omitempty"`
	CompanyName      interface{}   `json:"companyName,omitempty"`
	JobName          interface{}   `json:"jobName,omitempty"`
	JobId            string        `json:"jobId,omitempty"`
	Score            interface{}   `json:"score,omitempty"`
	TpJobId          interface{}   `json:"tpJobId,omitempty"`
	StarNum          interface{}   `json:"starNum,omitempty"`
	ConfirmDays      interface{}   `json:"confirmDays,omitempty"`
	IsApply          interface{}   `json:"isApply,omitempty"`
	CompStarNum      interface{}   `json:"compStarNum,omitempty"`
	CompScore        interface{}   `json:"compScore,omitempty"`
	CompComment      interface{}   `json:"compComment,omitempty"`
	CompState        interface{}   `json:"compState,omitempty"`
	Apply            interface{}   `json:"apply,omitempty"`
	LevelEntity      interface{}   `json:"levelEntity,omitempty"`
	FormFieldDtoList []interface{} `json:"formFieldDtoList,omitempty"`
	FieldEntityList  []interface{} `json:"fieldEntityList,omitempty"`
	Feedback         interface{}   `json:"feedback,omitempty"`
	HandleWay        interface{}   `json:"handleWay,omitempty"`
	IsWarning        int           `json:"isWarning,omitempty"`
	WarningType      interface{}   `json:"warningType,omitempty"`
	T                string        `json:"t,omitempty"`
}

func SubmitDataFunc(data SubmitData) map[string]interface{} {
	return map[string]interface{}{
		"address":          data.Address,
		"applyId":          data.ApplyId,
		"applyName":        data.ApplyName,
		"attachmentList":   data.AttachmentList,
		"commentNum":       data.CommentNum,
		"commentContent":   data.CommentContent,
		"content":          data.Content,
		"createBy":         data.CreateBy,
		"createTime":       data.CreateTime,
		"depName":          data.DepName,
		"reject":           data.Reject,
		"endTime":          data.EndTime,
		"headImg":          data.HeadImg,
		"yearmonth":        data.Yearmonth,
		"imageList":        data.ImageList,
		"isFine":           data.IsFine,
		"latitude":         data.Latitude,
		"gpmsSchoolYear":   data.GpmsSchoolYear,
		"longitude":        data.Longitude,
		"planId":           data.PlanId,
		"planName":         data.PlanName,
		"reportId":         data.ReportId,
		"reportType":       data.ReportType,
		"reportTime":       data.ReportTime,
		"isOnTime":         data.IsOnTime,
		"schoolId":         data.SchoolId,
		"startTime":        data.StartTime,
		"state":            data.State,
		"studentId":        data.StudentId,
		"studentNumber":    data.StudentNumber,
		"supportNum":       data.SupportNum,
		"title":            data.Title,
		"url":              data.Url,
		"username":         data.Username,
		"weeks":            data.Weeks,
		"videoUrl":         data.VideoUrl,
		"videoTitle":       data.VideoTitle,
		"attachments":      data.Attachments,
		"companyName":      data.CompanyName,
		"jobName":          data.JobName,
		"jobId":            data.JobId,
		"score":            data.Score,
		"tpJobId":          data.TpJobId,
		"starNum":          data.StarNum,
		"confirmDays":      data.ConfirmDays,
		"isApply":          data.IsApply,
		"compStarNum":      data.CompStarNum,
		"compScore":        data.CompScore,
		"compComment":      data.CompComment,
		"compState":        data.CompState,
		"apply":            data.Apply,
		"levelEntity":      data.LevelEntity,
		"formFieldDtoList": data.FormFieldDtoList,
		"fieldEntityList":  data.FieldEntityList,
		"feedback":         data.Feedback,
		"handleWay":        data.HandleWay,
		"isWarning":        data.IsWarning,
		"warningType":      data.WarningType,
		"t":                data.T,
	}
}
