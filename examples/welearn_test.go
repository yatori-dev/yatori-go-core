package examples

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/yatori-dev/yatori-go-core/v3/global"
	action "github.com/yatori-dev/yatori-go-core/v3/internal/aggregation/welearn"
	"github.com/yatori-dev/yatori-go-core/v3/internal/api/welearn"
	"github.com/yatori-dev/yatori-go-core/v3/utils"
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
	action.WeLearnCookieLoginAction(cache, os.Getenv("YATORI_COOKIE"))
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
