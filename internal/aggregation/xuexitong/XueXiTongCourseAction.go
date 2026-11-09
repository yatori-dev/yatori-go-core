package xuexitong

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/v3/internal/api/xuexitong"
	"github.com/yatori-dev/yatori-go-core/v3/utils"
	log2 "github.com/yatori-dev/yatori-go-core/v3/utils/log"
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
	CourseDataID   int     `json:"courseDataId"`
	ContentID      int     `json:"ContentID"`
	IsStart        bool    `json:"isstart"`        //是否开课了，开课了为true，没开课为false。一般来说开课才能刷，不然刷不了的
	State          int     `json:"state"`          //课程状态，0为正常，1为课程已经结束
	JobFinishCount int     `json:"jobFinishCount"` //完成的任务点数量
	JobCount       int     `json:"jobCount"`       //一共多少任务点
	JobRate        float64 `json:"jobRate"`        //完成进度
}

func (x *XueXiTCourse) ToString() string {
	return fmt.Sprintf(
		"XueXiTCourse{Cpi: %d, Key: %v, CourseID: %s,Teacher: %s, CourseName: %s, CourseImage: %s\nCourseDataID: %d, ContentID: %d}",
		x.Cpi, x.Key, x.CourseID, x.CourseTeacher, x.CourseName, x.CourseImage, x.CourseDataID, x.ContentID,
	)
}

// 拉取学习通所有课程列表并返回
func XueXiTPullCourseAction(cache *xuexitong.XueXiTUserCache) ([]XueXiTCourse, error) {
	return XueXiTPullCourseActionContext(context.Background(), cache)
}

