package yatori

import (
	"github.com/yatori-dev/yatori-go-core/aggregation/xuexitong"
	"github.com/yatori-dev/yatori-go-core/aggregation/yinghua"
	"github.com/yatori-dev/yatori-go-core/interfaces"
)

type YatoriUser struct {
	Account  string
	CacheMap map[string]any
	Password string
	PreUrl   string
}

type UserOpt func(*YatoriUser)

func (user *YatoriUser) On(accountType string) interfaces.IUser {
	switch accountType {
	case "XUEXITONG":
		return &xuexitong.XueXiTongUser{Account: user.Account, Password: user.Password}
	case "YINGHUA":
		return &yinghua.YingHuaUser{Account: user.Account, Password: user.Password, PreUrl: user.PreUrl}
	}
	return nil
}

func NewUser(account, password, url string, opt ...UserOpt) (*YatoriUser, error) {
	y := &YatoriUser{
		Account:  account,
		Password: password,
		PreUrl:   url,
	}
	for _, o := range opt {
		o(y)
	}
	return y, nil
}
