package examples

import (
	"fmt"
	"log"
	"strconv"
	"testing"
	"time"

	"github.com/yatori-dev/yatori-go-core/aggregation/xuexitong"
	"github.com/yatori-dev/yatori-go-core/aggregation/xuexitong/point"
	"github.com/yatori-dev/yatori-go-core/api/entity"
	xuexitongApi "github.com/yatori-dev/yatori-go-core/api/xuexitong"
	"github.com/yatori-dev/yatori-go-core/global"
	"github.com/yatori-dev/yatori-go-core/utils"
)

// TestLoginXueXiTo 测试学习通登录以及课程数据拉取
func TestLoginXueXiTo(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	user := global.Config.Users[1]
	userCache := xuexitongApi.XueXiTUserCache{
		Name:     user.Account,
		Password: user.Password,
	}
	err := xuexitong.XueXiTLoginAction(&userCache)
	if err != nil {
		log.Fatal(err)
	}
	//拉取课程列表并打印
	action, err := xuexitong.XueXiTPullCourseAction(&userCache)
	if err != nil {
		return
	}
	for _, v := range action {
		fmt.Println(v.ToString())
	}
}

// 测试学习通单课程详情
func TestCourseDetailXueXiTo(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	user := global.Config.Users[1]
	userCache := xuexitongApi.XueXiTUserCache{
		Name:     user.Account,
		Password: user.Password,
	}
	err := xuexitong.XueXiTLoginAction(&userCache)
	if err != nil {
		log.Fatal(err)
	}
	action, err := xuexitong.XueXiTPullCourseAction(&userCache)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(action[0].CourseID)
}

// TestCourseXueXiToChapter 用于测试学习通对应课程章节信息拉取
func TestCourseXueXiToChapter(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	user := global.Config.Users[1]
	userCache := xuexitongApi.XueXiTUserCache{
		Name:     user.Account,
		Password: user.Password,
	}

	err := xuexitong.XueXiTLoginAction(&userCache)
	if err != nil {
		log.Fatal(err)
	}
	//拉取对应课程信息
	course, err := xuexitong.XueXiTPullCourseAction(&userCache)
	var index int
	for i, v := range course {
		if v.ChatID == "261619055656961" {
			index = i
			break
		}
	}
	fmt.Println("name:" + course[index].CourseName)
	fmt.Println("courseID:" + course[index].CourseID)
	//拉取对应课程的章节信息
	key, _ := strconv.Atoi(course[index].Key)
	chapter, err := xuexitong.PullCourseChapterAction(&userCache, course[index].Cpi, key)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("chatid:" + chapter.ChatID)
	for _, item := range chapter.Knowledge {
		fmt.Println("ID:" + strconv.Itoa(item.ID))
		fmt.Println("章节名称:" + item.Name)
		fmt.Println("标签:" + item.Label)
		fmt.Println("层级" + strconv.Itoa(item.Layer))
	}
}

func TestXueXiToChapterPoint(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	user := global.Config.Users[1]
	userCache := xuexitongApi.XueXiTUserCache{
		Name:     user.Account,
		Password: user.Password,
	}

	err := xuexitong.XueXiTLoginAction(&userCache)
	if err != nil {
		log.Fatal(err)
	}
	//拉取对应课程信息
	course, err := xuexitong.XueXiTPullCourseAction(&userCache)
	var index int
	for i, v := range course {
		if v.CourseID == "261619055656961" {
			index = i
			break
		}
	}
	key, _ := strconv.Atoi(course[index].Key)
	action, _ := xuexitong.PullCourseChapterAction(&userCache, course[index].Cpi, key)
	var nodes []int
	for _, item := range action.Knowledge {
		nodes = append(nodes, item.ID)
	}

	userId, _ := strconv.Atoi(userCache.UserID)
	courseId, _ := strconv.Atoi(course[index].CourseID)

	pointAction, err := xuexitong.ChapterFetchPointAction(&userCache,
		nodes,
		&action,
		key, userId, course[index].Cpi, courseId)
	if err != nil {
		log.Fatal(err)
	}
	for i, item := range pointAction.Knowledge {
		fmt.Println(i)
		fmt.Println("ID:" + strconv.Itoa(item.ID))
		fmt.Println("章节名称:" + item.Name)
		fmt.Println("标签:" + item.Label)
		fmt.Println("层级" + strconv.Itoa(item.Layer))
		fmt.Println("总节点或未完成" + strconv.Itoa(item.PointTotal))
		fmt.Println("完成节点" + strconv.Itoa(item.PointFinished))
	}
}

