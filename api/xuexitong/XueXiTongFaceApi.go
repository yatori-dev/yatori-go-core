package xuexitong

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	log2 "github.com/yatori-dev/yatori-go-core/utils/log"
	"image"
	"image/jpeg"
	"io/ioutil"
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
	//req.Header.Add("Cookie", " fid=5339; _uid=348625454; _d=1748507296144; UID=348625454; vc3=VksOHe9Jcepoyb%2F1zUUZcj48Q24R2ChAShc98TF6eQx6W1kPavuEosmRIotMFiEpq04q2%2FeLFFGrvL4R2jG7ppnTY1ntggtqDVANrEeUz6f2y4UaIpqItBlucT1W5fMVoK60DC4CIkODHR%2BZPiYIuBMHA3GzB9pP6C8QrtxtsPs%3D6b76dd0fcba78eb963ddb266cc8e47f6; uf=569b376a64ccf0313129ca082ab4eaeede7e7778b17f9ae8265c811413bbd05ba698eb83c701a3b8db92082134c30573913b662843f1f4ade9295d8c89b08ad0f44425e20f927c6b94405ac272c83515fb98ce0e6210c3884a878d0a9a7b05dad8a8d0ca21d204eb3ad59b143144275b3d7e9258df4ffdf630409e216a2b096279123a9828d1f8e0; cx_p_token=fbd8a60b4a0d4ac6c103d937fe2fe90f; p_auth_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOiIzNDg2MjU0NTQiLCJsb2dpblRpbWUiOjE3NDg1MDcyOTYxNDYsImV4cCI6MTc0OTExMjA5Nn0.akLuZmVAalyIWDy8xDai-PLbM6Dkv4-0bUGfJIziopY; xxtenc=256f12e17e3f57e301008b366801437c; DSSTASH_LOG=C_38-UN_4533-US_348625454-T_1748507296146; source=num2; spaceFid=5339; spaceRoleId=3; tl=1; k8s=1748507308.412.12905.922105; jrose=BCABAAD7A1882C83BEEFD98740BE4683.mooc-1248283859-d0tq3; route=f9c314690d8e5d436efa7770254d0199; jrose=9E53AB2E95196F2A0541494D6D1FEF02.mooc-1248283859-jmxwc; k8s=1748596640.961.8327.803851; route=440ceb57420433374ff0504da9778fc7")
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")
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
	//url1 := "https://mooc1.chaoxing.com/mooc-ans/qr/produce?chapterId=" + chapterId + "&courseId=" + courseId + "&clazzid=" + clazzid + "&cpi=" + cpi + "&enc=" + enc + "&mooc2=1&uuid=" + uuidVal + "&videojobid=" + videojobid + "&chaptervideoobjectid=" + chaptervideoobjectid + "&videoCollectTime=0"
	//method1 := "GET"
	//
	//client1 := &http.Client{}
	//req1, err1 := http.NewRequest(method1, url1, nil)
	//
	//if err1 != nil {
	//	fmt.Println(err)
	//	return "", "", nil
	//}
	//for _, cookie := range cache.cookies {
	//	req.AddCookie(cookie)
	//}
	////req.Header.Add("Cookie", "tl=1; fid=5339; _uid=348625454; _d=1748513600505; UID=348625454; vc3=SiYd6lgojnaFlZpLJJBtcGdxrKtJwRucOxESqFSopRDr%2BhuD1kaISAN%2BYYbRyBzw7jSvMUpaXB%2FhEZFGTZ02iS7Tqtqmn5BGub5qSMJQ7JNtpwxEWj7QHuW0YTyL8wyrRS2KXLLluxfAd1hR%2FbTMJhZrHFBENnJFXnuvsdTuwRQ%3Dfdb0688bce66c9e827c0c44e57e274c6; uf=569b376a64ccf0313129ca082ab4eaeede7e7778b17f9ae8265c811413bbd05ba698eb83c701a3b8823c7d7bef8bc618913b662843f1f4ade9295d8c89b08ad0f44425e20f927c6b94405ac272c83515fb98ce0e6210c3884a878d0a9a7b05dad8a8d0ca21d204eb3ad59b143144275b95dfc474decc5b9a2cfce7bdc815a7d879123a9828d1f8e0; cx_p_token=1965b35a7496b40a03bd0085ad5284ac; p_auth_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOiIzNDg2MjU0NTQiLCJsb2dpblRpbWUiOjE3NDg1MTM2MDA1MDcsImV4cCI6MTc0OTExODQwMH0.gTgeghUuSwW75JNmyxiCd_I1pMHk2EOOR6kHp9OkOzQ; xxtenc=256f12e17e3f57e301008b366801437c; DSSTASH_LOG=C_38-UN_4533-US_348625454-T_1748513600508; k8s=1748523387.424.3082.42948; jrose=17C17148C482DE991CBE252E75C48A2E.mooc-1248283859-kqdd5; route=f537d772be8122bff9ae56a564b98ff6; writenote=yes; videojs_id=7142498; jrose=C18E3EF6DEE9443E9438AE7D90BE63E0.mooc-1248283859-jmxwc; k8s=1748596640.961.8327.803851; route=440ceb57420433374ff0504da9778fc7")
	//req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")
	//req.Header.Add("Accept", "*/*")
	//req.Header.Add("Host", "mooc1.chaoxing.com")
	//req.Header.Add("Connection", "keep-alive")
	//
	//res1, err1 := client1.Do(req1)
	//if err1 != nil {
	//	fmt.Println(err1)
	//	return "", "", nil
	//}
	//defer res.Body.Close()
	//
	//body1, err := ioutil.ReadAll(res1.Body)
	//if err != nil {
	//	fmt.Println(err)
	//	return "", "", nil
	//}
	//fmt.Println(string(body1))
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
// 获取人脸状态
func (cache *XueXiTUserCache) GetCourseFaceQrStateApi(uuid, enc, clazzid, courseid, cpi string) (string, error) {

	url := "https://mooc1.chaoxing.com/mooc-ans/qr/getqrstatus?uuid=" + uuid + "&enc=" + enc + "&clazzid=" + clazzid + "&courseid=" + courseid + "&cpi=" + cpi + "&ismooc2=1&v=0&pageHeader=-1&taskrefId=&workOrExam="
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
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")
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
