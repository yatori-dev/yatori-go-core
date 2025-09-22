package examples

import (
	"testing"

	"github.com/yatori-dev/yatori-go-core/api/welearn"
	"github.com/yatori-dev/yatori-go-core/global"
	"github.com/yatori-dev/yatori-go-core/utils"
)

func TestWeLearnLogin(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	user := global.Config.Users[36]
	cache := &welearn.WeLearnCache{
		Account:  user.Account,
		Password: user.Password,
	}

	cache.WeLearnLoginApi()

}
