package xuexitong

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/thedevsaddam/gojsonq"
	log2 "github.com/yatori-dev/yatori-go-core/utils/log"
	"image"
	"image/jpeg"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// 用于获取云盘token（用于人脸）
func (cache *XueXiTUserCache) GetFaceUpLoadToken() (string, error) {

	url := "https://pan-yz.chaoxing.com/api/token/uservalid"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36 Edg/136.0.0.0")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "mooc1.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	for _, cookie := range res.Cookies() {
		if cookie.Name == "puid" {
			fmt.Println(cookie.Value)
		}
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	return string(body), nil
}

// 获取历史人脸图片
func (cache *XueXiTUserCache) GetHistoryFaceImg(puid string) (string, image.Image, error) {
	//获取puid
	if puid == "" {
		for _, cookie := range cache.cookies {
			if cookie.Name == "UID" { //获取puid
				puid = cookie.Value
				break
			}
		}
	}
	hash := md5.Sum([]byte(puid + "uWwjeEKsri"))
	enc := hex.EncodeToString(hash[:])
	url := "https://passport2-api.chaoxing.com/api/getUserFaceid?enc=" + enc + "&token=4faa8662c59590c6f43ae9fe5b002b42&_time=" + fmt.Sprintf("%d", time.Now().Unix())
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return "", nil, err
	}
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36 Edg/136.0.0.0")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "passport2-api.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", nil, err
	}
	if strconv.Itoa(int(gojsonq.New().JSONString(string(body)).Find("result").(float64))) != "1" {
		return "", nil, nil
	}
	//如果为空
	if gojsonq.New().JSONString(string(body)).Find("data.http").(string) == "" {
		return "", nil, errors.New("没有历史人脸")
	}
	//图片获取段
	methodImg := "GET"
	clientImg := &http.Client{}
	reqImg, errImg := http.NewRequest(methodImg, gojsonq.New().JSONString(string(body)).Find("data.http").(string), nil)

	if errImg != nil {
		fmt.Println(errImg)
		return "", nil, errImg
	}
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}
	reqImg.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36 Edg/136.0.0.0")
	reqImg.Header.Add("Accept", "*/*")
	reqImg.Header.Add("Host", "passport2-api.chaoxing.com")
	reqImg.Header.Add("Connection", "keep-alive")

	resImg, errImg := clientImg.Do(reqImg)
	if errImg != nil {
		fmt.Println(errImg)
		return "", nil, errImg
	}
	defer res.Body.Close()

	// 解码图片
	img, _, errImg := image.Decode(resImg.Body)
	if errImg != nil {
		return "", nil, fmt.Errorf("图片解码失败: %w", err)
	}
	return string(body), img, nil
}

// 上传人脸图片
func (cache *XueXiTUserCache) UploadFaceImageApi(token string, image image.Image) (string, error) {

	url := "https://pan-yz.chaoxing.com/upload"
	method := "POST"

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	//获取puid
	puid := ""
	for _, cookie := range cache.cookies {
		if cookie.Name == "UID" { //获取puid
			puid = cookie.Value
			break
		}
	}
	// 添加参数字段
	_ = writer.WriteField("uploadtype", "face") //还有一种normal类型，一般用于上传文件
	_ = writer.WriteField("_token", token)
	_ = writer.WriteField("puid", puid)

	part, err := writer.CreateFormFile("file", fmt.Sprintf("%d", time.Now().UnixMilli())+".jpg")

	if err != nil {
		return "", err
	}
	err = jpeg.Encode(part, image, nil)
	if err != nil {
		return "", err
	}

	writer.Close()

	client := &http.Client{}
	req, err := http.NewRequest(method, url, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Linux; Android 12; SM-N9006 Build/V417IR; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/95.0.4638.74 Mobile Safari/537.36 (schild:e9b05c3f9fb49fef2f516e86ac3c4ff1) (device:SM-N9006) Language/zh_CN com.chaoxing.mobile/ChaoXingStudy_3_6.3.7_android_phone_10822_249 (@Kalimdor)_4627cad9c4b6415cba5dc6cac39e6c96")
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "mooc1-api.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	defer resp.Body.Close()

	// 解析响应 JSON
	var jsonResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&jsonResp)
	if err != nil {
		return "", err
	}

	log2.Print(log2.DEBUG, "人脸上传 resp: ", jsonResp)

	if jsonResp["result"] != true {
		return "", fmt.Errorf("人脸上传失败")
	}

	objectId, _ := jsonResp["objectId"].(string)
	data, _ := jsonResp["data"].(map[string]interface{})
	previewUrl, _ := data["previewUrl"].(string)

	log2.Print(log2.DEBUG, "人脸上传成功 ", objectId, " ", previewUrl)
	return objectId, nil
}

// 根据PUID查找人脸图片上传
func (cache *XueXiTUserCache) UploadFaceImageForPUIDApi(puid string) (string, error) {
	if puid == "" {
		cookies := cache.cookies
		for _, cookie := range cookies {
			if cookie.Name == "puid" {
				puid = cookie.Value
			}
		}
	}
	return "", nil
}

