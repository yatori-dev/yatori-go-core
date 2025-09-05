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
	user := global.Config.Users[0]
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
	user := global.Config.Users[0]
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
		if v.CourseName == "软件工程" {
			index = i
			break
		}
	}
	fmt.Println("name:" + course[index].CourseName)
	fmt.Println("courseID:" + course[index].CourseID)
	//拉取对应课程的章节信息
	key, _ := strconv.Atoi(course[index].Key)
	chapter, _, err := xuexitong.PullCourseChapterAction(&userCache, course[index].Cpi, key)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("chatid:" + chapter.ChatID)

	for i, item := range chapter.Knowledge {
		fmt.Println(i)
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
		if v.CourseName == "形势与政策" {
			index = i
			break
		}
	}
	key, _ := strconv.Atoi(course[index].Key)
	action, _, _ := xuexitong.PullCourseChapterAction(&userCache, course[index].Cpi, key)
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
	user := global.Config.Users[13]
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
		if v.CourseName == "现代仪器分析技术" {
			index = i
			break
		}
	}
	if err != nil {
		log.Fatal(err)
	}
	key, _ := strconv.Atoi(course[index].Key)
	action, _, _ := xuexitong.PullCourseChapterAction(&userCache, course[index].Cpi, key)
	var nodes []int
	for _, item := range action.Knowledge {
		nodes = append(nodes, item.ID)
	}
	courseId, _ := strconv.Atoi(course[index].CourseID)
	_, fetchCards, err := xuexitong.ChapterFetchCardsAction(&userCache, &action, nodes, 7, courseId, key, course[index].Cpi)
	if err != nil {
		log.Fatal(err)
	}
	//var (
	//	videoDTO entity.PointVideoDto
	//)
	// 处理返回的任务点对象
	videoDTOs, _, _ := entity.ParsePointDto(fetchCards)

	//card3, err := xuexitong.PageMobileChapterCardAction(
	//	&userCache, key, courseId, videoDTOs[3].KnowledgeID, videoDTOs[3].CardIndex, course[index].Cpi)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//videoDTOs[3].AttachmentsDetection(card3)
	//fmt.Println(videoDTOs[3])
	//
	//card4, err := xuexitong.PageMobileChapterCardAction(
	//	&userCache, key, courseId, videoDTOs[4].KnowledgeID, videoDTOs[4].CardIndex, course[index].Cpi)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//videoDTOs[4].AttachmentsDetection(card4)
	//fmt.Println(videoDTOs[4])
	for _, videoDTO := range videoDTOs {
		card, _, err := xuexitong.PageMobileChapterCardAction(
			&userCache, key, courseId, videoDTO.KnowledgeID, videoDTO.CardIndex, course[index].Cpi)
		if err != nil {
			log.Fatal(err)
		}
		videoDTO.AttachmentsDetection(card)
		fmt.Println(videoDTO)
	}
	fmt.Println(videoDTOs)
	//videoDTO = fetchCards[0].PointVideoDto
	//videoCourseId, _ := strconv.Atoi(videoDTO.CourseID)
	//videoClassId, _ := strconv.Atoi(videoDTO.ClassID)
	//if courseId == videoCourseId && key == videoClassId {
	//	// 测试只对单独一个卡片测试
	//	card, err := xuexitong.PageMobileChapterCardAction(&userCache, key, courseId, videoDTO.KnowledgeID, videoDTO.CardIndex, course[index].Cpi)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//	videoDTO.AttachmentsDetection(card)
	//	fmt.Println(videoDTO)
	//	point.ExecuteVideo(&userCache, &videoDTO)
	//} else {
	//	log.Fatal("任务点对象错误")
	//}
}

