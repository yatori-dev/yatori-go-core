package examples

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/aggregation/xuexitong"
	"github.com/yatori-dev/yatori-go-core/aggregation/xuexitong/point"
	"github.com/yatori-dev/yatori-go-core/api/entity"
	xuexitongApi "github.com/yatori-dev/yatori-go-core/api/xuexitong"
	"github.com/yatori-dev/yatori-go-core/global"
	"github.com/yatori-dev/yatori-go-core/utils"
	log2 "github.com/yatori-dev/yatori-go-core/utils/log"
	"golang.org/x/net/html"
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
	videoDTOs, _, _, _, _, _ := entity.ParsePointDto(fetchCards)

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

	videoDTOs, workDTOs, documentDTOs, _, _, _ := entity.ParsePointDto(fetchCards)
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

		questionAction, err1 := xuexitong.ParseWorkQuestionAction(&userCache, &workDTOs[0])
		if err1 != nil && strings.Contains(err1.Error(), "已截止，不能作答") {
			fmt.Println("该试卷已截止，已自动跳过")
			return
		}
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
		answerAction, _ := xuexitong.WorkNewSubmitAnswerAction(&userCache, questionAction, true)
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

// 解析HTML数字实体
func TestUnicodeToText(t *testing.T) {
	unicode := html.UnescapeString(`9.4&#23567;&#33410;&#27979;&#39564;`)
	fmt.Println(unicode)
}

