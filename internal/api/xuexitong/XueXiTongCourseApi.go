package xuexitong

//注意Api类文件主需要写最原始的接口请求和最后的json的string形式返回，不需要用结构体序列化。
//序列化和具体的功能实现请移步到Action代码文件中
import (
	"bytes"
	"context"
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/yatori-dev/yatori-go-core/v3/core"
	"github.com/yatori-dev/yatori-go-core/v3/utils"
)

// CourseListApi 拉取对应账号的课程数据
func (cache *XueXiTUserCache) CourseListApi(retry int, lastErr error) (string, error) {
	return cache.CourseListApiContext(context.Background(), retry, lastErr)
}

func (cache *XueXiTUserCache) CourseListApiContext(ctx context.Context, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	method := "GET"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{},
	}

	//如果开启了IP代理，那么就直接添加代理
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse("http://" + cache.ProxyIP) // 设置代理
		}
	}
	client := newRequestClient(tr)
	req, err := http.NewRequestWithContext(ctx, method, "https://mooc1-api.chaoxing.com/mycourse/backclazzdata", nil)

	if err != nil {
		return "", err
	}
	cache.applyCourseSession(client, req, GetUA("mobile"))
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	//cache.cookies = append(cache.cookies, res.Cookies()...)
	utils.CookiesAddNoRepetition(&cache.cookies, res.Cookies()) //赋值cookie
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	if (res.Request != nil && res.Request.URL != nil && res.Request.URL.Hostname() == "passport2.chaoxing.com") ||
		(bytes.Contains(body, []byte(`id="phone"`)) && bytes.Contains(body, []byte(`id="pwd"`))) {
		return "", core.NotLoggedIn(core.XueXiTong, "courses")
	}
	if err := validateResponse("拉取课程列表", res, body); err != nil {
		return "", err
	}
	if strings.Contains(string(body), "很抱歉，您所浏览的页面不存在") { //防止学习通抽风
		lastErr = &RemoteError{Operation: "拉取课程列表", StatusCode: res.StatusCode, BodySummary: responseSummary(body)}
		return cache.CourseListApiContext(ctx, retry-1, lastErr)
	}
	return string(body), nil
}

// 拉取课程完成度状态
func (cache *XueXiTUserCache) CourseCompleteStatusApi(courseListData string, retry int, lastErr error) (string, error) {
	return cache.CourseCompleteStatusApiContext(context.Background(), courseListData, retry, lastErr)
}

func (cache *XueXiTUserCache) CourseCompleteStatusApiContext(ctx context.Context, courseListData string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	urlStr := "https://mooc2-ans.chaoxing.com/mooc2-ans/mycourse/stu-job-info?clazzPersonStr=" + url.QueryEscape(courseListData)
	//urlStr := "https://mooc2-ans.chaoxing.com/mooc2-ans/mycourse/stu-job-info?clazzPersonStr=134350229_407555221%252C125743273_407555221%252C127063689_407555221%252C126701067_407555221%252C125755386_407555221%252C125888882_407555221%252C125783661_454194591%252C124859308_454194591%252C124554421_454194591%252C116272688_407555221%252C117108284_407555221%252C117784832_407555221%252C116370785_407555221%252C116370660_407555221%252C117687599_407555221"
	method := "GET"
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{},
	}
	client := newRequestClient(tr)
	req, err := http.NewRequestWithContext(ctx, method, urlStr, nil)

	if err != nil {
		return "", err
	}
	req.Header.Add("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Add("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Pragma", "no-cache")
	//req.Header.Add("Referer", "https://mooc2-ans.chaoxing.com/mooc2-ans/visit/interaction?moocDomain=https://mooc1-1.chaoxing.com/mooc-ans")
	req.Header.Add("Sec-Fetch-Dest", "empty")
	req.Header.Add("Sec-Fetch-Mode", "cors")
	req.Header.Add("Sec-Fetch-Site", "same-origin")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Linux; Android 11; MI10 Build/OPM1.171019.019; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/71.0.3578.99 Mobile Safari/537.36 (schild:5e5510ce86e012a7f489e7c488fc17b4) (device:MI10) Language/zh_CN com.chaoxing.mobile/ChaoXingStudy_3_6.6.4_android_phone_10831_263 (@Kalimdor)_c86f59bf72a9e4a0540b390d77d3ec3d Edg/142.0.0.0")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Add("sec-ch-ua", "\"\"")
	req.Header.Add("sec-ch-ua-mobile", "?1")
	req.Header.Add("sec-ch-ua-platform", "\"\"")
	req.Header.Add("Host", "mooc2-ans.chaoxing.com")
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	if err := validateResponse("拉取课程完成状态", res, body); err != nil {
		return "", err
	}
	return string(body), nil
}
