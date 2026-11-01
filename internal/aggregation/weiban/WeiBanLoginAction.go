package weiban

import (
	"fmt"
	"strings"

	ddddocr "github.com/Changbaiqi/ddddocr-go/utils"
	"github.com/thedevsaddam/gojsonq"
	ort "github.com/yalue/onnxruntime_go"
	"github.com/yatori-dev/yatori-go-core/v3/internal/api/weiban"
	"github.com/yatori-dev/yatori-go-core/v3/utils"
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
	if strings.TrimSpace(cache.TenantCode) == "" {
		return "", fmt.Errorf("未找到学校对应的 tenantCode：%s", cache.School)
	}
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
	if err := validateWeiBanLogin(api); err != nil {
		return api, err
	}

	return api, nil
}

func validateWeiBanLogin(body string) error {
	if strings.TrimSpace(body) == "" {
		return fmt.Errorf("微伴登录返回为空")
	}
	if success, ok := gojsonq.New().JSONString(body).Find("success").(bool); ok && !success {
		return fmt.Errorf("微伴登录失败：%s", body)
	}
	if code, ok := gojsonq.New().JSONString(body).Find("code").(float64); ok && int(code) != 200 && int(code) != 0 {
		return fmt.Errorf("微伴登录失败：%s", body)
	}
	return nil
}
