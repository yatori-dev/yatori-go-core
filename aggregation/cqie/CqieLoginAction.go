package cqie

import (
	"errors"
	cqieApi "github.com/Yatori-Dev/yatori-go-core/api/cqie"
	"github.com/Yatori-Dev/yatori-go-core/utils"
	"github.com/Yatori-Dev/yatori-go-core/utils/log"
	"github.com/thedevsaddam/gojsonq"
	ort "github.com/yalue/onnxruntime_go"
)

// CqieLoginAction 登录API聚合整理
// {"refresh_code":1,"status":false,"msg":"账号密码不正确"}
// {"_code": 1, "status": false,"msg": "账号登录超时，请重新登录", "result": {}}
func CqieLoginAction(cache *cqieApi.CqieUserCache) error {
	for {
		path, cookie := cache.VerificationCodeApi() //获取验证码
		cache.SetCookie(cookie)
		img, _ := utils.ReadImg(path)                                  //读取验证码图片
		codeResult := utils.AutoVerification(img, ort.NewShape(1, 26)) //自动识别
		utils.DeleteFile(path)                                         //删除验证码文件
		cache.SetVerCode(codeResult)                                   //填写验证码
		jsonStr, _ := cache.LoginApi()                                 //执行登录
		log.Print(log.DEBUG, "["+cache.Account+"] "+"LoginAction---"+jsonStr)
		if gojsonq.New().JSONString(jsonStr).Find("msg") == "验证码有误！" {
			continue
		} else if int(gojsonq.New().JSONString(jsonStr).Find("code").(float64)) != 200 {
			return errors.New(gojsonq.New().JSONString(jsonStr).Find("msg").(string))
		}
		cache.SetAccess_Token(gojsonq.New().JSONString(jsonStr).Find("data.access_token").(string))
		cache.SetToken(gojsonq.New().JSONString(jsonStr).Find("data.user.token").(string))
		cache.SetUserId(gojsonq.New().JSONString(jsonStr).Find("data.user.userId").(string))
		cache.SetAppId(gojsonq.New().JSONString(jsonStr).Find("data.user.appId").(string))
		cache.SetIpaddr(gojsonq.New().JSONString(jsonStr).Find("data.user.ipaddr").(string))
		cache.SetDeptId(gojsonq.New().JSONString(jsonStr).Find("data.user.deptId").(string))
		log.Print(log.INFO, "["+cache.Account+"] "+" 登录成功")
		break
	}
	return nil
}
