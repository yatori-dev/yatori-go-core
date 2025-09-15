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

type KetangxNode struct {
	SectId     string `json:"sectId"`
	Title      string `json:"title"`
	EnterNum   string `json:"enterNum"`   //参与人数
	IsComplete bool   `json:"isComplete"` //该任务点是否完成，true为完成，false为未完成
	Type       string `json:"type"`
}

func PullCourseListAction(cache *ketangx.KetangxUserCache) []KetangxCourse {
	course, err2 := cache.PullCourseListHTMLApi()
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

// 拉取对应课程视屏列表
func PullNodeListAction(cache *ketangx.KetangxUserCache, course *KetangxCourse) []KetangxNode {
	html, err := cache.PullVideoListHTMLApi(course.ActivityId)
	videoList := []KetangxNode{}
	if err != nil {
		fmt.Println(err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader([]byte(html)))
	doc.Find("li.wis-leftNodeItem[sectli]").Each(func(i int, s *goquery.Selection) {
		text := s.Find("div.wis-iconActive-tit").Text()
		if text != "视频" && text != "文档" {
			return //如果该节点不是视频或文档
		}

		videoData := KetangxNode{}
		//SectId
		sectId, ok1 := s.Attr("sectli")
		if ok1 {
			videoData.SectId = sectId
		}
		//视屏节点标题
		status, ok2 := s.Find("img.iconNodeStatus").Attr("src")
		if ok2 {
			videoData.IsComplete = status == "/Content/ZHYX/images/icon/icon-WanCheng.png"
		}
		videoData.Title = s.Find("div.leftNodeItemInfo-tit").Text()
		videoData.EnterNum = s.Find("span.NodeItemInfo-msgTxt").Text()
		videoData.Type = text
		videoList = append(videoList, videoData)
	})
	return videoList
}

// 直接完成视屏
func CompleteVideoAction(cache *ketangx.KetangxUserCache, video *KetangxNode) (string, error) {
	_, err2 := cache.SignVideoStatusApi(video.SectId) //学习视屏任务点时先进行标记任务点
	if err2 != nil {
		return "", err2
	}
	api, err := cache.CompleteVideoApi(video.SectId, cache.Id, 114514, 114514)
	if err != nil {
		return "", err
	}
	return api, nil
}
