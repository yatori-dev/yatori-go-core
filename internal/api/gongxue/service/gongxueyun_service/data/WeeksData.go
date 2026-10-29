package data

type WeeksData struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data []struct {
		IsDefault int    `json:"isDefault"`
		Weeks     string `json:"weeks"`
		StartTime string `json:"startTime"`
		EndTime   string `json:"endTime"`
	} `json:"data"`
	Flag int `json:"flag"`
}
