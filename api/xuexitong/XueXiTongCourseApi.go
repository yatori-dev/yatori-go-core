package xuexitong

//注意Api类文件主需要写最原始的接口请求和最后的json的string形式返回，不需要用结构体序列化。
//序列化和具体的功能实现请移步到Action代码文件中
import (
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"net/url"
)

// CourseListApi 拉取对应账号的课程数据
func (cache *XueXiTUserCache) CourseListApi() (string, error) {

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
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36 Edg/136.0.0.0")

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	cache.cookies = append(cache.cookies, res.Cookies()...)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
