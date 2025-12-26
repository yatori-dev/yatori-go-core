package examples

import (
	"fmt"
	"testing"
	"time"

	ddddocr "github.com/Changbaiqi/ddddocr-go/utils"
	"github.com/thedevsaddam/gojsonq"
	qsxt "github.com/yatori-dev/yatori-go-core/aggregation/qingshuxuetang"
	"github.com/yatori-dev/yatori-go-core/api/qingshuxuetang"
	"github.com/yatori-dev/yatori-go-core/global"
	"github.com/yatori-dev/yatori-go-core/utils"
)

// 青书学堂登录
func TestQsxtLogin(t *testing.T) {
	utils.YatoriCoreInit()
	setup()
	user := global.Config.Users[74]
	cache := qingshuxuetang.QsxtUserCache{
		Account:  user.Account,
		Password: user.Password,
	}
	//Cookie登录
	action, err2 := qsxt.QsxtLoginAction(&cache)
	if err2 != nil {
		panic(err2)
	}
	fmt.Println(action)
	//qsxt.QsxtCookieLoginAction(&cache)
	courseList, err := qsxt.PullCourseListAction(&cache)
	if err != nil {
		panic(err)
	}
	for _, course := range courseList {
		if course.StudyStatusName != "在修" { //过滤非在修课程
			continue
		}
		if course.CourseName != "药理学(专升本)" {
			continue
		}
		fmt.Println(course)
		//过滤不能学习的课程
		if !course.AllowLearn {
			continue
		}

		//workList, err2 := qsxt.PullWorkListAction(&cache, course)
		//if err2 != nil {
		//	panic(err2)
		//}
		//for _, work := range workList {
		//	action, err3 := qsxt.WriteWorkAction(&cache, work, true)
		//	if err3 != nil {
		//		panic(err3)
		//	}
		//	fmt.Println(action)
		//}
		videoList, err1 := qsxt.PullCourseNodeListAction(&cache, course)
		if err1 != nil {
			panic(err1)
		}
		if course.CoursewareLearnGainScore < course.CoursewareLearnTotalScore {
			for _, node := range videoList {
				fmt.Println(node)
				if node.NodeType == "chapter" {
					continue
				}
				startId, err2 := node.StartStudyTimeAction(&cache)
				if err2 != nil {
					fmt.Println(err2)
				}
				studyTime := 0 //当前累计学习了多久
				maxTime := 600 //最大学习多长时间
				for {
					time.Sleep(60 * time.Second)
					submitResult, err3 := node.SubmitStudyTimeAction(&cache, startId, false)
					if err3 != nil {
						fmt.Println(err3)
					}
					fmt.Println(submitResult)
					studyTime += 60
					if studyTime >= maxTime {
						break
					}
				}
				submitResult, err3 := node.SubmitStudyTimeAction(&cache, startId, true)
				if err3 != nil {
					fmt.Println(err3)
				}
				fmt.Println(submitResult)
			}
		}
		materialList, err2 := qsxt.PullCourseMaterialListAction(&cache, course)
		if err2 != nil {
			panic(err2)
		}
		if course.CourseMaterialsLearnGainCore < course.CourseMaterialsLearnTotalCore {
			for _, node := range materialList {
				startId, err2 := node.StartStudyTimeAction(&cache)
				if err2 != nil {
					fmt.Println(err2)
				}
				studyTime := 0 //当前累计学习了多久
				for {
					if course.CourseMaterialsLearnGainCore >= course.CourseMaterialsLearnTotalCore {
						break
					}
					time.Sleep(60 * time.Second)
					submitResult, err3 := node.SubmitStudyTimeAction(&cache, startId, false)
					if err3 != nil {
						fmt.Println(err3)
					}
					fmt.Println(submitResult)
					studyTime += 60

					qsxt.UpdateCourseScore(&cache, &course)
				}
				submitResult, err3 := node.SubmitStudyTimeAction(&cache, startId, true)
				if err3 != nil {
					fmt.Println(err3)
				}
				fmt.Println(submitResult)
			}
		}
	}
}

func TestQsxtPullCodeList(t *testing.T) {
	utils.YatoriCoreInit()
	setup()
	user := global.Config.Users[60]
	cache := qingshuxuetang.QsxtUserCache{
		Account:  user.Account,
		Password: user.Password,
	}
	for i := 0; i < 50; i++ {
		pullCodeJson, err := cache.QsxtPhoneValidationCodeApi(3, nil)
		if err != nil {
			panic(err)
		}
		if codeImgBase64, ok := gojsonq.New().JSONString(pullCodeJson).Find("data.code").(string); ok {
			image, err := utils.Base64ToImage(codeImgBase64)
			if err != nil {
				panic(err)
			}
			utils.SaveImageAsJPEG(image, "./qsxt_code/qsxt_code_"+fmt.Sprintf("%d", i)+".png")
			verification := ddddocr.AutoOCRVerification(image)
			fmt.Println(verification)
		}
	}

}
