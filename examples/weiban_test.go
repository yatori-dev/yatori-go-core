package examples

import (
	"fmt"
	"testing"

	"github.com/yatori-dev/yatori-go-core/v3/global"
	action "github.com/yatori-dev/yatori-go-core/v3/internal/aggregation/weiban"
	"github.com/yatori-dev/yatori-go-core/v3/internal/api/weiban"
	"github.com/yatori-dev/yatori-go-core/v3/utils"
)

// 测试微伴登录
func TestWeiBanLogin(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	cache := weiban.WeiBanCache{
		School:   "浙江师范大学",
		Account:  global.Config.Users[0].Account,
		Password: global.Config.Users[0].Password,
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
