package xuexitong

import (
	"errors"
	"fmt"
	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/api/xuexitong"
	log2 "github.com/yatori-dev/yatori-go-core/utils/log"
	"image"
	"net/http"
)

// PassFaceAction 过人脸
func PassFaceAction3(cache *xuexitong.XueXiTUserCache, courseId, classId, cpi, chapterId, enc, videojobid, chaptervideoobjectid, mid, videoRandomCollectTime string, face image.Image) (string, string, string, string, error) {
	//uuid, qrEnc, err := cache.GetFaceQrCodeApi2(courseId, classId, cpi)
	uuid, qrEnc, err := cache.GetFaceQrCodeApi3(courseId, classId, chapterId, cpi, enc, videojobid, chaptervideoobjectid)
	if err != nil {
		return "", "", "", "", err
	}
	if uuid == "" || qrEnc == "" {
		return "", "", "", "", errors.New("uui或qrEnc为空")
	}
	//获取token
	tokenJson, err := cache.GetFaceUpLoadToken()

	token := gojsonq.New().JSONString(tokenJson).Find("_token").(string)
	if err != nil {
		return "", "", "", "", err
	}

	//上传人脸
	ObjectId, err := cache.UploadFaceImageApi(token, face)
	if err != nil {
		return "", "", "", "", err
	}
	if ObjectId == "" {
		return "", "", "", "", errors.New("ObjectId is empty")
	}
	plan3Api, err := cache.GetCourseFaceQrPlan3Api(uuid, classId, courseId, qrEnc, ObjectId)
	//plan3Api, err := cache.GetCourseFaceQrPlan1Api(courseId, classId, uuid, ObjectId, qrEnc, "0")
	passMsg := gojsonq.New().JSONString(plan3Api).Find("msg")
	if err != nil {
		return "", "", "", "", err
	}
	if passMsg != nil {
		if passMsg != "通过" {
			return "", "", "", "", errors.New(plan3Api)
		}
	}
	//获取人脸状态
	stateApi, err := cache.GetCourseFaceQrStateApi(uuid, qrEnc, classId, courseId, cpi, mid, videojobid, videoRandomCollectTime, chapterId)
	if err != nil {
		return "", "", "", "", err
	}
	stateCode := gojsonq.New().JSONString(plan3Api).Find("code")
	if stateCode != nil {

		if int(stateCode.(float64)) != 0 {
			return "", "", "", "", errors.New(stateApi)
		}
	}
	successEnc := gojsonq.New().JSONString(stateApi).Find("videoFaceCaptureSuccessEnc").(string)
	return uuid, qrEnc, ObjectId, successEnc, nil
}

// PassVerAnd202 绕过验证码和状态202情况
func PassVerAnd202(cache *xuexitong.XueXiTUserCache) {
	//重新登录逻辑
	log2.Print(log2.DEBUG, "触发验证码或者202，正在绕过...")
	cache.SetCookies([]*http.Cookie{})
	cache.LoginApi()      //重新登录设置cookie
	cache.CourseListApi() //重新设置k8s值
	log2.Print(log2.DEBUG, "重新登录后cookie值>>", fmt.Sprintf("%+v", cache.GetCookies()))
}
