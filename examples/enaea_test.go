package examples

import (
	"fmt"
	"testing"
	time2 "time"

	"github.com/yatori-dev/yatori-go-core/aggregation/enaea"
	enaeaApi "github.com/yatori-dev/yatori-go-core/api/enaea"
	"github.com/yatori-dev/yatori-go-core/global"
	"github.com/yatori-dev/yatori-go-core/utils"
)

// 测试学习公社登录
func TestENAEALogin(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	users := global.Config.Users[9]
	cache := enaeaApi.EnaeaUserCache{Account: users.Account, Password: users.Password}
	_, err := enaea.EnaeaLoginAction(&cache)
	if err != nil {
		t.Error(err)
	}
}

// 测试学习公社拉取所有学习项目
func TestENAEAPullProject(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	users := global.Config.Users[3]
	cache := enaeaApi.EnaeaUserCache{Account: users.Account, Password: users.Password}
	_, err := enaea.EnaeaLoginAction(&cache)
	if err != nil {
		t.Error(err)
	}
	action, err := enaea.ProjectListAction(&cache)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(action)
}

// 测试学习公社根据学习项目拉取所对应的课程
func TestENAEAPullProjectForCourse(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	users := global.Config.Users[3]
	cache := enaeaApi.EnaeaUserCache{Account: users.Account, Password: users.Password}
	_, err := enaea.EnaeaLoginAction(&cache)
	if err != nil {
		t.Error(err)
	}
	action, err := enaea.ProjectListAction(&cache)
	if err != nil {
		t.Error(err)
	}
	enaea.CourseListAction(&cache, action[0].CircleId)
}

// 测试学习公社根据学习项目拉取所对应的视频并学习
func TestENAEAPullProjectForVideo(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	users := global.Config.Users[19]
	cache := enaeaApi.EnaeaUserCache{Account: users.Account, Password: users.Password}
	_, err := enaea.EnaeaLoginAction(&cache)
	if err != nil {
		t.Error(err)
	}
	action, err := enaea.ProjectListAction(&cache)
	if err != nil {
		t.Error(err)
	}
	courseList, err := enaea.CourseListAction(&cache, action[0].CircleId)
	for _, v := range courseList {
		videoList, err := enaea.VideoListAction(&cache, &v)
		if err != nil {
			t.Error(err)
		}
		for _, video := range videoList {
			err := enaea.StatisticTicForCCVideAction(&cache, &video)
			if err != nil {
				t.Error(err)
			}
			//如果学过了，那么跳过
			if video.StudyProgress >= 100 {
				continue
			}
			for {
				//err := enaea.SubmitStudyTimeAction(&cache, &video, time2.Now().UnixMilli(), 0)
				err := enaea.SubmitStudyTimeAction(&cache, &video, 60, 1)
				if err != nil {
					t.Error(err)
				}
				err2 := enaea.StatisticTicForCCVideAction(&cache, &video)
				if err2 != nil {
					t.Error(err)
				}
				if video.StudyProgress >= 100 { //如果学习完毕那么跳过 break
				}
				time2.Sleep(time2.Second * 16)

			}
		}

	}
}

// 测试学习公社暴力模式视频学习
func TestENAEAPullProjectForFastVideo(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	users := global.Config.Users[54]
	cache := enaeaApi.EnaeaUserCache{Account: users.Account, Password: users.Password}
	//_, err := enaea.EnaeaLoginAction(&cache)
	err := enaea.EnaeaCookieLoginAction(&cache, "sensorsdata2015jssdkcross=%7B%22distinct_id%22%3A%221994dc3d83473b-03eb7ded9951ae6-4c657b58-1639680-1994dc3d835684%22%2C%22first_id%22%3A%22%22%2C%22props%22%3A%7B%22%24latest_traffic_source_type%22%3A%22%E7%9B%B4%E6%8E%A5%E6%B5%81%E9%87%8F%22%2C%22%24latest_search_keyword%22%3A%22%E6%9C%AA%E5%8F%96%E5%88%B0%E5%80%BC_%E7%9B%B4%E6%8E%A5%E6%89%93%E5%BC%80%22%2C%22%24latest_referrer%22%3A%22%22%7D%2C%22identities%22%3A%22eyIkaWRlbnRpdHlfY29va2llX2lkIjoiMTk5NGRjM2Q4MzQ3M2ItMDNlYjdkZWQ5OTUxYWU2LTRjNjU3YjU4LTE2Mzk2ODAtMTk5NGRjM2Q4MzU2ODQifQ%3D%3D%22%2C%22history_login_id%22%3A%7B%22name%22%3A%22%22%2C%22value%22%3A%22%22%7D%2C%22%24device_id%22%3A%221994dc3d83473b-03eb7ded9951ae6-4c657b58-1639680-1994dc3d835684%22%7D; HWWAFSESID=ea88688f743a66d215; HWWAFSESTIME=1761020037430; UM_distinctid=19a04fba3b72ab-05799ce31b3b2a-4c657b58-190500-19a04fba3b810a9; ASUSS=MDk3QzY2NEJGNjhGRkY3OUNGMTU1ODBDRkQ2MzY0RjMucGFzc3BvcnQ0MToxOTIuMTY4LjAuMjM5OjExMjExOjE0MzQ3NjQ5OnBjdXNlcl9waG05cHh0eToxNzYxMDIxMTg3OTU3Ok56QXpZVE5rTnpVeU5UUTVZMkUyWW1aallUZGhPV0kxTWpGbE1UTXlNemM9; hasUnReadAnwser=false; hasUnreadNotice=false; hasUnreadPlan=\"\"; hasUnreadBrief=false; hasUnreadManual=false; LOGIN_USER_NAME=pcuser_phm9pxty; SCFUCKP_14347649_308349_course=1d210a1f-a346-43ec-a453-45f32cf40bf9; SCFUCKP_14347649_304163_course=84e279b1-1ac9-4373-ae59-bbc0ef224b56; JSESSIONID=45263795AFC3FF0EF01C1D7EBE790AAE-n2.web76; SCFUCKP_14347649_308348_course=968168ce-1b22-46b8-be44-2dce318a8c85")
	if err != nil {
		t.Error(err)
	}
	action, err := enaea.ProjectListAction(&cache)
	if err != nil {
		t.Error(err)
	}
	courseList, err := enaea.CourseListAction(&cache, action[0].CircleId)
	for _, v := range courseList {
		videoList, err := enaea.VideoListAction(&cache, &v)
		if err != nil {
			t.Error(err)
		}
		//if v.CourseTitle != "人民教育家陶行知的爱国故事——我是一个中国人，要为中国做出一些贡献" {
		//	continue
		//}
		for _, video := range videoList {
			err := enaea.StatisticTicForCCVideAction(&cache, &video)
			if err != nil {
				t.Error(err)
			}

			//如果学过了，那么跳过
			//if video.StudyProgress >= 100 {
			//	continue
			//}
			for {
				//err := enaea.SubmitStudyTimeAction(&cache, &video, time2.Now().UnixMilli(), 0)
				err := enaea.SubmitStudyTimeAction(&cache, &video, 20, 1)
				if err != nil {
					t.Error(err)
				}
				err2 := enaea.StatisticTicForCCVideAction(&cache, &video)
				if err2 != nil {
					t.Error(err)
				}
				if video.StudyProgress >= 100 { //如果学习完毕那么跳过
					break
				}
				time2.Sleep(time2.Second * 1)

			}
		}

	}
}
