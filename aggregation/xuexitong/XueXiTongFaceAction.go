package xuexitong

import (
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/api/xuexitong"
	"github.com/yatori-dev/yatori-go-core/utils"
	log2 "github.com/yatori-dev/yatori-go-core/utils/log"

	_ "golang.org/x/image/webp"
)

// loadFaceImage 拉取人脸图片: 本地库优先(多张轮换) → 单张兼容 → 服务器历史人脸
//
// 本地库: ./assets/faces/{账号}/ 目录下的图片文件(jpg/jpeg/png/webp)，
// 存在多张时随机选一张轮换(降低同一张被风控的概率)；
// 兼容旧单张 ./assets/faces/{账号}.jpg；
// 都没有时拉取服务器历史人脸。
func loadFaceImage(cache *xuexitong.XueXiTUserCache) (image.Image, error) {
	// 1) 多张目录(前端上传)
	dirPath := filepath.Join(".", "assets", "faces", cache.Name)
	if entries, err := os.ReadDir(dirPath); err == nil {
		var candidates []string
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			switch strings.ToLower(filepath.Ext(entry.Name())) {
			case ".jpg", ".jpeg", ".png", ".webp":
				candidates = append(candidates, filepath.Join(dirPath, entry.Name()))
			}
		}
		if len(candidates) > 0 {
			if img, err := loadFaceImageFile(candidates[rand.Intn(len(candidates))]); err == nil {
				return img, nil
			}
		}
	}
	// 2) 兼容旧单张
	singlePath := filepath.Join(".", "assets", "faces", cache.Name+".jpg")
	if exists, _ := utils.PathExists(singlePath); exists {
		return utils.LoadImage(singlePath)
	}
	// 3) 服务器历史人脸
	_, img, err := cache.GetHistoryFaceImg("")
	if err != nil {
		return nil, err
	}
	if img == nil {
		return nil, errors.New("没有历史人脸")
	}
	return img, nil
}

func loadFaceImageFile(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("图片解码失败: %w", err)
	}
	return img, nil
}

// fetchFaceUploadToken 拉取人脸上传token
func fetchFaceUploadToken(cache *xuexitong.XueXiTUserCache) (string, error) {
	tokenJson, err := cache.GetFaceUpLoadToken()
	if err != nil {
		return "", err
	}
	token, _ := gojsonq.New().JSONString(tokenJson).Find("_token").(string)
	if token == "" {
		return "", fmt.Errorf("获取上传token失败: %s", responseSummary(tokenJson))
	}
	return token, nil
}

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
	token, err := fetchFaceUploadToken(cache)
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
	//更换基准FaceId为新上传的照片(新版APP过人脸的核心: 先换基准再比对，绕过历史基准照片限制)
	if _, err := cache.ChangeFaceIdApi(ObjectId, 3, nil); err != nil {
		return "", "", "", "", err
	}
	//暂时先用plan3的过人脸（进入课程扫通过的）
	//plan2Api, err := cache.GetCourseFaceQrPlan3Api(classId, courseId, uuid, qrEnc, cpi, ObjectId)
	plan2Api, err := cache.PassFaceQrPlanPhoneNewApi(classId, courseId, chapterId, cpi, ObjectId)
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
	//拉取用户照片(本地库多张轮换优先 → 单张 → 服务器历史人脸)
	faceImg, err := loadFaceImage(cache)
	if err != nil {
		return "", err
	}
	disturbImage := utils.ImageRGBDisturb(faceImg)
	//disturbImage := utils.ProcessImageDisturb(faceImg)
	//disturbImage := utils.ImageRGBDisturbAdjust(faceImg, 10)
	//utils.SaveImageAsJPEG(disturbImage, "./assets/18106919661.jpg")
	//cache.GetCourseFaceStart(classId, courseId, chapterId, cpi) //没事别放开，测试用的

	//获取token
	token, err := fetchFaceUploadToken(cache)
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
	time.Sleep(1 * time.Second)
	//更换基准FaceId为新上传的照片(新版APP过人脸的核心: 先换基准再比对，绕过历史基准照片限制)
	if _, err := cache.ChangeFaceIdApi(ObjectId, 3, nil); err != nil {
		return "", err
	}
	time.Sleep(1 * time.Second)
	plan3Api, err := cache.PassFaceQrPlanPhoneNewApi(classId, courseId, chapterId, cpi, ObjectId)
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

// 手机端过人脸接口(老接口）
func PassFacePhoneOldAction(cache *xuexitong.XueXiTUserCache, courseId, classId, cpi, chapterId, videojobid, mid, videoRandomCollectTime string) (string /*识别状态*/, error) {
	//拉取用户照片(本地库多张轮换优先 → 单张 → 服务器历史人脸)
	faceImg, err := loadFaceImage(cache)
	if err != nil {
		return "", err
	}
	disturbImage := utils.ImageRGBDisturb(faceImg)
	//disturbImage := utils.ProcessImageDisturb(faceImg)
	//disturbImage := utils.ImageRGBDisturbAdjust(faceImg, 10)
	//utils.SaveImageAsJPEG(disturbImage, "./assets/18106919661.jpg")
	//cache.GetCourseFaceStart(classId, courseId, chapterId, cpi) //没事别放开，测试用的

	//获取token
	token, err := fetchFaceUploadToken(cache)
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
	time.Sleep(1 * time.Second)
	//更换基准FaceId为新上传的照片(新版APP过人脸的核心: 先换基准再比对，绕过历史基准照片限制)
	if _, err := cache.ChangeFaceIdApi(ObjectId, 3, nil); err != nil {
		return "", err
	}
	time.Sleep(1 * time.Second)
	//plan3Api, err := cache.GetCourseFaceQrPlan4Api( classId,courseId, chapterId,"", "", ObjectId)
	plan3Api, err := cache.PassFaceQrPlanPhoneOldApi(classId, courseId, chapterId, cpi, ObjectId)
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
	token, err := fetchFaceUploadToken(cache)
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
	//更换基准FaceId为新上传的照片(新版APP过人脸的核心: 先换基准再比对，绕过历史基准照片限制)
	if _, err := cache.ChangeFaceIdApi(ObjectId, 3, nil); err != nil {
		return "", "", "", "", err
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