func TestXueXiToChapterCord(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	user := global.Config.Users[1]
	userCache := xuexitongApi.XueXiTUserCache{
		Name:     user.Account,
		Password: user.Password,
	}

	err := xuexitong.XueXiTLoginAction(&userCache)
	if err != nil {
		log.Fatal(err)
	}
	//拉取对应课程信息
	course, err := xuexitong.XueXiTPullCourseAction(&userCache)
	var index int
	for i, v := range course {
		if v.ChatID == "261619055656961" {
			index = i
			break
		}
	}
	if err != nil {
		log.Fatal(err)
	}
	key, _ := strconv.Atoi(course[index].Key)
	action, _ := xuexitong.PullCourseChapterAction(&userCache, course[index].Cpi, key)
	var nodes []int
	for _, item := range action.Knowledge {
		nodes = append(nodes, item.ID)
	}
	courseId, _ := strconv.Atoi(course[index].CourseID)
	_, fetchCards, err := xuexitong.ChapterFetchCardsAction(&userCache, &action, nodes, 27, courseId, key, course[index].Cpi)
	if err != nil {
		log.Fatal(err)
	}
	var (
		videoDTO entity.PointVideoDto
	)
	// 处理返回的任务点对象
	videoDTO = fetchCards[0].PointVideoDto
	videoCourseId, _ := strconv.Atoi(videoDTO.CourseID)
	videoClassId, _ := strconv.Atoi(videoDTO.ClassID)
	if courseId == videoCourseId && key == videoClassId {
		// 测试只对单独一个卡片测试
		card, err := xuexitong.PageMobileChapterCardAction(&userCache, key, courseId, videoDTO.KnowledgeID, videoDTO.CardIndex, course[index].Cpi)
		if err != nil {
			log.Fatal(err)
		}
		videoDTO.AttachmentsDetection(card)
		fmt.Println(videoDTO)
		point.ExecuteVideo(&userCache, &videoDTO)
	} else {
		log.Fatal("任务点对象错误")
	}
}

func TestXueXiToChapterCardWork(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	user := global.Config.Users[1]
	userCache := xuexitongApi.XueXiTUserCache{
		Name:     user.Account,
		Password: user.Password,
	}

	err := xuexitong.XueXiTLoginAction(&userCache)
	if err != nil {
		log.Fatal(err)
	}

	//拉取对应课程信息
	course, err := xuexitong.XueXiTPullCourseAction(&userCache)
	var index int
	for i, v := range course {
		if v.ChatID == "259500784287745" {
			index = i
			break
		}
	}

	key, _ := strconv.Atoi(course[index].Key)
	action, _ := xuexitong.PullCourseChapterAction(&userCache, course[index].Cpi, key)
	var nodes []int
	for _, item := range action.Knowledge {
		nodes = append(nodes, item.ID)
	}
	courseId, _ := strconv.Atoi(course[index].CourseID)
	fmt.Println(course[index].CourseDataID)
	_, fetchCards, err := xuexitong.ChapterFetchCardsAction(&userCache, &action, nodes, 82, courseId,
		key, course[index].Cpi)

	videoDTOs, workDTOs, documentDTOs := entity.ParsePointDto(fetchCards)
	fmt.Println(videoDTOs)
	fmt.Println(workDTOs)
	fmt.Println(documentDTOs)
	videoCourseId, _ := strconv.Atoi(videoDTOs[0].CourseID)
	videoClassId, _ := strconv.Atoi(videoDTOs[0].ClassID)

	if courseId == videoCourseId && key == videoClassId {
		// 测试只对单独一个卡片测试
		card, err := xuexitong.PageMobileChapterCardAction(
			&userCache,
			key,
			courseId,
			videoDTOs[0].KnowledgeID,
			videoDTOs[0].CardIndex,
			course[index].Cpi)
		if err != nil {
			log.Fatal(err)
		}
		videoDTOs[0].AttachmentsDetection(card)
		workDTOs[0].AttachmentsDetection(card)
		fmt.Println(videoDTOs)
		fmt.Println(workDTOs)

		fromAction, _ := xuexitong.WorkPageFromAction(&userCache, &workDTOs[0])

		for _, input := range fromAction {
			fmt.Printf("Name: %s, Value: %s, Type: %s, ID: %s\n", input.Name, input.Value, input.Type, input.ID)
		}

		xuexitong.ParseWorkQuestionAction(&userCache, &workDTOs[0])
	} else {
		log.Fatal("任务点对象错误")
	}
}

