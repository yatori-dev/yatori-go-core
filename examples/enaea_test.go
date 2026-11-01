package examples

import (
	"fmt"
	"os"
	"testing"
	time2 "time"

	"github.com/yatori-dev/yatori-go-core/v3/global"
	"github.com/yatori-dev/yatori-go-core/v3/internal/aggregation/enaea"
	enaeaApi "github.com/yatori-dev/yatori-go-core/v3/internal/api/enaea"
	"github.com/yatori-dev/yatori-go-core/v3/utils"
)

// 测试学习公社登录
func TestENAEALogin(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	users := global.Config.Users[9]
	cache := enaeaApi.EnaeaUserCache{Account: users.Account, Password: users.Password}
	_, err := enaea.EnaeaLoginAction(&cache)
	if err != nil {
		t.Error(err)
	}
}

// 测试学习公社拉取所有学习项目
func TestENAEAPullProject(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	users := global.Config.Users[3]
	cache := enaeaApi.EnaeaUserCache{Account: users.Account, Password: users.Password}
	_, err := enaea.EnaeaLoginAction(&cache)
	if err != nil {
		t.Error(err)
	}
	action, err := enaea.ProjectListAction(&cache)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(action)
}

// 测试学习公社根据学习项目拉取所对应的课程
func TestENAEAPullProjectForCourse(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	users := global.Config.Users[3]
	cache := enaeaApi.EnaeaUserCache{Account: users.Account, Password: users.Password}
	_, err := enaea.EnaeaLoginAction(&cache)
	if err != nil {
		t.Error(err)
	}
	action, err := enaea.ProjectListAction(&cache)
	if err != nil {
		t.Error(err)
	}
	enaea.CourseListAction(&cache, action[0].CircleId)
}

// 测试学习公社根据学习项目拉取所对应的视频并学习
func TestENAEAPullProjectForVideo(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	users := global.Config.Users[19]
	cache := enaeaApi.EnaeaUserCache{Account: users.Account, Password: users.Password}
	_, err := enaea.EnaeaLoginAction(&cache)
	if err != nil {
		t.Error(err)
	}
	action, err := enaea.ProjectListAction(&cache)
	if err != nil {
		t.Error(err)
	}
	courseList, err := enaea.CourseListAction(&cache, action[0].CircleId)
	for _, v := range courseList {
		videoList, err := enaea.VideoListAction(&cache, &v)
		if err != nil {
			t.Error(err)
		}
		for _, video := range videoList {
			err := enaea.StatisticTicForCCVideAction(&cache, &video)
			if err != nil {
				t.Error(err)
			}
			//如果学过了，那么跳过
			if video.StudyProgress >= 100 {
				continue
			}
			for {
				//err := enaea.SubmitStudyTimeAction(&cache, &video, time2.Now().UnixMilli(), 0)
				err := enaea.SubmitStudyTimeAction(&cache, &video, 60, 1)
				if err != nil {
					t.Error(err)
				}
				err2 := enaea.StatisticTicForCCVideAction(&cache, &video)
				if err2 != nil {
					t.Error(err)
				}
				if video.StudyProgress >= 100 { //如果学习完毕那么跳过 break
				}
				time2.Sleep(time2.Second * 16)

			}
		}

	}
}

// 测试学习公社暴力模式视频学习
func TestENAEAPullProjectForFastVideo(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	users := global.Config.Users[54]
	cache := enaeaApi.EnaeaUserCache{Account: users.Account, Password: users.Password}
	//_, err := enaea.EnaeaLoginAction(&cache)
	err := enaea.EnaeaCookieLoginAction(&cache, os.Getenv("YATORI_COOKIE"))
	if err != nil {
		t.Error(err)
	}
	action, err := enaea.ProjectListAction(&cache)
	if err != nil {
		t.Error(err)
	}
	courseList, err := enaea.CourseListAction(&cache, action[0].CircleId)
	for _, v := range courseList {
		videoList, err := enaea.VideoListAction(&cache, &v)
		if err != nil {
			t.Error(err)
		}
		//if v.CourseTitle != "人民教育家陶行知的爱国故事——我是一个中国人，要为中国做出一些贡献" {
		//	continue
		//}
		for _, video := range videoList {
			err := enaea.StatisticTicForCCVideAction(&cache, &video)
			if err != nil {
				t.Error(err)
			}

			//如果学过了，那么跳过
			//if video.StudyProgress >= 100 {
			//	continue
			//}
			for {
				//err := enaea.SubmitStudyTimeAction(&cache, &video, time2.Now().UnixMilli(), 0)
				err := enaea.SubmitStudyTimeAction(&cache, &video, 20, 1)
				if err != nil {
					t.Error(err)
				}
				err2 := enaea.StatisticTicForCCVideAction(&cache, &video)
				if err2 != nil {
					t.Error(err)
				}
				if video.StudyProgress >= 100 { //如果学习完毕那么跳过
					break
				}
				time2.Sleep(time2.Second * 1)

			}
		}

	}
}
