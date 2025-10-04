package weiban

import (
	"fmt"

	ddddocr "github.com/Changbaiqi/ddddocr-go/utils"
	"github.com/thedevsaddam/gojsonq"
	ort "github.com/yalue/onnxruntime_go"
	"github.com/yatori-dev/yatori-go-core/api/weiban"
	"github.com/yatori-dev/yatori-go-core/utils"
)

// 登录
func WeiBanLoginAction(cache *weiban.WeiBanCache) (string, error) {
	tenantListJson, err := cache.PullTenantCodeApi()
	if err != nil {
		return "", err
	}
	//拉取学校code码
	tenantList := gojsonq.New().JSONString(tenantListJson).Find("data")
	if datas, ok := tenantList.([]interface{}); ok {
		for _, data := range datas {
			if index, ok1 := data.(map[string]interface{}); ok1 {
				list := index["list"].([]interface{})
				for _, item := range list {
					if obj, ok2 := item.(map[string]interface{}); ok2 {
						if obj["name"] == cache.School {
							cache.TenantCode = obj["code"].(string)
							break
						}
					}
				}
			}
		}
	}
	fmt.Println(tenantList)
	// 拉取验证码
	codePath, err := cache.PullCapterApi(3, nil)
	if err != nil {
		return "", err
	}
	img, _ := utils.ReadImg(codePath) //读取验证码图片
	//verification := utils.AutoVerification(img, ort.NewShape(1, 18))
	verification := ddddocr.SemiOCRVerification(img, ort.NewShape(1, 18))
	cache.VerifyCode = verification

	api, err := cache.LoginApi()
	if err != nil {
		return "", err
	}
	fmt.Println(api)

	return "", nil
}