// 测试拉取作业
func TestXueXiToChapterCardWork(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	user := global.Config.Users[0]
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
		if v.CourseName == "软件工程" {
			index = i
			break
		}
	}

	key, _ := strconv.Atoi(course[index].Key)
	action, _, _ := xuexitong.PullCourseChapterAction(&userCache, course[index].Cpi, key)
	var nodes []int
	for _, item := range action.Knowledge {
		nodes = append(nodes, item.ID)
	}
	courseId, _ := strconv.Atoi(course[index].CourseID)
	fmt.Println(course[index].CourseDataID)
	_, fetchCards, err := xuexitong.ChapterFetchCardsAction(&userCache, &action, nodes, 51, courseId,
		key, course[index].Cpi)

	videoDTOs, workDTOs, documentDTOs := entity.ParsePointDto(fetchCards)
	fmt.Println(videoDTOs)
	fmt.Println(workDTOs)
	fmt.Println(documentDTOs)
	videoCourseId, _ := strconv.Atoi(workDTOs[0].CourseID)
	videoClassId, _ := strconv.Atoi(workDTOs[0].ClassID)

	if courseId == videoCourseId && key == videoClassId {
		// 测试只对单独一个卡片测试
		card, _, err := xuexitong.PageMobileChapterCardAction(
			&userCache,
			key,
			courseId,
			workDTOs[0].KnowledgeID,
			workDTOs[0].CardIndex,
			course[index].Cpi)
		if err != nil {
			log.Fatal(err)
		}
		workDTOs[0].AttachmentsDetection(card)
		fmt.Println(workDTOs)

		fromAction, _ := xuexitong.WorkPageFromAction(&userCache, &workDTOs[0])

		for _, input := range fromAction {
			fmt.Printf("Name: %s, Value: %s, Type: %s, ID: %s\n", input.Name, input.Value, input.Type, input.ID)
		}

		questionAction := xuexitong.ParseWorkQuestionAction(&userCache, &workDTOs[0])
		for i := range questionAction.Choice {
			q := &questionAction.Choice[i] // 获取指向切片元素的指针

			message := xuexitong.AIProblemMessage(q.Type.String(), q.Text, entity.ExamTurn{
				XueXChoiceQue: *q,
			})
			aiSetting := global.Config.Setting.AiSetting
			q.AnswerAIGet(userCache.UserID,
				aiSetting.AiUrl, aiSetting.Model, aiSetting.AiType, message, aiSetting.APIKEY)
		}
		for i, que := range questionAction.Choice {
			fmt.Println(fmt.Sprintf("%d. %v", i, que.Answers))
		}

		for i := range questionAction.Fill {
			q := &questionAction.Fill[i]
			message := xuexitong.AIProblemMessage(q.Type.String(), q.Text, entity.ExamTurn{
				XueXFillQue: *q,
			})
			aiSetting := global.Config.Setting.AiSetting
			q.AnswerAIGet(userCache.UserID,
				aiSetting.AiUrl, aiSetting.Model, aiSetting.AiType, message, aiSetting.APIKEY)
		}
		for i, que := range questionAction.Fill {
			for j := range que.OpFromAnswer {
				fmt.Println(fmt.Sprintf("%d%v. %v", i, j, que.OpFromAnswer[j]))
			}
		}
		for i := range questionAction.Judge {
			q := &questionAction.Judge[i]
			message := xuexitong.AIProblemMessage(q.Type.String(), q.Text, entity.ExamTurn{
				XueXJudgeQue: *q,
			})
			aiSetting := global.Config.Setting.AiSetting
			q.AnswerAIGet(userCache.UserID,
				aiSetting.AiUrl, aiSetting.Model, aiSetting.AiType, message, aiSetting.APIKEY)
		}
		for i, que := range questionAction.Judge {
			fmt.Println(fmt.Sprintf("%d. %v", i, que.Answers))
		}

		for i := range questionAction.Short {
			q := &questionAction.Short[i]
			message := xuexitong.AIProblemMessage(q.Type.String(), q.Text, entity.ExamTurn{
				XueXShortQue: *q,
			})
			aiSetting := global.Config.Setting.AiSetting
			q.AnswerAIGet(userCache.UserID,
				aiSetting.AiUrl, aiSetting.Model, aiSetting.AiType, message, aiSetting.APIKEY)
		}
		for i, que := range questionAction.Short {
			for j := range que.OpFromAnswer {
				fmt.Println(fmt.Sprintf("%d%v. %v", i, j, que.OpFromAnswer[j]))
			}
		}
		answerAction := xuexitong.WorkNewSubmitAnswerAction(&userCache, questionAction, true)
		println(answerAction)
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
		action, _, err := xuexitong.PullCourseChapterAction(&userCache, course.Cpi, key) //获取对应章节信息
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
			//videoDTOs, workDTOs, documentDTOs := entity.ParsePointDto(fetchCards)
			//if videoDTOs == nil && workDTOs == nil && documentDTOs == nil {
			//	log.Println("没有可学习的内容")
			//}

			documentDTOs := entity.GroupPointDtos[entity.PointDocumentDto](fetchCards, func(dto entity.PointDocumentDto) bool {
				return dto.IsSetted()
			})
			// 暂时只测试视频
			if documentDTOs != nil {
				for _, documentDTO := range documentDTOs {
					card, _, err := xuexitong.PageMobileChapterCardAction(
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

// 遍历所有课程并刷取
func TestXueXiToFlushCourse(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	user := global.Config.Users[25]

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

		if course.CourseName != "算法与程序的奥秘" {
			continue
		}
		// 6c444b8d5c6203ee2f2aef4b76f5b2ce qrcEnc

		key, _ := strconv.Atoi(course.Key)
		action, ok, err := xuexitong.PullCourseChapterAction(&userCache, course.Cpi, key) //获取对应章节信息
		fmt.Println(action.ChatID)
		if err != nil {
			log.Fatal(err)
		}
		if !ok {
			continue
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
			log.Printf("ID.%d(%s/%s)正在执行任务点\n",
				item,
				pointAction.Knowledge[index].Label, pointAction.Knowledge[index].Name)

			_, fetchCards, err := xuexitong.ChapterFetchCardsAction(&userCache, &action, nodes, index, courseId, key, course.Cpi)

			if err != nil {
				log.Fatal(err)
			}
			videoDTOs, workDTOs, documentDTOs := entity.ParsePointDto(fetchCards)
			if videoDTOs == nil && workDTOs == nil && documentDTOs == nil {
				log.Println("没有可学习的内容")
			}

			// 视频刷取
			if videoDTOs != nil && true {
				for _, videoDTO := range videoDTOs {
					card, enc, err := xuexitong.PageMobileChapterCardAction(
						&userCache, key, courseId, videoDTO.KnowledgeID, videoDTO.CardIndex, course.Cpi)
					if err != nil {
						log.Fatal(err)
					}
					videoDTO.AttachmentsDetection(card)
					if videoDTO.IsPassed == true { //过滤完成的
						continue
					}
					videoDTO.Enc = enc
					point.ExecuteVideo(&userCache, &videoDTO, key, course.Cpi) //常规
					//point.ExecuteFastVideo(&userCache, &videoDTO) //秒刷
					time.Sleep(5 * time.Second)
				}
			}
			// 文档刷取
			if documentDTOs != nil && true {
				for _, documentDTO := range documentDTOs {
					card, _, err := xuexitong.PageMobileChapterCardAction(
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
			}
			//作业刷取
			if workDTOs != nil && false {
				for _, workDTO := range workDTOs {

					//以手机端拉取章节卡片数据
					mobileCard, _, _ := xuexitong.PageMobileChapterCardAction(&userCache, key, courseId, workDTO.KnowledgeID, workDTO.CardIndex, course.Cpi)
					workDTO.AttachmentsDetection(mobileCard)
					fromAction, _ := xuexitong.WorkPageFromAction(&userCache, &workDTO)
					for _, input := range fromAction {
						fmt.Printf("Name: %s, Value: %s, Type: %s, ID: %s\n", input.Name, input.Value, input.Type, input.ID)
					}
					questionAction := xuexitong.ParseWorkQuestionAction(&userCache, &workDTO)
					fmt.Println(questionAction)
					for i := range questionAction.Choice {
						q := &questionAction.Choice[i] // 获取对应选项
						message := xuexitong.AIProblemMessage(q.Type.String(), "", entity.ExamTurn{
							XueXChoiceQue: *q,
						})
						aiSetting := global.Config.Setting.AiSetting //获取AI设置
						q.AnswerAIGet(userCache.UserID, aiSetting.AiUrl, aiSetting.Model, aiSetting.AiType, message, aiSetting.APIKEY)
					}
					for i, que := range questionAction.Choice {
						fmt.Println(fmt.Sprintf("%d. %v", i, que.Answers))
					}
					xuexitong.WorkNewSubmitAnswerAction(&userCache, questionAction, false)
				}
			}

		}
	}
}

// 测试扫人脸
func TestFaceQrScan(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	user := global.Config.Users[10]
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
	uuid, qrEnc, err := userCache.GetFaceQrCodeApi1("245211886", "103839833", "891797235", "283918535")
	fmt.Println(uuid, qrEnc, err)
	//过人脸
	api, _ := userCache.GetCourseFaceQrPlan1Api("245211886", "103839833", uuid, "16eeb4b1d6d733a08785449c8d9784f7", qrEnc, "0")
	fmt.Println(api)
	//api, _ := userCache.GetCourseFaceQrApi("2c261aa3-d428-414c-a619-56535f85c8", "105533723")
	//fmt.Println(api)
}

// 测试扫人脸
func TestFaceQrScanPlan1(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	user := global.Config.Users[21]
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
		if course.CourseName != "中国传统玉文化与美玉鉴赏" {
			continue
		}
		// 6c444b8d5c6203ee2f2aef4b76f5b2ce qrcEnc

		key, _ := strconv.Atoi(course.Key)
		action, ok, err := xuexitong.PullCourseChapterAction(&userCache, course.Cpi, key) //获取对应章节信息
		fmt.Println(action.ChatID)
		if err != nil {
			log.Fatal(err)
		}
		if !ok {
			continue
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
			log.Printf("ID.%d(%s/%s)正在执行任务点\n",
				item,
				pointAction.Knowledge[index].Label, pointAction.Knowledge[index].Name)

			_, fetchCards, err := xuexitong.ChapterFetchCardsAction(&userCache, &action, nodes, index, courseId, key, course.Cpi)

			if err != nil {
				log.Fatal(err)
			}
			videoDTOs, workDTOs, documentDTOs := entity.ParsePointDto(fetchCards)
			if videoDTOs == nil && workDTOs == nil && documentDTOs == nil {
				log.Println("没有可学习的内容")
			}

			// 视频刷取
			if videoDTOs != nil && true {
				for _, videoDTO := range videoDTOs {
					card, _, err := xuexitong.PageMobileChapterCardAction(
						&userCache, key, courseId, videoDTO.KnowledgeID, videoDTO.CardIndex, course.Cpi)
					if err != nil {
						log.Fatal(err)
					}
					videoDTO.AttachmentsDetection(card)
					point.ExecuteVideo(&userCache, &videoDTO, key, course.Cpi) //常规
					//point.ExecuteFastVideo(&userCache, &videoDTO) //秒刷
					time.Sleep(5 * time.Second)
				}
			}
		}
	}
}
