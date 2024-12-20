package xuexitong

//注意Api类文件主需要写最原始的接口请求和最后的json的string形式返回，不需要用结构体序列化。
//序列化和具体的功能实现请移步到Action代码文件中
import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

// CourseListApi 拉取对应账号的课程数据
func (cache *XueXiTUserCache) CourseListApi() (string, error) {

	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, ApiPullCourses, nil)

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

// 获取人脸的必要数据
func (cache *XueXiTUserCache) GetFaceQrCodeApi(courseId, clazzid, chapterId, cpi string) (string, string, error) {

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
func (cache *XueXiTUserCache) GetCourseFaceQrPlan2Api(classId, courseId, uuid, objectId, qrcEnc, failCount string) (string, error) {
	return "", nil
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
