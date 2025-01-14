package api

const (
	// API_LOGIN_WEB 接口-web端登录
	API_LOGIN_WEB = "https://passport2.chaoxing.com/fanyalogin"

	// API_QRCREATE 接口-激活二维码key并返回二维码图片
	API_QRCREATE = "https://passport2.chaoxing.com/createqr"

	// API_QRLOGIN 接口-web端二维码登录
	API_QRLOGIN = "https://passport2.chaoxing.com/getauthstatus"

	// API_CLASS_LST 接口-课程列表
	API_CLASS_LST = "https://mooc1-api.chaoxing.com/mycourse/backclazzdata"

	// API_SSO_LOGIN 接口-SSO二步登录 (用作获取登录信息)
	API_SSO_LOGIN = "https://sso.chaoxing.com/apis/login/userLogin4Uname.do"

	// PAGE_LOGIN SSR页面-登录 用于提取二维码key
	PAGE_LOGIN = "https://passport2.chaoxing.com/login"

	// API_FACE_IMAGE 接口-获取预上传人脸图片
	API_FACE_IMAGE = "https://passport2-api.chaoxing.com/api/getUserFaceid"

	// URL_QRLOGIN 二维码登录 url
	URL_QRLOGIN = "https://passport2.chaoxing.com/toauthlogin"

	// PAGE_MOBILE_CHAPTER_CARD SSR页面-客户端章节任务卡片
	PAGE_MOBILE_CHAPTER_CARD = "https://mooc1-api.chaoxing.com/knowledge/cards"
)
