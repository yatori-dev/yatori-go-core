package action

import (
	"github.com/yatori-dev/yatori-go-core/api/xuexitong"
	"github.com/yatori-dev/yatori-go-core/api/yinghua"
	"github.com/yatori-dev/yatori-go-core/global"
)

// MainFunc TODO 后续接口 待完成 先用interface{}占位
type MainFunc interface {
	YingHua() LoginInterface
	XueXiT() LoginInterface
}

type LoginInterface interface {
	LoginAction() interface{}
}

type YatoriCache struct {
	yinghua.YingHuaUserCache
	xuexitong.XueXiTUserCache
	currentCacheType string // 标识当前使用的Cache类型
}

func ActionCache(input int) YatoriCache {
	user := global.Config.Users[input]
	switch user.AccountType {
	case "YINGHUA":
		cache := yinghua.NewCache(input)
		return YatoriCache{
			YingHuaUserCache: cache,
			currentCacheType: user.AccountType,
		}
	case "XUEXITONG":
		cache := xuexitong.NewCache(input)
		return YatoriCache{
			XueXiTUserCache:  cache,
			currentCacheType: user.AccountType,
		}
	default:
		return YatoriCache{currentCacheType: "error"}
	}
}

func (y YatoriCache) YingHua() LoginInterface {
	return y
}

func (y YatoriCache) XueXiT() LoginInterface {
	return y
}
