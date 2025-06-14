package examples

import (
	"fmt"
	"github.com/yatori-dev/yatori-go-core/global"
	"github.com/yatori-dev/yatori-go-core/utils"
	"github.com/yatori-dev/yatori-go-core/yatori"
	"testing"
)

func TestInterfacesTest(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	user1 := global.Config.Users[0]
	user, _ := yatori.NewUser(user1.Account, user1.Password, user1.URL)
	login, err := user.On(user1.AccountType).Login()
	if err != nil {
		panic(err)
	}
	fmt.Println(login)
	userInfo, err := user.On(user1.AccountType).CourseList()
	fmt.Println(userInfo)
}
