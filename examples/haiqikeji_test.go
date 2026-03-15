package examples

import (
	"fmt"
	"testing"

	zxcpks2 "github.com/yatori-dev/yatori-go-core/aggregation/haiqikeji"
	"github.com/yatori-dev/yatori-go-core/api/haiqikeji"
	"github.com/yatori-dev/yatori-go-core/global"
	"github.com/yatori-dev/yatori-go-core/utils"
)

// 测试登录
func TestHqkjLogin(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	user := global.Config.Users[75]
	cache := haiqikeji.HqkjUserCache{PreUrl: "https://swxy.haiqikeji.com/", Account: user.Account, Password: user.Password}
	zxcpks2.HqkjLoginAction(&cache)
	courseList, err := zxcpks2.HqkjCourseListAction(&cache)
	if err != nil {
		t.Error(err)
	}
	for _, course := range courseList {
		//fmt.Println(course)
		nodeList, err := zxcpks2.HqkjNodeListAction(&cache, course)
		if err != nil {
			t.Error(err)
		}
		for _, node := range nodeList {
			fmt.Println(node)
			zxcpks2.HqkjSubmitFastSutdyTimeAction(&cache, node)
		}
	}
}