func TestXueXiToChapterCardDocument(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	user := global.Config.Users[7]
	userCache := xuexitongApi.XueXiTUserCache{
		Name:     user.Account,
		Password: user.Password,
	}

	err := xuexitong.XueXiTLoginAction(&userCache)
	if err != nil {
		log.Fatal(err)
	}

	courseList, err := xuexitong.XueXiTPullCourseAction(&userCache) //拉取所有课程
	for _, course := range courseList {                             //遍历课程
		key, _ := strconv.Atoi(course.Key)
		action, err := xuexitong.PullCourseChapterAction(&userCache, course.Cpi, key) //获取对应章节信息
		if err != nil {
			log.Fatal(err)
		}
		var nodes []int
		for _, item := range action.Knowledge {
			nodes = append(nodes, item.ID)
		}
		courseId, _ := strconv.Atoi(course.CourseID)
		userId, _ := strconv.Atoi(userCache.UserID)
		// 检测节点完成情况
		pointAction, err := xuexitong.ChapterFetchPointAction(&userCache, nodes, &action, key, userId, course.Cpi, courseId)
		if err != nil {
			log.Fatal(err)
		}
		var isFinished = func(index int) bool {
			if index < 0 || index >= len(pointAction.Knowledge) {
				return false
			}
			i := pointAction.Knowledge[index]
			return i.PointTotal >= 0 && i.PointTotal == i.PointFinished
		}

		for index, item := range nodes {
			if isFinished(index) {
				log.Printf("ID.%d(%s/%s)任务点已完成忽略\n",
					item,
					pointAction.Knowledge[index].Label, pointAction.Knowledge[index].Name)
				time.Sleep(500 * time.Millisecond)
				continue
			}
			_, fetchCards, err := xuexitong.ChapterFetchCardsAction(&userCache, &action, nodes, index, courseId, key, course.Cpi)
			if err != nil {
				log.Fatal(err)
			}
			videoDTOs, workDTOs, documentDTOs := entity.ParsePointDto(fetchCards)
			if videoDTOs == nil && workDTOs == nil && documentDTOs == nil {
				log.Println("没有可学习的内容")
			}
			// 暂时只测试视频
			if documentDTOs != nil {
				for _, documentDTO := range documentDTOs {
					card, err := xuexitong.PageMobileChapterCardAction(
						&userCache, key, courseId, documentDTO.KnowledgeID, documentDTO.CardIndex, course.Cpi)
					if err != nil {
						log.Fatal(err)
					}
					documentDTO.AttachmentsDetection(card)
					point.ExecuteDocument(&userCache, &documentDTO)
					if err != nil {
						log.Fatal(err)
					}
					time.Sleep(5 * time.Second)
				}
			} else {
				log.Println("暂时仅对文档刷取")
			}
		}
	}
}

