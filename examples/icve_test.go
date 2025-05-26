package examples

import (
	"github.com/yatori-dev/yatori-go-core/api/icve"
	"github.com/yatori-dev/yatori-go-core/global"
	"github.com/yatori-dev/yatori-go-core/utils"
	"testing"
)

// 测试登录
func TestIcveLogin(t *testing.T) {
	utils.YatoriCoreInit()
	setup()
	user := global.Config.Users[16]
	userCache := icve.IcveUserCache{
		Account:  user.Account,
		Password: user.Password,
	}
	userCache.IcveLoginApi()
}

// 测试拉取课程
func TestIcveCourseList(t *testing.T) {
	utils.YatoriCoreInit()
	setup()
	user := global.Config.Users[16]
	userCache := icve.IcveUserCache{
		Account:  user.Account,
		Password: user.Password,
	}
	userCache.IcveLoginApi()
	userCache.CourseListApi()
}
