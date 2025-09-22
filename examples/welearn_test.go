package examples

import (
	"fmt"
	"testing"
	"time"

	action "github.com/yatori-dev/yatori-go-core/aggregation/welearn"
	"github.com/yatori-dev/yatori-go-core/api/welearn"
	"github.com/yatori-dev/yatori-go-core/global"
	"github.com/yatori-dev/yatori-go-core/utils"
)

func TestWeLearnLogin(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	user := global.Config.Users[39]
	cache := &welearn.WeLearnUserCache{
		Account:  user.Account,
		Password: user.Password,
	}
	//action.WeLearnCookieLoginAction(cache, "ASP.NET_SessionId=vk5s2gayoes3ndnh2h0lgpyr; _ga=GA1.2.954519684.1758557982; _gid=GA1.2.157205280.1758557982; expandable=-1c; acw_tc=ac11000117585658551191538eee27a4f5dbd114bbc6390b05a3a4ee045eef; .AspNet.Cookies=JzojDdmXcylZPhTZq7fW3YGHA1FpeG4C_34uMBIwJERMENHMWgHiRVQTAffQmYEAwydb8QRlWO3MHzJ_ITGhJWF0h04xGWUtoLTJIhkzqXRhfOb-7DubbtmzYGvRHKKbWrpeZsRXue79nFiexsUsKyspSTmAdVfETNPVzbdlpQAvVtIMoD9UFivV77jF6n7I8V7kwU4T6f-_kfUbzYh3pPoqGjUcmUIR4SXbBH8fHNo7e7E8bx-hRF0X4vevKidjMpquQVLgSbcAsdId-f4-MISJJ2cktF1mvuURd-Y9DSIuS2HPuzRgdDhPtoxI52yms0yNi0x67SFoAsClIpknS8A_R4he-42RvGvlJss68md8a0TphSsWOHrthUQNOWXm8pq5QrAYqIG1Osi9JAIbVaF2y6OROPA6ODDI8HVr0YcxbgS0e5wUME31U7MvCLZ00pQWsv_QoHm3BxepJ3Lt5RBW3_2IVQiDqF6pD2BeyaATOtZbzqnuY00gkiXm4FFlo16UMcXEwUAkqqqpF9l1VcqjM87HWZDd1Nip51mMpJsRnXtb0pVc76w7CZioAlhjVd77fA-vteNTM5ZX93uPgH-CrYjANF1eChJbXH5LSiyXOzYDZQdHJGA7Vj2icJGQ8j4mm_ZDNUfV08GLfxYsqH6Rh_Zq_aYiCiQJiTRr0l71SXYBr0DDmpImx4s5lt4Z343BPwxpGLHJZ42xyqLjQc79zCBB_dCuBwGy4VL_2lgqZ2yrS-wlyBRU2Iu5n_wjzlX_Xpezsbav24eTdMkTbdkZai3fX5Z2_oYpIDrR1ohGhQ4GhG8dbzr8niygVDwK-9TOnmQTWju4zw4qLpZnAPR_q_mc8CHcqCq0ctoUUksxYy1axFTULWUv7e8eFptQvOyUsyOfzixxGNDxoTi4Wq9YzwUa2jNayWEi9meZeCME7hhTrrKwlsnIY-lhNH_qUYjmPCrfDFVHPY7RY-uDKyL7fQDUkdVgEssYnHCncE1YdZ8KnhkDyxgSxB-L3XKukyd8wW2kZ5uy3vlRW2lQpZoNd5w3Swepek_Ib6e05X1Tqr0IIK6uQMSI4GWu6jYTq9TJNpjWoyhuAkV01gN8JWYEpMLuX7gKCRPKJLkXI02Pleh0rzAJLSaQ2ckMVq1Gt8JxyYY_3IsPJkuhtex_IKn8FbihmT-mMNpMa73VgoLOfpGHXfGS6lPEkped_BHvlWww6sptjqOgRFslPzXnMxWQxxWiBpowznsMSdLXjOQ33jkkjxX2TnQhxT1Uxtn6A4zp91WUXkDnSDXKawMlN-8r3-RPygct9BeGUVgT3Wr2lGoMf3WgYB8_rVkTavwiiNFjFaZxQg2jpOpjFTIHC19nOeB4ickHhia7HG1WRu7XGyQ95JVWS_iCPLijcIAHyVPSKCOCUqUnxXbZBHkftmu3M3x41hOKsucf6SWL6JkC3EYd0jtpI343yUiZgFi8iKa2vVL9zGGKaY8dL1Af7ALqb5XeC2MMc95v_ehBgdXfSlKLjo8MktEYJlPK5WUqBbN2lIms4mgB9bPzoiQUC5S9eNVf8--8jptpz1o6ym66xPtMa4yLcaDV90OeU2AemX8rOKsT-dWNnecDm689yUMkNx2Kjd7ftQ8S1-r3x11HhN91MZ0VMmg_tnU-xMQ0eqI_901jeeWwjqWCSqrNOAh-ObTFZlwxFa9mAsfIabcTgxNKTMQNNicFxUVPwMVA2N1Uk7UBxoHgISNxDRvdrl9Pg8sXKf69RqM5AaCdnjdv8FryIuMWpUYQHKzKuhMMdJdTLjw2KkBA1DPR5FQ1931C9U7v3LC7YIxJfLc34iYo3ZJMKueRwMVVFbYUS6o0FKq4PyfffgrLLZUGI6S_8xQgHaipIKcwu8gSVEDrflBHwyKj3xNmAd6WWJRuB933N3nVWlhPXgYgK1secYdIeHfpPIxdvm-i55CLEBEhekgAIBxSKJrROS7loLW0LO5n1tRWYYEYHI173Wxr9vzuG4Ftm-mvo5_FggVGQlsnoWaneHpoSfQDgMylFsggRnnA1bgYUTOjhVRXaRqqlOucvX3_ex2nbeyO1catkpt1RJ3K8RnfECWxCEycyxMgq_hb_MWmdlvGDHSCr9xVvhOJzM3GygyPsT4A5QkA3qRKNgJDsYS44DYuOHi1RMYxIsQhIOqtq3dBwLkaU9RkaY5II8gpCE6l3RxFzYwih3CYjaJ10x-tmfHeEwSMwG2_8EMgW4wCRSaH6htRbAmF8Q0kri4jSeZxZLG87pYuB3usozDtuQm9GHSDLjsulzoMeYK-fPw4y1rdcPH2qTp-apzNk5o2oMMdJTqksXAfMPdnBcy7rWf8Fkd22fe9xEcxCKU3KplzBwCzVF1OheqaiahQ3mpcqWN4fRD3a2wQv7KMZ4jIIyUrgp-6xRHx_0i0dq64SZaqQuOfJCuuYIHvGksn48LWZjZzei8atGRwZdm__Ugvxyjf1jzQJNgVJQSYqSP2Jet7GT9rzL7ZLQgZAnEzsPXxLilZUBtVZMftsBZcxa6c1K70sIITFAL0fkOqQyNp3k2gMcHAFrI2QHEbaZe3LdR162bBmo6jhUUTdqPGKAFULelalI-nr3AuBBgo6oyc4cdZcpP0s0NX-Wo8qX3wS2RQoJ-8W0bYStaO_f6KQuJ1COVS5m6QBg39RDPa1BJphk3jAwrxbfD8JuwXdRTpBKQv0BE9VzF1Nrmo9p7BdjOhPT4oFQOqBNYbBDWW-ImbnVUNv-Y6UcEqdcHIaE9Opg; area=dbG; _ga_PNJRS2N8S4=GS2.2.s1758563308$o3$g1$t1758567186$j21$l0$h0")
	err := action.WeLearnLoginAction(cache)
	if err != nil {
		t.Error(err)
	}
	courseList, err := action.WeLearnPullCourseListAction(cache)
	if err != nil {
		t.Error(err)
	}
	for _, course := range courseList {
		chapters, err1 := action.WeLearnPullCourseChapterAction(cache, course)
		if err1 != nil {
			t.Error(err1)
		}
		for _, chapter := range chapters {
			points, err2 := action.WeLearnPullChapterPointAction(cache, course, chapter)
			if err2 != nil {
				t.Error(err2)
			}
			for _, point := range points {
				fmt.Println(point)
				//action.WeLearnCompletePointAction(cache, course, point)
				StudyTime(cache, course, point)
			}
		}
	}
}

// 刷学习时长
func StudyTime(cache *welearn.WeLearnUserCache, course action.WeLearnCourse, point action.WeLearnPoint) {
	//completationStatus,
	_, progressMeasure, sessionTime, totalTime, scaled, err := action.WeLearnSubmitStudyTimeAction(cache, course, point)
	if err != nil {
		fmt.Println(err)
	}
	endTime := 300
	//比阈值大就直接返回
	if totalTime > endTime {
		return
	}
	for {
		api, err1 := cache.KeepPointSessionPlan1Api(course.Cid, point.Id, course.Uid, course.ClassId, sessionTime, totalTime, 3, nil)
		if err1 != nil {
			fmt.Println(err1)
		}
		fmt.Println(totalTime, api)

		if sessionTime >= endTime {
			break
		}
		sessionTime += 60
		totalTime += 60
		time.Sleep(time.Duration(60) * time.Second)
	}
	submitApi2, err2 := cache.SubmitStudyPlan2Api(course.Cid, point.Id, course.Uid, scaled, course.ClassId, progressMeasure, "completed", 3, nil)
	if err2 != nil {
		fmt.Println(err2)
	}
	fmt.Println(submitApi2)

}
