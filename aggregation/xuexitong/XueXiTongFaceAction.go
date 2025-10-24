package xuexitong

import (
	"errors"
	"fmt"
	"image"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/api/xuexitong"
	"github.com/yatori-dev/yatori-go-core/utils"
	log2 "github.com/yatori-dev/yatori-go-core/utils/log"
)

// 人脸识别1
func PassFaceAction1(cache *xuexitong.XueXiTUserCache, courseId, classId, cpi, chapterId, enc, videojobid, chaptervideoobjectid, mid, videoRandomCollectTime string, face image.Image) (string, string, string, string /*识别状态*/, error) {
	//uuid, qrEnc, err := cache.GetFaceQrCodeApi2(courseId, classId, cpi)
	uuid, qrEnc, err := cache.GetFaceQrCodeApi3(courseId, classId, chapterId, cpi, enc, videojobid, chaptervideoobjectid)
	if err != nil {
		return "", "", "", "", err
	}
	if uuid == "" || qrEnc == "" {
		return "", "", "", "", errors.New("uuid或qrEnc为空")
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
	//暂时先用plan3的过人脸（进入课程扫通过的）
	//plan2Api, err := cache.GetCourseFaceQrPlan3Api(classId, courseId, uuid, qrEnc, cpi, ObjectId)
	plan2Api, err := cache.PassFaceQrPlanPhoneApi(classId, courseId, chapterId, cpi, ObjectId)
	//plan3Api, err := cache.GetCourseFaceQrPlan1Api(courseId, classId, uuid, ObjectId, qrEnc, "0")
	log2.Print(log2.DEBUG, plan2Api)
	passMsg := gojsonq.New().JSONString(plan2Api).Find("msg")
	if err != nil {
		return "", "", "", "", err
	}
	if passMsg != nil {
		if passMsg != "识别通过" {
			return "", "", "", "", errors.New(plan2Api)
		}
	}
	//获取人脸状态
	stateApi, err := cache.GetCourseFaceQrStateApi(uuid, qrEnc, classId, courseId, cpi, mid, videojobid, videoRandomCollectTime, chapterId)
	if err != nil {
		return "", "", "", "", err
	}
	stateCode := gojsonq.New().JSONString(plan2Api).Find("code")
	if stateCode != nil {

		if int(stateCode.(float64)) != 0 {
			return "", "", "", "", errors.New(stateApi)
		}
	}
	//successEnc := gojsonq.New().JSONString(stateApi).Find("videoFaceCaptureSuccessEnc")
	status := strconv.Itoa(int(gojsonq.New().JSONString(stateApi).Find("status").(float64)))
	return uuid, qrEnc, ObjectId, status, nil
}

// 手机端过人脸接口
func PassFacePhoneAction(cache *xuexitong.XueXiTUserCache, courseId, classId, cpi, chapterId, videojobid, mid, videoRandomCollectTime string) (string /*识别状态*/, error) {
	//拉取用户照片
	localFaceExists, _ := utils.PathExists("./assets/faces/" + cache.Name + ".jpg")
	var faceImg image.Image
	if localFaceExists {
		img, err2 := utils.LoadImage("./assets/faces/" + cache.Name + ".jpg")
		if err2 != nil {
			log2.Print(log2.INFO, err2)
			//os.Exit(0)
			return "", err2
		}
		faceImg = img
	} else {
		pullJson, img, err2 := cache.GetHistoryFaceImg("")
		if err2 != nil {
			log2.Print(log2.DEBUG, pullJson, err2)
			//os.Exit(0)
			return "", err2
		}
		faceImg = img
	}
	disturbImage := utils.ImageRGBDisturb(faceImg)
	//disturbImage := utils.ProcessImageDisturb(faceImg)
	//disturbImage := utils.ImageRGBDisturbAdjust(faceImg, 10)
	//utils.SaveImageAsJPEG(disturbImage, "./assets/18106919661.jpg")
	//cache.GetCourseFaceStart(classId, courseId, chapterId, cpi) //没事别放开，测试用的

	//获取token
	tokenJson, err := cache.GetFaceUpLoadToken()

	token := gojsonq.New().JSONString(tokenJson).Find("_token").(string)
	if err != nil {
		return "", err
	}
	time.Sleep(1 * time.Second) //隔一下
	//上传人脸
	ObjectId, err := cache.UploadFaceImageApi(token, disturbImage)
	if err != nil {
		return "", err
	}
	if ObjectId == "" {
		return "", errors.New("ObjectId is empty")
	}
	time.Sleep(2 * time.Second)
	plan3Api, err := cache.PassFaceQrPlanPhoneApi(classId, courseId, chapterId, cpi, ObjectId)
	//暂时先用plan3的过人脸（进入课程扫通过的）

	passMsg := gojsonq.New().JSONString(plan3Api).Find("msg")
	if err != nil {
		return "", err
	}

	if strings.Contains(plan3Api, "活体检测不通过") {
		return "", errors.New(plan3Api)
	}
	if strings.Contains(plan3Api, "用户图片信息出错") {
		return "", errors.New(plan3Api)
	}
	if passMsg != nil {
		if passMsg != "通过" && passMsg != "识别通过" {
			return "", errors.New(plan3Api)
		}
	}
	//获取人脸状态

	return ObjectId, nil
}

// PassFacePCAction PC过人脸
func PassFacePCAction(cache *xuexitong.XueXiTUserCache, courseId, classId, cpi, chapterId, enc, videojobid, chaptervideoobjectid, mid, videoRandomCollectTime string, face image.Image) (string, string, string, string, error) {
	//uuid, qrEnc, err := cache.GetFaceQrCodeApi2(courseId, classId, cpi)
	uuid, qrEnc, err := cache.GetFaceQrCodeApi3(courseId, classId, chapterId, cpi, enc, videojobid, chaptervideoobjectid)

	if err != nil {
		return "", "", "", "", err
	}
	if uuid == "" || qrEnc == "" {
		return "", "", "", "", errors.New("uuid或qrEnc为空")
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
	//plan3Api, err := cache.PassFaceQrPlanPhoneApi(classId, courseId, chapterId, cpi, ObjectId)
	//暂时先用plan3的过人脸（进入课程扫通过的）
	plan3Api, err := cache.GetCourseFaceQrPlan3Api(classId, courseId, uuid, qrEnc, cpi, ObjectId)

	//plan3Api, err := cache.GetCourseFaceQrPlan3Api(uuid, classId, courseId, qrEnc, ObjectId)
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

// ReLogin 重登
func ReLogin(cache *xuexitong.XueXiTUserCache) {
	//重新登录逻辑
	log2.Print(log2.DEBUG, "触发验证码或者202，正在绕过...")
	cache.SetCookies([]*http.Cookie{})
	loginResult, _ := cache.LoginApi(5) //重新登录设置cookie
	//如果登录成功
	if gojsonq.New().JSONString(loginResult).Find("status") == nil {
		//如果失败
		if gojsonq.New().JSONString(loginResult).Find("msg2") != nil {
			log2.Print(log2.INFO, gojsonq.New().JSONString(loginResult).Find("msg2").(string))
		} else {
			log2.Print(log2.INFO, loginResult)
		}
	}
	cache.CourseListApi(5, nil) //重新设置k8s值
	log2.Print(log2.DEBUG, "重新登录后cookie值>>", fmt.Sprintf("%+v", cache.GetCookies()))
}
