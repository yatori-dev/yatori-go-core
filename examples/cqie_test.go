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
