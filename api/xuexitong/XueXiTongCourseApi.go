package xuexitong

//注意Api类文件主需要写最原始的接口请求和最后的json的string形式返回，不需要用结构体序列化。
//序列化和具体的功能实现请移步到Action代码文件中
import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/yatori-dev/yatori-go-core/utils"
	"github.com/yatori-dev/yatori-go-core/utils/log"
)

// CourseListApi 拉取对应账号的课程数据
func (cache *XueXiTUserCache) CourseListApi(retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	method := "GET"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}

	//如果开启了IP代理，那么就直接添加代理
	if cache.IpProxySW {
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse(cache.ProxyIP) // 设置代理
		}
	}
	client := &http.Client{
		Transport: tr,
	}
	req, err := http.NewRequest(method, ApiPullCourses, nil)

	if err != nil {
		return "", err
	}
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}
	//req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36 Edg/136.0.0.0")
	req.Header.Add("User-Agent", GetUA("mobile"))
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
	if strings.Contains(string(body), "很抱歉，您所浏览的页面不存在") { //防止学习通抽风
		return cache.CourseListApi(retry-1, err)
	}
	return string(body), nil
}

// 拉取课程完成度状态
func (cache *XueXiTUserCache) CourseCompleteStatusApi(courseListData string, retry int, lastErr error) (string, error) {
	urlStr := "https://mooc2-ans.chaoxing.com/mooc2-ans/mycourse/stu-job-info?clazzPersonStr=" + url.QueryEscape(courseListData)
	//urlStr := "https://mooc2-ans.chaoxing.com/mooc2-ans/mycourse/stu-job-info?clazzPersonStr=134350229_407555221%252C125743273_407555221%252C127063689_407555221%252C126701067_407555221%252C125755386_407555221%252C125888882_407555221%252C125783661_454194591%252C124859308_454194591%252C124554421_454194591%252C116272688_407555221%252C117108284_407555221%252C117784832_407555221%252C116370785_407555221%252C116370660_407555221%252C117687599_407555221"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, urlStr, nil)

	if err != nil {
		log.Print(log.INFO, fmt.Sprintf("[%s]", cache.Name), err.Error())
		return "", nil
	}
	req.Header.Add("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Add("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Pragma", "no-cache")
	req.Header.Add("Referer", "https://mooc2-ans.chaoxing.com/mooc2-ans/visit/interaction?moocDomain=https://mooc1-1.chaoxing.com/mooc-ans")
	req.Header.Add("Sec-Fetch-Dest", "empty")
	req.Header.Add("Sec-Fetch-Mode", "cors")
	req.Header.Add("Sec-Fetch-Site", "same-origin")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Linux; Android 11; MI10 Build/OPM1.171019.019; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/71.0.3578.99 Mobile Safari/537.36 (schild:5e5510ce86e012a7f489e7c488fc17b4) (device:MI10) Language/zh_CN com.chaoxing.mobile/ChaoXingStudy_3_6.6.4_android_phone_10831_263 (@Kalimdor)_c86f59bf72a9e4a0540b390d77d3ec3d Edg/142.0.0.0")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Add("sec-ch-ua", "\"\"")
	req.Header.Add("sec-ch-ua-mobile", "?1")
	req.Header.Add("sec-ch-ua-platform", "\"\"")
	//req.Header.Add("Cookie", "tl=1; fanyamoocs=11401F839C536D9E; fid=1668; _uid=225366429; _d=1763716664776; UID=225366429; vc3=Ht0raJ7eU0q2giGKRgmVATj3sotta9iUNmRgCnCAdaL%2FwKKa8OgNvt9Opc5OtFZz5NGxXEtkkp1V7g3kAXvCtD%2FSACOBdX6P%2FWXftDFvk6AOP4WAkaGdGeYiB7D%2B3sxaqqqP8cGD60b1bQpgVqAvcxuM0bNAb1s%2BRv2IQ0KPCvE%3D8ecaf7ed46f69fb6b172d8c4dae310b6; uf=b2d2c93beefa90dcd9da4b7d2fedbd7afe4a1d9d29436c4e6cff1904f69931641a8c565fab7db4413c0e0903ab46345681a6c9ddee30899fd807a544f7930b6aed1e6c11a143bb563b0339d97cdac4ba8e51c3c7f17d1921e5851b744f8aa02c9fb3947ed09a594c7285f94493657dcbfdf06e8c02a6dd9b628e52e343eee961a6ef692ee0b1b6b684d250e6e5420be370b5a05e402d2a6370184964ffe8c27c7e9bd78ebd57f02adff861248a798ee3b1f899d50c1c3fa3aa2ebad65cd196bb; cx_p_token=d3fbaa8a1a8b17e6ef266fa0b994204e; p_auth_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOiIyMjUzNjY0MjkiLCJsb2dpblRpbWUiOjE3NjM3MTY2NjQ3NzgsImV4cCI6MTc2NDMyMTQ2NH0.AEJpPS3fD-07ec6CZwGrCjOgRkIEtxJ56MP7PheOf7M; xxtenc=d04c17efd580d691db044016bfedc2f9; DSSTASH_LOG=C_38-UN_644-US_225366429-T_1763716664778; source=\"\"; thirdRegist=0; k8s=1763716667.926.23122.851800; jrose=B03FA5831DA7B56F6C13C619DD42FB3C.mooc2-2039891087-cb8zn; route=1ab934bb3bbdaaef56ce3b0da45c52ed; jrosehead=1B2A2A395C4E0544EE3B000BAA9ECDCD.mooc-portal-3714944895-tkjc6")
	req.Header.Add("Host", "mooc2-ans.chaoxing.com")
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}

	res, err := client.Do(req)
	if err != nil {
		log.Print(log.INFO, fmt.Sprintf("[%s]", cache.Name), err.Error())
		return "", nil
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Print(log.INFO, fmt.Sprintf("[%s]", cache.Name), err.Error())
		return "", nil
	}
	return string(body), nil
}