// 遍历所有课程并刷取
func TestXueXiToFlushCourse(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	user := global.Config.Users[66]

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

		//if course.CourseName != "解读中国经济发展的密码" {
		//	continue
		//}
		if course.CourseName != "大学生礼仪与形象设计" {
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
			if i.PointTotal == 0 && i.PointFinished == 0 {
				err2 := xuexitong.EnterChapterForwardCallAction(&userCache, strconv.Itoa(courseId), strconv.Itoa(key), strconv.Itoa(pointAction.Knowledge[index].ID), strconv.Itoa(course.Cpi))
				if err2 != nil {
					log.Fatal(err2)
				}
				return false
			}
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
			if pointAction.Knowledge[index].Label != "6.5" {
				//fmt.Println("断点")
				continue
			}
			_, fetchCards, err := xuexitong.ChapterFetchCardsAction(&userCache, &action, nodes, index, courseId, key, course.Cpi)

			if err != nil {
				log.Fatal(err)
			}
			videoDTOs, workDTOs, documentDTOs, hyperlinkDTOs, liveDTOs, bbsDTOs := entity.ParsePointDto(fetchCards)
			if videoDTOs == nil && workDTOs == nil && documentDTOs == nil && hyperlinkDTOs == nil && liveDTOs == nil && bbsDTOs == nil {
				log.Println("没有可学习的内容")
			}

			// 视频刷取
			if videoDTOs != nil && false {
				for _, videoDTO := range videoDTOs {
					card, enc, err := xuexitong.PageMobileChapterCardAction(
						&userCache, key, courseId, videoDTO.KnowledgeID, videoDTO.CardIndex, course.Cpi)

					if err != nil {
						log.Fatal(err)
					}
					videoDTO.AttachmentsDetection(card)
					//if videoDTO.IsPassed == true { //过滤完成的
					//	continue
					//}
					if !videoDTO.IsJob {
						fmt.Println("(", videoDTO.Title, ")", "该视频为非任务点，已自动跳过")
						continue
					}
					videoDTO.Enc = enc
					point.ExecuteVideoTest(&userCache, &videoDTO, key, course.Cpi) //常规
					//point.ExecuteFastVideo(&userCache, &videoDTO) //秒刷
					time.Sleep(5 * time.Second)
				}
			}
			// 文档刷取
			if documentDTOs != nil && false {
				for _, documentDTO := range documentDTOs {
					card, _, err := xuexitong.PageMobileChapterCardAction(
						&userCache, key, courseId, documentDTO.KnowledgeID, documentDTO.CardIndex, course.Cpi)
					if err != nil {
						log.Fatal(err)
					}
					documentDTO.AttachmentsDetection(card)

					if !documentDTO.IsJob {
						log.Printf("(%s)该文档非任务点或已完成，已自动跳过\n", documentDTO.Title)
						continue
					}
					document, err1 := point.ExecuteDocument(&userCache, &documentDTO)
					if err1 != nil {
						log.Fatal(err1)
					}
					log2.Print(log2.INFO, "(", documentDTO.Title, ")", document)
					time.Sleep(5 * time.Second)
				}
			}
			//作业刷取
			if workDTOs != nil && true {
				for _, workDTO := range workDTOs {

					//以手机端拉取章节卡片数据
					mobileCard, _, _ := xuexitong.PageMobileChapterCardAction(&userCache, key, courseId, workDTO.KnowledgeID, workDTO.CardIndex, course.Cpi)

					workDTO.AttachmentsDetection(mobileCard)
					//fromAction, _ := xuexitong.WorkPageFromAction(&userCache, &workDTO)
					//for _, input := range fromAction {
					//	fmt.Printf("Name: %s, Value: %s, Type: %s, ID: %s\n", input.Name, input.Value, input.Type, input.ID)
					//}
					questionAction, err1 := xuexitong.ParseWorkQuestionAction(&userCache, &workDTO)
					if err1 != nil && strings.Contains(err1.Error(), "已截止，不能作答") {
						fmt.Println("该试卷已截止，已自动跳过")
						continue
					}
					fmt.Println(questionAction)
					for i := range questionAction.Choice {
						q := &questionAction.Choice[i] // 获取对应选项
						message := xuexitong.AIProblemMessage(questionAction.Title, q.Type.String(), entity.ExamTurn{
							XueXChoiceQue: *q,
						})

						aiSetting := global.Config.Setting.AiSetting //获取AI设置
						q.AnswerAIGet(userCache.UserID, aiSetting.AiUrl, aiSetting.Model, aiSetting.AiType, message, aiSetting.APIKEY)
						//q.AnswerExternalGet("http://localhost:8083")
					}
					//判断题
					for i := range questionAction.Judge {
						q := &questionAction.Judge[i] // 获取对应选项
						message := xuexitong.AIProblemMessage(questionAction.Title, q.Type.String(), entity.ExamTurn{
							XueXJudgeQue: *q,
						})

						aiSetting := global.Config.Setting.AiSetting //获取AI设置
						q.AnswerAIGet(userCache.UserID, aiSetting.AiUrl, aiSetting.Model, aiSetting.AiType, message, aiSetting.APIKEY)
					}
					//填空题
					for i := range questionAction.Fill {
						q := &questionAction.Fill[i] // 获取对应选项
						message := xuexitong.AIProblemMessage(questionAction.Title, q.Type.String(), entity.ExamTurn{
							XueXFillQue: *q,
						})
						aiSetting := global.Config.Setting.AiSetting //获取AI设置
						q.AnswerAIGet(userCache.UserID, aiSetting.AiUrl, aiSetting.Model, aiSetting.AiType, message, aiSetting.APIKEY)
						//q.AnswerExternalGet("http://localhost:8083")
					}
					//简答题
					for i := range questionAction.Short {
						q := &questionAction.Short[i] // 获取对应选项
						message := xuexitong.AIProblemMessage(questionAction.Title, q.Type.String(), entity.ExamTurn{
							XueXShortQue: *q,
						})
						aiSetting := global.Config.Setting.AiSetting //获取AI设置
						q.AnswerAIGet(userCache.UserID, aiSetting.AiUrl, aiSetting.Model, aiSetting.AiType, message, aiSetting.APIKEY)
						//q.AnswerExternalGet("http://localhost:8083")
					}
					//名词解释
					for i := range questionAction.TermExplanation {
						q := &questionAction.TermExplanation[i] // 获取对应选项
						message := xuexitong.AIProblemMessage(questionAction.Title, q.Type.String(), entity.ExamTurn{
							XueXTermExplanationQue: *q,
						})
						aiSetting := global.Config.Setting.AiSetting //获取AI设置
						q.AnswerAIGet(userCache.UserID, aiSetting.AiUrl, aiSetting.Model, aiSetting.AiType, message, aiSetting.APIKEY)
					}
					//论述题
					for i := range questionAction.Essay {
						q := &questionAction.Essay[i] // 获取对应选项
						message := xuexitong.AIProblemMessage(questionAction.Title, q.Type.String(), entity.ExamTurn{
							XueXEssayQue: *q,
						})
						aiSetting := global.Config.Setting.AiSetting //获取AI设置
						q.AnswerAIGet(userCache.UserID, aiSetting.AiUrl, aiSetting.Model, aiSetting.AiType, message, aiSetting.APIKEY)
					}

					//连线题题
					for i := range questionAction.Matching {
						q := &questionAction.Matching[i] // 获取对应选项
						message := xuexitong.AIProblemMessage(questionAction.Title, q.Type.String(), entity.ExamTurn{
							XueXMatchingQue: *q,
						})
						aiSetting := global.Config.Setting.AiSetting //获取AI设置
						q.AnswerAIGet(userCache.UserID, aiSetting.AiUrl, aiSetting.Model, aiSetting.AiType, message, aiSetting.APIKEY)
					}

					//其他题（按论述题方式进行）
					for i := range questionAction.Other {
						q := &questionAction.Other[i] // 获取对应选项
						message := xuexitong.AIProblemMessage(questionAction.Title, q.Type.String(), entity.ExamTurn{
							XueXOtherQue: *q,
						})
						aiSetting := global.Config.Setting.AiSetting //获取AI设置
						q.AnswerAIGet(userCache.UserID, aiSetting.AiUrl, aiSetting.Model, aiSetting.AiType, message, aiSetting.APIKEY)
					}

					xuexitong.AnswerFixedPattern(questionAction.Choice, questionAction.Judge)
					answerAction, _ := xuexitong.WorkNewSubmitAnswerAction(&userCache, questionAction, true)
					fmt.Printf("%s答题完成，返回信息：%s\n", questionAction.Title, answerAction)
				}
			}
			//外链任务点刷取
			if hyperlinkDTOs != nil && true {
				for _, hyperlinkDTO := range hyperlinkDTOs {
					card, _, err := xuexitong.PageMobileChapterCardAction(
						&userCache, key, courseId, hyperlinkDTO.KnowledgeID, hyperlinkDTO.CardIndex, course.Cpi)
					if err != nil {
						log.Fatal(err)
					}
					hyperlinkDTO.AttachmentsDetection(card)

					document, err1 := point.ExecuteHyperlink(&userCache, &hyperlinkDTO)
					if err1 != nil {
						log.Fatal(err1)
					}
					log2.Print(log2.INFO, "(", hyperlinkDTO.Title, ")", document)
					time.Sleep(5 * time.Second)
				}
			}
			//直播任务
			if liveDTOs != nil && true {
				for _, liveDTO := range liveDTOs {
					card, _, err := xuexitong.PageMobileChapterCardAction(
						&userCache, key, courseId, liveDTO.KnowledgeID, liveDTO.CardIndex, course.Cpi)
					if err != nil {
						log.Fatal(err)
					}
					liveDTO.AttachmentsDetection(card)
					if !liveDTO.IsJob {
						log.Printf("(%s)该直播非任务点或已完成，已自动跳过\n", liveDTO.Title)
						continue
					}
					point.ExecuteLiveTest(&userCache, &liveDTO)
					time.Sleep(5 * time.Second)
				}
			}

			//直播任务
			if bbsDTOs != nil && false {
				for _, bbsDTO := range bbsDTOs {
					card, _, err := xuexitong.PageMobileChapterCardAction(
						&userCache, key, courseId, bbsDTO.KnowledgeID, bbsDTO.CardIndex, course.Cpi)
					if err != nil {
						log.Fatal(err)
					}
					bbsDTO.AttachmentsDetection(card)
					if !bbsDTO.IsJob {
						log.Printf("(%s)该讨论非任务点或已完成，已自动跳过\n", bbsDTO.Title)
						continue
					}
					aiSetting := global.Config.Setting.AiSetting //获取AI设置
					point.ExecuteBbsTest(&userCache, &bbsDTO, aiSetting)
					//point.ExecuteLiveTest(&userCache, &liveDTO)
					time.Sleep(5 * time.Second)
				}
			}
		}
	}
}

