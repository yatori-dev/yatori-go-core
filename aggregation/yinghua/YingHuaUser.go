package yinghua

import (
	"encoding/json"
	"errors"
	"github.com/thedevsaddam/gojsonq"
	ort "github.com/yalue/onnxruntime_go"
	"github.com/yatori-dev/yatori-go-core/api/yinghua"
	"github.com/yatori-dev/yatori-go-core/interfaces"
	"github.com/yatori-dev/yatori-go-core/utils"
	"github.com/yatori-dev/yatori-go-core/utils/log"
	"net/http"
	"strings"
)

type YingHuaUser struct {
	Account  string
	Password string
	PreUrl   string
	CacheMap map[string]any
}

func (user *YingHuaUser) Login() (map[string]any, error) {
	for {
		user.CacheMap = make(map[string]any)
		path, cookies := yinghua.VerificationCodeApi(user.PreUrl, nil, 5) //获取验证码
		if path == "" {                                                   //如果path为空，那么可能是账号问题
			return nil, errors.New("无法正常获取对应网站验证码，请检查对应url是否正常")
		}
		user.CacheMap["cookies"] = cookies                                                                                                                              //寄存cookie
		img, _ := utils.ReadImg(path)                                                                                                                                   //读取验证码图片
		codeResult := utils.AutoVerification(img, ort.NewShape(1, 18))                                                                                                  //自动识别
		utils.DeleteFile(path)                                                                                                                                          //删除验证码文件
		user.CacheMap["verCode"] = codeResult                                                                                                                           //填写验证码
		jsonStr, _ := yinghua.LoginApi(user.PreUrl, user.Account, user.Password, user.CacheMap["verCode"].(string), user.CacheMap["cookies"].([]*http.Cookie), 10, nil) //执行登录
		log.Print(log.DEBUG, "["+user.Account+"] "+"LoginAction---"+jsonStr)
		if gojsonq.New().JSONString(jsonStr).Find("msg") == "验证码有误！" {
			continue
		} else if gojsonq.New().JSONString(jsonStr).Find("redirect") == nil {
			if gojsonq.New().JSONString(jsonStr).Find("msg") == nil { // 如果登录不成功并且msg也返回空那么直接重新再登录一遍
				continue
			}
			return nil, errors.New(gojsonq.New().JSONString(jsonStr).Find("msg").(string))
		}
		user.CacheMap["token"] = strings.Split(
			strings.Split(
				gojsonq.New().JSONString(jsonStr).
					Find("redirect").(string), "token=")[1], "&")[0]
		user.CacheMap["sign"] = strings.Split(
			gojsonq.New().JSONString(jsonStr).Find("redirect").(string), "&sign=")[1]
		log.Print(log.DEBUG, "["+user.Account+"] "+" 登录成功")
		var resultMap map[string]any
		json.Unmarshal([]byte(jsonStr), &resultMap)
		return resultMap, nil
	}
}

func (user *YingHuaUser) UserInfo() (map[string]any, error) {
	//TODO implement me
	panic("implement me")
}

func (user *YingHuaUser) CacheData() (map[string]any, error) {
	//TODO implement me
	panic("implement me")
}

func (user *YingHuaUser) CourseList() ([]interfaces.ICourse, error) {
	//TODO implement me
	panic("implement me")
}
