package examples

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
)

func Test_PullVerData(t *testing.T) {
	PullVerData()
}

func PullVerData() {

	url := "https://turing.captcha.qcloud.com/cap_union_prehandle?aid=196632980&protocol=https&accver=1&showtype=popup&ua=TW96aWxsYS81LjAgKFdpbmRvd3MgTlQgMTAuMDsgV2luNjQ7IHg2NCkgQXBwbGVXZWJLaXQvNTM3LjM2IChLSFRNTCwgbGlrZSBHZWNrbykgQ2hyb21lLzE0MC4wLjAuMCBTYWZhcmkvNTM3LjM2IEVkZy8xNDAuMC4wLjA%253D&noheader=1&fb=1&aged=0&enableAged=0&enableDarkMode=0&grayscale=1&clientype=2&cap_cd=&uid=&lang=zh-cn&entry_url=https%253A%252F%252Fsso.icve.com.cn%252Fsso%252Fauth&elder_captcha=0&js=%252FtgJCap.977ef8c3.js&login_appid=&wb=1&subsid=3&callback=_aq_55600&sess="
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36 Edg/140.0.0.0")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "turing.captcha.qcloud.com")
	req.Header.Add("Connection", "keep-alive")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}
