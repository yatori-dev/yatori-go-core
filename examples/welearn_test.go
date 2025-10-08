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
	user := global.Config.Users[46]
	cache := &welearn.WeLearnUserCache{
		Account:  user.Account,
		Password: user.Password,
	}
	action.WeLearnCookieLoginAction(cache, "_ga=GA1.2.954519684.1758557982; ASP.NET_SessionId=mn23hb3qnyphakadsv3rm1sd; area=dbB; _gid=GA1.2.1159277000.1759937553; .AspNet.Cookies=qlTwLZKsTCeiGlUhN14dLNCzkzH8xQadU4fMtWhlVHKMSbyBOdeH-tC1zdggAM1_kAHf63USMYZ7Iwgpk3A6GKjq1_SSU1qjbe2LeHU0261YfPIe-nhbXJ9LXnzjNXsAKppTqDHfnPA6H3M3PhJip87XotDsK6iJ0O6IK28T70j1I7BKWNgEBdQQMI5zAyq6uatd9Vah_VRDcuC7IqcgIyWrJKnyXl_RXk6yEZOjStdRG_SnieXQOCLrk8MtegXG1wqaIrPG6-FkIaqLbH8K728JH7WSINnabUSf5rVPPwc_Efp8c5aNcdSMD8kFqC5ZTKxf7rxcEq1RSafOUHyHaNnppgwQWcqCssZNLUbaJq3KOz7wWD49pkyhODBROt1fKWa6VSXb0b3gXRWAvdGYy3QPyWBAIgr5Y6WFoBIwnsKJgzhqU5jMaAznCsx-aaht5_szizRcwMKwuTkORFQ2sC6o-oE0XPQ67zBnNROQ_nySY4z2kLw5iQtgFbbWeKCLl4XhNg8BAF6tkn3RDphQ3hLgJFW-wusNJqR4dpwX9rvlBqgvGqJwkSBdad81A4M5JyIKyRO5cLt_4rbe4Gg4tqY3QuutnQXdwHPaaIqG6qsExWBdyxmKH-xoSUtWgvC947QT_4pPOTYm-kZMbY99SS24vXIrGSdlvDb8Jd4V2XYvKMedm65qgOGZViKiz-tdIQbM0Yq4WiqXot1KiTsb67u-nGzRtHDium_i_XvKMJeWECsq68i2jSITcenlRwgpoRwjyK_kW1Fce-gzjpYl97cGhgDgUszag5nNXC2Nj_6-MuCEs6n4Fj39EIp_BSl39ohIxih5MqUw1yraqe4PnvLjkfeJh5ailB2eFJLPfxc-vJLthsIycEKtSXSD8MnUVBwMJ7pIzyaZDz-0nAlwwPLbBP6qNaHaSMXswgq6Wlzy2IYLxlJEB2YwXXU_RTJSch_-iKjmUJ8D8NcJK8SpCbozyFZkHpg7ve7VIDLZL0EO9ABAaSC7v6LN35SF4KmdWvDfTmYgAoX9v2s7SgrzqA19kMHvqUDLU87dC8K0xLjR4oavPJ0oBrcUe5wcAh5VeQsfrjUmKWZQO7OkBEiU9k9ZTpD8izBwn3cxW1Z4PpqFYbNOoUKqetMYEvTWS76aEHQhQ1ssUIhCulkq_P-K8GEgMbsffAy_fkV8Tmrl-6s3JQzZROq7jh16ZoByz5BFbN-EjJc-h0FFJLjXSsdRqt7Qq0XvS6j-OeXnxEwxRODLur3_x2hK06Kh3EksM1NX9t9hRbVI2MlFMmrr7-2wpgNDX4NDxND0BJ80ZXFp-5NLflX1ShynKhQdQD_y-A09fJvgN9UYWeQHhBmO_eJKgmsCeBK_rK2nikIzkKKYJqDnpLmOOmH0qf8F6tnCPyis5t7Z2n0L3S0XUltYMa-Bcz4UZmm0bIhxLUc_r3rBt8ScS9e62FHjNPhLvF4aatBgkku89i_XXen19v-pdnUQdh1sN88skeCDaupSb2Guu-saVzeJxbbs-5qVSojdMNuWSqkqh-bW5bl34bWLu9Nu6SmnhrMgUhZtB02cDyMQR3i13bb_p2dHSb_OGAWP_SLD-Wc5-RNXzdGnieEsR_N7G8RHzY0lJ0lp82oYAZGl4zSR02Y4yGLYECtuO6x2HiL-1zv5knIv0p95CqA2YrW-xQlawImDCLMsVCZ1HjYdoLxp2gLR_elbauH5AOWyyocvSp2BenNaeP-QcMsN5QDpxPofIQ2BeVKgB8ahH3u5aU4WlmRF0jVJJohYGaqqRAkrpC0HdlV7HdCosSSNtZnD3iQRlO8GaqYWyvcNzBhU-BsDCirNys27yVUHsIN0rhdwDVa-3u2BxJ3TH-znUr6b4bRWiRVYtICxjagMZ3gxY3PyjXbfGRMa1qpw0PkV-9cAvzgSyF5vCP_RcWoqp5OCTur8Joxhp3m5RnbnB0visAmWxsaBlGHt_qXC0lBvaLROyursVOWFgXy3gy81amIZxjH5Nol2kzNqfAr8DTUA-XxNIRUTMgzbRvnU_gSJHl4W0wQ_BVq7ItspRLExDWn7WnXXJdOUqkmOr9zkTUkErJa4OvCWXulVVveWa13XEBJzr8wlm10omJZVw5CAUI1wSmNjr9y-GJ8dW7HUNWeUY8mYz22V1jwODevgIH2HIcuBoOUGQtoaQaKkmsoYJGhw1qVhz4MmBGLNJRFnM3QF2mda1SQ01_Jr4CAkWqEvN8vZlgG-3uqr-vR5LP8K8mAy21DYq0wimlHCHa_-VsFYdokFBlSyscuPia3w_lQ1SzfGOD_ZgEJHSYmV5sGvwqMeTCdvVQGVjRuEo5SM-PiUtxbjHN8MoP1lhZIScWQCdV2CXe5P6zMZNMSsKBz4gjx7KfoOFlwh-bHhscZpINUJUSztOh0fZ0qiuYENmZ-cbtsk4mHq95WRtDkC9f4AUx9BvK_vxrRwLgoEw-VC_nhw9M1uUlPjkEsvbrzytD-njTx6zTecZfCDIUJkI2G91nvUFIxc_yfrNhXem6gRyRQueYOnHTv4FM0LymPVFo3v2gr2GKq6UzFcrjXitqLpdUBWR97-DGAXqZaQZV_MndN3_PMwrqTSq0NWbTgxRdSmrGrVXUPhDi2xmkYga1VQRYICzTi9izvoun58VqswOdIlesTIaWScNx3wHbXf7d0DIxj6zrSObq8a5BfKx0MgORugvLvcO90zuuWEbZqX9dzeal8fZzQaaiwnyNgUnM1GD7jxdZdi5N5iMtAsnzFybCuUprdrGGnmnHSglfLcEF8fXtTuiveLePXdC909E89U2YHS4YQXViHwij-v876dERzHuQ; acw_tc=1a0c639217599420811763697e666a8fc770a9379cb65d52a5c57f38914ae1; expandable=-1c; _ga_PNJRS2N8S4=GS2.2.s1759940188$o8$g1$t1759942574$j60$l0$h0")
	//err := action.WeLearnLoginAction(cache)
	//if err != nil {
	//	t.Error(err)
	//}
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
			fmt.Println(points)
			for _, point := range points {
				fmt.Println(point)
				action.WeLearnCompletePointAction(cache, course, point)
				//StudyTime(cache, course, point)
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
