package examples

import (
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"testing"
	time2 "time"

	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/aggregation/yinghua"
	yinghuaApi "github.com/yatori-dev/yatori-go-core/api/yinghua"
	"github.com/yatori-dev/yatori-go-core/global"
	"github.com/yatori-dev/yatori-go-core/utils"
	log2 "github.com/yatori-dev/yatori-go-core/utils/log"
)

// 账号登录测试
func TestLogin(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	user := global.Config.Users[0]
	cache := yinghuaApi.YingHuaUserCache{
		PreUrl:   user.URL,
		Account:  user.Account,
		Password: user.Password,
	}
	error := yinghua.YingHuaLoginAction(&cache)
	if error != nil {

	}
}

// 测试获取课程列表
func TestPullCourseList(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	user := global.Config.Users[0]
	cache := yinghuaApi.YingHuaUserCache{
		PreUrl:   user.URL,
		Account:  user.Account,
		Password: user.Password,
	}

	error := yinghua.YingHuaLoginAction(&cache)
	if error != nil {

	}
	list, _ := yinghua.CourseListAction(&cache)
	for _, item := range list {
		log2.Print(log2.INFO, "课程：", item.Id, " ", item.Name, " ", strconv.FormatFloat(item.Progress, 'b', 5, 32), " ", item.StartDate.String(), " ", strconv.Itoa(item.VideoCount), " ", strconv.Itoa(item.VideoLearned))

	}
}

// 测试拉取对应课程的视屏列表
func TestPullCourseVideoList(t *testing.T) {
	utils.YatoriCoreInit()
	log2.NOWLOGLEVEL = log2.INFO //设置日志登记为DEBUG
	//测试账号
	setup()
	user := global.Config.Users[17]
	cache := yinghuaApi.YingHuaUserCache{
		PreUrl:   user.URL,
		Account:  user.Account,
		Password: user.Password,
		//IpProxySW: true,
		//ProxyIP:   "http://localhost:7899",
	}

	err := yinghua.YingHuaLoginAction(&cache)
	if err != nil {

	}
	list, _ := yinghua.CourseListAction(&cache)
	for _, courseItem := range list {
		log2.Print(log2.INFO, " ", courseItem.Id, " ", courseItem.Name, " ", strconv.FormatFloat(courseItem.Progress, 'b', 5, 32), " ", courseItem.StartDate.String(), " ", strconv.Itoa(courseItem.VideoCount), " ", strconv.Itoa(courseItem.VideoLearned))
		videoList, _ := yinghua.VideosListAction(&cache, courseItem) //拉取视屏列表动作
		for _, videoItem := range videoList {
			log2.Print(log2.INFO, " ", "视屏：", videoItem.CourseId, " ", videoItem.Id, " ", videoItem.Name, " ", strconv.Itoa(int(videoItem.VideoDuration)))
		}
	}

}

// 用于登录保活
func keepAliveLogin(UserCache yinghuaApi.YingHuaUserCache) {
	ticker := time2.NewTicker(time2.Second * 60)
	for {
		select {
		case <-ticker.C:
			api := yinghuaApi.KeepAliveApi(UserCache, 8)
			log2.Print(log2.INFO, " ", "登录保活状态：", api)
		}
	}
	//for {
	//	api := yinghuaApi.KeepAliveApi(UserCache)
	//	log2.Print(log2.INFO, " ", "登录保活状态：", api)
	//	time2.Sleep(time2.Second * 60)
	//}
}

var wg sync.WaitGroup

// 刷视频的抽离函数
func videoListStudy(UserCache yinghuaApi.YingHuaUserCache, course yinghua.YingHuaCourse) {
	videoList, _ := yinghua.VideosListAction(&UserCache, course) //拉取对应课程的视屏列表

	// 提交学时
	for _, video := range videoList {
		log2.Print(log2.INFO, " ", video.Name)
		time := video.ViewedDuration //设置当前观看时间为最后看视屏的时间
		studyId := "0"
		for {
			if video.Progress == 100 {
				break //如果看完了，也就是进度为100那么直接跳过
			}
			sub, _ := yinghuaApi.SubmitStudyTimeApi(UserCache, video.Id, studyId, time, 5, nil) //提交学时
			if gojsonq.New().JSONString(sub).Find("msg") != "提交学时成功!" {
				time2.Sleep(5 * time2.Second)
				continue
			}
			resStudyId := gojsonq.New().JSONString(sub).Find("result.data.studyId")
			if resStudyId != nil {
				studyId = strconv.Itoa(int(resStudyId.(float64)))
			}

			log2.Print(log2.INFO, " ", video.Name, " ", "提交状态：", gojsonq.New().JSONString(sub).Find("msg").(string), " ", "观看时间：", strconv.Itoa(time)+"/"+strconv.Itoa(video.VideoDuration), " ", "观看进度：", fmt.Sprintf("%.2f", float32(time)/float32(video.VideoDuration)*100), "%")
			time += 5
			time2.Sleep(5 * time2.Second)
			if time > video.VideoDuration {
				break //如果看完该视屏则直接下一个
			}
		}
	}
	wg.Done()
}

