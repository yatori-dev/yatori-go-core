package examples

import (
	"fmt"
	"testing"
	"time"

	cqie "github.com/yatori-dev/yatori-go-core/aggregation/cqie"
	cqieApi "github.com/yatori-dev/yatori-go-core/api/cqie"
	"github.com/yatori-dev/yatori-go-core/global"
	"github.com/yatori-dev/yatori-go-core/utils"
)

// 测试加密函数
func TestCqieEncrypted(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	users := global.Config.Users[4]
	// 调用函数进行加密
	accEncrypted := utils.CqieEncrypt(users.Account)
	passEncrypted := utils.CqieEncrypt(users.Password)
	// 输出加密后的数据
	fmt.Printf("Encrypted data: %x\n", accEncrypted)
	fmt.Printf("Encrypted data: %x\n", passEncrypted)
}

// TestCqieLogin 登录测试函数
func TestCqieLogin(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	users := global.Config.Users[4]
	cache := cqieApi.CqieUserCache{Account: users.Account, Password: users.Password}
	cqie.CqieLoginAction(&cache)
}

// 测试拉取课程
func TestCqiePullCourse(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	users := global.Config.Users[4]

	cache := cqieApi.CqieUserCache{Account: users.Account, Password: users.Password} //构建用户
	cqie.CqieLoginAction(&cache)                                                     //登录
	courseList, _ := cqie.CqiePullCourseListAction(&cache)                           //拉取课程列表
	for _, course := range courseList {
		cqie.PullCourseVideoListAction(&cache, &course)
	}
}

// 测试拉取课程所有视屏
func TestCqiePullCourseVideos(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	users := global.Config.Users[4]

	cache := cqieApi.CqieUserCache{Account: users.Account, Password: users.Password} //构建用户
	cqie.CqieLoginAction(&cache)                                                     //登录
	courseList, _ := cqie.CqiePullCourseListAction(&cache)                           //拉取课程列表
	for _, course := range courseList {
		videos, err := cqie.PullCourseVideoListAction(&cache, &course)
		if err != nil {
			panic(err)
		}
		fmt.Println(videos)
	}
}

// 测试刷视频-常规计时
func TestCqieVideosBrush(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	users := global.Config.Users[4]

	cache := cqieApi.CqieUserCache{Account: users.Account, Password: users.Password} //构建用户
	//token := "eyJhbGciOiJIUzUxMiJ9.eyIwIjoiMSIsInVzZXJfaWQiOiJiODc5N2FkNjdhMGNmZDk2N2ViNGJhOWM4ODBkOWY5MCIsImFwcElkIjoiMjAyNDEyMDIwMTA3NTY5MTY4MiIsInVzZXJfa2V5IjoiYmU5ZDFjZjAtMWExNi00ZTUxLWE0YTgtYTQwYzRhNDA0ODFjIiwidXNlcm5hbWUiOiLlrovlhYPlhbUifQ.Z5on9z6u2kODw737WpIHwcQQr1G1GVMAYcwivh0zmdYBOx9i9shiLJTS8cwQpPLL9RJm2rCYzD5LnovO2nRQxQ"
	//cqie.CqieLoginTokenAction(&cache, token)
	cqie.CqieLoginAction(&cache)
	//登录
	courseList, _ := cqie.CqiePullCourseListAction(&cache) //拉取课程列表
	for _, course := range courseList {
		//videos, err := cqie.PullCourseVideoListAction(&cache, &course)
		videos, err := cqie.PullCourseVideoListAndProgress(&cache, &course)
		fmt.Println("正在学习课程：" + course.CourseName)
		if err != nil {
			panic(err)
		}
		for _, video := range videos {
			cqie.PullCourseVideoListAction(&cache, &course) //每刷一次课就拉一次视屏
			if video.VideoName == "303-修改表结构" {
				fmt.Println("断点")
			}
			nowTime := time.Now()
			if video.StudyTime >= video.TimeLength {
				fmt.Println(video.VideoName, "刷课完毕")
				continue
			}
			startPos := video.StudyTime
			stopPos := video.StudyTime
			maxPos := video.StudyTime
			err = cqie.SaveVideoStudyTimeAction(&cache, &video, startPos, stopPos) //每次刷课前都得先获取一遍，因为要获取学习分配的id
			fmt.Println("正在学习视屏：" + video.VideoName)
			for {
				if maxPos >= video.TimeLength+3 { //+3是为了防止漏时
					startPos = video.TimeLength
					stopPos = video.TimeLength
					maxPos = video.TimeLength
					break
				}
				if stopPos >= maxPos {
					maxPos = startPos + 3
				}
				fmt.Println(startPos, stopPos, maxPos)
				err := cqie.SubmitStudyTimeAction(&cache, &video, nowTime, startPos, stopPos, maxPos)
				if err != nil {
					fmt.Println(err)
				}
				startPos = startPos + 3
				stopPos = stopPos + 3
				time.Sleep(100 * time.Millisecond)
			}
			err = cqie.SaveVideoStudyTimeAction(&cache, &video, startPos, stopPos) //学完之后保存学习点
			if err != nil {
				panic(err)
			}
		}
	}
}

