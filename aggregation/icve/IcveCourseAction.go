package icve

import (
	"errors"
	log2 "log"
	"strconv"

	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/api/icve"
)

type IcveCourse struct {
	Id           string
	CourseId     string
	CourseName   string
	SchoolName   string
	TeacherName  string
	CourseType   string //课程类型，是资源库课程，还是职教云课程
	CourseInfoId string
	WeekStr      string //是否进行中

}

// 课程任务节点
type IcveCourseNode struct {
	Id       string
	CourseId string
	Name     string //任务点名称
	FileType string //任务点类型
	IsLook   bool   //是否看过

}

// 拉取资源库课程
func PullZYKCourseAction(cache *icve.IcveUserCache) ([]IcveCourse, error) {
	courseList := make([]IcveCourse, 0)
	courseResult, err := cache.PullZykCourseApi()
	if err != nil {
		log2.Fatal(err)
	}
	resultCode := gojsonq.New().JSONString(courseResult).Find("code")
	if resultCode == nil {
		return []IcveCourse{}, errors.New(courseResult)
	}
	if int(resultCode.(float64)) != 200 {
		return []IcveCourse{}, errors.New(courseResult)
	}
	rowsJson := gojsonq.New().JSONString(courseResult).Find("rows")
	if rows, ok := rowsJson.([]interface{}); ok {
		for _, row := range rows {
			if courseData, ok1 := row.(map[string]interface{}); ok1 {
				course := IcveCourse{
					CourseType: "资源库",
				}
				id := courseData["id"]
				if id != nil {
					course.Id = id.(string)
				}
				courseId := courseData["courseId"]
				if courseId != nil {
					course.CourseId = courseId.(string)
				}
				courseName := courseData["courseName"]
				if courseName != nil {
					course.CourseName = courseName.(string)
				}
				schoolName := courseData["schoolName"]
				if schoolName != nil {
					course.SchoolName = schoolName.(string)
				}
				courseInfoId := courseData["courseInfoId"]
				if courseInfoId != nil {
					course.CourseInfoId = courseInfoId.(string)
				}
				weekStr := courseData["weekStr"]
				if weekStr != nil {
					course.WeekStr = weekStr.(string)
				}
				courseList = append(courseList, course)
			}
		}
	}
	return courseList, nil
}

// 拉取任务节点
func PullZYKCourseNodeAction(cache *icve.IcveUserCache, course IcveCourse) ([]IcveCourseNode, error) {
	nodeList := make([]IcveCourseNode, 0)
	//拉取根目录
	rootResult, err := cache.PullRootNodeListApi(course.CourseInfoId)
	if err != nil {
		log2.Fatal(err)
	}

	rootJson := gojsonq.New().JSONString(rootResult).Get()
	if rootJson == "" {
		return []IcveCourseNode{}, errors.New(rootResult)
	}
	if root, ok := rootJson.([]interface{}); ok {
		for nodeJson := range root {
			if nodeData, ok1 := root[nodeJson].(map[string]interface{}); ok1 {
				//parentId := nodeData["id"].(string)
				nodes, err1 := pullNode(cache, nodeData, course)
				if err1 != nil {
					log2.Fatal(err1)
				}
				nodeList = append(nodeList, nodes...)
			}
		}
	}
	return nodeList, nil
}

// 递归拉取节点
func pullNode(cache *icve.IcveUserCache, root map[string]interface{}, course IcveCourse) ([]IcveCourseNode, error) {
	nodeList := make([]IcveCourseNode, 0)
	parentId := root["id"].(string)
	fileType := root["fileType"]
	switch fileType {
	case "父节点":
		nodeResult, err1 := cache.PullNodeListApi(1, parentId, course.CourseInfoId)
		if err1 != nil {
			log2.Fatal(err1)
		}
		//继续递归
		rootJson := gojsonq.New().JSONString(nodeResult).Get()
		if nodes, ok := rootJson.([]interface{}); ok {
			for _, nodeJson := range nodes {
				if node, ok1 := nodeJson.(map[string]interface{}); ok1 {
					result, err2 := pullNode(cache, node, course)
					if err2 != nil {
						log2.Fatal(err2)
					}
					nodeList = append(nodeList, result...)
				}
			}
		}
	case "子节点":
		level, err := strconv.Atoi(root["level"].(string))
		if err != nil {
			log2.Fatal(err)
		}
		nodeResult, err1 := cache.PullNodeListApi(level, parentId, course.CourseInfoId)
		if err1 != nil {
			log2.Fatal(err1)
		}
		//继续递归
		rootJson := gojsonq.New().JSONString(nodeResult).Get()
		if nodes, ok := rootJson.([]interface{}); ok {
			for _, nodeJson := range nodes {
				if node, ok1 := nodeJson.(map[string]interface{}); ok1 {
					result, err2 := pullNode(cache, node, course)
					if err2 != nil {
						log2.Fatal(err2)
					}
					nodeList = append(nodeList, result...)
				}
			}
		}
	case "docx":
		node := IcveCourseNode{
			Id:       root["id"].(string),
			CourseId: root["courseId"].(string),
			Name:     root["name"].(string),
			FileType: root["fileType"].(string),
		}

		isLook := root["isLook"]
		if isLook != nil {
			node.IsLook = isLook.(bool)
		}
		nodeList = append(nodeList, node)
	case "pptx":
		node := IcveCourseNode{
			Id:       root["id"].(string),
			CourseId: root["courseId"].(string),
			Name:     root["name"].(string),
			FileType: root["fileType"].(string),
		}

		isLook := root["isLook"]
		if isLook != nil {
			node.IsLook = isLook.(bool)
		}
		nodeList = append(nodeList, node)
	}

	return nodeList, nil
}
