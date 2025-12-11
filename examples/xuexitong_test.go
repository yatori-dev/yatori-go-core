package examples

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/aggregation/xuexitong"
	"github.com/yatori-dev/yatori-go-core/aggregation/xuexitong/point"
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
	videoDTOs, _, _, _, _, _ := xuexitongApi.ParsePointDto(fetchCards)

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

	videoDTOs, workDTOs, documentDTOs, _, _, _ := xuexitongApi.ParsePointDto(fetchCards)
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

			message := xuexitong.AIProblemMessage(q.Type.String(), q.Text, xuexitongApi.ExamTurn{
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
			message := xuexitong.AIProblemMessage(q.Type.String(), q.Text, xuexitongApi.ExamTurn{
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
			message := xuexitong.AIProblemMessage(q.Type.String(), q.Text, xuexitongApi.ExamTurn{
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
			message := xuexitong.AIProblemMessage(q.Type.String(), q.Text, xuexitongApi.ExamTurn{
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

			documentDTOs := xuexitongApi.GroupPointDtos[xuexitongApi.PointDocumentDto](fetchCards, func(dto xuexitongApi.PointDocumentDto) bool {
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
	user := global.Config.Users[71]

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
		if course.CourseName != "形势与政策" {
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
				//if pointAction.Knowledge[index].Label == "6.3" {
				//	fmt.Println("断点")
				//}
			}
			log.Printf("ID.%d(%s/%s)正在执行任务点\n",
				item,
				pointAction.Knowledge[index].Label, pointAction.Knowledge[index].Name)
			//if pointAction.Knowledge[index].Label != "5.2" {
			//	//fmt.Println("断点")
			//	continue
			//}
			//if pointAction.Knowledge[index].Name != "学术评价" {
			//	continue
			//}
			_, fetchCards, err := xuexitong.ChapterFetchCardsAction(&userCache, &action, nodes, index, courseId, key, course.Cpi)

			if err != nil {
				log.Fatal(err)
			}
			videoDTOs, workDTOs, documentDTOs, hyperlinkDTOs, liveDTOs, bbsDTOs := xuexitongApi.ParsePointDto(fetchCards)
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
			if documentDTOs != nil && true {
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
			if workDTOs != nil && false {
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
						message := xuexitong.AIProblemMessage(questionAction.Title, q.Type.String(), xuexitongApi.ExamTurn{
							XueXChoiceQue: *q,
						})

						//aiSetting := global.Config.Setting.AiSetting //获取AI设置
						//q.AnswerAIGet(userCache.UserID, aiSetting.AiUrl, aiSetting.Model, aiSetting.AiType, message, aiSetting.APIKEY)
						q.AnswerXXTAIGet(&userCache, questionAction.ClassId, questionAction.CourseId, questionAction.Cpi, message)
						//q.AnswerExternalGet("http://localhost:8083")
					}
					//判断题
					for i := range questionAction.Judge {
						q := &questionAction.Judge[i] // 获取对应选项
						message := xuexitong.AIProblemMessage(questionAction.Title, q.Type.String(), xuexitongApi.ExamTurn{
							XueXJudgeQue: *q,
						})

						//aiSetting := global.Config.Setting.AiSetting //获取AI设置
						//q.AnswerAIGet(userCache.UserID, aiSetting.AiUrl, aiSetting.Model, aiSetting.AiType, message, aiSetting.APIKEY)
						q.AnswerXXTAIGet(&userCache, questionAction.ClassId, questionAction.CourseId, questionAction.Cpi, message)
					}
					//填空题
					for i := range questionAction.Fill {
						q := &questionAction.Fill[i] // 获取对应选项
						message := xuexitong.AIProblemMessage(questionAction.Title, q.Type.String(), xuexitongApi.ExamTurn{
							XueXFillQue: *q,
						})
						//aiSetting := global.Config.Setting.AiSetting //获取AI设置
						//q.AnswerAIGet(userCache.UserID, aiSetting.AiUrl, aiSetting.Model, aiSetting.AiType, message, aiSetting.APIKEY)
						q.AnswerXXTAIGet(&userCache, questionAction.ClassId, questionAction.CourseId, questionAction.Cpi, message)
						//q.AnswerExternalGet("http://localhost:8083")
					}
					//简答题
					for i := range questionAction.Short {
						q := &questionAction.Short[i] // 获取对应选项
						message := xuexitong.AIProblemMessage(questionAction.Title, q.Type.String(), xuexitongApi.ExamTurn{
							XueXShortQue: *q,
						})
						//aiSetting := global.Config.Setting.AiSetting //获取AI设置
						//q.AnswerAIGet(userCache.UserID, aiSetting.AiUrl, aiSetting.Model, aiSetting.AiType, message, aiSetting.APIKEY)
						q.AnswerXXTAIGet(&userCache, questionAction.ClassId, questionAction.CourseId, questionAction.Cpi, message)
						//q.AnswerExternalGet("http://localhost:8083")
					}
					//名词解释
					for i := range questionAction.TermExplanation {
						q := &questionAction.TermExplanation[i] // 获取对应选项
						message := xuexitong.AIProblemMessage(questionAction.Title, q.Type.String(), xuexitongApi.ExamTurn{
							XueXTermExplanationQue: *q,
						})
						//aiSetting := global.Config.Setting.AiSetting //获取AI设置
						//q.AnswerAIGet(userCache.UserID, aiSetting.AiUrl, aiSetting.Model, aiSetting.AiType, message, aiSetting.APIKEY)
						q.AnswerXXTAIGet(&userCache, questionAction.ClassId, questionAction.CourseId, questionAction.Cpi, message)
					}
					//论述题
					for i := range questionAction.Essay {
						q := &questionAction.Essay[i] // 获取对应选项
						message := xuexitong.AIProblemMessage(questionAction.Title, q.Type.String(), xuexitongApi.ExamTurn{
							XueXEssayQue: *q,
						})
						aiSetting := global.Config.Setting.AiSetting //获取AI设置
						q.AnswerAIGet(userCache.UserID, aiSetting.AiUrl, aiSetting.Model, aiSetting.AiType, message, aiSetting.APIKEY)
					}

					//连线题题
					for i := range questionAction.Matching {
						q := &questionAction.Matching[i] // 获取对应选项
						message := xuexitong.AIProblemMessage(questionAction.Title, q.Type.String(), xuexitongApi.ExamTurn{
							XueXMatchingQue: *q,
						})
						aiSetting := global.Config.Setting.AiSetting //获取AI设置
						q.AnswerAIGet(userCache.UserID, aiSetting.AiUrl, aiSetting.Model, aiSetting.AiType, message, aiSetting.APIKEY)
					}

					//其他题（按论述题方式进行）
					for i := range questionAction.Other {
						q := &questionAction.Other[i] // 获取对应选项
						message := xuexitong.AIProblemMessage(questionAction.Title, q.Type.String(), xuexitongApi.ExamTurn{
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

// 考试测试
func TestXueXiToExam(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	user := global.Config.Users[68]

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

		if course.CourseName != "大学教育" {
			continue
		}
		examList, err1 := xuexitong.PullExamListAction(&userCache, course)
		if err1 != nil {
			log.Fatal(err1)
		}
		// 打印结果
		for _, exam := range examList {
			if exam.Status != "待做" {
				continue
			}
			err2 := xuexitong.EnterExamAction(&userCache, &exam)
			if err2 != nil {
				log.Fatal(err2)
			}
			err3 := xuexitong.PullExamPaperAction(&userCache, &exam)
			if err3 != nil {
				log.Fatal(err3)
			}

		}
	}
}

// 试卷题目截取测试
func TestXXTExamPaperPull(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	paperHtml, err := utils.ReadFileAsString("./学习通考试试题页面.html")
	if err != nil {
		log.Fatal(err)
	}
	_, err1 := xuexitong.HtmlPaperTurnEntity(paperHtml)
	if err1 != nil {
		log.Fatal(err1)
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

// inf_enc加密参数
func TestXXTBBsInf_enc(t *testing.T) {
	sign := InfEncSign(map[string]string{
		"token":     "4faa8662c59590c6f43ae9fe5b002b42",
		"_time":     "1765296675494",
		"_c_0_":     "7dd0787e7a0f4b9c93f14fa26e9728cb",
		"puid":      "346635955",
		"uuid":      "fa1968e2-b717-4547-bdab-4d2d1fe37bf4",
		"tag":       "classId134204187",
		"maxW":      "1080",
		"topicUUID": "da22aad94ce2410a99984067d25eeb10",
		"anonymous": "0",
	}, []string{"token", "_time", "_c_0_", "puid", "uuid", "tag", "maxW", "topicUUID", "anonymous"})
	jmstr := `token=4faa8662c59590c6f43ae9fe5b002b42&_time=1765296675494&_c_0_=7dd0787e7a0f4b9c93f14fa26e9728cb&puid=346635955&uuid=fa1968e2-b717-4547-bdab-4d2d1fe37bf4&tag=classId134204187&maxW=1080&topicUUID=da22aad94ce2410a99984067d25eeb10&anonymous=0&DESKey=Z(AfY@XS`
	sum := md5.Sum([]byte(jmstr))
	fmt.Println(hex.EncodeToString(sum[:]))
	fmt.Println(sign)
}

//func dfs(strs []string,res []string){
//	sign := InfEncSign()
//	if
//}

// InfEncSign 移动端为参数添加 inf_enc 签名
func InfEncSign(params map[string]string, order []string) string {
	const DESKey = "Z(AfY@XS"

	parts := make([]string, 0, len(order))
	for _, k := range order {
		// 跳过不存在的 key（或你也可以要求都存在）
		v, ok := params[k]
		if !ok {
			continue
		}
		// 使用 url.QueryEscape 与 Python urlencode 行为兼容（空格 -> +）
		parts = append(parts, k+"="+url.QueryEscape(v))
	}

	// 拼接并加上 DESKey
	query := strings.Join(parts, "&") + "&DESKey=" + DESKey

	// md5
	sum := md5.Sum([]byte(query))
	return hex.EncodeToString(sum[:])
}

// 移动端_c_0参数生成
func ParamFor_c_0_Generete() string {
	u := uuid.New()
	c0 := strings.ReplaceAll(u.String(), "-", "")
	return c0
}
func MakeParams(orig map[string]string) map[string]string {
	params := make(map[string]string)
	for k, v := range orig {
		params[k] = v
	}

	// 生成 _c_0_：UUID，无 “-”
	u := uuid.New()
	c0 := strings.ReplaceAll(u.String(), "-", "")
	params["_c_0_"] = c0

	// time
	params["_time"] = fmt.Sprintf("%d", time.Now().UnixMilli())

	// 然后计算 inf_enc（按 key 排序 urlencode + DESKey 拼接 MD5）
	// ... (你之前实现 inf_enc 的逻辑) ...

	return params
}

// 手机端阅读Enc参数生成
func TestReadEncParam(t *testing.T) {
	jsonBody, err := BuildSpecialReport(
		"256268467",        // userId
		"218403954",        // courseId
		"437042861",        // chapterId
		"132232726",        // classId
		"c8f68a62e7ef7fa3", // deviceId
		1626,               // wordCount
		11,                 // interactCount
		271,                // readSeconds
		map[string]any{ // event 内容（你随便填）
			"ts":     time.Now().UnixMilli(),
			"scroll": 1200,
		},
	)
	settings := map[string]any{
		"f": "readPoint",
		"u": "339543304",
		"s": "",
		"d": url.QueryEscape(`{"a":null,"r":"218403954,437042861","t":"special","l":1,"f":"0","wc":1626,"ic":11,"v":2,"s":2,"h":271,"e":"H4sIAAAAAAAAA32SS5LDIAxET5NtSv/Pdub+dxokHBw7YVhQFH4WrVbHg36RYeyiOXZ48E+cd+Z1tqf1AoYI9vzCqE7Gk4OYhF6MFGNOY6eYkJqkjA+26mQzWIzbk8cSVtDQlMXQyQB3HUdXSkp/MVRKhP29j3knUufs0sxpiWiEi7HFMO565e6SZ80JCRIrK56NVAmVNg1iB8GCFC/lxa4j0LrzuiPeyYrFqO9GgG1dNUmGG3vnCKTfgpvpbZ5vTG+7xfhducxglH48gmEZohoIuqD5Y8ky3XilnR6InvocHwWJgwOvOrYYDPjO9GAOZnjZKdBwE5OQawwCP+MT3tH8LxoSdT7e/6zds9TEdvLqr7YFSLLxoGVNyHUnfXoZl/hgyyoH6a512nrLyx9Cn92TDAQAAA==","ext":"{\"_from_\":\"256268467_132232726_339543304_c8f68a62e7ef7fa3a704d6b031d19697\",\"rtag\":\"1054242600_477554005_read-218403954\"}"}`),
		"t": "20251211204612158",
	}

	enc := GenerateEnc(settings)

	fmt.Println(enc)

	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(jsonBody)
}

// 生成 enc 参数（完全匹配超星 JS）
func GenerateEnc(g map[string]any) string {
	// Step 1: key 排序
	keys := make([]string, 0, len(g))
	for k := range g {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Step 2: 拼接 value
	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(toString(g[k]))
	}
	h := sb.String()

	// Step 3: 加盐 MD5
	final := h + "NrRzLDpWB2JkeodIVAn4"

	sum := md5.Sum([]byte(final))
	return strings.ToUpper(hex.EncodeToString(sum[:]))
}

// 把 interface{} 转换成 JS 里的字符串表现形式
func toString(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case int, int32, int64, float32, float64:
		return fmt.Sprintf("%v", val)
	default:
		// JSON.stringify 等效
		b, _ := json.Marshal(val)
		return string(b)
	}
}

// 生成与 JS 完全一致的时间戳 yyyyMMddHHmmssSSS
func GenT() string {
	now := time.Now()
	return now.Format("20060102150405") + fmt.Sprintf("%03d", now.Nanosecond()/1e6)
}

// ReportBody 上报 JSON 的结构体
type ReportBody struct {
	A   any    `json:"a"`
	R   string `json:"r"` // courseId,chapterId
	T   string `json:"t"` // "special"
	L   int    `json:"l"` // level
	F   string `json:"f"`
	Wc  int    `json:"wc"`  // 字数
	Ic  int    `json:"ic"`  // 行为次数
	V   int    `json:"v"`   // 协议版本
	S   int    `json:"s"`   // 状态
	H   int    `json:"h"`   // 阅读秒数
	E   string `json:"e"`   // gzip + base64 数据
	Ext string `json:"ext"` // ext json (string)
}

// Gzip+Base64
func gzipBase64(input []byte) (string, error) {
	var buf bytes.Buffer
	g := gzip.NewWriter(&buf)
	_, err := g.Write(input)
	if err != nil {
		return "", err
	}
	g.Close()
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// 自动生成 ext 字段
func buildExt(userId, courseId, classId, device string, chapterId string) string {
	ext := map[string]string{
		"_from_": fmt.Sprintf("%s_%s_%s_%s", userId, courseId, classId, device),
		"rtag":   fmt.Sprintf("%s_%s_read-%s", userId, courseId, chapterId),
	}
	b, _ := json.Marshal(ext)
	return string(b)
}

// 构建上报 JSON（你以后只用填参数就能生成）
func BuildSpecialReport(
	userId string,
	courseId string,
	chapterId string,
	classId string,
	deviceId string,
	wordCount int,
	interactCount int,
	readSeconds int,
	event map[string]any,
) (string, error) {

	// 将 event 做 gzip + base64
	eventBytes, _ := json.Marshal(event)
	eStr, err := gzipBase64(eventBytes)
	if err != nil {
		return "", err
	}

	// 打 ext
	extStr := buildExt(userId, courseId, classId, deviceId, chapterId)

	body := ReportBody{
		A:   nil,
		R:   fmt.Sprintf("%s,%s", courseId, chapterId),
		T:   "special",
		L:   1,
		F:   "0",
		Wc:  wordCount,
		Ic:  interactCount,
		V:   2,
		S:   2,
		H:   readSeconds,
		E:   eStr,
		Ext: extStr,
	}

	jsonBytes, _ := json.Marshal(body)
	return string(jsonBytes), nil
}
