package examples

import (
	"fmt"
	"github.com/yatori-dev/yatori-go-core/aggregation/yinghua"
	"github.com/yatori-dev/yatori-go-core/global"
	"github.com/yatori-dev/yatori-go-core/interfaces"
	"github.com/yatori-dev/yatori-go-core/utils"
	"testing"
)

func TestInterfacesTest(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	users := global.Config.Users[12]
	var user interfaces.IUser = &yinghua.YingHuaUser{PreUrl: users.URL, Account: users.Account, Password: users.Password}
	login, err := user.Login()
	fmt.Println(login, err)
}