// 获取人脸的必要数据（老的）
func (cache *XueXiTUserCache) GetFaceQrCodeApi1(courseId, clazzid, chapterId, cpi string) (string, string, error) {

	url := "https://mooc1.chaoxing.com/mooc-ans/mycourse/studentstudyAjax?courseId=" + courseId + "&clazzid=" + clazzid + "&chapterId=" + chapterId + "&cpi=" + cpi + "&verificationcode=&mooc2=1&toComputer=false&microTopicId=0"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return "", "", nil
	}
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36 Edg/136.0.0.0")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "mooc1.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", "", nil
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", "", nil
	}
	var uuid string
	var qrcEnc string
	uuidPattern := `<input type="hidden" value="([^"]+)" id="uuid"/>`
	uuidRegexp := regexp.MustCompile(uuidPattern)
	uuidMatcher := uuidRegexp.FindStringSubmatch(string(body))
	if len(uuidMatcher) > 0 {
		uuid = uuidMatcher[1]
	}
	qrcEncPattern := `<input type="hidden" value="([^"]+)" id="qrcEnc"/>`
	qrcEncRegexp := regexp.MustCompile(qrcEncPattern)
	qrcEncMatcher := qrcEncRegexp.FindStringSubmatch(string(body))
	if len(qrcEncMatcher) > 0 {
		qrcEnc = qrcEncMatcher[1]
	}
	return uuid, qrcEnc, nil
}

// 获取人脸的必要数据
func (cache *XueXiTUserCache) GetFaceQrCodeApi2(courseId, clazzId, cpi string) (string, string, error) {

	url := "https://mooc1.chaoxing.com/visit/stucoursemiddle?" + "courseid=" + courseId + "&clazzid=" + clazzId + "&cpi=" + cpi + "&ismooc2=1"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return "", "", nil
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Linux; Android 12; SM-N9006 Build/V417IR; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/95.0.4638.74 Mobile Safari/537.36 (schild:e9b05c3f9fb49fef2f516e86ac3c4ff1) (device:SM-N9006) Language/zh_CN com.chaoxing.mobile/ChaoXingStudy_3_6.3.7_android_phone_10822_249 (@Kalimdor)_4627cad9c4b6415cba5dc6cac39e6c96")
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "mooc1-api.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", "", nil
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", "", nil
	}
	//fmt.Println(string(body))
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return "", "", err
	}
	uuidFirst := doc.Find("input#uuid").First()
	uuidVal, _ := uuidFirst.Attr("value")
	qrcEncFirst := doc.Find("input#qrcEnc").First()
	qrcEncVal, _ := qrcEncFirst.Attr("value")
	//if qrcEncVal == "" {
	//	uploadEnc := doc.Find("input#uploadEnc").First()
	//	qrcEncVal, _ = uploadEnc.Attr("value")
	//}
	//if uuidVal == "" {
	//	uploadUid := doc.Find("input#uploadUid").First()
	//	uuidVal, _ = uploadUid.Attr("value")
	//}
	return uuidVal, qrcEncVal, nil
}

// 拉人脸数据3（课程中）
func (cache *XueXiTUserCache) GetFaceQrCodeApi3(courseId, clazzid, chapterId, cpi, enc, videojobid, chaptervideoobjectid string) (string, string, error) {

	url := "https://mooc1.chaoxing.com/mycourse/studentstudy?chapterId=" + chapterId + "&courseId=" + courseId + "&clazzid=" + clazzid + "&cpi=" + cpi + "&enc=" + enc + "&mooc2=1"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return "", "", err
	}
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}

	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36 Edg/136.0.0.0")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "mooc1.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", "", err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", "", err
	}

	//fmt.Println(string(body))
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return "", "", err
	}
	uuidFirst := doc.Find("input#uuid").First()
	uuidVal, _ := uuidFirst.Attr("value")
	qrcEncFirst := doc.Find("input#qrcEnc").First()
	qrcEncVal, _ := qrcEncFirst.Attr("value")

	//第二步---------------------------
	url1 := "https://mooc1.chaoxing.com/mooc-ans/qr/produce?uuid=" + uuidVal + "&enc=" + qrcEncVal + "&clazzid=" + clazzid + "&videojobid=" + videojobid + "&chaptervideoobjectid=" + chaptervideoobjectid + "&videoCollectTime=0"
	method1 := "GET"

	client1 := &http.Client{}
	req1, err1 := http.NewRequest(method1, url1, nil)

	if err1 != nil {
		fmt.Println(err)
		return "", "", nil
	}
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}
	req1.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")
	req1.Header.Add("Accept", "*/*")
	req1.Header.Add("Host", "mooc1.chaoxing.com")
	req1.Header.Add("Connection", "keep-alive")

	res1, err1 := client1.Do(req1)
	if err1 != nil {
		fmt.Println(err1)
		return "", "", nil
	}
	defer res.Body.Close()

	body1, err := ioutil.ReadAll(res1.Body)
	if err != nil {
		fmt.Println(err)
		return "", "", nil
	}
	//fmt.Println(string(body1))
	jsonStatus := gojsonq.New().JSONString(string(body1)).Find("status")
	if jsonStatus == nil {
		return "", "", nil
	}
	if jsonStatus.(bool) == false {
		return "", "", nil
	}
	newEnc := gojsonq.New().JSONString(string(body1)).Find("newEnc").(string)
	newUuid := gojsonq.New().JSONString(string(body1)).Find("newUuid").(string)
	return newUuid, newEnc, nil
}

