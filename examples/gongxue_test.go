package examples

import (
	"github.com/yatori-dev/yatori-go-core/aggregation/gongxue"
	"testing"
)

// 工学云打卡例子
func TestGongxueSignTest(t *testing.T) {
	user := gongxue.SignUser{
		Username:  "账号",
		Password:  "密码",
		Country:   "China",
		Province:  "BeiJing",
		City:      "城市",
		Area:      "区域",
		Longitude: "131.073308", //经度
		Latitude:  "25.938855",  //纬度
		Address:   "金日成路",       //详细地址
		PlayType:  "END",        //下班为END，上班为START
	}
	gongxue.GongXueSignAction(user)
}
