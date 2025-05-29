package xuexitong

import (
	"errors"
	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/api/xuexitong"
	"image"
)

// PassFaceAction 过人脸
func PassFaceAction(cache *xuexitong.XueXiTUserCache, courseId, classId, cpi string, face image.Image) error {
	uuid, qrEnc, err := cache.GetFaceQrCodeApi2(courseId, classId, cpi)
	if err != nil {
		return err
	}
	if uuid == "" || qrEnc == "" {
		return errors.New("uui或qrEnc为空")
	}
	//获取token
	tokenJson, err := cache.GetFaceUpLoadToken()

	token := gojsonq.New().JSONString(tokenJson).Find("_token").(string)
	if err != nil {
		return err
	}
	//上传人脸
	ObjectId, err := cache.UploadFaceImageApi(token, face)
	if err != nil {
		return err
	}
	if ObjectId == "" {
		return errors.New("ObjectId is empty")
	}
	plan3Api, err := cache.GetCourseFaceQrPlan3Api(uuid, classId, courseId, qrEnc, ObjectId)
	passMsg := gojsonq.New().JSONString(plan3Api).Find("msg")
	if err != nil {
		return err
	}
	if passMsg != nil {
		if passMsg != "通过" {
			return errors.New(plan3Api)
		}
	}
	//获取人脸状态
	stateApi, err := cache.GetCourseFaceQrStateApi(uuid, qrEnc, classId, courseId, cpi)
	if err != nil {
		return err
	}
	stateCode := gojsonq.New().JSONString(plan3Api).Find("code")
	if stateCode != nil {

		if int(stateCode.(float64)) != 0 {
			return errors.New(stateApi)
		}
	}
	return nil
}
