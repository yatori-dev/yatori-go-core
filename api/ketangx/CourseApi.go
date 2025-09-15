package ketangx

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/yatori-dev/yatori-go-core/utils"
)

// 拉取课程对应列表HTML
func (cache *KetangxUserCache) PullCourseListHTMLApi() (string, error) {
	url := "https://www.ketangx.cn/Activity/Query"
	method := "POST"

	payload := strings.NewReader("actType=2&actStart=&actClose=&formId=&classId=&actKey=&actState=&timeId=" + fmt.Sprintf("%d", time.Now().UnixMilli()))

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "www.ketangx.cn")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	for _, cookie := range cache.Cookies {
		req.AddCookie(cookie)
	}

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
	cache.Cookies = res.Cookies()

	//fmt.Println(string(body))
	return string(body), nil
}

// 拉取课程对应视屏列表HTML
func (cache *KetangxUserCache) PullVideoListHTMLApi(courseId string) (string, error) {

	url := "https://www.ketangx.cn/DoAct/ActIndex/" + courseId + "?_=" + fmt.Sprintf("%d", time.Now().UnixMilli())
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	for _, cookie := range cache.Cookies {
		req.AddCookie(cookie)
	}
	//req.Header.Add("Cookie", "ASP.NET_SessionId=4rjzprtyowdj0zg321zymxzt; acw_tc=0b32973617579566456903578e9210c223548f6939dbd7444be03da24735bd; ZHYX=702720bb490846b8aec2b34500dae627_15213625522_2; SERVERID=698319db3a2920f24616a79b4e94f782|1757958314|1757956645; SERVERID=698319db3a2920f24616a79b4e94f782|1757958778|1757956645; ZHYX=702720bb490846b8aec2b34500dae627_15213625522_2")
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "www.ketangx.cn")
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
	//fmt.Println(string(body))
	return string(body), nil
}

// 标记任务点状态API，视屏学习前必须要先调用这个
func (cache *KetangxUserCache) SignVideoStatusApi(sectId string) (string, error) {

	url := "https://www.ketangx.cn/DoAct/GetSection?id=" + sectId + "&_=" + fmt.Sprintf("%d", time.Now().UnixMilli())
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("Cookie", "ASP.NET_SessionId=4rjzprtyowdj0zg321zymxzt; ZHYX=702720bb490846b8aec2b34500dae627_15213625522_2; acw_tc=0b32973617579140663533976e9240ae7b9485a649812e7944a71bfd3f310b; SERVERID=698319db3a2920f24616a79b4e94f782|1757914376|1757914066; SERVERID=6ca4eae317291c9402a558d8073fd296|1757963692|1757963692; ZHYX=702720bb490846b8aec2b34500dae627_15213625522_2")
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "www.ketangx.cn")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Referer", "https://www.ketangx.cn/DoAct/GetSection?id=5df477aeb91a42ef83e0b34e00e88825&_=1757963262117")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return string(body), nil
}

// 完成视频任务点API
func (cache *KetangxUserCache) CompleteVideoApi(sectId, userId string, studyTime, duration int) (string, error) {

	url := "https://www.ketangx.cn/Common/SetDuration"
	method := "POST"

	payload := strings.NewReader("studyData%5BSectId%5D=" + sectId + "&studyData%5BUserId%5D=" + userId + "&studyData%5BStudyTime%5D=" + fmt.Sprintf("%d", studyTime) + "&studyData%5BDuraion%5D=" + fmt.Sprintf("%d", duration))

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return "", nil
	}
	for _, cookie := range cache.Cookies {
		req.AddCookie(cookie)
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "www.ketangx.cn")
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
	//fmt.Println(string(body))
	return string(body), nil
}
