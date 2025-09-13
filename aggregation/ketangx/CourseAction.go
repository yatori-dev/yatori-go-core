package ketangx

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"github.com/yatori-dev/yatori-go-core/api/ketangx"
)

type KetangxCourse struct {
	Title      string  `json:"title"`       //课程标题
	Progress   float64 `json:"progress"`    //课程学习进度
	ActivityId string  `json:"activity_id"` //课程ID
}

func PullCourseAction(cache *ketangx.KetangxUserCache) []KetangxCourse {
	course, err2 := cache.PullCourse()
	if err2 != nil {
		fmt.Println(err2)
	}
	courseList := []KetangxCourse{}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader([]byte(course)))
	if err != nil {
		fmt.Println(err)
	}
	doc.Find("div.course-history").Each(func(i int, s *goquery.Selection) {
		title := ""
		activityId := ""
		var progress float64 = 0
		val, exists := s.Attr("title")
		if exists {
			title = val
		}

		val1, exists1 := s.Attr("activityid")
		if exists1 {
			activityId = val1
		}

		text := s.Find(".progress-label .num").Text()
		resProgress, err3 := strconv.ParseFloat(text, 64)
		if err3 == nil {
			progress = resProgress
		}
		courseList = append(courseList, KetangxCourse{
			Title:      title,
			ActivityId: activityId,
			Progress:   progress,
		})
	})
	return courseList
}
