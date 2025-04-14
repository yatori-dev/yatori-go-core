package action

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/yatori-dev/yatori-go-core/api/xuexitong"
	"github.com/yatori-dev/yatori-go-core/entity"
	log2 "github.com/yatori-dev/yatori-go-core/utils/log"
	"log"
	"sort"
	"strconv"
	"strings"
)

type XueXiT struct {
	Cpi           int    `json:"cpi"`      // 用户唯一标识
	Key           string `json:"key"`      // classID 在课程API中为key
	CourseID      string `json:"courseId"` // 课程ID
	ChatID        string `json:"chatId"`
	CourseTeacher string `json:"courseTeacher"` // 课程老师
	CourseName    string `json:"courseName"`    //课程名
	CourseImage   string `json:"courseImage"`
	// 两个标识 暂时不知道有什么用
	CourseDataID int          `json:"courseDataId"`
	ContentID    int          `json:"ContentID"`
	ChaptersList ChaptersList // Chapter 标识
}

type ChaptersList struct {
	ChatID    string          `json:"chatid"`
	Knowledge []KnowledgeItem `json:"knowledge"`
}

// KnowledgeItem 结构体用于存储 knowledge 中的每个项目
type KnowledgeItem struct {
	JobCount      int           `json:"jobcount"` // 作业数量
	IsReview      int           `json:"isreview"` // 是否为复习
	Attachment    []interface{} `json:"attachment"`
	IndexOrder    int           `json:"indexorder"` // 节点顺序
	Name          string        `json:"name"`       // 章节名称
	ID            int           `json:"id"`
	Label         string        `json:"label"`        // 节点标签
	Layer         int           `json:"layer"`        // 节点层级
	ParentNodeID  int           `json:"parentnodeid"` // 父节点 ID
	Status        string        `json:"status"`       // 节点状态
	PointTotal    int
	PointFinished int
}

type XueXiTInterface interface {
	CourseList() []XueXiT
}

func (cache YatoriCache) CourseList() []XueXiT {
	courses, err := cache.XueXiTUserCache.CourseListApi()
	if err != nil {
		log2.Print(log2.INFO, "["+cache.Name+"] "+" 拉取失败")
	}
	var xueXiTCourse entity.XueXiTCourseJson
	err = json.Unmarshal([]byte(courses), &xueXiTCourse)
	if err != nil {
		log2.Print(log2.INFO, "["+cache.Name+"] "+" 解析失败")
		panic(err)
	}
	log2.Print(log2.INFO, "["+cache.Name+"] "+" 课程数量："+strconv.Itoa(len(xueXiTCourse.ChannelList)))
	// log2.Print(log2.INFO, "["+cache.Name+"] "+courses)

	var courseList = make([]XueXiT, 0)
	for i, channel := range xueXiTCourse.ChannelList {
		var flag = false
		if channel.Content.Course.Data == nil && i >= 0 && i < len(xueXiTCourse.ChannelList) {
			xueXiTCourse.ChannelList = append(xueXiTCourse.ChannelList[:i], xueXiTCourse.ChannelList[i+1:]...)
			continue
		}
		var (
			teacher      string
			courseName   string
			courseDataID int
			classId      string
			courseID     string
			courseImage  string
		)

		for _, v := range channel.Content.Course.Data {
			teacher = v.Teacherfactor
			courseName = v.Name
			courseDataID = v.Id
			userID := strings.Split(v.CourseSquareUrl, "userId=")[1]
			cache.UserID = userID
			classId = strings.Split(strings.Split(v.CourseSquareUrl, "classId=")[1], "&userId")[0]
			courseID = strings.Split(strings.Split(v.CourseSquareUrl, "courseId=")[1], "&personId")[0]
			courseImage = v.Imageurl
		}

		course := XueXiT{
			Cpi:           channel.Cpi,
			Key:           classId,
			CourseID:      courseID,
			ChatID:        channel.Content.Chatid,
			CourseTeacher: teacher,
			CourseName:    courseName,
			CourseImage:   courseImage,
			CourseDataID:  courseDataID,
			ContentID:     channel.Content.Id,
		}
		for _, course := range courseList {
			if course.CourseID == courseID {
				flag = true
				break
			}
		}
		if flag {
			continue
		}
		courseList = append(courseList, course)
	}
	return courseList
}