// 测试获取指定视屏并且刷课
func TestBrushOneLesson(t *testing.T) {
	utils.YatoriCoreInit()
	log2.NOWLOGLEVEL = log2.INFO //设置日志登记为DEBUG
	//测试账号
	setup()
	user := global.Config.Users[17]
	cache := yinghuaApi.YingHuaUserCache{
		PreUrl:   user.URL,
		Account:  user.Account,
		Password: user.Password,
		//IpProxySW: true,
		//ProxyIP:   "http://localhost:7899",
	}

	err := yinghua.YingHuaLoginAction(&cache) // 登录
	if err != nil {
		log.Fatal(err) //登录失败则直接退出
	}
	go keepAliveLogin(cache)                    //携程保活
	list, _ := yinghua.CourseListAction(&cache) //拉取课程列表
	for _, item := range list {
		wg.Add(1)
		log2.Print(log2.INFO, " ", item.Id, " ", item.Name, " ", strconv.FormatFloat(item.Progress, 'b', 5, 32), " ", item.StartDate.String(), " ", strconv.Itoa(item.VideoCount), " ", strconv.Itoa(item.VideoLearned))
		go videoListStudy(cache, item) //多携程刷课
	}
	wg.Wait()
}

// 测试获取单个课程的详细信息
func TestCourseDetail(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	user := global.Config.Users[47]
	cache := yinghuaApi.YingHuaUserCache{
		PreUrl:   user.URL,
		Account:  user.Account,
		Password: user.Password,
	}

	err := yinghua.YingHuaLoginAction(&cache) // 登录
	if err != nil {
		log.Fatal(err) //登录失败则直接退出
	}
	fmt.Println(cache.GetToken())
	action, _ := yinghua.CourseDetailAction(&cache, "1012027")
	fmt.Println(action)
	if err != nil {
		log.Fatal(err)
	}

}

// 测试获取考试的信息
func TestExamDetail(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	user := global.Config.Users[47]
	cache := yinghuaApi.YingHuaUserCache{
		PreUrl:   user.URL,
		Account:  user.Account,
		Password: user.Password,
	}
	fmt.Println(cache)

	err := yinghua.YingHuaLoginAction(&cache) // 登录
	if err != nil {
		log.Fatal(err) //登录失败则直接退出
	}
	list, _ := yinghua.CourseListAction(&cache) //拉取课程列表
	//list[0]
	action, err := yinghua.VideosListAction(&cache, list[0])
	if err != nil {
		log.Fatal(err)
	}
	for _, node := range action {
		if node.Name != "期末考试（补考）" {
			continue
		}
		fmt.Println(node)
		//api := yinghuaApi.ExamDetailApi(cache, node.Id)
		detailAction, _ := yinghua.ExamDetailAction(&cache, node.Id)
		yinghua.StartExamAction(&cache, detailAction[0], global.Config.Setting.AiSetting.AiUrl, global.Config.Setting.AiSetting.Model, global.Config.Setting.AiSetting.APIKEY, global.Config.Setting.AiSetting.AiType, 0)
		//{"_code":9,"status":false,"msg":"考试测试时间还未开始","result":{}}
		exam, _ := yinghuaApi.StartExam(cache, node.CourseId, node.Id, detailAction[0].ExamId, 3, nil)
		fmt.Println(detailAction)
		fmt.Println(exam)
	}
}

// 测试获取作业信息并写作业
func TestWorkDetail(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	user := global.Config.Users[48]
	cache := yinghuaApi.YingHuaUserCache{
		PreUrl:   user.URL,
		Account:  user.Account,
		Password: user.Password,
	}

	err := yinghua.YingHuaLoginAction(&cache) // 登录
	if err != nil {
		log.Fatal(err) //登录失败则直接退出
	}
	list, _ := yinghua.CourseListAction(&cache) //拉取课程列表
	//list[0]
	action, err := yinghua.VideosListAction(&cache, list[0])
	if err != nil {
		log.Fatal(err)
	}
	for _, node := range action {
		if node.Name != "JAVA程序设计A-作业" {
			continue
		}
		fmt.Println(node)
		//获取作业详细信息
		detailAction, _ := yinghua.WorkDetailAction(&cache, node.Id)
		////{"_code":9,"status":false,"msg":"考试测试时间还未开始","result":{}}
		//开始写作业
		yinghua.StartWorkAction(&cache, detailAction[0], global.Config.Setting.AiSetting.AiUrl, global.Config.Setting.AiSetting.Model, global.Config.Setting.AiSetting.APIKEY, global.Config.Setting.AiSetting.AiType, 0)
		fmt.Println(detailAction)
		//打印最终分数
		s, error := yinghua.WorkedFinallyScoreAction(&cache, detailAction[0])
		if error != nil {
			log.Fatal(error)
		}
		fmt.Println("最高分：", s)
	}
}

