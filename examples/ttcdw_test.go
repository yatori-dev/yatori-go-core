package examples

import (
	"fmt"
	"log"
	"testing"

	ttcdw "github.com/yatori-dev/yatori-go-core/aggregation/ttcdw"
	ttcdwApi "github.com/yatori-dev/yatori-go-core/api/ttcdw"
	"github.com/yatori-dev/yatori-go-core/global"
	"github.com/yatori-dev/yatori-go-core/utils"
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
	user := global.Config.Users[53]
	cache := ttcdwApi.TtcdwUserCache{Account: user.Account, Password: user.Password}
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

// TTCDW测试刷课
func TestTtcdwTestCourseBrush(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	user := global.Config.Users[58]
	//cache := ttcdwApi.TtcdwUserCache{Account: user.Account, Password: user.Password}
	//err2 := ttcdw.TTCDWLoginAction(&cache) //登录账号
	cache := ttcdwApi.TtcdwUserCache{Account: user.Account, Password: "CNZZDATA1273209267=2118013400-1757946520-%7C1760943926; HWWAFSESID=6146c73229b89198f1; HWWAFSESTIME=1761842593979; passport=https://www.ttcdw.cn/p/passport; u-lastLoginTime=1761842830610; u-activeState=1; u-mobileState=0; u-mobile=13875993221; u-preLoginTime=1761842763611; u-token=eyJhbGciOiJIUzI1NiJ9.eyJqdGkiOiIwY2NmNzc2MC05ZDJlLTQ1MjAtODdlMS0yMDhjZWUwOGFjYzIiLCJpYXQiOjE3NjE4NDI4MzAsInN1YiI6Ijk1NTYwNDIxNTQ3NzE1Nzg4OCIsImlzcyI6Imd1b3JlbnQiLCJhdHRlc3RTdGF0ZSI6MCwic3JjIjoid2ViIiwiYWN0aXZlU3RhdGUiOjEsIm1vYmlsZSI6IjEzODc1OTkzMjIxIiwicGxhdGZvcm1JZCI6IjEzMTQ1ODU0OTgzMzExIiwiYWNjb3VudCI6IjEzODc1OTkzMjIxIiwiZXhwIjoxNzYxODc4ODMwfQ.ymsB6x5ENCGyxg-J5jGQvDbbK3y_9VCp2N-MFwWL1YY; u-token-legacy=eyJhbGciOiJIUzI1NiJ9.eyJqdGkiOiIwY2NmNzc2MC05ZDJlLTQ1MjAtODdlMS0yMDhjZWUwOGFjYzIiLCJpYXQiOjE3NjE4NDI4MzAsInN1YiI6Ijk1NTYwNDIxNTQ3NzE1Nzg4OCIsImlzcyI6Imd1b3JlbnQiLCJhdHRlc3RTdGF0ZSI6MCwic3JjIjoid2ViIiwiYWN0aXZlU3RhdGUiOjEsIm1vYmlsZSI6IjEzODc1OTkzMjIxIiwicGxhdGZvcm1JZCI6IjEzMTQ1ODU0OTgzMzExIiwiYWNjb3VudCI6IjEzODc1OTkzMjIxIiwiZXhwIjoxNzYxODc4ODMwfQ.ymsB6x5ENCGyxg-J5jGQvDbbK3y_9VCp2N-MFwWL1YY; u-id=955604215477157888; u-account=13875993221; ufo-urn=MTM4NzU5OTMyMjE=; ufo-un=5Y2i5a2j5p2+; ufo-id=955604215477157888; u-name=web_user_FQYX2QyH; ufo-nk=5Y2i5a2j5p2%2B"}
	err2 := ttcdw.TTCDWCookieLoginAction(&cache) //登录账号
	if err2 != nil {
		log.Fatal(err2)
	}

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
	videos, err := ttcdw.PullVideoListAction(&cache, projects[0], classRooms[0], courses[0])
	if err != nil {
		log.Fatal(err)
	}

	//ttcdw.PullVideoAction(&cache, projects[0])
	fmt.Println("Action:", projects)
	fmt.Println("ClassRoom:", classRooms)
	fmt.Println("Course:", courses)
	fmt.Println("Videos:", videos)
}
