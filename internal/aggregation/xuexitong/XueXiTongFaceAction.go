package xuexitong

import (
	"encoding/json"
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
	"github.com/yatori-dev/yatori-go-core/internal/api/xuexitong"
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

// faceCheckResult 判断人脸校验接口响应是否通过，并返回失败原因。
//
// 判定顺序必须是先 status/result/success、后 msg：
// 老接口(uploadInfo)成功时既可能返回字面量 true，也可能返回
// {"status":true,"msg":"上传成功"} 这类响应；如果先看 msg，
// 会因为 msg 不是"通过"而把已经成功的校验误判为失败。
func faceCheckResult(body string) (bool, string) {
	trimmed := strings.TrimSpace(body)
	if trimmed == "" {
		return false, "接口返回空响应"
	}
	// 老接口(uploadInfo)成功时会直接返回字面量 true
	switch strings.ToLower(trimmed) {
	case "true", `"true"`, "1":
		return true, ""
	case "false", `"false"`, "0":
		return false, "接口返回 false"
	}
	if strings.Contains(body, "活体检测不通过") {
		return false, "活体检测不通过"
	}
	if strings.Contains(body, "用户图片信息出错") {
		return false, "用户图片信息出错"
	}

	// 兼容 JSONP 包裹的响应
	jsonBody := trimmed
	if start, end := strings.Index(jsonBody, "{"), strings.LastIndex(jsonBody, "}"); start >= 0 && end > start {
		jsonBody = jsonBody[start : end+1]
	}
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(jsonBody), &payload); err != nil {
		// 登录页/验证码页等非JSON响应不能当作通过，否则会造成
		// "接口通过但章节仍要求人脸"的假成功
		return false, "非JSON响应: " + responseSummary(body)
	}

	// status/result/success 是明确结果，优先于 msg
	for _, key := range []string{"status", "result", "success"} {
		value, exists := payload[key]
		if !exists || value == nil {
			continue
		}
		if s, ok := value.(string); ok && strings.TrimSpace(s) == "" {
			continue
		}
		if passed, known := faceBoolValue(value); known {
			if passed {
				return true, ""
			}
			return false, faceFailureMessage(payload, body)
		}
		// 字段存在但含义不明确(如状态机数字)，保持旧版行为：不算失败
		return true, ""
	}

	// 仅返回 code 的接口：code=0 视为成功
	if code, exists := payload["code"]; exists {
		switch v := code.(type) {
		case float64:
			if int(v) == 0 {
				return true, ""
			}
		case string:
			if v == "0" {
				return true, ""
			}
		}
		return false, faceFailureMessage(payload, body)
	}

	// 只有 msg 的接口：通过/成功类文案算通过，其余算失败
	if msg := faceFailureMessage(payload, body); msg != "" {
		if isPositiveFaceMessage(msg) {
			return true, ""
		}
		return false, msg
	}
	return true, ""
}

// faceBoolValue 把接口里常见的布尔表达(true/false、1/0、"true"/"false")统一解析
func faceBoolValue(value interface{}) (bool, bool) {
	switch v := value.(type) {
	case bool:
		return v, true
	case string:
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "true", "1":
			return true, true
		case "false", "0":
			return false, true
		}
	case float64:
		if v == 0 {
			return false, true
		}
		if v == 1 {
			return true, true
		}
	}
	return false, false
}

// faceFailureMessage 从响应里提取失败原因，提取不到时回退为原始响应摘要
func faceFailureMessage(payload map[string]interface{}, body string) string {
	for _, key := range []string{"msg", "message", "errorMsg", "error"} {
		if msg, ok := payload[key].(string); ok && strings.TrimSpace(msg) != "" {
			return strings.TrimSpace(msg)
		}
	}
	return responseSummary(body)
}

// isPositiveFaceMessage 判断仅返回 msg 的接口是否为成功文案
func isPositiveFaceMessage(msg string) bool {
	lower := strings.ToLower(msg)
	for _, word := range []string{"通过", "成功", "正确", "success"} {
		if strings.Contains(lower, word) {
			return true
		}
	}
	return false
}

