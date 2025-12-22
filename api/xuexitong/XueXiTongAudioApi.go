package xuexitong

import (
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

// 移动端提交音频
func (cache *XueXiTUserCache) AudioPhoneSubmitTimeApi(p *PointVideoDto, playingTime int, isdrag int /*提交模式，0代表正常视屏播放提交，2代表暂停播放状态，3代表着点击开始播放状态*/, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	clipTime := fmt.Sprintf("0_%d", p.Duration)
	hash := md5.Sum([]byte(fmt.Sprintf("[%s][%s][%s][%s][%d][%s][%d][%s]",
		p.ClassID, cache.UserID, p.JobID, p.ObjectID, playingTime*1000, "d_yHJ!$pdA~5", p.Duration*1000, clipTime)))
	enc := hex.EncodeToString(hash[:])
	//url := "https://mooc1-api.chaoxing.com/mooc-ans/multimedia/log?objectId=90d77e2b106bfa28dbe0e00bf4e7d3c2&clazzId=134204187&userid=346635955&jobid=176552392921124&duration=1&otherInfo=nodeId_1088037085-cpi_411545273-rt_d-ds_false-ff_1-vt_0-v_5-enc_43bcc41f311ac76e7f3a62175e32f41d&courseId=258101827&dtype=Audio&view=json&playingTime=1&isdrag=4&enc=42165394c2d241f07ca1ff0ee797f84b&_dc=1765524099658"
	urlStr := "https://mooc1-api.chaoxing.com/mooc-ans/multimedia/log?objectId=" + p.ObjectID + "&clazzId=" + p.ClassID + "&userid=" + cache.UserID + "&jobid=" + p.JobID + "&duration=" + fmt.Sprintf("%d", p.Duration) + "&otherInfo=" + p.OtherInfo + "&courseId=" + p.CourseID + "&dtype=Audio&view=json&playingTime=" + fmt.Sprintf("%d", playingTime) + "&isdrag=" + fmt.Sprintf("%d", isdrag) + "&enc=" + enc + "&_dc=" + fmt.Sprintf("%d", time.Now().UnixMilli())
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
	req.Header.Add("User-Agent", GetUA("mobile"))
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Add("Accept-Language", "zh-CN,en-US;q=0.9")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "mooc1-api.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")
	for _, cookie := range cache.cookies {
		req.AddCookie(cookie)
	}

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
	//fmt.Println(string(body))
	return string(body), nil
}
