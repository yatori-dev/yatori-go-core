package gongxue

import (
	"github.com/yatori-dev/yatori-go-core/api/gongxue/entity"
	"github.com/yatori-dev/yatori-go-core/api/gongxue/global"
	"github.com/yatori-dev/yatori-go-core/api/gongxue/service"
)

type SignUser struct {
	Username  string `gorm:"unique;not null;column:username"` // 唯一，非空
	Password  string `gorm:"not null;column:password"`        // 非空
	Country   string `gorm:"column:country"`                  //国家
	Province  string `gorm:"column:province"`                 //省份
	City      string `gorm:"column:city"`                     //城市
	Area      string `gorm:"column:area"`                     //地区
	Longitude string `gorm:"column:longitude"`                //精度
	Latitude  string `gorm:"column:latitude"`                 //纬度
	Address   string `gorm:"column:address"`                  //详细地址
	PlayType  string //上班还是下班START为上班，END为下班
}

// GongXueSignAction 工学云签到动作
func GongXueSignAction(signUser SignUser) {
	signEntity := entity.SignEntity{
		Username:  signUser.Username,
		Password:  signUser.Password,
		Country:   signUser.Country,
		Province:  signUser.Province,
		City:      signUser.City,
		Area:      signUser.Area,
		Longitude: signUser.Longitude,
		Latitude:  signUser.Latitude,
		Address:   signUser.Address,
	}
	m := service.CreateMoguDing(signEntity)
	if err := m.HandleCaptcha(); err != nil {
		//utils.SendMail(m.Email, "Block-Error", err.Error())
		global.Log.Error(err.Error())
		return
	}
	if err := m.Login(); err != nil {
		//utils.SendMail(m.Email, "Login-Error-测试邮件请勿回复", err.Error())
		global.Log.Error(err.Error())
		return
	}

	if err := m.GetPlanId(); err != nil {
		global.Log.Warning(err.Error())
		return
	}
	if err := m.GetJobInfo(); err != nil {
		global.Log.Warn("Failed to get job info: %v", err)
		return
	}

	m.CusSignIn(signUser.PlayType)
}
