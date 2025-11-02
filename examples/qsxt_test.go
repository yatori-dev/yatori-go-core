package examples

import (
	"fmt"
	"testing"
	"time"

	qsxt "github.com/yatori-dev/yatori-go-core/aggregation/qingshuxuetang"
	"github.com/yatori-dev/yatori-go-core/api/qingshuxuetang"
	"github.com/yatori-dev/yatori-go-core/global"
	"github.com/yatori-dev/yatori-go-core/utils"
)

// 青书学堂登录
func TestQsxtLogin(t *testing.T) {
	utils.YatoriCoreInit()
	setup()
	user := global.Config.Users[60]
	cache := qingshuxuetang.QsxtUserCache{
		Account:  user.Account,
		Password: user.Password,
	}
	//Cookie登录
	qsxt.QsxtCookieLoginAction(&cache)
	courseList, err := qsxt.PullCourseListACtion(&cache)
	if err != nil {
		panic(err)
	}
	for _, course := range courseList {
		if course.StudyStatusName != "在修" { //过滤非在修课程
			continue
		}
		if course.CourseName != "先秦两汉散文专题(专升本)" {
			continue
		}
		fmt.Println(course)
		//过滤不能学习的课程
		if !course.AllowLearn {
			continue
		}
		nodeList, err1 := qsxt.PullCourseNodeListAction(&cache, course)
		if err1 != nil {
			panic(err1)
		}
		for _, node := range nodeList {
			fmt.Println(node)
			if node.NodeType == "chapter" {
				continue
			}
			startId, err2 := qsxt.StartStudyTimeAction(&cache, node)
			if err2 != nil {
				fmt.Println(err2)
			}
			studyTime := 0 //当前累计学习了多久
			maxTime := 600 //最大学习多长时间
			for {
				time.Sleep(60 * time.Second)
				submitResult, err3 := qsxt.SubmitStudyTimeAction(&cache, node, startId, false)
				if err3 != nil {
					fmt.Println(err3)
				}
				fmt.Println(submitResult)
				studyTime += 60
				if studyTime >= maxTime {
					break
				}
			}
			submitResult, err3 := qsxt.SubmitStudyTimeAction(&cache, node, startId, true)
			if err3 != nil {
				fmt.Println(err3)
			}
			fmt.Println(submitResult)
		}
	}
}
