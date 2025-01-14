package data

type BlockRes struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		CaptchaId           interface{} `json:"captchaId"`
		ProjectCode         interface{} `json:"projectCode"`
		CaptchaType         interface{} `json:"captchaType"`
		CaptchaOriginalPath interface{} `json:"captchaOriginalPath"`
		CaptchaFontType     interface{} `json:"captchaFontType"`
		CaptchaFontSize     interface{} `json:"captchaFontSize"`
		SecretKey           string      `json:"secretKey"`
		OriginalImageBase64 string      `json:"originalImageBase64"`
		Point               interface{} `json:"point"`
		JigsawImageBase64   string      `json:"jigsawImageBase64"`
		WordList            interface{} `json:"wordList"`
		PointList           interface{} `json:"pointList"`
		PointJson           interface{} `json:"pointJson"`
		Token               string      `json:"token"`
		Result              bool        `json:"result"`
		CaptchaVerification interface{} `json:"captchaVerification"`
		ClientUid           interface{} `json:"clientUid"`
		Ts                  interface{} `json:"ts"`
		BrowserInfo         interface{} `json:"browserInfo"`
	} `json:"data"`
}
