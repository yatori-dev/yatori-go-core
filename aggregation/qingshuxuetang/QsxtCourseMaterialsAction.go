package qingshuxuetang

import (
	"fmt"
	"strconv"

	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/api/qingshuxuetang"
)

type QsxtCourseMaterial struct {
	Id          string
	BookId      string
	Name        string
	CoverImg    string
	Url         string
	Isbn        string
	Author      string
	Publisher   string
	CxEbookDxid string
	CxEbookKey  string
	EbookType   string
	Free        bool
	ClassId     string
	CourseId    string
	SemesterId  string
	SchoolId    string
}

// 课程资料列表拉取
func PullCourseMaterialListAction(cache *qingshuxuetang.QsxtUserCache, course QsxtCourse) ([]QsxtCourseMaterial, error) {
	materialList := []QsxtCourseMaterial{}
	materialListJson, err := cache.PullCourseMaterialsListApi(course.SemesterId, course.ClassId, course.SchoolId, "1", "0", course.CourseId, 5, nil)
	if err != nil {
		return nil, err
	}
	//异常处理
	pullStatus := gojsonq.New().JSONString(materialListJson).Find("hr")
	if pullStatus == nil {
		return nil, fmt.Errorf(materialListJson)
	}
	if int(pullStatus.(float64)) != 0 {
		return nil, fmt.Errorf(materialListJson)
	}
	materialsJson := gojsonq.New().JSONString(materialListJson).Find("data")
	if materialJson, ok := materialsJson.(map[string]interface{}); ok {
		if ebooksJson, ok := materialJson["materials"].([]interface{}); ok {
			for _, ebookJson := range ebooksJson {
				if ebook, ok := ebookJson.(map[string]interface{}); ok {
					material := QsxtCourseMaterial{
						Id:         strconv.Itoa(int(ebook["id"].(float64))),
						BookId:     strconv.Itoa(int(ebook["id"].(float64))),
						Name:       ebook["name"].(string),
						ClassId:    course.ClassId,
						CourseId:   course.CourseId,
						SemesterId: course.SemesterId,
						SchoolId:   course.SchoolId,
					}

					if free, ok := ebook["free"].(bool); ok {
						material.Free = free
					}
					if isbn, ok := ebook["isbn"].(string); ok {
						material.Isbn = isbn
					}
					if author, ok := ebook["author"].(string); ok {
						material.Author = author
					}
					if publisher, ok := ebook["publisher"].(string); ok {
						material.Publisher = publisher
					}
					if cxEbookDxid, ok := ebook["cxEbookDxid"].(string); ok {
						material.CxEbookDxid = cxEbookDxid
					}
					if cxEbookKey, ok := ebook["cxEbookKey"].(string); ok {
						material.CxEbookKey = cxEbookKey
					}
					materialList = append(materialList, material)
				}
			}
		}
		if ebooksJson, ok := materialJson["ebooks"].([]interface{}); ok {
			for _, ebookJson := range ebooksJson {
				if ebook, ok := ebookJson.(map[string]interface{}); ok {
					material := QsxtCourseMaterial{
						Id:         strconv.Itoa(int(ebook["id"].(float64))),
						BookId:     ebook["bookId"].(string),
						Name:       ebook["name"].(string),
						CoverImg:   ebook["coverImg"].(string),
						ClassId:    course.ClassId,
						CourseId:   course.CourseId,
						SemesterId: course.SemesterId,
						SchoolId:   course.SchoolId,
					}
					if free, ok := ebook["free"].(bool); ok {
						material.Free = free
					}
					if isbn, ok := ebook["isbn"].(string); ok {
						material.Isbn = isbn
					}
					if author, ok := ebook["author"].(string); ok {
						material.Author = author
					}
					if publisher, ok := ebook["publisher"].(string); ok {
						material.Publisher = publisher
					}
					if cxEbookDxid, ok := ebook["cxEbookDxid"].(string); ok {
						material.CxEbookDxid = cxEbookDxid
					}
					if cxEbookKey, ok := ebook["cxEbookKey"].(string); ok {
						material.CxEbookKey = cxEbookKey
					}
					materialList = append(materialList, material)
				}
			}
		}
	}
	return materialList, nil
}

// 开始学习
func (node QsxtCourseMaterial) StartStudyTimeAction(cache *qingshuxuetang.QsxtUserCache) (string, error) {
	startResult, err := cache.StartStudyApi(node.ClassId, node.BookId, "12", node.CourseId, node.SemesterId, node.SchoolId, 3, nil)
	if err != nil {
		return "", err
	}
	//异常处理
	pullStatus := gojsonq.New().JSONString(startResult).Find("hr")
	if pullStatus == nil {
		return "", fmt.Errorf(startResult)
	}
	if int(pullStatus.(float64)) != 0 {
		return "", fmt.Errorf(startResult)
	}

	startId := gojsonq.New().JSONString(startResult).Find("data").(string)
	return startId, nil
}

// 提交学时
func (node QsxtCourseMaterial) SubmitStudyTimeAction(cache *qingshuxuetang.QsxtUserCache, startId string, isEnd bool) (string, error) {
	submitResult, err1 := cache.SubmitStudyTimeApi(node.SchoolId, startId, 0, isEnd, 3, nil)
	if err1 != nil {
		return "", err1
	}
	return submitResult, nil
}