// 测试拉取人脸照片
func TestPullFaceImg(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	user := global.Config.Users[55]
	userCache := xuexitongApi.XueXiTUserCache{
		Name:     user.Account,
		Password: user.Password,
	}
	err := xuexitong.XueXiTLoginAction(&userCache)
	if err != nil {
		log.Fatal(err)
	}
	//拉取用户照片
	_, img, _ := userCache.GetHistoryFaceImg("")
	utils.SaveImageAsJPEG(img, "./assets/18106919661.jpg")
}

// 测试扫人脸
func TestFaceQrScan(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	user := global.Config.Users[56]
	userCache := xuexitongApi.XueXiTUserCache{
		Name:     user.Account,
		Password: user.Password,
	}
	err := xuexitong.XueXiTLoginAction(&userCache)
	if err != nil {
		log.Fatal(err)
	}
	//拉取用户照片
	_, img, _ := userCache.GetHistoryFaceImg("")
	//上传人脸
	//disturbImage := utils.ImageRGBDisturb(img)
	disturbImage := utils.ImageRGBDisturbAdjust(img, 15)
	//disturbImage := utils.ProcessImageDisturb(img)
	//获取token
	tokenJson, err := userCache.GetFaceUpLoadToken()
	token := gojsonq.New().JSONString(tokenJson).Find("_token").(string)
	ObjectId, err := userCache.UploadFaceImageApi(token, disturbImage)

	//plan3是点击进入课程时候的人脸识别
	planApi, err := userCache.GetCourseFaceQrPlan3Api("128609334", "255665643", "c3ece4f5-e61b-454e-931a-e52a5124cc57", "18120f93a46ed8a614c6923c834b5e18", "492936718", ObjectId)
	//planApi, err := userCache.PassFaceQrPlanPhoneApi("128609334", "255665643", "44736e4b27f8981de19e222bf29c969d", "492936718", ObjectId)
	log.Println(planApi, err)
}

func TestXXTQuestionSelect(t *testing.T) {
	resSe := []string{"1", "2", "3", "4"}

	options := make(map[string]string)
	letter := []string{"A", "B", "C", "D", "E", "F"}
	for i, _ := range resSe {
		options[letter[i]] = ""
	}
	options["C"] = "选项1"
	options["D"] = "选项2"
	options["A"] = "选项3"
	options["B"] = "选项4"
	for k, v := range options {
		fmt.Println(k, v)
	}
}