func XueXiTPullCourseActionContext(ctx context.Context, cache *xuexitong.XueXiTUserCache) ([]XueXiTCourse, error) {
	courses, err := cache.CourseListApiContext(ctx, 3, nil)
	if err != nil {
		return nil, fmt.Errorf("[%s] 拉取课程失败: %w", cache.Name, err)
	}

	//判断是否触发验证码
	if strings.Contains(courses, "输入验证码") {
		log2.Print(log2.DEBUG, utils.RunFuncName(), "触发验证码，正在进行AI智能识别绕过.....")
		if err := passImageCaptcha(ctx, cache); err != nil {
			return nil, fmt.Errorf("[%s] 课程验证码处理失败: %w", cache.Name, err)
		}
		courses, err = cache.CourseListApiContext(ctx, 3, nil)
		if err != nil {
			return nil, fmt.Errorf("[%s] 验证码通过后拉取课程失败: %w", cache.Name, err)
		}
		log2.Print(log2.DEBUG, utils.RunFuncName(), "绕过成功")
	}

	var xueXiTCourse xuexitong.XueXiTCourseJson
	err = json.Unmarshal([]byte(courses), &xueXiTCourse)
	if err != nil {
		return nil, fmt.Errorf("课程列表响应无法解析: %w", err)
	}
	log2.Print(log2.DEBUG, "["+cache.Name+"] "+" 课程数量："+strconv.Itoa(len(xueXiTCourse.ChannelList)))
	// log2.Print(log2.INFO, "["+cache.Name+"] "+courses)

	var courseList = make([]XueXiTCourse, 0)
	keyCourse := map[string]int{}
	for i, channel := range xueXiTCourse.ChannelList {
		var flag = false
		if channel.Content.Course.Data == nil && i >= 0 && i < len(xueXiTCourse.ChannelList) {
			//xueXiTCourse.ChannelList = append(xueXiTCourse.ChannelList[:i], xueXiTCourse.ChannelList[i+1:]...)
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

		cpi := channel.Cpi
		if cpi <= 0 {
			cpi = channel.Content.Cpi
		}
		course := XueXiTCourse{
			Cpi:           cpi,
			Key:           classId,
			CourseID:      courseID,
			ChatID:        channel.Content.Chatid,
			CourseTeacher: teacher,
			CourseName:    courseName,
			CourseImage:   courseImage,
			CourseDataID:  courseDataID,
			ContentID:     channel.Content.Id,
			IsStart:       channel.Content.Isstart,
			State:         channel.Content.State,
		}
		for _, cr := range courseList {
			if cr.Key == course.Key {
				flag = true
				break
			}
		}
		if flag {
			continue
		}
		keyCourse[course.Key] = len(courseList) //添加映射，方便后续处理
		courseList = append(courseList, course)
	}

	// 拉取课程完成度状态。
	// 课程多时接口对 clazzPersonStr 的长度/内容很敏感，这里分批查询；
	// 某一批失败只影响该批进度，不再让整个课程列表拉取失败。
	fillCourseStatuses(ctx, cache, courseList, keyCourse)

	return courseList, nil
}

type ChaptersList struct {
	ChatID    string          `json:"chatid"`
	IsStart   bool            `json:"isstart"` //是否开始
	Bbsid     string          `json:"bbsid"`
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
	return PullCourseChapterActionContext(context.Background(), cache, cpi, key)
}

func PullCourseChapterActionContext(ctx context.Context, cache *xuexitong.XueXiTUserCache, cpi, key int) (chaptersList ChaptersList, ok bool, err error) {
	//拉取对应课程的章节信息
	chapter, err := cache.PullChapterContext(ctx, cpi, key, 3, nil)
	if err != nil {
		return ChaptersList{}, false, errors.New("[" + cache.Name + "] " + " 拉取章节失败" + err.Error())
	}

	type chapterKnowledge struct {
		JobCount   int    `json:"jobcount"`
		IsReview   int    `json:"isreview"`
		IndexOrder int    `json:"indexorder"`
		Name       string `json:"name"`
		ID         int    `json:"id"`
		Label      string `json:"label"`
		Layer      int    `json:"layer"`
		ParentNode int    `json:"parentnodeid"`
		Status     string `json:"status"`
		Attachment struct {
			Data []interface{} `json:"data"`
		} `json:"attachment"`
	}
	var response struct {
		Data []struct {
			ChatID  string `json:"chatid"`
			IsStart bool   `json:"isstart"`
			BbsID   string `json:"bbsid"`
			Course  struct {
				Data []struct {
					Knowledge struct {
						Data []chapterKnowledge `json:"data"`
					} `json:"knowledge"`
				} `json:"data"`
			} `json:"course"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(chapter), &response); err != nil {
		return ChaptersList{}, false, fmt.Errorf("[%s] 解析课程章节失败: %w", cache.Name, err)
	}
	if len(response.Data) == 0 {
		return ChaptersList{}, false, fmt.Errorf("[%s] 课程章节响应缺少 data", cache.Name)
	}
	chapterData := response.Data[0]
	if len(chapterData.Course.Data) == 0 {
		return ChaptersList{}, false, fmt.Errorf("[%s] 课程章节响应缺少 course.data", cache.Name)
	}

	knowledgeData := chapterData.Course.Data[0].Knowledge.Data
	knowledgeItems := make([]KnowledgeItem, 0, len(knowledgeData))
	for _, item := range knowledgeData {
		knowledgeItems = append(knowledgeItems, KnowledgeItem{
			JobCount:     item.JobCount,
			IsReview:     item.IsReview,
			Attachment:   item.Attachment.Data,
			IndexOrder:   item.IndexOrder,
			Name:         item.Name,
			ID:           item.ID,
			Label:        item.Label,
			Layer:        item.Layer,
			ParentNodeID: item.ParentNode,
			Status:       item.Status,
		})
	}
	chaptersList = ChaptersList{
		ChatID:    chapterData.ChatID,
		IsStart:   chapterData.IsStart,
		Bbsid:     chapterData.BbsID,
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
	return ChapterFetchPointActionContext(context.Background(), cache, nodes, chapters, clazzID, userID, cpi, courseID)
}

func ChapterFetchPointActionContext(ctx context.Context, cache *xuexitong.XueXiTUserCache,
	nodes []int,
	chapters *ChaptersList,
	clazzID, userID, cpi, courseID int,
) (ChaptersList, error) {
	status, err := cache.FetchChapterPointStatusContext(ctx, nodes, clazzID, userID, cpi, courseID, 3, nil)
	if err != nil {
		return ChaptersList{}, fmt.Errorf("[%s] 获取章节状态失败: %w", cache.Name, err)
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

const courseStatusBatchSize = 20

// fillCourseStatuses 分批拉取课程完成度。单批失败会拆半重试，
// 最坏情况下只跳过该课程的进度，不影响课程列表本身。
func fillCourseStatuses(ctx context.Context, cache *xuexitong.XueXiTUserCache, courseList []XueXiTCourse, keyCourse map[string]int) {
	if len(courseList) == 0 {
		return
	}
	pending := make(map[int][]int, 4)
	for i := range courseList {
		if strings.TrimSpace(courseList[i].Key) == "" || courseList[i].Cpi <= 0 {
			continue
		}
		pending[courseList[i].Cpi] = append(pending[courseList[i].Cpi], i)
	}
	failed := 0
	for _, indexes := range pending {
		for start := 0; start < len(indexes); start += courseStatusBatchSize {
			end := start + courseStatusBatchSize
			if end > len(indexes) {
				end = len(indexes)
			}
			_, batchFailed := fillCourseStatusBatch(ctx, cache, courseList, keyCourse, indexes[start:end])
			failed += batchFailed
		}
	}
	if failed > 0 {
		log2.Print(log2.INFO, "["+cache.Name+"] ", log2.BoldRed,
			"有 ", strconv.Itoa(failed), " 门课程的完成状态拉取失败，进度按 0% 处理")
	}
}

func fillCourseStatusBatch(ctx context.Context, cache *xuexitong.XueXiTUserCache, courseList []XueXiTCourse, keyCourse map[string]int, indexes []int) (okCount, failedCount int) {
	if len(indexes) == 0 {
		return 0, 0
	}
	err := fetchCourseStatusBatch(ctx, cache, courseList, keyCourse, indexes)
	if err == nil {
		return len(indexes), 0
	}
	if len(indexes) == 1 {
		log2.Print(log2.DEBUG, "["+cache.Name+"] 拉取课程完成状态失败，已跳过 ",
			courseList[indexes[0]].CourseName, ": ", err)
		return 0, 1
	}
	// 拆半定位坏掉的 clazzPersonStr，避免一条坏数据拖垮整批。
	mid := len(indexes) / 2
	okLeft, failedLeft := fillCourseStatusBatch(ctx, cache, courseList, keyCourse, indexes[:mid])
	okRight, failedRight := fillCourseStatusBatch(ctx, cache, courseList, keyCourse, indexes[mid:])
	return okLeft + okRight, failedLeft + failedRight
}

func fetchCourseStatusBatch(ctx context.Context, cache *xuexitong.XueXiTUserCache, courseList []XueXiTCourse, keyCourse map[string]int, indexes []int) error {
	parts := make([]string, 0, len(indexes))
	for _, index := range indexes {
		parts = append(parts, fmt.Sprintf("%s_%d", courseList[index].Key, courseList[index].Cpi))
	}
	courseStatusListJson, err := cache.CourseCompleteStatusApiContext(ctx, strings.Join(parts, ","), 5, nil)
	if err != nil {
		return err
	}
	courseStatusList := gojsonq.New().JSONString(courseStatusListJson).Find("jobArray")
	if courseStatusList == nil {
		log2.Print(log2.DEBUG, "["+cache.Name+"] 无法拉取课程任务点进度数据: ", courseStatusListJson)
		return nil
	}
	items, ok := courseStatusList.([]interface{})
	if !ok {
		return nil
	}
	for _, item := range items {
		statusData, ok := item.(map[string]any)
		if !ok {
			continue
		}
		clazzID := courseStatusString(statusData["clazzId"])
		if clazzID == "" {
			continue
		}
		keyed, ok := keyCourse[clazzID]
		if !ok || keyed < 0 || keyed >= len(courseList) {
			continue
		}
		courseList[keyed].JobFinishCount = int(courseStatusFloat(statusData["jobFinishCount"]))
		courseList[keyed].JobRate = courseStatusFloat(statusData["jobRate"])
		courseList[keyed].JobCount = int(courseStatusFloat(statusData["jobCount"]))
	}
	return nil
}

func courseStatusString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case float64:
		return strconv.Itoa(int(v))
	case int:
		return strconv.Itoa(v)
	case json.Number:
		return v.String()
	default:
		return ""
	}
}

func courseStatusFloat(value any) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case json.Number:
		f, _ := v.Float64()
		return f
	default:
		return 0
	}
}
