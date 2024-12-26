package examples

import (
	"fmt"
	ttcdw "github.com/yatori-dev/yatori-go-core/aggregation/ttcdw"
	ttcdwApi "github.com/yatori-dev/yatori-go-core/api/ttcdw"
	"github.com/yatori-dev/yatori-go-core/global"
	"github.com/yatori-dev/yatori-go-core/utils"
	"log"
	"testing"
)

/*
	{
	    "companyCode": "D387ED042DF13283",
	    "userId": "527283745945702400:876955629390692352",
	    "resId": 263696,
	    "courseId": "3086",
	    "courseType": "share",
	    "tickerTime": 1734944953704,
	    "md5": "[\"mGKNmQbR9KynQGQV83UgaA==\"]"
	}

["3y41F5QeTMvbZt+njam/cne6F8LV3K6vKaU8D5hioQe8ZHprx+dPoRvaafaKs2tf+QJzdsWlQsRYdK0yChRi4aXAV0YEEq+FYxQJyw+CfrPtuvm6nDh+92pXbCeetY/MD/2f0zdbFh0=","/Xs6cIlSR7i43HDbNBjcgt6vC30boHdwQZqf8+bTkpPbyFxKe157zsGLv0TqFABTkL2uJINT0FWCk6q5XDo71PPCA4i+11L+mSrER1rhL/SrIwkH9o94hnZRoT0XZUNSMYE4HWQmwf4=","AUPB8KqydTA="]
*/
//测试提交学时加密
func TestEnc(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	data := `{"companyCode":"D387ED042DF13283","userId":"527283745945702400:876955629390692352","resId":263918,"courseId":"3086","courseType":"share","tickerTime":1734963388575,"md5":"[\"m9t4w5EvLSVf9QAXaaTROw==\"]"}`
	encDataStr, err := ttcdwApi.EncData(data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Encrypted Data:", encDataStr)
}

// TTCDW测试登录
func TestTtcdwTestLogin(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	cache := ttcdwApi.TtcdwUserCache{Account: global.Config.Users[6].Account, Password: global.Config.Users[6].Password}
	cache.TtcdwLoginApi() //登录账号

	projects, err := ttcdw.PullProjectAction(&cache) //拉取项目
	if err != nil {
		log.Fatal(err)
	}
	classRooms, err := ttcdw.PullClassRoomAction(&cache, projects[0]) //拉取ClassRoom
	if err != nil {
		log.Fatal(err)
	}

	courses, err := ttcdw.PullCourseAction(&cache, classRooms[0]) //拉取对应的课程
	if err != nil {
		log.Fatal(err)
	}

	//ttcdw.PullVideoAction(&cache, projects[0])
	fmt.Println("Action:", projects)
	fmt.Println("ClassRoom:", classRooms)
	fmt.Println("Course:", courses)
}
