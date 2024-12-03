package examples

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/yatori-dev/yatori-go-core/api/entity"

	"github.com/yatori-dev/yatori-go-core/aggregation/xuexitong"
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
	xuexitong.XueXiTPullCourseAction(&userCache)
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
	action, err := xuexitong.XueXiTCourseDetailForCourseIdAction(&userCache, "261619055656961")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(action)
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
	course, err := xuexitong.XueXiTCourseDetailForCourseIdAction(&userCache, "261619055656961")
	fmt.Println("name:" + course.CourseName)
	fmt.Println("courseID:" + course.CourseId)
	//拉取对应课程的章节信息
	cpi, _ := strconv.Atoi(course.Cpi)
	key, _ := strconv.Atoi(course.ClassId)
	chapter, err := xuexitong.PullCourseChapterAction(&userCache, cpi, key)
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
	course, err := xuexitong.XueXiTCourseDetailForCourseIdAction(&userCache, "261619055656961")
	if err != nil {
		log.Fatal(err)
	}
	cpi, _ := strconv.Atoi(course.Cpi)
	key, _ := strconv.Atoi(course.ClassId)
	action, _ := xuexitong.PullCourseChapterAction(&userCache, cpi, key)
	var nodes []int
	for _, item := range action.Knowledge {
		nodes = append(nodes, item.ID)
	}

	userId, _ := strconv.Atoi(course.UserId)
	courseId, _ := strconv.Atoi(course.CourseId)

	pointAction, err := xuexitong.ChapterFetchPointAction(&userCache,
		nodes,
		&action,
		key, userId, cpi, courseId)
	if err != nil {
		log.Fatal(err)
	}
	for _, item := range pointAction.Knowledge {
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
	course, err := xuexitong.XueXiTCourseDetailForCourseIdAction(&userCache, "261619055656961")
	if err != nil {
		log.Fatal(err)
	}
	cpi, _ := strconv.Atoi(course.Cpi)
	key, _ := strconv.Atoi(course.ClassId)
	action, _ := xuexitong.PullCourseChapterAction(&userCache, cpi, key)
	var nodes []int
	for _, item := range action.Knowledge {
		nodes = append(nodes, item.ID)
	}
	courseId, _ := strconv.Atoi(course.CourseId)
	_, fetchCards, err := xuexitong.ChapterFetchCardsAction(&userCache, &action, nodes, 1, courseId, key, cpi)
	if err != nil {
		log.Fatal(err)
	}
	var (
		videoDTO entity.PointVideoDto
	)
	// 处理返回的任务点对象
	fmt.Println(fetchCards[0])
	videoDTO = fetchCards[0].PointVideoDto
	videoCourseId, _ := strconv.Atoi(videoDTO.CourseID)
	videoClassId, _ := strconv.Atoi(videoDTO.ClassID)
	if courseId == videoCourseId && key == videoClassId {
		// 测试只对单独一个卡片测试
		card, err := xuexitong.PageMobileChapterCardAction(&userCache, key, courseId, videoDTO.KnowledgeID, videoDTO.CardIndex, cpi)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(card)
		flag, _ := videoDTO.AttachmentsDetection(card)
		if flag {
			fmt.Println(videoDTO)
		}
		state, err := userCache.VideoDtoFetch(&videoDTO)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(state)

	} else {
		log.Fatal("任务点对象错误")
	}
}

// 测试apifox直接生成的请求是否有误乱码现象，测试结果为没有
func TestXueXiToChapterPointPostTest(t *testing.T) {

	url := "https://mooc1-api.chaoxing.com/job/myjobsnodesmap"
	method := "POST"

	payload := strings.NewReader("view=json&nodes=705040658%2C705040670%2C705040660%2C705040672%2C705040678%2C705040685%2C705040689%2C705040695%2C705040699%2C705040703%2C705040706%2C705040710%2C705040711%2C705040714%2C705040715%2C705040717%2C705040719%2C705040721%2C705040723%2C705040724%2C705040727%2C705040729%2C705040730%2C705040732%2C705040734%2C705040736%2C705040738%2C705040740%2C705040742%2C705040661%2C705040673%2C705040680%2C705040686%2C705040691%2C705040696%2C705040744%2C705040662%2C705040675%2C705040682%2C705040688%2C705040693%2C705040698%2C705040702%2C705040704%2C705040708%2C705040745%2C705040663%2C705040677%2C705040664%2C705040668%2C705040683%2C705040666%2C705040746%2C705040667%2C705040747&clazzid=107333284&time=1732895144419&userid=253568561&cpi=283918535&courseid=246742628")

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Cookie", "JSESSIONID=3A20D3BF7DF2C75D4D6C1DA699C6BBDD; ; lv=1; Domain=.chaoxing.com; Expires=Sun, 29-Dec-2024 15:45:45 GMT; fid=7950; _uid=253568561; uf=dff23984ef72c20b30666cf0e946ecc33479c99d647b66eb0739c1c75ceb267e91302755acfa2e35b709492af46199f2913b662843f1f4ad6d92e371d7fdf644912d7e82d6dfb9c83341727ca9b9739ff81fb60e07cc87334d375be2df9727eda85bca34ac51dfd0; _d=1732895145034; UID=253568561; vc=A1F0301E380A6A41CC2FB68F4D00E62F; ; vc2=BBAC913D793C70A546983170BA3C8047; ; vc3=aLoMhatIBUmtIvzzJ7j8GndWR8FRk5ueKc0mmc5fWqL9WeZO2Os7H3BIeAyM%2By9EmpPA9zDkKnhIpIzYGLDKFABXc0DwG2cOFRUrd1H%2Fuu1tj3CgYYhWp44kw7zzFNq%2BY8daoSwXiDQtdyNJYj70bN4cJKF41Z17M%2BxznNNgnAQ%3D51ca1eb4ac4d500089a5f0a57ebeac76; ; cx_p_token=25f58ac703ef16deb0f19556de35b1cc; p_auth_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOiIyNTM1Njg1NjEiLCJsb2dpblRpbWUiOjE3MzI4OTUxNDUwMzUsImV4cCI6MTczMzQ5OTk0NX0.mtlvBRiBMMsZUSDBlFjL6vc2yQdvyO1_CL8qWJOy5Zc; ; xxtenc=6f4d61a148b41994c442e7e73e493404; DSSTASH_LOG=C_38-UN_7293-US_253568561-T_1732895145035; route=26e346b982eea47de2f6652532e77800")
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "mooc1-api.chaoxing.com")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}
