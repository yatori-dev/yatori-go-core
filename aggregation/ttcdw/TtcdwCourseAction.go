package ttcdw

import (
	"errors"
	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/api/ttcdw"
	"strconv"
	"strings"
)

type TtcdwProject struct {
	CourseProjectId string //工程ID
	ClassId         string //班级ID
	Name            string //工程名称
	studyState      string //学习状态
	OrgId           string //不知道啥用的ID
}
type TtcdwClassRoom struct {
	Name      string //名称
	Title     string //是必修还是选修等
	ItemId    string //itemId
	SegmentId string
}

type TtcdwCourse struct {
	CourseId           string
	Name               string
	Progress           float32 //进度
	Duration           int     //总时长
	OriginalId         string  //originalID
	TotalStudyProgress float32 // 已学习的时长
	CompanyCode        string  //公司对应代码编号
	MD5                string  //对应md5值
	ShortCourseId      string  //课程短ID
	UserId             string  //用户ID
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

// 拉取所有ClassRoom
func PullClassRoomAction(cache *ttcdw.TtcdwUserCache, project TtcdwProject) ([]TtcdwClassRoom, error) {
	var classRooms []TtcdwClassRoom
	classRoom, err := cache.PullClassRoomApi(project.CourseProjectId, project.ClassId, 5, nil)
	if err != nil {
		return nil, err
	}
	//fmt.Println(classRoom)
	//如果获取失败
	if gojsonq.New().JSONString(classRoom).Find("success") != true {
		return classRooms, errors.New(classRoom)
	}
	jsonList := gojsonq.New().JSONString(classRoom).Find("data")
	// 断言为切片并遍历
	if items, ok := jsonList.([]interface{}); ok {
		for _, item := range items {
			// 每个 item 是 map[string]interface{} 类型
			if obj, ok := item.(map[string]interface{}); ok {
				nodeList := obj["itemList"]
				if nodes, ok := nodeList.([]interface{}); ok {
					for _, node := range nodes {
						if obj1, ok := node.(map[string]interface{}); ok {
							classRooms = append(classRooms, TtcdwClassRoom{
								Name:      obj["name"].(string),
								Title:     obj1["title"].(string),
								ItemId:    obj1["id"].(string),
								SegmentId: obj1["segmentId"].(string),
							})
						}
					}
				}
			}
		}
	}

	return classRooms, nil
}

// 拉取所有课程
func PullCourseAction(cache *ttcdw.TtcdwUserCache, class TtcdwClassRoom) ([]TtcdwCourse, error) {
	var courses []TtcdwCourse
	coursesApi, err := cache.PullCourseApi(class.SegmentId, class.ItemId, 5, nil)
	if err != nil {
		return nil, err
	}
	//fmt.Println(classRoom)
	//如果获取失败
	if gojsonq.New().JSONString(coursesApi).Find("success") != true {
		return courses, errors.New(coursesApi)
	}
	jsonList := gojsonq.New().JSONString(coursesApi).Find("data")
	// 断言为切片并遍历
	if items, ok := jsonList.([]interface{}); ok {
		for _, item := range items {
			// 每个 item 是 map[string]interface{} 类型
			if obj, ok := item.(map[string]interface{}); ok {
				infoApi, err := cache.PullCourseInfoApi(class.SegmentId, obj["id"].(string), 5, nil)
				if err != nil {
					return nil, err
				}
				if gojsonq.New().JSONString(infoApi).Find("success") != true {
					return courses, errors.New(infoApi)
				}
				info := gojsonq.New().JSONString(infoApi).Find("data.course.thirdCourseUrl")
				var companyCode string
				var MD5 string
				var shortCourseId string
				var userId string
				if info != nil {
					kv := strings.Split(info.(string), "?")[1]
					companyCode = strings.Split(strings.Split(kv, "companyCode=")[1], "&")[0]
					MD5 = strings.Split(strings.Split(kv, "md5=")[1], "&")[0]
					shortCourseId = strings.Split(strings.Split(kv, "courseId=")[1], "&")[0]
					userId = strings.Split(strings.Split(kv, "userId=")[1], "&")[0]
				}

				progress, _ := strconv.ParseFloat(obj["progress"].(string), 32)
				duration, _ := strconv.Atoi(obj["duration"].(string))
				courses = append(courses, TtcdwCourse{
					CourseId:      obj["id"].(string),
					Name:          obj["name"].(string),
					Progress:      float32(progress),
					Duration:      duration,
					OriginalId:    obj["originalId"].(string),
					ShortCourseId: shortCourseId,
					MD5:           MD5,
					CompanyCode:   companyCode,
					UserId:        userId,
				})
			}
		}
	}
	return courses, nil
}
func PullVideoAction(cache *ttcdw.TtcdwUserCache) ([]TtcdwVideo, error) {
	var videos []TtcdwVideo

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
