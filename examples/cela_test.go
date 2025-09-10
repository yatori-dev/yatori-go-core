package examples

import (
	"fmt"
	"testing"

	"github.com/thedevsaddam/gojsonq"
	ort "github.com/yalue/onnxruntime_go"
	"github.com/yatori-dev/yatori-go-core/api/cela"
	"github.com/yatori-dev/yatori-go-core/global"
	"github.com/yatori-dev/yatori-go-core/utils"
)

func Test_MD5(t *testing.T) {
	//sum := md5.Sum([]byte("Nlj123456"))
	//fmt.Printf("%x", sum)
	utils.YatoriCoreInit()
	setup()
	user := global.Config.Users[30]
	cache := cela.CelaUserCache{
		Account:  user.Account,
		Password: user.Password,
	}
	//cache.Code = "YgAH"
	//初始化登录
	cache.InitLoginDataApi()
	//获取验证码
	captchaPath, err := cache.GetCaptchaApi()
	if err != nil {
		t.Error(err)
	}
	img, _ := utils.ReadImg(captchaPath)                           //读取验证码图片
	codeResult := utils.AutoVerification(img, ort.NewShape(1, 25)) //自动识别
	cache.Code = codeResult                                        //设置验证码
	checkResult, err := cache.CheckCaptchaApi()                    //检测验证码是否过了
	if err != nil {
		t.Error(err)
	}
	fmt.Println(checkResult)
	data := cache.LoginApi()
	find := gojsonq.New().JSONString(data).Find("data")
	if find != nil {
		cache.GetLoginAfterData(find.(string))
	}
}
