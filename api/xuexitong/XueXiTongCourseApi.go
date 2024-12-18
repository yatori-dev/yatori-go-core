package xuexitong

//注意Api类文件主需要写最原始的接口请求和最后的json的string形式返回，不需要用结构体序列化。
//序列化和具体的功能实现请移步到Action代码文件中
import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
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

// 获取人脸二维码信息（未完成）
func (cache *XueXiTUserCache) GetCourseFaceQrApi(classId, courseId, uuid, objectId, qrcEnc, failCount string) (string, error) {
	//req, err := http.NewRequest(method, "https://mooc1-api.chaoxing.com/qr/updateqrstatus"+uuid+"&clazzId2="+classId, nil)
	req, err := http.PostForm("https://mooc1-api.chaoxing.com/qr/updateqrstatus", url.Values{
		"clazzId":       {classId},
		"courseId":      {courseId},
		"uuid":          {uuid},
		"objectId":      {objectId},
		"qrcEnc":        {qrcEnc},
		"failCount":     {failCount},
		"compareResult": {"0"},
	})
	if err != nil {
		return "", err
	}
	req.Header.Add("Cookie", cache.cookie)
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")

	if err != nil {
		fmt.Println("Error sending request:", err)
		return "", err
	}
	defer req.Body.Close()

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return "", err
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
