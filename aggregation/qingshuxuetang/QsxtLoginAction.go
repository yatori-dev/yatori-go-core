package qingshuxuetang

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	ddddocr "github.com/Changbaiqi/ddddocr-go/utils"
	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/api/qingshuxuetang"
	"github.com/yatori-dev/yatori-go-core/utils"
)

// 登录
func QsxtLoginAction(cache *qingshuxuetang.QsxtUserCache) (string, error) {
	for {
		pullCodeJson, err := cache.QsxtPhoneValidationCodeApi(3, nil)
		if err != nil {
			return "", err
		}
		hr := gojsonq.New().JSONString(pullCodeJson).Find("hr")
		if hr == nil {
			return "", errors.New(pullCodeJson)
		}
		if int(hr.(float64)) != 0 {
			return "", errors.New(pullCodeJson)
		}
		if sessionId, ok := gojsonq.New().JSONString(pullCodeJson).Find("data.sessionId").(string); ok {
			cache.VerCodeSession = sessionId
		}
		if codeImgBase64, ok := gojsonq.New().JSONString(pullCodeJson).Find("data.code").(string); ok {
			image, err := utils.Base64ToImage(codeImgBase64)
			if err != nil {
				return "", err
			}
			//utils.SaveImageAsJPEG(image, "./qsxt_code.png")
			//verification := ddddocr.AutoOCRVerification(image)
			dets, err := ddddocr.AutoDetectionForCalc(image, 7)
			if err != nil {
				return "", err
			}
			sort.Slice(dets, func(i, j int) bool {
				if dets[i].BBox.Min.X < dets[j].BBox.Min.X {
					return true
				}
				return false
			})
			calc, err := ddddocr.AutoCalc(dets)
			if err != nil {
				continue
			}
			cache.VerCode = fmt.Sprintf("%d", calc)
			//fmt.Println(calc)
		}
		login_json, err := cache.QsxtPhoneLoginApi(3, nil)
		//fmt.Println(login_json)
		if err != nil {
			return "", err
		}
		if message, ok := gojsonq.New().JSONString(login_json).Find("message").(string); ok {
			if strings.Contains(message, "图片验证码答案错误") {
				continue
			}
		}
		//如果登录成功，则直接返回
		if token, ok := gojsonq.New().JSONString(login_json).Find("data.token").(string); ok {
			cache.Token = token
			return login_json, nil
		}
		return "", errors.New(login_json)
	}
}

func QsxtCookieLoginAction(cache *qingshuxuetang.QsxtUserCache) {
	cache.Token = cache.Password
}
