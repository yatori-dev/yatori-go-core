package xuexitong

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
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

	client := &http.Client{}
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

//func (cache *XueXiTUserCache) AudioSubmitReport(objectId, clazzId, userId, jobId, duration, otherInfo, courseId string, retry int, lastErr error) (string, error) {
//	if retry < 0 {
//		return "", lastErr
//	}
//	//url := "https://mooc1-api.chaoxing.com/mooc-ans/multimedia/log?objectId=90d77e2b106bfa28dbe0e00bf4e7d3c2&clazzId=134204187&userid=346635955&jobid=176552392921124&duration=1&otherInfo=nodeId_1088037085-cpi_411545273-rt_d-ds_false-ff_1-vt_0-v_5-enc_43bcc41f311ac76e7f3a62175e32f41d&courseId=258101827&dtype=Audio&view=json&playingTime=1&isdrag=4&enc=42165394c2d241f07ca1ff0ee797f84b&_dc=1765524099658"
//	urlStr := "https://mooc1-api.chaoxing.com/mooc-ans/multimedia/log?objectId=" + objectId + "&clazzId=" + clazzId + "&userid=" + userId + "&jobid=" + jobId + "&duration=" + duration + "&otherInfo=" + otherInfo + "&courseId=" + courseId + "&dtype=Audio&view=json&playingTime=1&isdrag=4&enc=42165394c2d241f07ca1ff0ee797f84b&_dc=1765524099658"
//	method := "GET"
//
//	client := &http.Client{}
//	req, err := http.NewRequest(method, urlStr, nil)
//
//	if err != nil {
//		fmt.Println(err)
//		return "", err
//	}
//	req.Header.Add("User-Agent", "Mozilla/5.0 (Linux; Android 8.1.0; MI 5X Build/OPM1.171019.019; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/71.0.3578.99 Mobile Safari/537.36 (schild:ce5175d20950c8ee955fb03246f762da) (device:MI 5X) Language/zh_CN com.chaoxing.mobile/ChaoXingStudy_3_6.7.2_android_phone_10936_311 (@Kalimdor)_76c82452584d47e39ab79aa54ea86554")
//	req.Header.Add("X-Requested-With", "XMLHttpRequest")
//	req.Header.Add("Referer", "https://mooc1-api.chaoxing.com/ananas/modules/audio/index_wap_new.html?v=12025-1128-0958")
//	req.Header.Add("Accept-Language", "zh-CN,en-US;q=0.9")
//	req.Header.Add("Cookie", "fid=10596; _uid=346635955; UID=346635955; xxtenc=f8c84ceb53bc45f40b7d9bfaaa413810; fidsCount=1; _industry=5; sso_role=3; _d=1765522850072; vc3=Ew%2BiOaP1EdjH8zAX6DFZ8Mb75TfoqJhkmBPGzLBirgYbITJpIf%2FoU3TmPEp0Fyx7VZSp1zihtojQOdfHXsTlV57P88FPbnofFqQtChWkqLEpsJN1j14W02iUPpL41%2FIaewuhhPTEpk9Xt3HRaPYYDgAEUyKkHVmlOt8kkKMeddE%3D13cc09c9ed0c87fdfcb38263937e5153; uf=b2d2c93beefa90dc495549838143a13b264677447b1a2384b8cd17c4874b05f58a6d64458f94593c09d0f7e006063d352288078c94f43c22c49d67c0c30ca5043ad701c8b4cc548c0234d89f51c3dccfb0f1a1db51ab43f5fb98ce0e6210c3884a878d0a9a7b05da6103a97f8cd189bceec201d6ac0516c3fcd9e6069140d14ef8513768fe2f921d26f3eaec88c8ef25da9735baa04d8d5fce71fc6e59483dd39b16e3a709730613350421afcd29fbcc640c6e56482cdf12e9fdc681bdf07734; cx_p_token=d17169f0c16b7e73e3b8728ad8d530ff; p_auth_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOiIzNDY2MzU5NTUiLCJsb2dpblRpbWUiOjE3NjU1MjI4NTAwNzMsImV4cCI6MTc2NjEyNzY1MH0.m_tTUBOZMsllf-hXlqzGDnm-HJ63N2R8gZ4_TjTDN4g; DSSTASH_LOG=C_38-UN_10038-US_346635955-T_1765522850074; sso_t=1765522850072; sso_v=19ceff89d5dbd7e65c6c8bcfc06b2722; KI4SO_SERVER_EC=RERFSWdRQWdsckJiQXZ5ZmdkWW10b244Qm1NV0laT3hRa1VSeHFjMk0rZG1TME55eG5ya2NpR3J2%0ATW5ZQXhUK0RtU2djWlRjU2NTVQp1NG13bWtscWxaVWNOQlhLbEhtUzROMzEyZ0l3cWIwN3dlclJt%0ARDFtdWE0OVd4bXBHSTZoZFFXNy9qQlRKb2wzY1V2R0dNNjFTRWxPCjhKbkRyZHlUQjNPT1pld0pz%0ANjhyZFR3TFlKaDViZk5OU3pNajNvY29hcU12bVBycExsckV6TWJLYkEvdFhVaTgwMTYzRHRKZUd2%0ARUgKaW55cFE3ZW1aNW9oUGRsVWp6SHVORHFmVXE1ZFdlRXMxaUw1L05DNHhqSllPL3Q1STBLbUxW%0ASTBDK0ZjdUlETU1FSFVXTEJGeWpmQQpVbGo3MVc4K1F5STI1cFFaTTN1VGh6VmJrblFqRXNucjB4%0ANmxoQnNwdjJGNkhQcXcvdE5QWGhidVBpWmIxeEIvZ1F1NCtMaUF6REJCCjFGaXdSY28zd0ZKWSs5%0AVnZQa01pTnVhVUdkS0Y0RlI4bFpQcy9nQ2dHcHc2MTVQandySXhEY1BUSGIxMkpkUEN5VUxVczBk%0AYjVhYWYKVGVrPT9hcHBJZD0xJmtleUlkPTE%3D; _tid=300631019; sso_puid=346635955; k8s=1765522955.933.25603.292680; jrose=09B33A7835F552DC78B95D92E381FA22.mooc-3229109492-6tp9v; route=2fe558bdb0a1aea656e6ca70ad0cad20")
//	req.Header.Add("Accept", "*/*")
//	req.Header.Add("Host", "mooc1-api.chaoxing.com")
//	req.Header.Add("Connection", "keep-alive")
//
//	res, err := client.Do(req)
//	if err != nil {
//		fmt.Println(err)
//		return "", err
//	}
//	defer res.Body.Close()
//
//	body, err := ioutil.ReadAll(res.Body)
//	if err != nil {
//		fmt.Println(err)
//		return "", err
//	}
//	//fmt.Println(string(body))
//	return string(body), nil
//}
