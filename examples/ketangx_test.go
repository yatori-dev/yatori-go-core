package examples

import (
	"fmt"
	"testing"

	action "github.com/yatori-dev/yatori-go-core/aggregation/ketangx"
	"github.com/yatori-dev/yatori-go-core/api/ketangx"
	"github.com/yatori-dev/yatori-go-core/global"
	"github.com/yatori-dev/yatori-go-core/utils"
)

// 测试登录
func Test_KetangxLogin(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	user := global.Config.Users[31]
	cache := ketangx.KetangxUserCache{
		Account:  user.Account,
		Password: user.Password,
	}

	cache.LoginApi()
	courseAction := action.PullCourseAction(&cache)
	for _, course := range courseAction {
		fmt.Println(course)
	}
}
