package examples

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"testing"
	"time"

	action "github.com/yatori-dev/yatori-go-core/aggregation/ketangx"
	"github.com/yatori-dev/yatori-go-core/api/ketangx"
	"github.com/yatori-dev/yatori-go-core/global"
	"github.com/yatori-dev/yatori-go-core/utils"
)

var pv = "v1.32.0"
var href = "https://www.ketangx.cn/DoAct/Index/d1bdd8da7e094e34a443b34e00aa447f#/section/4fe4c81c30ef4eb3baebb34e00e8471b"

// 获取PID
func getPID() string {
	return fmt.Sprintf("%dX%d", time.Now().UnixMilli(), 1e6*rand.Int()+1e6)
}

// 获取UID
func getUid(vid string) string {
	return vid[:10]
}
func getSign(rtas, pid, vid, pd, r, cT string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(rtas+pid+vid+pd+r+cT)))
}

func Test_KetangxPID(t *testing.T) {
	pid := "1757909840335X1342862"
	vid := "a2cf165d12d19123561ed300e4cc2ff1_a"
	r := "19"
	cT := "20"
	fmt.Println(getPID())
	fmt.Println(getUid(getPID()))
	fmt.Println(getSign("rtas.net", pid, vid, "0", r, cT))
}

// 测试登录
func Test_KetangxLogin(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	user := global.Config.Users[32]
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
