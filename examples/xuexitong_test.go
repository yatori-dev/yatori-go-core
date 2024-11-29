package examples

import (
	"fmt"
	"github.com/Yatori-Dev/yatori-go-core/aggregation/xuexitong"
	xuexitongApi "github.com/Yatori-Dev/yatori-go-core/api/xuexitong"
	"github.com/Yatori-Dev/yatori-go-core/global"
	"github.com/Yatori-Dev/yatori-go-core/utils"
	"log"
	"strconv"
	"testing"
)

// TestLoginXueXiTo 测试学习通登录以及课程数据拉取
func TestLoginXueXiTo(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	user := global.Config.Users[1]
	userCache := xuexitongApi.XueXiTUserCache{
		Name:     user.Account,
		Password: user.Password,
	}
	err := xuexitong.XueXiTLoginAction(&userCache)
	if err != nil {
		log.Fatal(err)
	}
	//拉取课程列表并打印
	xuexitong.XueXiTPullCourseAction(&userCache)
}

// 测试学习通单课程详情
func TestCourseDetailXueXiTo(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	user := global.Config.Users[1]
	userCache := xuexitongApi.XueXiTUserCache{
		Name:     user.Account,
		Password: user.Password,
	}
	err := xuexitong.XueXiTLoginAction(&userCache)
	if err != nil {
		log.Fatal(err)
	}
	action, err := xuexitong.XueXiTCourseDetailForCourseIdAction(&userCache, "261619055656961")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(action)
}

// TestCourseXueXiToChapter 用于测试学习通对应课程章节信息拉取
func TestCourseXueXiToChapter(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	user := global.Config.Users[1]
	userCache := xuexitongApi.XueXiTUserCache{
		Name:     user.Account,
		Password: user.Password,
	}

	err := xuexitong.XueXiTLoginAction(&userCache)
	if err != nil {
		log.Fatal(err)
	}
	//拉取对应课程信息
	course, err := xuexitong.XueXiTCourseDetailForCourseIdAction(&userCache, "261619055656961")
	fmt.Println("name:" + course.CourseName)
	fmt.Println("courseID:" + course.CourseId)
	//拉取对应课程的章节信息
	cpi, _ := strconv.Atoi(course.Cpi)
	key, _ := strconv.Atoi(course.ClassId)
	chapter, err := xuexitong.PullCourseChapterAction(&userCache, cpi, key)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("chatid:" + chapter.ChatID)
	for _, item := range chapter.Knowledge {
		fmt.Println("ID:" + strconv.Itoa(item.ID))
		fmt.Println("章节名称:" + item.Name)
		fmt.Println("标签:" + item.Label)
		fmt.Println("层级" + strconv.Itoa(item.Layer))
	}
}

func TestXueXiToChapterPoint(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	user := global.Config.Users[1]
	userCache := xuexitongApi.XueXiTUserCache{
		Name:     user.Account,
		Password: user.Password,
	}

	err := xuexitong.XueXiTLoginAction(&userCache)
	if err != nil {
		log.Fatal(err)
	}
	course, err := xuexitong.XueXiTCourseDetailForCourseIdAction(&userCache, "261619055656961")
	if err != nil {
		log.Fatal(err)
	}
	cpi, _ := strconv.Atoi(course.Cpi)
	key, _ := strconv.Atoi(course.ClassId)
	action, _ := xuexitong.PullCourseChapterAction(&userCache, cpi, key)
	var nodes []int
	for _, item := range action.Knowledge {
		nodes = append(nodes, item.ID)
	}
	userId, _ := strconv.Atoi(course.UserId)
	courseId, _ := strconv.Atoi(course.CourseId)
	fmt.Println(nodes)
	fmt.Println(key)
	fmt.Println(userId)
	fmt.Println(cpi)
	fmt.Println(courseId)
	body, err := userCache.FetchChapterPointStatus(nodes, key, userId, cpi, courseId)
	fmt.Println(body)
}
