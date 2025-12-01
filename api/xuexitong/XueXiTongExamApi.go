package xuexitong

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

// PullExamListHtmlApi 拉取邮箱考试列表
func (cache *XueXiTUserCache) PullExamListHtmlApi(courseId string, classId string, cpi string, retry int, lastErr error) (string, error) {
	if retry < 0 {
		return "", lastErr
	}
	urlStr := "https://mooc1-api.chaoxing.com/mooc-ans/exam/phone/task-list?courseId=" + courseId + "&classId=" + classId + "&cpi=" + cpi
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}
	client := &http.Client{
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return errors.New("太多重定向")
			}

			// 复制 Cookie
			if len(via) > 0 {
				for _, c := range via[0].Cookies() {
					req.AddCookie(c)
				}
			}
			return nil // 允许重定向
		},
	}
	req, err := http.NewRequest("GET", urlStr, nil)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Linux; Android 8.1.0; MI 5X Build/OPM1.171019.019; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/71.0.3578.99 Mobile Safari/537.36 (schild:ce5175d20950c8ee955fb03246f762da) (device:MI 5X) Language/zh_CN com.chaoxing.mobile/ChaoXingStudy_3_6.7.2_android_phone_10936_311 (@Kalimdor)_76c82452584d47e39ab79aa54ea86554")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	req.Header.Add("Upgrade-Insecure-Requests", "1")
	req.Header.Add("accept-language", "zh_CN")
	req.Header.Add("X-Requested-With", "com.chaoxing.mobile")
	req.Header.Add("Cookie", "k8s=1764505505.713.24258.990048; route=1ab934bb3bbdaaef56ce3b0da45c52ed; fid=123402; _uid=410496399; UID=410496399; xxtenc=bac7e84bc401921a0c9bbbc730d6a802; fidsCount=1; _industry=6; sso_role=3; _tid=362220710; sso_puid=410496399; wfwIncode=xw85779; wfwfid=123402; spaceFid=123402; source=num2; wfwEnc=EFD203BB75DE4CE9DDE7348502FBDF88; _d=1764567030436; vc3=ZMa3ZNbJZxzO27Le2nMWqYNHxFC1CsL%2BMiQVphB%2FkiABCazrlQ3v7CpOvpx9ZtMRW%2BzDfNg5kuXRHrQ6H2C8PHF2pc0G4UsTAH%2BcgnAJEMFnbnWYR3a5TTiZP3jlFjh4Bx0d0wzm1iMyoTndqd5DWJsmQFJjwr0b3U7Ckr0uGxQ%3Def0b614ffd3367c1fc06afe55e074c0c; uf=b2d2c93beefa90dc51b37d0932243830be3902fbfb1fc62ce5e0881cbf78ed4a7cc932f0757146e08b64b4d624315739179900d10480c3f3ea4a1670a3a8352fe9295d8c89b08ad0f44425e20f927c6b97cb2aec7f1e5a9afb98ce0e6210c3884a878d0a9a7b05da6103a97f8cd189bccc5d4a9b0d1b6a7a1939dc78d50c971ae2ead16ef2f50a8d7420ea0e76898b41da9735baa04d8d5fce71fc6e59483dd39b16e3a7097306134bf1828e50f6a9b65f822a7404fce8bfe9fdc681bdf07734; cx_p_token=3a75006e93978cdefe29431a1d503240; p_auth_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOiI0MTA0OTYzOTkiLCJsb2dpblRpbWUiOjE3NjQ1NjcwMzA0MzgsImV4cCI6MTc2NTE3MTgzMH0.8Rnwx7-fAHepel9EwqQKf0h48126-AngAhnZqZ5YzB4; DSSTASH_LOG=C_38-UN_2010-US_410496399-T_1764567030438; sso_t=1764567030436; sso_v=586660f4fc7a775c7920a0ff2d6d4c9b; KI4SO_SERVER_EC=RERFSWdRQWdsckJiQXZ5ZmdkWW10bkdSREo2RU9sU1hpRldOcmRNcVkzcGZhejZxVm1pLzNlbDZi%0AZEc0MVhuWkRtU2djWlRjU2NTVQp1NG13bWtscWxkc3ZDODB1b05kYVB2b29OblVmd253N3dlclJt%0ARDFtdWE0OVd4bXBHSTZoVnhLTCswbDhsSW1RZFpncDlWNGgzNGMzCkFZZUlFck8zajRzTVBLcU1C%0AM0RYZEtodUhpeStxWG1paHdJYTByQ3B1NlBmSlpXSEhUb21wRWs4QkdRcFJWalNhcGVra0NzWCtR%0AWlcKUVNwUHdzV3hJbS9rOURPcEJkbXJNMGlFRE5uQTFqQnhsazFqL2oyNHhqSllPL3Q1STBLbUxW%0ASTBDK0ZjdUlETU1FSFVXTEJtMVE4TQp3NFBOeXMyTWVKRHl5VjlBVmkrRFUzMHRYWGlMekVmMGdl%0ASEVsRWZUa2hrSml1TUhNZlV0MUZpMDFWWUlablFjYllWdlhBKzRyWDVHCkhGN08wOEZaS1p2OXdS%0ASWNKVFozajdMRTJwMFJYTDI4c0pyUDR2NVZncFpPY0xVeDdqUTZuMUt1WFNqeGQ0MHluWC9jQ2ZU%0AeTlNZmUKV3VjPT9hcHBJZD0xJmtleUlkPTE%3D; jrose=208D25304E08B788854AC10FCE760EA7.mooc-p4-3965218474-jf1q6")
	req.Header.Add("Host", "mooc1-api.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")

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