// 秒刷版本
func TestCqieVideosBrushFast(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	users := global.Config.Users[69]

	cache := cqieApi.CqieUserCache{Account: users.Account, Password: users.Password} //构建用户
	//cqie.CqieLoginAction(&cache)
	token := "eyJhbGciOiJIUzUxMiJ9.eyIwIjoiMSIsInVzZXJfaWQiOiI1MTY4MmNjMGVhMjk3MWJiNjI3Nzg5NTQ5NzJmNjYxMyIsImFwcElkIjoiMjAyNTEyMDQwMTc0MjY5ODMxNiIsInVzZXJfa2V5IjoiOGNhZmMzNzAtN2Y2ZC00OWQyLTgzZTctNTcyYTY1MWYxODA2IiwidXNlcm5hbWUiOiLlrovnqIvplKYifQ.1gYFFMTq7dJRwPVXVEIXthNRL3YQAuaokrkrZUk6A3ppuR4Azxm_6VeHVGYABmiGkc17Lc6JuNJbyqsDPFGQbg"
	cqie.CqieLoginTokenAction(&cache, token)               //登录
	courseList, _ := cqie.CqiePullCourseListAction(&cache) //拉取课程列表
	for _, course := range courseList {
		if course.CourseName != "数据结构与算法" {
			continue
		}
		//videos, err := cqie.PullCourseVideoListAction(&cache, &course)
		videos, err := cqie.PullCourseVideoListAndProgress(&cache, &course)
		fmt.Println("正在学习课程：" + course.CourseName)
		if err != nil {
			panic(err)
		}
		for _, video := range videos {
			//if !strings.Contains(video.VideoName, "909") {
			//	//fmt.Println("断点")
			//	continue
			//}
			cqie.PullCourseVideoListAction(&cache, &course) //每刷一次课就拉一次视屏
			nowTime := time.Now()
			//if video.StudyTime >= video.TimeLength {
			//	fmt.Println(video.VideoName, "刷课完毕")
			//	continue
			//}
			startPos := video.TimeLength
			stopPos := video.TimeLength
			maxPos := video.TimeLength + 3
			err = cqie.SaveVideoStudyTimeAction(&cache, &video, startPos, stopPos) //每次刷课前都得先获取一遍，因为要获取学习分配的id
			fmt.Println("正在学习视屏：" + video.VideoName)

			err := cqie.SubmitStudyTimeAction(&cache, &video, nowTime, startPos, stopPos, maxPos)

			err = cqie.SaveVideoStudyTimeAction(&cache, &video, startPos, stopPos-3) //学完之后保存学习点
			if err != nil {
				panic(err)
			}
		}
	}
}

// TestCourse 用于测试CQIE视屏刷课
func TestCourse(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	users := global.Config.Users[4]
	cache := cqieApi.CqieUserCache{Account: users.Account, Password: users.Password}
	cqie.CqieLoginAction(&cache)

	startPos := 0
	stopPos := 3
	maxPos := 3
	for {
		if stopPos >= maxPos {
			maxPos = startPos + 3
		}
		fmt.Println(startPos, stopPos, maxPos)
		// cqieApi.SubmitStudyTimeApi(&cache,"","","","", startPos, stopPos, maxPos)
		startPos = startPos + 3
		stopPos = stopPos + 3
		time.Sleep(3 * time.Second)
	}
}
