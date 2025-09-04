package examples

import (
	"testing"

	"github.com/yatori-dev/yatori-go-core/api/mooc"
	"github.com/yatori-dev/yatori-go-core/global"
	"github.com/yatori-dev/yatori-go-core/utils"
)

// MOOC登录接口测试
func TestPowGetP(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	//构建用户结构体
	cache := mooc.MOOCUserCache{
		Account:   global.Config.Users[22].Account,
		Password:  global.Config.Users[22].Password,
		IpProxySW: false,
		ProxyIP:   "",
	}
	cache.InitCookiesApi() //初始化Cookie
	cache.GtApi()          //通过gt接口获取必要登录参数
	cache.PowGetPApi()     //通过powGetP接口获取必要登录参数
	cache.LoginApi()       //登录

}