// 过人脸（第一版）
func (cache *XueXiTUserCache) GetCourseFaceQrPlan1Api(courseId, classId, uuid, objectId, qrcEnc, failCount string) (string, error) {

	url := "https://mooc1-api.chaoxing.com/qr/updateqrstatus"
	method := "POST"

	payload := strings.NewReader("clazzId=" + classId + "&courseId=" + courseId + "&uuid=" + uuid + "&qrcEnc=" + qrcEnc + "&objectId=" + objectId + "&failCount=" + failCount + "&compareResult=0")

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Linux; Android 12; SM-N9006 Build/V417IR; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/95.0.4638.74 Mobile Safari/537.36 (schild:e9b05c3f9fb49fef2f516e86ac3c4ff1) (device:SM-N9006) Language/zh_CN com.chaoxing.mobile/ChaoXingStudy_3_6.3.7_android_phone_10822_249 (@Kalimdor)_4627cad9c4b6415cba5dc6cac39e6c96")

	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "mooc1-api.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	return string(body), nil
}

// 过人脸（第二版）
func (cache *XueXiTUserCache) GetCourseFaceQrPlan2Api(classId, courseId, knowledgeId, cpi, objectId /*人脸上传id*/ string) (string, error) {

	url := "https://mooc1-api.chaoxing.com/mooc-ans/facephoto/clientfacecheckstatus?" + "courseId=" + courseId + "&clazzId=" + classId + "&cpi=" + cpi + "&chapterId=" + knowledgeId + "&objectId=" + objectId + "&type=1"
	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Linux; Android 12; SM-N9006 Build/V417IR; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/95.0.4638.74 Mobile Safari/537.36 (schild:e9b05c3f9fb49fef2f516e86ac3c4ff1) (device:SM-N9006) Language/zh_CN com.chaoxing.mobile/ChaoXingStudy_3_6.3.7_android_phone_10822_249 (@Kalimdor)_4627cad9c4b6415cba5dc6cac39e6c96")
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "mooc1-api.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")
	//req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", nil
	}

	return string(body), nil
}

// 过人脸（第三版）
func (cache *XueXiTUserCache) GetCourseFaceQrPlan3Api(uuid, clazzId, courseId, qrcEnc, objectId /*人脸的objectId*/ string) (string, error) {
	url := "https://mooc1-api.chaoxing.com/qr/updateqrstatus?uuid2=" + uuid + "&clazzId2=" + clazzId
	method := "POST"

	payload := strings.NewReader("clazzId=" + clazzId + "&courseId=" + courseId + "&uuid=" + uuid + "&qrcEnc=" + qrcEnc + "&objectId=" + objectId)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		return "", nil
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Linux; Android 12; SM-N9006 Build/V417IR; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/95.0.4638.74 Mobile Safari/537.36 (schild:e9b05c3f9fb49fef2f516e86ac3c4ff1) (device:SM-N9006) Language/zh_CN com.chaoxing.mobile/ChaoXingStudy_3_6.3.7_android_phone_10822_249 (@Kalimdor)_4627cad9c4b6415cba5dc6cac39e6c96")
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "mooc1-api.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// 获取人脸状态（二维码状态）
//
//	func (cache *XueXiTUserCache) GetCourseFaceQrStateApi() (string, error) {
//		method := "GET"
//
//		client := &http.Client{}
//		req, err := http.NewRequest(method, "https://mooc1-api.chaoxing.com/knowledge/uploadInfo", nil)
//
//		if err != nil {
//			return "", err
//		}
//		req.Header.Add("Cookie", cache.cookie)
//		req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")
//
//		res, err := client.Do(req)
//		if err != nil {
//			return "", err
//		}
//		defer res.Body.Close()
//
//		body, err := ioutil.ReadAll(res.Body)
//		if err != nil {
//			return "", err
//		}
//		return string(body), nil
//	}
//
// 获取人脸状态{"code":"0","failCount":"90","videoFaceCaptureSuccessEnc":"2416cd8e0f5949d4b4d66da05aafb15a","compareResult":"0","status":2}
func (cache *XueXiTUserCache) GetCourseFaceQrStateApi(uuid, enc, clazzid, courseid, cpi, mid, videoObjectId, videoRandomCollectTime, chapterId string) (string, error) {
	url := "https://mooc1.chaoxing.com/mooc-ans/qr/getqrstatus?uuid=" + uuid + "&enc=" + enc + "&clazzid=" + clazzid + "&courseid=" + courseid + "&cpi=" + cpi + "&collectionTime=0&mid=" + mid + "&videoObjectId=" + videoObjectId + "&videoRandomCollectTime=" + videoRandomCollectTime + "&chapterId=" + chapterId
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36 Edg/136.0.0.0")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "mooc1.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	log2.Print(log2.DEBUG, string(body))
	return string(body), nil
}
