package data

type CheckData struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		CaptchaId           interface{} `json:"captchaId"`
		ProjectCode         interface{} `json:"projectCode"`
		CaptchaType         string      `json:"captchaType"`
		CaptchaOriginalPath interface{} `json:"captchaOriginalPath"`
		CaptchaFontType     interface{} `json:"captchaFontType"`
		CaptchaFontSize     interface{} `json:"captchaFontSize"`
		SecretKey           interface{} `json:"secretKey"`
		OriginalImageBase64 interface{} `json:"originalImageBase64"`
		Point               interface{} `json:"point"`
		JigsawImageBase64   interface{} `json:"jigsawImageBase64"`
		WordList            interface{} `json:"wordList"`
		PointList           interface{} `json:"pointList"`
		PointJson           string      `json:"pointJson"`
		Token               string      `json:"token"`
		Result              bool        `json:"result"`
		CaptchaVerification interface{} `json:"captchaVerification"`
		ClientUid           interface{} `json:"clientUid"`
		Ts                  interface{} `json:"ts"`
		BrowserInfo         interface{} `json:"browserInfo"`
	} `json:"data"`
}