// passFacePhoneCheck 依次尝试手机端人脸校验接口。
// 不同账号/课程可能只接受其中某一套接口(例如其中一套返回"非法请求")，
// 主接口被拒时自动切换备用接口，避免直接失败。
//
// 注意：这里不能在"活体检测不通过"/"用户图片信息出错"时提前 break。
// 前面的 clientfacecheckstatus 接口带活体检测参数，用静态照片调用会报活体失败；
// 真正能过的往往是最后的静默比对接口(uploadInfo 老接口)，
// 提前退出会导致明明老接口返回 true 却判定整次人脸校验失败。
func passFacePhoneCheck(cache *xuexitong.XueXiTUserCache, classId, courseId, chapterId, cpi, objectId string) (bool, string) {
	type candidate struct {
		name string
		call func() (string, error)
	}
	candidates := []candidate{
		{"clientfacecheckstatus(GET)", func() (string, error) {
			return cache.PassFaceQrPlanPhoneNewApi(classId, courseId, chapterId, cpi, objectId)
		}},
		{"clientfacecheckstatus(POST)", func() (string, error) {
			return cache.GetCourseFaceQrPlan5Api(classId, courseId, chapterId, cpi, objectId)
		}},
		{"clientfacecheckstatus(POST2)", func() (string, error) {
			return cache.PassFaceQrPlanPhoneNew2Api(classId, courseId, chapterId, cpi, objectId)
		}},
		{"uploadInfo(老接口)", func() (string, error) {
			return cache.PassFaceQrPlanPhoneOldApi(classId, courseId, chapterId, cpi, objectId)
		}},
	}
	var failures []string
	for index, candidate := range candidates {
		if index > 0 {
			// 接口切换之间保持拟人化间隔
			utils.HumanDelay(500*time.Millisecond, 1200*time.Millisecond)
		}
		body, err := candidate.call()
		if err != nil {
			failures = append(failures, candidate.name+": "+err.Error())
			continue
		}
		log2.Print(log2.DEBUG, "人脸校验接口 ", candidate.name, " 响应: ", responseSummary(body))
		passed, reason := faceCheckResult(body)
		if passed {
			log2.Print(log2.DEBUG, "人脸校验通过接口: ", candidate.name)
			return true, ""
		}
		failures = append(failures, candidate.name+": "+reason)
	}
	return false, strings.Join(failures, "; ")
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

	utils.HumanDelay(500*time.Millisecond, 1200*time.Millisecond)
	//上传人脸
	ObjectId, err := cache.UploadFaceImageApi(token, face)
	if err != nil {
		return "", "", "", "", err
	}
	if ObjectId == "" {
		return "", "", "", "", errors.New("ObjectId is empty")
	}
	utils.HumanDelay(500*time.Millisecond, 1200*time.Millisecond)
	//更换基准FaceId为新上传的照片(新版APP过人脸的核心: 先换基准再比对，绕过历史基准照片限制)
	if _, err := cache.ChangeFaceIdApi(ObjectId, 3, nil); err != nil {
		return "", "", "", "", err
	}
	utils.HumanDelay(600*time.Millisecond, 1400*time.Millisecond)
	//多接口依次校验: 主接口被拒(如"非法请求")时自动切换备用接口
	passed, detail := passFacePhoneCheck(cache, classId, courseId, chapterId, cpi, ObjectId)
	if !passed {
		return "", "", "", "", fmt.Errorf("人脸校验失败: %s", detail)
	}
	//获取人脸状态
	stateApi, err := cache.GetCourseFaceQrStateApi(uuid, qrEnc, classId, courseId, cpi, mid, videojobid, videoRandomCollectTime, chapterId)
	if err != nil {
		return "", "", "", "", err
	}
	//successEnc := gojsonq.New().JSONString(stateApi).Find("videoFaceCaptureSuccessEnc")
	status := ""
	if statusVal := gojsonq.New().JSONString(stateApi).Find("status"); statusVal != nil {
		if statusFloat, ok := statusVal.(float64); ok {
			status = strconv.Itoa(int(statusFloat))
		}
	}
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
	utils.HumanDelay(800*time.Millisecond, 1800*time.Millisecond) //隔一下
	//上传人脸
	ObjectId, err := cache.UploadFaceImageApi(token, disturbImage)
	if err != nil {
		return "", err
	}
	if ObjectId == "" {
		return "", errors.New("ObjectId is empty")
	}
	utils.HumanDelay(600*time.Millisecond, 1500*time.Millisecond)
	//更换基准FaceId为新上传的照片(新版APP过人脸的核心: 先换基准再比对，绕过历史基准照片限制)
	if _, err := cache.ChangeFaceIdApi(ObjectId, 3, nil); err != nil {
		return "", err
	}
	utils.HumanDelay(800*time.Millisecond, 1600*time.Millisecond)
	//多接口依次校验: 主接口被拒(如"非法请求")时自动切换备用接口
	passed, detail := passFacePhoneCheck(cache, classId, courseId, chapterId, cpi, ObjectId)
	if !passed {
		return "", fmt.Errorf("人脸校验失败: %s", detail)
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
	utils.HumanDelay(800*time.Millisecond, 1800*time.Millisecond) //隔一下
	//上传人脸
	ObjectId, err := cache.UploadFaceImageApi(token, disturbImage)
	if err != nil {
		return "", err
	}
	if ObjectId == "" {
		return "", errors.New("ObjectId is empty")
	}
	utils.HumanDelay(600*time.Millisecond, 1500*time.Millisecond)
	//更换基准FaceId为新上传的照片(新版APP过人脸的核心: 先换基准再比对，绕过历史基准照片限制)
	if _, err := cache.ChangeFaceIdApi(ObjectId, 3, nil); err != nil {
		return "", err
	}
	utils.HumanDelay(800*time.Millisecond, 1600*time.Millisecond)
	//plan3Api, err := cache.GetCourseFaceQrPlan4Api( classId,courseId, chapterId,"", "", ObjectId)
	plan3Api, err := cache.PassFaceQrPlanPhoneOldApi(classId, courseId, chapterId, cpi, ObjectId)
	//暂时先用plan3的过人脸（进入课程扫通过的）
	if err != nil {
		return "", err
	}
	log2.Print(log2.DEBUG, "老接口uploadInfo响应: ", responseSummary(plan3Api))
	//统一走 faceCheckResult：status:true 优先于 msg，
	//避免老接口返回 {"status":true,"msg":"..."} 却被误判为失败
	if passed, reason := faceCheckResult(plan3Api); !passed {
		return "", fmt.Errorf("老接口人脸校验失败: %s", reason)
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

	utils.HumanDelay(500*time.Millisecond, 1200*time.Millisecond)
	//上传人脸
	ObjectId, err := cache.UploadFaceImageApi(token, face)
	if err != nil {
		return "", "", "", "", err
	}
	if ObjectId == "" {
		return "", "", "", "", errors.New("ObjectId is empty")
	}
	utils.HumanDelay(500*time.Millisecond, 1200*time.Millisecond)
	//更换基准FaceId为新上传的照片(新版APP过人脸的核心: 先换基准再比对，绕过历史基准照片限制)
	if _, err := cache.ChangeFaceIdApi(ObjectId, 3, nil); err != nil {
		return "", "", "", "", err
	}
	utils.HumanDelay(600*time.Millisecond, 1400*time.Millisecond)
	//plan3Api, err := cache.PassFaceQrPlanPhoneApi(classId, courseId, chapterId, cpi, ObjectId)
	//暂时先用plan3的过人脸（进入课程扫通过的）
	plan3Api, err := cache.GetCourseFaceQrPlan3Api(classId, courseId, uuid, qrEnc, cpi, ObjectId)

	if err != nil {
		return "", "", "", "", err
	}
	//PC扫码接口被拒(如"非法请求")时，回退手机端多接口校验
	if passed, reason := faceCheckResult(plan3Api); !passed {
		ok, detail := passFacePhoneCheck(cache, classId, courseId, chapterId, cpi, ObjectId)
		if !ok {
			return "", "", "", "", fmt.Errorf("人脸校验失败: %s; PC接口: %s", detail, reason)
		}
	}
	//获取人脸状态
	stateApi, err := cache.GetCourseFaceQrStateApi(uuid, qrEnc, classId, courseId, cpi, mid, videojobid, videoRandomCollectTime, chapterId)
	if err != nil {
		return "", "", "", "", err
	}
	stateCode := gojsonq.New().JSONString(plan3Api).Find("code")
	if stateCode != nil {
		if code, ok := stateCode.(float64); ok && int(code) != 0 {
			return "", "", "", "", errors.New(stateApi)
		}
	}
	successEnc, _ := gojsonq.New().JSONString(stateApi).Find("videoFaceCaptureSuccessEnc").(string)

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
