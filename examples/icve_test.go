package examples

import (
	"fmt"
	log2 "log"
	"testing"

	action "github.com/yatori-dev/yatori-go-core/aggregation/icve"
	"github.com/yatori-dev/yatori-go-core/api/icve"
	"github.com/yatori-dev/yatori-go-core/global"
	"github.com/yatori-dev/yatori-go-core/utils"
)

// 测试登录
func TestIcveLogin(t *testing.T) {
	utils.YatoriCoreInit()
	setup()
	user := global.Config.Users[44]
	cache := icve.IcveUserCache{
		Account:  user.Account,
		Password: user.Password,
	}
	//userCache.IcveLoginApi()
	err := action.IcveLoginAction(&cache)
	if err != nil {
		fmt.Println(err)
	}

}

// 测试拉取课程
func TestIcveCourseList(t *testing.T) {
	utils.YatoriCoreInit()
	setup()
	user := global.Config.Users[45]
	cache := icve.IcveUserCache{
		Account:  user.Account,
		Password: user.Password,
	}
	err := action.IcveLoginAction(&cache)
	if err != nil {
		fmt.Println(err)
	}
	courseList, err := action.PullZYKCourseAction(&cache)
	if err != nil {
		fmt.Println(err)
	}
	for _, course := range courseList {
		fmt.Println(course)
		nodeList, err1 := action.PullZYKCourseNodeAction(&cache, course)
		if err1 != nil {
			fmt.Println(err1)
		}
		for _, node := range nodeList {
			if node.Speed >= 100 {
				continue
			}
			fmt.Println(node)
			result, err2 := action.SubmitZYKStudyTimeAction(&cache, node)
			if err2 != nil {
				fmt.Println(err2)
			}
			log2.Printf("(%s)学习状态：%s", node.Name, result)
		}
	}
	//userCache.CourseListApi()
}