// 测试外部挂载题库
func TestApiQueBack(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	user := global.Config.Users[12]
	cache := yinghuaApi.YingHuaUserCache{
		PreUrl:   user.URL,
		Account:  user.Account,
		Password: user.Password,
	}

	error := yinghua.YingHuaLoginAction(&cache) // 登录
	if error != nil {
		log.Fatal(error) //登录失败则直接退出
	}
	list, _ := yinghua.CourseListAction(&cache) //拉取课程列表
	//list[0]
	action, error := yinghua.VideosListAction(&cache, list[0])
	if error != nil {
		log.Fatal(error)
	}
	for _, node := range action {
		if node.Name != "绪论作业" {
			continue
		}
		fmt.Println(node)
		//获取作业详细信息
		detailAction, _ := yinghua.WorkDetailAction(&cache, node.Id)
		////{"_code":9,"status":false,"msg":"考试测试时间还未开始","result":{}}
		//开始写作业
		//yinghua.StartWorkAction(&cache, detailAction[0], global.Config.Setting.AiSetting.AiUrl, global.Config.Setting.AiSetting.Model, global.Config.Setting.AiSetting.APIKEY, global.Config.Setting.AiSetting.AiType, 1)
		yinghua.StartWorkForExternalAction(&cache, "http://localhost:8083", detailAction[0], 0)
		fmt.Println(detailAction)
		//打印最终分数
		s, error := yinghua.WorkedFinallyScoreAction(&cache, detailAction[0])
		if error != nil {
			log.Fatal(error)
		}
		fmt.Println("最高分：", s)
	}
}

// 测试账号是否可以正常链接
func TestApi(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	user := global.Config.Users[16]
	cache := yinghuaApi.YingHuaUserCache{
		PreUrl:   user.URL,
		Account:  user.Account,
		Password: user.Password,
	}
	urlStr := cache.PreUrl + fmt.Sprintf("/service/code?r=%d", time2.Now().Unix())
	method := "GET"

	tr := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse("http://localhost:7899") // 设置代理
		},
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}
	client := &http.Client{
		Timeout:   30 * time2.Second,
		Transport: tr,
	}

	req, err := http.NewRequest(method, urlStr, nil)
	//req.Header.Add("Cookie", cache.cookie)

	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("User-Agent", utils.DefaultUserAgent)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	fmt.Println(string(body))
}

// 测试验证码图片拉取
func TestCapterImg(t *testing.T) {
	// 设置随机种子并生成 [0,1) 随机小数
	rand.Seed(time2.Now().UnixNano())
	r := fmt.Sprintf("%.16f", rand.Float64())
	urlStr := fmt.Sprintf("https://bwgl.qiankj.com/service/code?r=%s", r)

	// 构建请求
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		log.Fatalf("请求创建失败: %v", err)
	}

	// 设置真实浏览器常见的 User-Agent
	req.Header.Set("User-Agent", utils.DefaultUserAgent)
	req.Header.Set("Accept", "image/avif,image/webp,image/apng,image/*,*/*;q=0.8")
	req.Header.Set("Referer", "https://bwgl.qiankj.com/")
	req.Header.Set("Connection", "keep-alive")

	// 发起请求
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 跳过证书验证，仅用于开发环境
		},
	}

	//如果开启了IP代理，那么就直接添加代理

	tr.Proxy = func(req *http.Request) (*url.URL, error) {
		return url.Parse("http://localhost:7899") // 设置代理
	}

	client := &http.Client{
		Timeout:   30 * time2.Second,
		Transport: tr,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("请求发送失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应是否为200 OK
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("请求失败，状态码: %d", resp.StatusCode)
	}

	// 打印 Content-Type 用于确认图片类型
	contentType := resp.Header.Get("Content-Type")
	fmt.Println("响应 Content-Type:", contentType)

	// 保存图片
	outFile, err := os.Create("./assets/code/captcha.png")
	if err != nil {
		log.Fatalf("创建文件失败: %v", err)
	}
	defer outFile.Close()

	// 正确复制响应体到文件
	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		log.Fatalf("写入文件失败: %v", err)
	}

	fmt.Println("✅ 验证码图片保存成功: captcha.png")
}
