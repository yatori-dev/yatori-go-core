package action

import (
	"errors"
	"github.com/thedevsaddam/gojsonq"
	ort "github.com/yalue/onnxruntime_go"
	"github.com/yatori-dev/yatori-go-core/api/xuexitong"
	"github.com/yatori-dev/yatori-go-core/api/yinghua"
	"github.com/yatori-dev/yatori-go-core/global"
	"github.com/yatori-dev/yatori-go-core/utils"
	log "github.com/yatori-dev/yatori-go-core/utils/log"
	"strings"
)

// MainFunc TODO 后续接口 待完成 先用interface{}占位
type MainFunc interface {
	XueXiT() (xuexitong.XueXiTUserCache, XueXiTInterface)
	YingHua() (yinghua.YingHuaUserCache, YingHuaInterface)
}

type CourseInterface interface {
	CourseList() []interface{}
}

type YatoriCache struct {
	yinghua.YingHuaUserCache
	xuexitong.XueXiTUserCache
	currentCacheType string // 标识当前使用的Cache类型
}

func ActionCache(input int) YatoriCache {
	user := global.Config.Users[input]
	switch user.AccountType {
	case "YINGHUA":
		cache := yinghua.NewCache(input)
		return YatoriCache{
			YingHuaUserCache: cache,
			currentCacheType: user.AccountType,
		}
	case "XUEXITONG":
		cache := xuexitong.NewCache(input)
		return YatoriCache{
			XueXiTUserCache:  cache,
			currentCacheType: user.AccountType,
		}
	default:
		return YatoriCache{currentCacheType: "error"}
	}
}

func (y YatoriCache) YingHua() (yinghua.YingHuaUserCache, YingHuaInterface) {
	cache := y.YingHuaUserCache
	for {
		path, cookie := cache.VerificationCodeApi(5) //获取验证码
		cache.SetCookie(cookie)
		img, _ := utils.ReadImg(path)                                  //读取验证码图片
		codeResult := utils.AutoVerification(img, ort.NewShape(1, 18)) //自动识别
		utils.DeleteFile(path)                                         //删除验证码文件
		cache.SetVerCode(codeResult)                                   //填写验证码
		jsonStr, _ := cache.LoginApi(5, nil)                           //执行登录
		log.Print(log.DEBUG, "["+cache.Account+"] "+"LoginAction---"+jsonStr)
		if gojsonq.New().JSONString(jsonStr).Find("msg") == "验证码有误！" {
			continue
		} else if gojsonq.New().JSONString(jsonStr).Find("redirect") == nil {
			return cache, errors.New(gojsonq.New().JSONString(jsonStr).Find("msg").(string))
		}
		cache.SetToken(
			strings.Split(
				strings.Split(
					gojsonq.New().JSONString(jsonStr).
						Find("redirect").(string), "token=")[1], "&")[0]) //设置Token
		cache.SetSign(
			strings.Split(
				gojsonq.New().JSONString(jsonStr).Find("redirect").(string), "&sign=")[1]) //设置签名
		log.Print(log.INFO, "["+cache.Account+"] "+" 登录成功")
		break
	}
	return y.YingHuaUserCache, y
}

func (y YatoriCache) XueXiT() (xuexitong.XueXiTUserCache, XueXiTInterface) {
	_, err := y.XueXiTUserCache.LoginApi()
	if err == nil {
		log.Print(log.INFO, "["+y.XueXiTUserCache.Name+"] "+" 登录成功")
	}
	return y.XueXiTUserCache, y
}
