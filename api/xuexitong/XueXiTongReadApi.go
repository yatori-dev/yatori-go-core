package xuexitong

import (
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// 提交累计阅读时间
func (cache *XueXiTUserCache) ReadSubmitTimeLog(p *PointDocumentDto, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	t := time.Now().Format("20060102150405") + "000" // 伪造毫秒值
	wc := rand.Intn(1000) + 4000                     //观看的字数
	h := rand.Intn(100) + 4000                       //高度
	e := ""
	d := `{"a":null,"r":"` + p.CourseID + `,` + fmt.Sprintf("%d", p.KnowledgeID) + `","t":"special","l":1,"f":0,"wc":` + fmt.Sprintf("%d", wc) + `,"ic":0,"v":2,"s":2,"h":` + fmt.Sprintf("%d", h) + `,"e":"` + e + `","ext":"{\"_from_\":\"256268467_132232726_339543304_c8f68a62e7ef7fa3a704d6b031d19697\",\"rtag\":\"1054242600_477554005_read-` + p.CourseID + `\"}"}`
	settings := map[string]any{
		"f": "readPoint",
		//"u": "339543304", //userId
		"u": cache.UserID, //userId
		"s": "",
		"d": url.QueryEscape(d),
		"t": t,
	}

	enc := GenerateReadEnc(settings)
	//url := "https://data-xxt.aichaoxing.com/analysis/ac_mark?&f=readPoint&u=339543304&d=%257B%2522a%2522%253Anull%252C%2522r%2522%253A%2522218403608%252C437039890%2522%252C%2522t%2522%253A%2522special%2522%252C%2522l%2522%253A1%252C%2522f%2522%253A0%252C%2522wc%2522%253A1457%252C%2522ic%2522%253A1%252C%2522v%2522%253A2%252C%2522s%2522%253A2%252C%2522h%2522%253A168.6666717529297%252C%2522e%2522%253A%2522H4sIAAAAAAAAA43QMQ6AMAgF0NO4GqDAh1XvfycZ6qIMdKDJz4ME7JAbLFVXRFUWyOn1xBJu9R3rsi%252FiCaIBym0CVEO5sr%252BJgcHA%252BMDYwOg25prKSt6YBW1SGXS%252BpwXDJCXRbds10jBjogckSOsm9QEAAA%253D%253D%2522%252C%2522ext%2522%253A%2522%257B%255C%2522_from_%255C%2522%253A%255C%2522256268467_132232726_339543304_c8f68a62e7ef7fa3a704d6b031d19697%255C%2522%252C%255C%2522rtag%255C%2522%253A%255C%25221054242600_477554005_read-218403608%255C%2522%257D%2522%257D&t=20251212142327644&enc=11091c1703d2ba723afdf2338f24b8ae"
	urlStr := "https://data-xxt.aichaoxing.com/analysis/ac_mark?&f=readPoint&u=" + cache.UserID + "&d=" + url.QueryEscape(d) + "&t=" + t + "&enc=" + enc
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
	req, err := http.NewRequest(method, urlStr, nil)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Origin", "https://mooc1-1.chaoxing.com")
	req.Header.Add("Pragma", "no-cache")
	req.Header.Add("Sec-Fetch-Dest", "empty")
	req.Header.Add("Sec-Fetch-Mode", "cors")
	req.Header.Add("Sec-Fetch-Site", "cross-site")
	req.Header.Add("User-Agent", GetUA("mobile"))
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Add("sec-ch-ua", "\"Microsoft Edge\";v=\"143\", \"Chromium\";v=\"143\", \"Not A(Brand\";v=\"24\"")
	req.Header.Add("sec-ch-ua-mobile", "?0")
	req.Header.Add("sec-ch-ua-platform", "\"Windows\"")
	req.Header.Add("Host", "data-xxt.aichaoxing.com")

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
	fmt.Println(string(body))
	return string(body), nil
}

// 生成 enc 参数（完全匹配超星 JS）
func GenerateReadEnc(g map[string]any) string {
	// Step 1: key 排序
	keys := make([]string, 0, len(g))
	for k := range g {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Step 2: 拼接 value
	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(toString(g[k]))
	}
	h := sb.String()

	// Step 3: 加盐 MD5
	final := h + "NrRzLDpWB2JkeodIVAn4"

	sum := md5.Sum([]byte(final))
	return strings.ToLower(hex.EncodeToString(sum[:]))
}

// 把 interface{} 转换成 JS 里的字符串表现形式
func toString(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case int, int32, int64, float32, float64:
		return fmt.Sprintf("%v", val)
	default:
		// JSON.stringify 等效
		b, _ := json.Marshal(val)
		return string(b)
	}
}
