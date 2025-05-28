package xuexitong

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/yatori-dev/yatori-go-core/utils"
	"image/jpeg"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"regexp"
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
	//req.Header.Add("Cookie", cache.cookie)
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")
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

// 上传人脸图片
func (cache *XueXiTUserCache) UploadFaceImage(token, imgPath string) (string, error) {
	image, _ := utils.LoadImage(imgPath)
	disturbImage := utils.ImageRGBDisturb(image)

	url := "https://pan-yz.chaoxing.com/upload"
	method := "POST"

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	//获取puid
	puid := "244955941"
	for _, cookie := range cache.cookies {
		if cookie.Name == "puid" {
			puid = cookie.Value
		}
	}
	// 添加参数字段
	_ = writer.WriteField("uploadtype", "face")
	_ = writer.WriteField("_token", token)
	_ = writer.WriteField("puid", puid)

	part, err := writer.CreateFormFile("file", fmt.Sprintf("%d", time.Now().UnixMilli())+".jpg")

	if err != nil {
		return "", err
	}
	err = jpeg.Encode(part, disturbImage, nil)
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
	req.Header.Add("Cookie", cache.cookie)
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

	log.Printf("人脸上传 resp: %+v\n", jsonResp)

	if jsonResp["result"] != true {
		return "", fmt.Errorf("人脸上传失败")
	}

	objectId, _ := jsonResp["objectId"].(string)
	data, _ := jsonResp["data"].(map[string]interface{})
	previewUrl, _ := data["previewUrl"].(string)

	log.Printf("人脸上传成功 I.%s/U.%s\n", objectId, previewUrl)
	return objectId, nil
}

// 根据PUID查找人脸图片上传
func (cache *XueXiTUserCache) UploadFaceImageForPUID(puid string) (string, error) {
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
	req.Header.Add("Cookie", cache.cookie)
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")
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
	req.Header.Add("Cookie", cache.cookie)
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

	//body, err := ioutil.ReadAll(res.Body)
	//if err != nil {
	//	fmt.Println(err)
	//	return "", "", nil
	//}
	//fmt.Println(string(body))
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", "", err
	}
	uuidFirst := doc.Find("input#uuid").First()
	uuidVal, _ := uuidFirst.Attr("value")
	qrcEncFirst := doc.Find("input#qrcEnc").First()
	qrcEncVal, _ := qrcEncFirst.Attr("value")
	return uuidVal, qrcEncVal, nil
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
	req.Header.Add("Cookie", cache.cookie)
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
	req.Header.Add("Cookie", cache.cookie)
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

// 获取人脸状态（二维码状态）
func (cache *XueXiTUserCache) GetCourseFaceQrStateApi() (string, error) {
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, "https://mooc1-api.chaoxing.com/knowledge/uploadInfo", nil)

	if err != nil {
		return "", err
	}
	req.Header.Add("Cookie", cache.cookie)
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")

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