func (x *XueXiT) GetChapter(cache xuexitong.XueXiTUserCache) (ok bool, err error) {

	key, _ := strconv.Atoi(x.Key)
	//拉取对应课程的章节信息
	chapter, err := cache.PullChapter(x.Cpi, key)
	if err != nil {
		return false, errors.New("[" + cache.Name + "] " + " 拉取章节失败")
	}

	var chapterMap map[string]interface{}
	err = json.Unmarshal([]byte(chapter), &chapterMap)
	if err != nil {
		return false, errors.New(fmt.Sprintf("Error parsing JSON: %s", err))
	}
	chapterMapJson, err := json.Marshal(chapterMap["data"])
	if len(chapterMapJson) == 2 {
		return false, errors.New("[" + cache.Name + "] " + " 课程获取失败")
	}
	// 解析 JSON 数据为 map 切片
	var chapterData []map[string]interface{}
	if err := json.Unmarshal(chapterMapJson, &chapterData); err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}

	chatid := chapterData[0]["chatid"].(string)
	// 提取 knowledge
	var knowledgeData []map[string]interface{}
	course, ok := chapterData[0]["course"].(map[string]interface{})
	if !ok {
		return ok, errors.New("[" + cache.Name + "] " + " 无法提取 course")
	}
	data, ok := course["data"].([]interface{})
	if !ok {
		return ok, errors.New("[" + cache.Name + "] " + " 无法提取 course data")
	}
	if len(data) > 0 {
		knowledge, ok := data[0].(map[string]interface{})["knowledge"].(map[string]interface{})["data"].([]interface{})
		if !ok {
			return ok, errors.New("[" + cache.Name + "] " + " 无法提取 knowledge data")
		}
		for _, item := range knowledge {
			knowledgeMap := item.(map[string]interface{})
			knowledgeData = append(knowledgeData, knowledgeMap)
		}
	} else {
		return false, errors.New("[" + cache.Name + "] " + " course data 为空")
	}

	// 将提取的数据封装到 CourseInfo 结构体中
	var knowledgeItems []KnowledgeItem
	for _, item := range knowledgeData {
		knowledgeItem := KnowledgeItem{
			JobCount:     int(item["jobcount"].(float64)),
			IsReview:     int(item["isreview"].(float64)),
			Attachment:   item["attachment"].(map[string]interface{})["data"].([]interface{}),
			IndexOrder:   int(item["indexorder"].(float64)),
			Name:         item["name"].(string),
			ID:           int(item["id"].(float64)),
			Label:        item["label"].(string),
			Layer:        int(item["layer"].(float64)),
			ParentNodeID: int(item["parentnodeid"].(float64)),
			Status:       item["status"].(string),
		}
		knowledgeItems = append(knowledgeItems, knowledgeItem)
	}
	x.ChaptersList = ChaptersList{
		ChatID:    chatid,
		Knowledge: knowledgeItems,
	}
	if len(x.ChaptersList.Knowledge) == 0 {
		log2.Print(log2.INFO, "["+cache.Name+"] "+"["+x.ChaptersList.ChatID+"] "+" 课程章节为空")
		return false, err
	}
	// 按照任务点节点重排顺序
	sort.Slice(x.ChaptersList.Knowledge, func(i, j int) bool {
		iLabelParts := strings.Split(x.ChaptersList.Knowledge[i].Label, ".")
		jLabelParts := strings.Split(x.ChaptersList.Knowledge[j].Label, ".")
		for k := range iLabelParts {
			if k >= len(jLabelParts) {
				return false // i has more parts, so it should come after j
			}
			iv, _ := strconv.Atoi(iLabelParts[k])
			jv, _ := strconv.Atoi(jLabelParts[k])
			if iv != jv {
				return iv < jv
			}
		}
		return len(iLabelParts) < len(jLabelParts)
	})
	fmt.Printf("获取课程章节成功 (共 %d 个)\n",
		len(x.ChaptersList.Knowledge)) //  [%s(Cou.%s/Cla.%s)]
	return true, nil
}
