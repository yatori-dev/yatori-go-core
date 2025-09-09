package examples

import (
	"testing"

	"github.com/yatori-dev/yatori-go-core/api/cela"
	"github.com/yatori-dev/yatori-go-core/global"
	"github.com/yatori-dev/yatori-go-core/utils"
)

func Test_MD5(t *testing.T) {
	//sum := md5.Sum([]byte("Nlj123456"))
	//fmt.Printf("%x", sum)
	utils.YatoriCoreInit()
	setup()
	user := global.Config.Users[30]
	cache := cela.CelaUserCache{
		Account:  user.Account,
		Password: user.Password,
	}
	cache.Code = "Ypnf"
	cache.LoginApi()
}