// 遍历所有课程对应视屏的例子
func TestXueXiToCourseForVideo(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	user := global.Config.Users[5]
	userCache := xuexitongApi.XueXiTUserCache{
		Name:     user.Account,
		Password: user.Password,
	}

	err := xuexitong.XueXiTLoginAction(&userCache)
	if err != nil {
		log.Fatal(err)
	}

	courseList, err := xuexitong.XueXiTPullCourseAction(&userCache) //拉取所有课程
	for _, course := range courseList {                             //遍历课程
		if course.CourseName != "名侦探柯南与化学探秘" {
			continue
		}
		// 6c444b8d5c6203ee2f2aef4b76f5b2ce qrcEnc

		key, _ := strconv.Atoi(course.Key)
		action, err := xuexitong.PullCourseChapterAction(&userCache, course.Cpi, key) //获取对应章节信息
		if err != nil {
			log.Fatal(err)
		}
		var nodes []int
		for _, item := range action.Knowledge {
			nodes = append(nodes, item.ID)
		}
		courseId, _ := strconv.Atoi(course.CourseID)
		userId, _ := strconv.Atoi(userCache.UserID)
		// 检测节点完成情况
		pointAction, err := xuexitong.ChapterFetchPointAction(&userCache, nodes, &action, key, userId, course.Cpi, courseId)
		if err != nil {
			log.Fatal(err)
		}
		var isFinished = func(index int) bool {
			if index < 0 || index >= len(pointAction.Knowledge) {
				return false
			}
			i := pointAction.Knowledge[index]
			return i.PointTotal >= 0 && i.PointTotal == i.PointFinished
		}

		for index, item := range nodes {
			if isFinished(index) {
				log.Printf("ID.%d(%s/%s)任务点已完成忽略\n",
					item,
					pointAction.Knowledge[index].Label, pointAction.Knowledge[index].Name)
				time.Sleep(500 * time.Millisecond)
				continue
			}
			_, fetchCards, err := xuexitong.ChapterFetchCardsAction(&userCache, &action, nodes, index, courseId, key, course.Cpi)
			if err != nil {
				log.Fatal(err)
			}
			videoDTOs, workDTOs, documentDTOs := entity.ParsePointDto(fetchCards)
			if videoDTOs == nil && workDTOs == nil && documentDTOs == nil {
				log.Println("没有可学习的内容")
			}
			// 暂时只测试视频
			if videoDTOs != nil {
				for _, videoDTO := range videoDTOs {
					card, err := xuexitong.PageMobileChapterCardAction(
						&userCache, key, courseId, videoDTO.KnowledgeID, videoDTO.CardIndex, course.Cpi)
					if err != nil {
						log.Fatal(err)
					}
					videoDTO.AttachmentsDetection(card)
					//point.ExecuteVideo(&userCache, &videoDTO) //常规
					point.ExecuteFastVideo(&userCache, &videoDTO) //秒刷
					time.Sleep(5 * time.Second)
				}
			} else {
				log.Println("暂时仅对视频刷取")
			}
		}
	}
}

// 测试扫人脸
func TestFaceQrScan(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	user := global.Config.Users[5]
	userCache := xuexitongApi.XueXiTUserCache{
		Name:     user.Account,
		Password: user.Password,
	}
	err := xuexitong.XueXiTLoginAction(&userCache)
	if err != nil {
		log.Fatal(err)
	}
	//拉取人脸必要数据
	//uuid, qrEnc, err := userCache.GetFaceQrCodeApi(course.CourseID, videoDTO.ClassID, strconv.Itoa(item), strconv.Itoa(course.Cpi))
	//过人脸
	api, _ := userCache.GetCourseFaceQrPlan1Api("245983363", "105533723", "48960c03-5e57-408c-bb97-1718b30032fd", "16eeb4b1d6d733a08785449c8d9784f7", "197217206d33b1af6d8118996e3762f8", "0")
	fmt.Println(api)
	//api, _ := userCache.GetCourseFaceQrApi("2c261aa3-d428-414c-a619-56535f85c8", "105533723")
	//fmt.Println(api)
}
