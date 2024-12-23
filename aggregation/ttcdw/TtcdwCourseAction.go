package ttcdw

import (
	"errors"
	"fmt"
	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/api/ttcdw"
)

type TtcdwProject struct {
	CourseProjectId string //工程ID
	ClassId         string //班级ID
	Name            string //工程名称
	studyState      string //学习状态
	OrgId           string //不知道啥用的ID
}
type TtcdwClassRoom struct {
}
type TtcdwVideo struct {
}

// 拉取项目
func PullProjectAction(cache *ttcdw.TtcdwUserCache) ([]TtcdwProject, error) {
	var projects []TtcdwProject
	api, err := cache.PullProjectApi(5, nil)
	if err != nil {
		return nil, err
	}
	//如果获取失败
	if gojsonq.New().JSONString(api).Find("success") != true {
		return projects, errors.New(api)
	}
	jsonList := gojsonq.New().JSONString(api).Find("data")
	// 断言为切片并遍历
	if items, ok := jsonList.([]interface{}); ok {
		for _, item := range items {
			// 每个 item 是 map[string]interface{} 类型
			if obj, ok := item.(map[string]interface{}); ok {
				projects = append(projects, TtcdwProject{
					CourseProjectId: obj["courseProjectId"].(string),
					ClassId:         obj["classId"].(string),
					Name:            obj["name"].(string),
					studyState:      obj["studyState"].(string),
					OrgId:           obj["orgId"].(string),
				})
			}
		}
	}
	return projects, nil
}

func PullVideoAction(cache *ttcdw.TtcdwUserCache, project TtcdwProject) ([]TtcdwVideo, error) {
	var videos []TtcdwVideo
	classRoom, err := cache.PullClassRoomApi(project.CourseProjectId, project.ClassId, 5, nil)
	if err != nil {
		return nil, err
	}
	fmt.Println(classRoom)
	//chapterHtml, err := cache.PullChapterListHtmlApi(5, nil)
	//if err != nil {
	//	return nil, err
	//}
	//secPattern := `data-secId="([^"]+)"`
	//secRegexp := regexp.MustCompile(secPattern)
	//secMatches := secRegexp.FindAllStringSubmatch(chapterHtml, -1)
	//for _, v := range secMatches {
	//	//secId := v[1] //获取章节编号
	//
	//}
	return videos, nil
}
