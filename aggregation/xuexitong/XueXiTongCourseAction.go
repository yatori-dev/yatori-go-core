package xuexitong

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/yatori-dev/yatori-go-core/api/entity"
	"github.com/yatori-dev/yatori-go-core/api/xuexitong"
	log2 "github.com/yatori-dev/yatori-go-core/utils/log"
)

type XueXiTCourse struct {
	Cpi           int    `json:"cpi"`      // 用户唯一标识
	Key           string `json:"key"`      // classID 在课程API中为key
	CourseID      string `json:"courseId"` // 课程ID
	ChatID        string `json:"chatId"`
	CourseTeacher string `json:"courseTeacher"` // 课程老师
	CourseName    string `json:"courseName"`    //课程名
	CourseImage   string `json:"courseImage"`
	// 两个标识 暂时不知道有什么用
	CourseDataID int `json:"courseDataId"`
	ContentID    int `json:"ContentID"`
}

func (x *XueXiTCourse) ToString() string {
	return fmt.Sprintf(
		"XueXiTCourse{Cpi: %d, Key: %v, CourseID: %s,Teacher: %s, CourseName: %s, CourseImage: %s\nCourseDataID: %d, ContentID: %d}",
		x.Cpi, x.Key, x.CourseID, x.CourseTeacher, x.CourseName, x.CourseImage, x.CourseDataID, x.ContentID,
	)
}

