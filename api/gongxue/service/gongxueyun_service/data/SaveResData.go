package data

type SaveData struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		CreateTime   string `json:"createTime"`
		AttendanceId string `json:"attendanceId"`
	} `json:"data"`
}
