package examples

import (
	"fmt"
	"testing"

	action "github.com/yatori-dev/yatori-go-core/aggregation/weiban"
	"github.com/yatori-dev/yatori-go-core/api/weiban"
	"github.com/yatori-dev/yatori-go-core/global"
	"github.com/yatori-dev/yatori-go-core/utils"
)

// 测试微伴登录
func TestWeiBanLogin(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	cache := weiban.WeiBanCache{
		School:   "西安文理学院",
		Account:  global.Config.Users[35].Account,
		Password: global.Config.Users[35].Password,
	}

	loginAction, err := action.WeiBanLoginAction(&cache)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(loginAction)

}

// 加密测试
func TestWeiBanEnc(t *testing.T) {
	encrypt, err := weiban.Encrypt(`{"keyNumber":"2**********001","password":"2**********001","tenantCode":"7****5","time":1758350222711,"verifyCode":"BRD8"}`)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(encrypt)
}