// 拉取学习通所有课程列表并返回
func XueXiTPullCourseAction(cache *xuexitong.XueXiTUserCache) ([]XueXiTCourse, error) {
	courses, err := cache.CourseListApi()
	if err != nil {
		log2.Print(log2.INFO, "["+cache.Name+"] "+" 拉取失败")
	}
	var xueXiTCourse entity.XueXiTCourseJson
	err = json.Unmarshal([]byte(courses), &xueXiTCourse)
	if err != nil {
		log2.Print(log2.INFO, "["+cache.Name+"] "+" 解析失败", courses)
		panic(err)
	}
	log2.Print(log2.INFO, "["+cache.Name+"] "+" 课程数量："+strconv.Itoa(len(xueXiTCourse.ChannelList)))
	// log2.Print(log2.INFO, "["+cache.Name+"] "+courses)

	var courseList = make([]XueXiTCourse, 0)
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

		course := XueXiTCourse{
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
	return courseList, nil
}

// TODO 优化掉
// XueXiTCourseDetailForCourseIdAction 根据课程ID拉取学习课程详细信息
//func XueXiTCourseDetailForCourseIdAction(cache *xuexitong.XueXiTUserCache, courseId string) (entity.XueXiTCourse, error) {
//	courses, err := cache.CourseListApi()
//	if err != nil {
//		return entity.XueXiTCourse{}, err
//	}
//	var xueXiTCourse entity.XueXiTCourseJson
//	err = json.Unmarshal([]byte(courses), &xueXiTCourse)
//	for _, channel := range xueXiTCourse.ChannelList {
//		if channel.Content.Chatid != courseId {
//			continue
//		}
//		//marshal, _ := json.Marshal()
//		sqUrl := channel.Content.Course.Data[0].CourseSquareUrl
//		courseId := strings.Split(strings.Split(sqUrl, "courseId=")[1], "&personId")[0]
//		personId := strings.Split(strings.Split(sqUrl, "personId=")[1], "&classId")[0]
//		classId := strings.Split(strings.Split(sqUrl, "classId=")[1], "&userId")[0]
//		userId := strings.Split(sqUrl, "userId=")[1]
//		course := entity.XueXiTCourse{
//			CourseName: channel.Content.Name,
//			ClassId:    classId,
//			CourseId:   courseId,
//			Cpi:        strconv.Itoa(channel.Cpi),
//			PersonId:   personId,
//			UserId:     userId}
//		return course, nil
//	}
//	log2.Print(log2.INFO, "["+cache.Name+"] "+" 课程不存在")
//	return entity.XueXiTCourse{}, nil
//}

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

// PullCourseChapterAction 获取对应课程的章节信息包括节点信息
func PullCourseChapterAction(cache *xuexitong.XueXiTUserCache, cpi, key int) (chaptersList ChaptersList, ok bool, err error) {
	//拉取对应课程的章节信息
	chapter, err := cache.PullChapter(cpi, key)
	if err != nil {
		return ChaptersList{}, false, errors.New("[" + cache.Name + "] " + " 拉取章节失败")
	}

	var chapterMap map[string]interface{}
	err = json.Unmarshal([]byte(chapter), &chapterMap)
	if err != nil {
		return ChaptersList{}, false, errors.New(fmt.Sprintf("Error parsing JSON: %s", err))
	}
	chapterMapJson, err := json.Marshal(chapterMap["data"])
	if len(chapterMapJson) == 2 {
		return ChaptersList{}, false, errors.New("[" + cache.Name + "] " + "[" + chaptersList.ChatID + "] " + " 课程获取失败")
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
		return ChaptersList{}, ok, errors.New("[" + cache.Name + "] " + "[" + chaptersList.ChatID + "] " + " 无法提取 course")
	}
	data, ok := course["data"].([]interface{})
	if !ok {
		return ChaptersList{}, ok, errors.New("[" + cache.Name + "] " + "[" + chaptersList.ChatID + "] " + " 无法提取 course data")
	}
	if len(data) > 0 {
		knowledge, ok := data[0].(map[string]interface{})["knowledge"].(map[string]interface{})["data"].([]interface{})
		if !ok {
			return ChaptersList{}, ok, errors.New("[" + cache.Name + "] " + "[" + chaptersList.ChatID + "] " + " 无法提取 knowledge data")
		}
		for _, item := range knowledge {
			knowledgeMap := item.(map[string]interface{})
			knowledgeData = append(knowledgeData, knowledgeMap)
		}
	} else {
		return ChaptersList{}, false, errors.New("[" + cache.Name + "] " + "[" + chaptersList.ChatID + "] " + " course data 为空")
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
	chaptersList = ChaptersList{
		ChatID:    chatid,
		Knowledge: knowledgeItems,
	}
	if len(chaptersList.Knowledge) == 0 {
		log2.Print(log2.DEBUG, "["+cache.Name+"] "+"["+chaptersList.ChatID+"] "+" 课程章节为空")
		//return ChaptersList{}, false, err
		return ChaptersList{}, false, errors.New("[" + cache.Name + "] " + "[" + chaptersList.ChatID + "] " + " 课程章节为空")
	}
	// 按照任务点节点重排顺序
	sort.Slice(chaptersList.Knowledge, func(i, j int) bool {
		iLabelParts := strings.Split(chaptersList.Knowledge[i].Label, ".")
		jLabelParts := strings.Split(chaptersList.Knowledge[j].Label, ".")
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
	log2.Print(log2.DEBUG, "["+cache.Name+"] "+"获取课程章节成功 (共 ", log2.Yellow, strconv.Itoa(len(chaptersList.Knowledge)), log2.Default, " 个) ")
	return chaptersList, true, nil
}

type ChapterPointDTO map[string]struct {
	ClickCount    int `json:"clickcount"`    // 是否还有节点
	FinishCount   int `json:"finishcount"`   // 已完成节点
	TotalCount    int `json:"totalcount"`    // 总节点
	OpenLock      int `json:"openlock"`      // 是否有锁
	UnFinishCount int `json:"unfinishcount"` // 未完成节点
}

// updatePointStatus 更新节点状态 单独对应ChaptersList每个KnowledgeItem
func (c *KnowledgeItem) updatePointStatus(chapterPoint ChapterPointDTO) {
	pointData, exists := chapterPoint[fmt.Sprintf("%d", c.ID)]
	if !exists {
		fmt.Printf("Chapter ID %d not found in API response\n", c.ID)
		return
	}
	// 当存在未完成节点 Item 中Total 记录数为未完成数数量
	// TotalCount == 0 没有节点 或者 属于顶级标签
	// 两种条件都不符合 则 记录此章节总结点数量
	if pointData.UnFinishCount != 0 && pointData.TotalCount == 0 {
		c.PointTotal = pointData.UnFinishCount
	} else {
		c.PointTotal = pointData.TotalCount
	}
	c.PointFinished = pointData.FinishCount
}

// ChapterFetchPointAction 对应章节的作业点信息 刷新KnowledgeItem中对应节点完成状态
func ChapterFetchPointAction(cache *xuexitong.XueXiTUserCache,
	nodes []int,
	chapters *ChaptersList,
	clazzID, userID, cpi, courseID int,
) (ChaptersList, error) {
	status, err := cache.FetchChapterPointStatus(nodes, clazzID, userID, cpi, courseID)
	if err != nil {
		log2.Print(log2.DEBUG, "["+cache.Name+"] "+" 获取章节状态失败")
	}
	var cp ChapterPointDTO
	if err := json.NewDecoder(bytes.NewReader([]byte(status))).Decode(&cp); err != nil {
		return ChaptersList{}, fmt.Errorf("failed to decode JSON response: %v", err)
	}

	for i := range chapters.Knowledge {
		chapters.Knowledge[i].updatePointStatus(cp)
	}
	//fmt.Println("任务点状态已更新")
	return *chapters, nil
}
