package yinghua

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/yatori-dev/yatori-go-core/api/entity"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/thedevsaddam/gojsonq"
	yinghuaApi "github.com/yatori-dev/yatori-go-core/api/yinghua"
	"github.com/yatori-dev/yatori-go-core/models/ctype"
	"github.com/yatori-dev/yatori-go-core/utils"
	"github.com/yatori-dev/yatori-go-core/utils/log"
)

// 课程必要数据得截取
type YingHuaCourse struct {
	Id           string    //课程ID
	Name         string    //课程名称
	Mode         int       //课程模式
	StartDate    time.Time //开始时间
	EndDate      time.Time //结束时间
	Progress     float64   //学习进度
	VideoCount   int       //视屏总数
	VideoLearned int       //已学习视屏数量
}

// 英华节点
type YingHuaNode struct {
	Id            string    //视屏Id
	CourseId      string    //对应课程的ID
	Name          string    //视屏名称
	VideoDuration int       //视屏时长
	NodeLock      int       //视屏解锁状态
	UnlockTime    time.Time //视屏解锁时间
	Progress      float32   //观看进度
	//Duration       int       //视屏时长
	ViewedDuration int  //观看时长
	State          int  //视屏状态
	TabVideo       bool //是否有视屏
	TabFile        bool //是否有文件
	TabVote        bool //是否有投票
	TabWork        bool //是否有作业
	TabExam        bool //是否有考试
}

// 考试节点信息
type YingHuaExam struct {
	Id          string    //ID
	ExamId      string    //考试ID
	NodeId      string    //节点ID
	CourseId    string    //课程ID
	Title       string    //考试标题名称
	StartTime   time.Time //考试开始时间
	EndTime     time.Time //考试结束时间
	LimitedTime float32   //考试限时
	Score       float32   //试卷总分

}

// 考试节点作业信息
type YingHuaWork struct {
	Id        string    //ID
	WorkId    string    //考试ID
	NodeId    string    //节点ID
	CourseId  string    //课程ID
	Title     string    //考试标题名称
	StartTime time.Time //考试开始时间
	EndTime   time.Time //考试结束时间
	Score     float32   //试卷总分
	Allow     int       //允许做题次数
	Frequency int       //不知道是啥参数
}

// 课程列表
func CourseListAction(cache *yinghuaApi.YingHuaUserCache) ([]YingHuaCourse, error) {
	var courseList []YingHuaCourse
	listJson, err := cache.CourseListApi(10, nil)
	if err != nil {
		return []YingHuaCourse{}, errors.New("获取数据失败:" + err.Error())
	}

	log.Print(log.DEBUG, `[`, cache.Account, `] `, `CourseListAction---`, listJson)
	//超时重登检测
	LoginTimeoutAfreshAction(cache, listJson)
	//如果获取失败
	if gojsonq.New().JSONString(listJson).Find("msg") != "获取数据成功" {
		return []YingHuaCourse{}, errors.New("获取数据失败:" + listJson)
	}
	jsonList := gojsonq.New().JSONString(listJson).Find("result.list")
	// 断言为切片并遍历
	if items, ok := jsonList.([]interface{}); ok {
		for _, item := range items {
			// 每个 item 是 map[string]interface{} 类型
			if obj, ok := item.(map[string]interface{}); ok {
				startTime, _ := time.Parse("2006-01-02", obj["startDate"].(string)) //时间转换
				endTime, _ := time.Parse("2006-01-02", obj["endDate"].(string))     //时间转换
				courseList = append(courseList, YingHuaCourse{Id: strconv.Itoa(int(obj["id"].(float64))), Name: obj["name"].(string), Mode: int(obj["mode"].(float64)), Progress: obj["progress"].(float64), StartDate: startTime, EndDate: endTime, VideoCount: int(obj["videoCount"].(float64)), VideoLearned: int(obj["videoLearned"].(float64))})
			}
		}
	} else {
		fmt.Println("无法将数据转换为预期的类型")
	}
	return courseList, nil
}

// CourseDetailAction 获取指定课程的信息
func CourseDetailAction(cache *yinghuaApi.YingHuaUserCache, courseId string) (YingHuaCourse, error) {
	courseDetailJson, err := cache.CourseDetailApi(courseId, 10, nil)
	if err != nil {
		return YingHuaCourse{}, errors.New("获取数据失败:" + err.Error())
	}

	//超时重登检测
	LoginTimeoutAfreshAction(cache, courseDetailJson)
	//如果获取失败
	if gojsonq.New().JSONString(courseDetailJson).Find("msg") != "获取数据成功" {
		return YingHuaCourse{}, errors.New("获取数据失败:" + courseDetailJson)
	}
	json := gojsonq.New().JSONString(courseDetailJson).Find("result.data")
	// 断言为切片并遍历
	if obj, ok := json.(map[string]interface{}); ok {
		startTime, _ := time.Parse("2006-01-02", obj["startDate"].(string)) //时间转换
		endTime, _ := time.Parse("2006-01-02", obj["endDate"].(string))     //时间转换
		return YingHuaCourse{Id: strconv.Itoa(int(obj["id"].(float64))), Name: obj["name"].(string), Mode: int(obj["mode"].(float64)), Progress: obj["progress"].(float64), StartDate: startTime, EndDate: endTime, VideoCount: int(obj["videoCount"].(float64)), VideoLearned: int(obj["videoLearned"].(float64))}, nil
	} else {
		fmt.Println("无法将数据转换为预期的类型")
	}
	return YingHuaCourse{}, nil
}

// VideosListAction 对应课程的视屏列表
// json1 {"_code":0,"status":true,"msg":"获取数据成功","result":{"list":[{"id":1097105,"name":"第1章 人工智能与人文智慧：创业者的行动","nodeList":[{"id":1458856,"name":"第一节 创业者让人工智能与人文智慧牵手","voteUrl":"","tabVideo":true,"tabFile":false,"tabVote":false,"tabWork":false,"tabExam":false,"videoDuration":"1235","unlockTime":"1970-01-01 08:00","nodeLock":0,"unlockTimeStamp":0,"videoState":2,"duration":"21分钟","index":"1.1","idx":1},{"id":1458857,"name":"第二节 人工智能创业架构与无尽的前沿","voteUrl":"","tabVideo":true,"tabFile":false,"tabVote":false,"tabWork":false,"tabExam":false,"videoDuration":"1432","unlockTime":"1970-01-01 08:00","nodeLock":0,"unlockTimeStamp":0,"videoState":2,"duration":"24分钟","index":"1.2","idx":2}],"idx":1},{"id":1097106,"name":"第2章 劳模人工智能与劳魔创业者","nodeList":[{"id":1458858,"name":"第一节 AI时代人类劳动式微了吗","voteUrl":"","tabVideo":true,"tabFile":false,"tabVote":false,"tabWork":false,"tabExam":false,"videoDuration":"1067","unlockTime":"1970-01-01 08:00","nodeLock":0,"unlockTimeStamp":0,"videoState":2,"duration":"18分钟","index":"2.1","idx":1},{"id":1458859,"name":"第二节 步骤一：人工智能成为好帮手","voteUrl":"","tabVideo":true,"tabFile":false,"tabVote":false,"tabWork":false,"tabExam":false,"videoDuration":"901","unlockTime":"1970-01-01 08:00","nodeLock":0,"unlockTimeStamp":0,"videoState":2,"duration":"15分钟","index":"2.2","idx":2},{"id":1458860,"name":"第三节 步骤二：人工智能成为好助手","voteUrl":"","tabVideo":true,"tabFile":false,"tabVote":false,"tabWork":false,"tabExam":false,"videoDuration":"790","unlockTime":"1970-01-01 08:00","nodeLock":0,"unlockTimeStamp":0,"videoState":2,"duration":"13分钟","index":"2.3","idx":3},{"id":1458861,"name":"第四节 步骤三：人工智能成为好推手","voteUrl":"","tabVideo":true,"tabFile":false,"tabVote":false,"tabWork":false,"tabExam":false,"videoDuration":"1065","unlockTime":"1970-01-01 08:00","nodeLock":0,"unlockTimeStamp":0,"videoState":2,"duration":"18分钟","index":"2.4","idx":4}],"idx":2},{"id":1097107,"name":"第3章 从人工智能性本善到上善若水创业伦理","nodeList":[{"id":1458862,"name":"第一节 人工智能之初应当性本善","voteUrl":"","tabVideo":true,"tabFile":false,"tabVote":false,"tabWork":false,"tabExam":false,"videoDuration":"969","unlockTime":"1970-01-01 08:00","nodeLock":0,"unlockTimeStamp":0,"videoState":2,"duration":"16分钟","index":"3.1","idx":1},{"id":1458863,"name":"第二节 人工智能伦理的难题与伦理方向的决定者","voteUrl":"","tabVideo":true,"tabFile":false,"tabVote":false,"tabWork":false,"tabExam":false,"videoDuration":"875","unlockTime":"1970-01-01 08:00","nodeLock":0,"unlockTimeStamp":0,"videoState":2,"duration":"15分钟","index":"3.2","idx":2},{"id":1458864,"name":"第三节 人工智能创业伦理","voteUrl":"","tabVideo":true,"tabFile":false,"tabVote":false,"tabWork":false,"tabExam":false,"videoDuration":"541","unlockTime":"1970-01-01 08:00","nodeLock":0,"unlockTimeStamp":0,"videoState":2,"duration":"9分钟","index":"3.3","idx":3},{"id":1458865,"name":"第四节 “上善若水”与人工智能创业伦理行动","voteUrl":"","tabVideo":true,"tabFile":false,"tabVote":false,"tabWork":false,"tabExam":false,"videoDuration":"1191","unlockTime":"1970-01-01 08:00","nodeLock":0,"unlockTimeStamp":0,"videoState":2,"duration":"20分钟","index":"3.4","idx":4}],"idx":3},{"id":1097108,"name":"第4章 人工智能创业生态系统的山、水、人","nodeList":[{"id":1458866,"name":"第一节 人工智能的那山、那水、那人","voteUrl":"","tabVideo":true,"tabFile":false,"tabVote":false,"tabWork":false,"tabExam":false,"videoDuration":"761","unlockTime":"1970-01-01 08:00","nodeLock":0,"unlockTimeStamp":0,"videoState":2,"duration":"13分钟","index":"4.1","idx":1},{"id":1458867,"name":"第二节 人工智能创业生态系统与都江堰工程","voteUrl":"","tabVideo":true,"tabFile":false,"tabVote":false,"tabWork":false,"tabExam":false,"videoDuration":"667","unlockTime":"1970-01-01 08:00","nodeLock":0,"unlockTimeStamp":0,"videoState":2,"duration":"11分钟","index":"4.2","idx":2},{"id":1458868,"name":"第三节 三大工程节点的启示","voteUrl":"","tabVideo":true,"tabFile":false,"tabVote":false,"tabWork":false,"tabExam":false,"videoDuration":"730","unlockTime":"1970-01-01 08:00","nodeLock":0,"unlockTimeStamp":0,"videoState":2,"duration":"12分钟","index":"4.3","idx":3}],"idx":4},{"id":1097109,"name":"第5章 人工智能思维对创业思维的挑战？蜂与蝇的启示","nodeList":[{"id":1458869,"name":"第一节 蜂和蝇的小实验","voteUrl":"","tabVideo":true,"tabFile":false,"tabVote":false,"tabWork":false,"tabExam":false,"videoDuration":"589","unlockTime":"1970-01-01 08:00","nodeLock":0,"unlockTimeStamp":0,"videoState":2,"duration":"10分钟","index":"5.1","idx":1},{"id":1458870,"name":"第二节 从管理思维到创业思维","voteUrl":"","tabVideo":true,"tabFile":false,"tabVote":false,"tabWork":false,"tabExam":false,"videoDuration":"431","unlockTime":"1970-01-01 08:00","nodeLock":0,"unlockTimeStamp":0,"videoState":2,"duration":"7分钟","index":"5.2","idx":2},{"id":1458871,"name":"第三节 人工智能思维","voteUrl":"","tabVideo":true,"tabFile":false,"tabVote":false,"tabWork":false,"tabExam":false,"videoDuration":"673","unlockTime":"1970-01-01 08:00","nodeLock":0,"unlockTimeStamp":0,"videoState":2,"duration":"11分钟","index":"5.3","idx":3},{"id":1458872,"name":"第四节 “玻璃瓶”是静止还是动态？","voteUrl":"","tabVideo":true,"tabFile":false,"tabVote":false,"tabWork":false,"tabExam":false,"videoDuration":"640","unlockTime":"1970-01-01 08:00","nodeLock":0,"unlockTimeStamp":0,"videoState":2,"duration":"11分钟","index":"5.4","idx":4},{"id":1458873,"name":"第五节 人机协作冲出铁笼","voteUrl":"","tabVideo":true,"tabFile":false,"tabVote":false,"tabWork":false,"tabExam":false,"videoDuration":"945","unlockTime":"1970-01-01 08:00","nodeLock":0,"unlockTimeStamp":0,"videoState":2,"duration":"16分钟","index":"5.5","idx":5}],"idx":5},{"id":1097110,"name":"第6章 人工智能与商业模式","nodeList":[{"id":1458874,"name":"第一节 商业模式：从画布到画脸","voteUrl":"","tabVideo":true,"tabFile":false,"tabVote":false,"tabWork":false,"tabExam":false,"videoDuration":"670","unlockTime":"1970-01-01 08:00","nodeLock":0,"unlockTimeStamp":0,"videoState":2,"duration":"11分钟","index":"6.1","idx":1},{"id":1458875,"name":"第二节 颜值派商业模式与实力派人工智能","voteUrl":"","tabVideo":true,"tabFile":false,"tabVote":false,"tabWork":false,"tabExam":false,"videoDuration":"822","unlockTime":"1970-01-01 08:00","nodeLock":0,"unlockTimeStamp":0,"videoState":0,"duration":"14分钟","index":"6.2","idx":2},{"id":1458876,"name":"第三节 完美愿景与残酷现实","voteUrl":"","tabVideo":true,"tabFile":false,"tabVote":false,"tabWork":false,"tabExam":false,"videoDuration":"828","unlockTime":"1970-01-01 08:00","nodeLock":1,"unlockTimeStamp":0,"videoState":0,"duration":"14分钟","index":"6.3","idx":3},{"id":1458877,"name":"第四节 人工智能美颜商业模式的误区","voteUrl":"","tabVideo":true,"tabFile":false,"tabVote":false,"tabWork":false,"tabExam":false,"videoDuration":"886","unlockTime":"1970-01-01 08:00","nodeLock":1,"unlockTimeStamp":0,"videoState":0,"duration":"15分钟","index":"6.4","idx":4}],"idx":6},{"id":1097111,"name":"第7章 人工智能与精益创业","nodeList":[{"id":1458878,"name":"第一节 人工智能技术：插曲还是调音","voteUrl":"","tabVideo":true,"tabFile":false,"tabVote":false,"tabWork":false,"tabExam":false,"videoDuration":"936","unlockTime":"1970-01-01 08:00","nodeLock":1,"unlockTimeStamp":0,"videoState":0,"duration":"16分钟","index":"7.1","idx":1},{"id":1458879,"name":"第二节 人工智能创业：由易及快","voteUrl":"","tabVideo":true,"tabFile":false,"tabVote":false,"tabWork":false,"tabExam":false,"videoDuration":"838","unlockTime":"1970-01-01 08:00","nodeLock":1,"unlockTimeStamp":0,"videoState":0,"duration":"14分钟","index":"7.2","idx":2},{"id":1458880,"name":"第三节 从精尖技术到精益启动：两种时间观","voteUrl":"","tabVideo":true,"tabFile":false,"tabVote":false,"tabWork":false,"tabExam":false,"videoDuration":"931","unlockTime":"1970-01-01 08:00","nodeLock":1,"unlockTimeStamp":0,"videoState":0,"duration":"16分钟","index":"7.3","idx":3},{"id":1458881,"name":"第四节 精益的时间艺术","voteUrl":"","tabVideo":true,"tabFile":false,"tabVote":false,"tabWork":false,"tabExam":false,"videoDuration":"796","unlockTime":"1970-01-01 08:00","nodeLock":1,"unlockTimeStamp":0,"videoState":0,"duration":"13分钟","index":"7.4","idx":4}],"idx":7},{"id":1097112,"name":"第8章 中国人工智能创业城市","nodeList":[{"id":1458882,"name":"第一节 五年五城五色土与北京","voteUrl":"","tabVideo":true,"tabFile":false,"tabVote":false,"tabWork":false,"tabExam":false,"videoDuration":"645","unlockTime":"1970-01-01 08:00","nodeLock":1,"unlockTimeStamp":0,"videoState":0,"duration":"11分钟","index":"8.1","idx":1},{"id":1458883,"name":"第二节 深圳与上海","voteUrl":"","tabVideo":true,"tabFile":false,"tabVote":false,"tabWork":false,"tabExam":false,"videoDuration":"724","unlockTime":"1970-01-01 08:00","nodeLock":1,"unlockTimeStamp":0,"videoState":0,"duration":"12分钟","index":"8.2","idx":2},{"id":1458884,"name":"第三节 成都与杭州","voteUrl":"","tabVideo":true,"tabFile":false,"tabVote":false,"tabWork":false,"tabExam":false,"videoDuration":"701","unlockTime":"1970-01-01 08:00","nodeLock":1,"unlockTimeStamp":0,"videoState":0,"duration":"12分钟","index":"8.3","idx":3},{"id":1458885,"name":"第四节 下一站坐标","voteUrl":"","tabVideo":true,"tabFile":false,"tabVote":false,"tabWork":false,"tabExam":false,"videoDuration":"771","unlockTime":"1970-01-01 08:00","nodeLock":1,"unlockTimeStamp":0,"videoState":0,"duration":"13分钟","index":"8.4","idx":4}],"idx":8},{"id":1097113,"name":"第9章 中国人工智能政策与创新创业政策","nodeList":[{"id":1458886,"name":"第一节 中国人工智能政策五年回顾","voteUrl":"","tabVideo":true,"tabFile":false,"tabVote":false,"tabWork":false,"tabExam":false,"videoDuration":"1109","unlockTime":"1970-01-01 08:00","nodeLock":1,"unlockTimeStamp":0,"videoState":0,"duration":"18分钟","index":"9.1","idx":1},{"id":1458887,"name":"第二节 中国双创政策五年概览","voteUrl":"","tabVideo":true,"tabFile":false,"tabVote":false,"tabWork":false,"tabExam":false,"videoDuration":"802","unlockTime":"1970-01-01 08:00","nodeLock":1,"unlockTimeStamp":0,"videoState":0,"duration":"13分钟","index":"9.2","idx":2}],"idx":9},{"id":1097114,"name":"第10章 人工智能创业与创业教育","nodeList":[{"id":1458888,"name":"第一节 人工智能教育与创业教育的融合","voteUrl":"","tabVideo":true,"tabFile":false,"tabVote":false,"tabWork":false,"tabExam":false,"videoDuration":"688","unlockTime":"1970-01-01 08:00","nodeLock":1,"unlockTimeStamp":0,"videoState":0,"duration":"11分钟","index":"10.1","idx":1},{"id":1458889,"name":"第二节 基于“理论-实践”的融合类型","voteUrl":"","tabVideo":true,"tabFile":false,"tabVote":false,"tabWork":false,"tabExam":false,"videoDuration":"805","unlockTime":"1970-01-01 08:00","nodeLock":1,"unlockTimeStamp":0,"videoState":0,"duration":"13分钟","index":"10.2","idx":2}],"idx":10},{"id":1097115,"name":"第11章 人工智能创业伦理专题探讨","nodeList":[{"id":1458890,"name":"第一节 人工智能创业伦理的新特征","voteUrl":"","tabVideo":true,"tabFile":false,"tabVote":false,"tabWork":false,"tabExam":false,"videoDuration":"771","unlockTime":"1970-01-01 08:00","nodeLock":1,"unlockTimeStamp":0,"videoState":0,"duration":"13分钟","index":"11.1","idx":1},{"id":1458891,"name":"第二节 人工智能创业伦理对创业者的新要求","voteUrl":"","tabVideo":true,"tabFile":false,"tabVote":false,"tabWork":false,"tabExam":false,"videoDuration":"1110","unlockTime":"1970-01-01 08:00","nodeLock":1,"unlockTimeStamp":0,"videoState":0,"duration":"19分钟","index":"11.2","idx":2},{"id":1458892,"name":"第三节 人工智能创业者的伦理决策与伦理领导能力提升","voteUrl":"","tabVideo":true,"tabFile":false,"tabVote":false,"tabWork":false,"tabExam":false,"videoDuration":"1182","unlockTime":"1970-01-01 08:00","nodeLock":1,"unlockTimeStamp":0,"videoState":0,"duration":"20分钟","index":"11.3","idx":3},{"id":1458893,"name":"第四节 人工智能创业伦理的治理平衡","voteUrl":"","tabVideo":true,"tabFile":false,"tabVote":false,"tabWork":false,"tabExam":false,"videoDuration":"966","unlockTime":"1970-01-01 08:00","nodeLock":1,"unlockTimeStamp":0,"videoState":0,"duration":"16分钟","index":"11.4","idx":4}],"idx":11},{"id":1097116,"name":"第12章 人工智能创业前沿趋势展望","nodeList":[{"id":1458894,"name":"第一节 人工智能创业的实践反思","voteUrl":"","tabVideo":true,"tabFile":false,"tabVote":false,"tabWork":false,"tabExam":false,"videoDuration":"1323","unlockTime":"1970-01-01 08:00","nodeLock":1,"unlockTimeStamp":0,"videoState":0,"duration":"22分钟","index":"12.1","idx":1},{"id":1458895,"name":"第二节 人工智能创业的教育创新","voteUrl":"","tabVideo":true,"tabFile":false,"tabVote":false,"tabWork":false,"tabExam":false,"videoDuration":"1063","unlockTime":"1970-01-01 08:00","nodeLock":1,"unlockTimeStamp":0,"videoState":0,"duration":"18分钟","index":"12.2","idx":2},{"id":1458896,"name":"第三节 人工智能创业的人才培养","voteUrl":"","tabVideo":true,"tabFile":false,"tabVote":false,"tabWork":false,"tabExam":false,"videoDuration":"892","unlockTime":"1970-01-01 08:00","nodeLock":1,"unlockTimeStamp":0,"videoState":0,"duration":"15分钟","index":"12.3","idx":3},{"id":1458897,"name":"第四节 人工智能创业的前沿展望","voteUrl":"","tabVideo":true,"tabFile":false,"tabVote":false,"tabWork":false,"tabExam":false,"videoDuration":"1627","unlockTime":"1970-01-01 08:00","nodeLock":1,"unlockTimeStamp":0,"videoState":0,"duration":"27分钟","index":"12.4","idx":4}],"idx":12}]}}
// json2 {"_code":0,"status":true,"msg":"获取数据成功","result":{"list":[{"id":1458856,"name":"第一节 创业者让人工智能与人文智慧牵手","courseId":1012027,"videoDuration":1235,"bid":"48570843","duration":1161,"progress":100,"state":2,"viewCount":2,"finalTime":"","viewedDuration":1161,"error":0,"errorMessage":"","beginTime":""},{"id":1458857,"name":"第二节 人工智能创业架构与无尽的前沿","courseId":1012027,"videoDuration":1432,"bid":"48571973","duration":1291,"progress":100,"state":2,"viewCount":1,"finalTime":"","viewedDuration":1291,"error":0,"errorMessage":"","beginTime":""},{"id":1458858,"name":"第一节 AI时代人类劳动式微了吗","courseId":1012027,"videoDuration":1067,"bid":"48573259","duration":961,"progress":100,"state":2,"viewCount":1,"finalTime":"","viewedDuration":961,"error":0,"errorMessage":"","beginTime":""},{"id":1458859,"name":"第二节 步骤一：人工智能成为好帮手","courseId":1012027,"videoDuration":901,"bid":"48574333","duration":811,"progress":100,"state":2,"viewCount":1,"finalTime":"","viewedDuration":811,"error":0,"errorMessage":"","beginTime":""},{"id":1458860,"name":"第三节 步骤二：人工智能成为好助手","courseId":1012027,"videoDuration":790,"bid":"48575214","duration":721,"progress":100,"state":2,"viewCount":1,"finalTime":"","viewedDuration":721,"error":0,"errorMessage":"","beginTime":""},{"id":1458861,"name":"第四节 步骤三：人工智能成为好推手","courseId":1012027,"videoDuration":1065,"bid":"48576005","duration":971,"progress":100,"state":2,"viewCount":1,"finalTime":"","viewedDuration":971,"error":0,"errorMessage":"","beginTime":""},{"id":1458862,"name":"第一节 人工智能之初应当性本善","courseId":1012027,"videoDuration":969,"bid":"48576942","duration":881,"progress":100,"state":2,"viewCount":1,"finalTime":"","viewedDuration":881,"error":0,"errorMessage":"","beginTime":""},{"id":1458863,"name":"第二节 人工智能伦理的难题与伦理方向的决定者","courseId":1012027,"videoDuration":875,"bid":"48577986","duration":791,"progress":100,"state":2,"viewCount":1,"finalTime":"","viewedDuration":791,"error":0,"errorMessage":"","beginTime":""},{"id":1458864,"name":"第三节 人工智能创业伦理","courseId":1012027,"videoDuration":541,"bid":"48579060","duration":491,"progress":100,"state":2,"viewCount":1,"finalTime":"","viewedDuration":491,"error":0,"errorMessage":"","beginTime":""},{"id":1458865,"name":"第四节 “上善若水”与人工智能创业伦理行动","courseId":1012027,"videoDuration":1191,"bid":"48579737","duration":1081,"progress":100,"state":2,"viewCount":1,"finalTime":"","viewedDuration":1081,"error":0,"errorMessage":"","beginTime":""}],"pageInfo":{"keyName":"page","page":1,"pageCount":5,"recordsCount":42,"onlyCount":42,"pageSize":10},"courseInfo":{"startDate":"2024-10-21","endDate":"2024-12-15","videoCount":"42","videoLearned":"19","state":1}}}
func VideosListAction(UserCache *yinghuaApi.YingHuaUserCache, course YingHuaCourse) ([]YingHuaNode, error) {
	var videoList []YingHuaNode
	videoSet := make(map[string]int)
	//接口一爬取视屏信息
	listJson, err := yinghuaApi.CourseVideListApi(*UserCache, course.Id, 10, nil)
	log.Print(log.DEBUG, `[`, UserCache.Account, `] `, `CourseListAction---`, listJson)
	//超时重登检测
	LoginTimeoutAfreshAction(UserCache, listJson)
	//如果获取失败
	if gojsonq.New().JSONString(listJson).Find("msg") != "获取数据成功" {
		return []YingHuaNode{}, errors.New("获取数据失败：" + err.Error())
	}
	jsonList := gojsonq.New().JSONString(listJson).Find("result.list")
	// 断言为切片并遍历
	if items, ok := jsonList.([]interface{}); ok {
		for _, item := range items {
			// 每个 item 是 map[string]interface{} 类型
			if obj, ok := item.(map[string]interface{}); ok {
				nodeList := obj["nodeList"]
				if nodes, ok := nodeList.([]interface{}); ok {
					for _, node := range nodes {
						if obj1, ok := node.(map[string]interface{}); ok {
							videoDuration := 0
							if obj1["videoDuration"] != nil {
								res, _ := strconv.Atoi(obj1["videoDuration"].(string))
								videoDuration = res
							}
							unlockTime, _ := time.Parse("2006-01-02 15:04", obj1["unlockTime"].(string)) //时间转换
							videoList = append(videoList, YingHuaNode{
								Id: strconv.Itoa(int(obj1["id"].(float64))),
								//CourseId:
								Name:          obj1["name"].(string),
								VideoDuration: int(videoDuration),
								NodeLock:      int(obj1["nodeLock"].(float64)),
								UnlockTime:    unlockTime,
								TabVideo:      obj1["tabVideo"].(bool),
								TabFile:       obj1["tabFile"].(bool),
								TabExam:       obj1["tabExam"].(bool),
								TabWork:       obj1["tabWork"].(bool),
							})
							videoSet[strconv.Itoa(int(obj1["id"].(float64)))] = len(videoList) - 1
						}
					}
				}
			}
		}
	} else {
		fmt.Println("无法将数据转换为预期的类型")
	}

	//接口二而爬取视屏信息
	signalSet := make(map[string]bool)
	for i := 1; i < 999; i++ {
		listJson1, err := yinghuaApi.VideWatchRecodeApi(*UserCache, course.Id, i, 10, nil)
		if err != nil {
			log.Print(log.INFO, `[`, UserCache.Account, `] `, log.BoldRed, err)
		}
		log.Print(log.DEBUG, `[`, UserCache.Account, `] `, `CourseListAction---`, listJson1)
		//如果获取失败
		if gojsonq.New().JSONString(listJson).Find("msg") != "获取数据成功" {
			return []YingHuaNode{}, errors.New("获取数据失败：" + err.Error())
		}
		jsonList1 := gojsonq.New().JSONString(listJson1).Find("result.list")
		// 断言为切片并遍历
		if items, ok := jsonList1.([]interface{}); ok {
			//如果为空表则直接跳出循环
			if len(items) == 0 {
				break
			}
			for _, item := range items {
				// 每个 item 是 map[string]interface{} 类型
				if obj, ok := item.(map[string]interface{}); ok {
					nodeId := strconv.Itoa(int(obj["id"].(float64)))
					//如果是最后一个了那么就退出
					_, isSignal := signalSet[nodeId]
					if isSignal {
						return videoList, nil
					}
					signalSet[nodeId] = true
					index, isOk := videoSet[nodeId]
					if isOk {
						videoList[index].CourseId = strconv.Itoa(int(obj["courseId"].(float64)))
						videoList[index].Progress = float32(obj["progress"].(float64))
						videoList[index].ViewedDuration = int(obj["viewedDuration"].(float64))
						videoList[index].State = int(obj["state"].(float64))
					}
				}
			}
		} else {
			log.Print(log.INFO, "无法将数据转换为预期的类型")
		}
	}

	return videoList, nil
}

// SubmitStudyTimeAction 提交学时
func SubmitStudyTimeAction(userCache *yinghuaApi.YingHuaUserCache, nodeId string /*对应视屏节点ID*/, studyId string /*学习分配ID*/, studyTime int /*提交的学时*/) (string, error) {
	//提交学时
	sub, err := yinghuaApi.SubmitStudyTimeApi(*userCache, nodeId, studyId, studyTime, 10, nil)
	//避免502情况
	if err != nil { //其他错误
		return "", err
	}
	return sub, nil
}

// ExamDetailAction 获取考试节点对应信息
// {"_code":9,"status":false,"msg":"考试测试时间还未开始","result":{}}
func ExamDetailAction(UserCache *yinghuaApi.YingHuaUserCache, nodeId string) ([]YingHuaExam, error) {
	var examList []YingHuaExam
	jsonStr, err := yinghuaApi.ExamDetailApi(*UserCache, nodeId, 10, nil)
	if err != nil {
		return []YingHuaExam{}, errors.New("获取数据失败" + err.Error())
	}
	//超时重登检测
	LoginTimeoutAfreshAction(UserCache, jsonStr)
	jsonData := gojsonq.New().JSONString(jsonStr).Find("result.list")

	log.Print(log.DEBUG, `[`, UserCache.Account, `] `, `CourseListAction---`, jsonData)

	//如果获取失败
	if gojsonq.New().JSONString(jsonStr).Find("msg") != "获取数据成功" {
		return []YingHuaExam{}, errors.New("获取数据失败" + jsonStr)
	}
	jsonList := gojsonq.New().JSONString(jsonStr).Find("result.list")
	// 断言为切片并遍历
	if items, ok := jsonList.([]interface{}); ok {
		for _, item := range items {
			// 每个 item 是 map[string]interface{} 类型
			if obj, ok := item.(map[string]interface{}); ok {
				startTime, _ := time.Parse("2006-01-02 15:04:05", obj["startTime"].(string)) //时间转换
				endTime, _ := time.Parse("2006-01-02 15:04:05", obj["endTime"].(string))     //时间转换
				examId := strings.Split(strings.Split(obj["url"].(string), "examId=")[1], "&token")[0]
				examList = append(examList, YingHuaExam{Id: strconv.Itoa(int(obj["id"].(float64))), Title: obj["title"].(string), Score: float32(obj["score"].(float64)), LimitedTime: float32(obj["limitedTime"].(float64)), StartTime: startTime, EndTime: endTime, CourseId: strconv.Itoa(int(obj["courseId"].(float64))), NodeId: strconv.Itoa(int(obj["nodeId"].(float64))), ExamId: examId})
			}
		}
	} else {
		fmt.Println("无法将数据转换为预期的类型")
	}
	return examList, nil
}

// randomAnswer 如果AI出问题那么直接随机返回答案
func randomAnswer(topic entity.YingHuaExamTopic) string {
	if topic.Type == "单选" {
		sct := rand.Intn(len(topic.Selects))
		return "[" + topic.Selects[sct].Value + "]"
	} else if topic.Type == "多选" {
		sct := "["
		for i := 0; i < len(topic.Selects); i++ {
			sct = sct + topic.Selects[i].Value
			if i != len(topic.Selects)-1 {
				sct += " "
			}
		}
		sct = sct + "]"
		return sct
	}
	return ""
}

// StartExamAction 开始考试
func StartExamAction(
	userCache *yinghuaApi.YingHuaUserCache,
	exam YingHuaExam,
	url, model, apiKey string,
	aiType ctype.AiType,
	isAutoSubExam int /*是否自动提交试卷，0为不自动提交，1为自动提交*/) error {
	//开始考试
	startExam, err := yinghuaApi.StartExam(*userCache, exam.CourseId, exam.NodeId, exam.ExamId, 10, nil)
	if err != nil {
		log.Print(log.INFO, err.Error())
		return errors.New(err.Error())
	}
	//如果开始考试状态异常则直接抛错
	if int(gojsonq.New().JSONString(startExam).Find("_code").(float64)) == 9 {
		return errors.New(gojsonq.New().JSONString(startExam).Find("msg").(string))
	}
	//开始答题
	api, err := yinghuaApi.GetExamTopicApi(*userCache, exam.NodeId, exam.ExamId, 8, nil)
	if int(gojsonq.New().JSONString(startExam).Find("_code").(float64)) == 9 {
		return errors.New(gojsonq.New().JSONString(startExam).Find("msg").(string))
	}
	//html转结构体
	topic := yinghuaApi.TurnExamTopic(api)
	//fmt.Println(topic)
	//遍历题目map,并回答问题
	var lastAnswer utils.Answer
	var lastProblem string
	for k, v := range topic.YingHuaExamTopics {
		//构建统一AI消息
		aiMessage := yinghuaApi.AIProblemMessage(exam.Title, entity.ExamTurn{
			YingHuaExamTopic: v,
		})
		aiAnswer, err := utils.AggregationAIApi(url, model, aiType, aiMessage, apiKey)
		answer := aiTurnYingHuaAnswer(userCache, aiAnswer, v)

		if err != nil {
			log.Print(log.INFO, `[`, userCache.Account, `] `, log.BoldRed, "Ai异常，返回信息：", err.Error())
			//os.Exit(0)
			log.Print(log.INFO, `[`, userCache.Account, `] `, log.BoldRed, "Ai异常", exam.Title, "第", v.Index, "随机填写答案", "题目内容：", v.Content)
			aiAnswer = randomAnswer(v)
		}
		//fmt.Println(aiAnswer)
		subWorkApi, err := yinghuaApi.SubmitExamApi(*userCache, exam.ExamId, k, answer, "0", 8, nil)
		if err != nil {
			log.Print(log.INFO, `[`, userCache.Account, `] `, log.BoldRed, "Ai异常，返回信息：", err.Error())
		}
		//如果提交答案服务器端返回信息异常
		if gojsonq.New().JSONString(subWorkApi).Find("msg") != "答题保存成功" {
			log.Print(log.INFO, log.BoldRed, `[`, userCache.Account, `] `, log.BoldRed, "提交答案异常，返回信息：", subWorkApi, "题目内容：", v, "AI回答信息：", aiAnswer)
		}
		lastAnswer = answer
		lastProblem = k
	}
	if isAutoSubExam == 1 {
		//结束考试
		subWorkApi, err := yinghuaApi.SubmitExamApi(*userCache, exam.ExamId, lastProblem, lastAnswer, "1", 8, nil)
		//如果结束做题服务器端返回信息异常
		if gojsonq.New().JSONString(subWorkApi).Find("msg") != "提交试卷成功" || err != nil {
			log.Print(log.INFO, log.BoldRed, `[`, userCache.Account, `] `, log.BoldRed, "提交试卷异常，返回信息：", subWorkApi, err.Error())
		}
	}
	return nil
}

// StartExamForExternalAction 开始考试
func StartExamForExternalAction(
	userCache *yinghuaApi.YingHuaUserCache,
	exam YingHuaExam,
	queBankUrl string,
	isAutoSubExam int /*是否自动提交试卷，0为不自动提交，1为自动提交*/) error {
	//开始考试
	startExam, err := yinghuaApi.StartExam(*userCache, exam.CourseId, exam.NodeId, exam.ExamId, 10, nil)
	if err != nil {
		log.Print(log.INFO, err.Error())
		return errors.New(err.Error())
	}
	//如果开始考试状态异常则直接抛错
	if int(gojsonq.New().JSONString(startExam).Find("_code").(float64)) == 9 {
		return errors.New(gojsonq.New().JSONString(startExam).Find("msg").(string))
	}
	//开始答题
	api, err := yinghuaApi.GetExamTopicApi(*userCache, exam.NodeId, exam.ExamId, 8, nil)
	if int(gojsonq.New().JSONString(startExam).Find("_code").(float64)) == 9 {
		return errors.New(gojsonq.New().JSONString(startExam).Find("msg").(string))
	}
	//html转结构体
	topic := yinghuaApi.TurnExamTopic(api)
	//fmt.Println(topic)
	//遍历题目map,并回答问题
	var lastAnswer utils.Answer
	var lastProblem string
	for k, v := range topic.YingHuaExamTopics {
		standardProblem := v.TurnProblem() //转为标准问题结构体
		answer, err := standardProblem.ApiQueRequest(queBankUrl, 5, nil)
		answer = AnswerTurnResult(answer, v) //转换答案
		//fmt.Println(aiAnswer)
		subWorkApi, err := yinghuaApi.SubmitExamApi(*userCache, exam.ExamId, k, answer, "0", 8, nil)
		if err != nil {
			log.Print(log.INFO, `[`, userCache.Account, `] `, log.BoldRed, "Ai异常，返回信息：", err.Error())
		}
		//如果提交答案服务器端返回信息异常
		if gojsonq.New().JSONString(subWorkApi).Find("msg") != "答题保存成功" {
			log.Print(log.INFO, log.BoldRed, `[`, userCache.Account, `] `, log.BoldRed, "提交答案异常，返回信息：", subWorkApi, "题目内容：", v, "外部接口回答信息：", answer)
		}
		lastAnswer = answer
		lastProblem = k
	}
	if isAutoSubExam == 1 {
		//结束考试
		subWorkApi, err := yinghuaApi.SubmitExamApi(*userCache, exam.ExamId, lastProblem, lastAnswer, "1", 8, nil)
		//如果结束做题服务器端返回信息异常
		if gojsonq.New().JSONString(subWorkApi).Find("msg") != "提交试卷成功" || err != nil {
			log.Print(log.INFO, log.BoldRed, `[`, userCache.Account, `] `, log.BoldRed, "提交试卷异常，返回信息：", subWorkApi, err.Error())
		}
	}
	return nil
}

// ExamFinallyScoreAction 获取最终作业分数
func ExamFinallyScoreAction(userCache *yinghuaApi.YingHuaUserCache, work YingHuaExam) (string, error) {
	detail, err := yinghuaApi.ExamFinallyDetailApi(*userCache, work.CourseId, work.NodeId, work.ExamId, 8, nil)
	if err != nil {
		log.Print(log.INFO, `[`, userCache.Account, `] `, log.BoldRed, err.Error())
	}
	scorePattern1 := `最终得分：[^\d]*([\d]+)[^分]*分`
	scoreRegexp1 := regexp.MustCompile(scorePattern1)
	scoreMatches1 := scoreRegexp1.FindAllStringSubmatch(detail, -1)
	for _, scoreMatch := range scoreMatches1 {
		score := scoreMatch[1] //最高分答题次数
		return score, nil
	}
	return "", errors.New("没有找到分数，可能成绩还未出")
}

// WorkDetailAction 获取作业节点对应信息
func WorkDetailAction(userCache *yinghuaApi.YingHuaUserCache, nodeId string) ([]YingHuaWork, error) {
	var workList []YingHuaWork
	jsonStr, err := yinghuaApi.WorkDetailApi(*userCache, nodeId, 8, nil)
	if err != nil {
		return []YingHuaWork{}, errors.New("获取数据失败" + err.Error())
	}

	//超时重登检测
	LoginTimeoutAfreshAction(userCache, jsonStr)
	jsonData := gojsonq.New().JSONString(jsonStr).Find("result.list")

	log.Print(log.DEBUG, `[`, userCache.Account, `] `, `CourseListAction---`, jsonData)

	//如果获取失败
	if gojsonq.New().JSONString(jsonStr).Find("msg") != "获取数据成功" {
		return []YingHuaWork{}, errors.New("获取数据失败" + jsonStr)
	}
	jsonList := gojsonq.New().JSONString(jsonStr).Find("result.list")
	// 断言为切片并遍历
	if items, ok := jsonList.([]interface{}); ok {
		for _, item := range items {
			// 每个 item 是 map[string]interface{} 类型
			if obj, ok := item.(map[string]interface{}); ok {
				startTime, _ := time.Parse("2006-01-02 15:04:05", obj["startTime"].(string)) //时间转换
				endTime, _ := time.Parse("2006-01-02 15:04:05", obj["endTime"].(string))     //时间转换
				workId := strings.Split(strings.Split(obj["url"].(string), "workId=")[1], "&token")[0]
				allow, _ := strconv.Atoi(obj["allow"].(string))
				frequency, _ := strconv.Atoi(obj["frequency"].(string))
				workList = append(workList, YingHuaWork{Id: strconv.Itoa(int(obj["id"].(float64))), Title: obj["title"].(string), Score: float32(obj["score"].(float64)), StartTime: startTime, EndTime: endTime, CourseId: strconv.Itoa(int(obj["courseId"].(float64))), NodeId: strconv.Itoa(int(obj["nodeId"].(float64))), WorkId: workId, Allow: allow, Frequency: frequency})
			}
		}
	} else {
		fmt.Println("无法将数据转换为预期的类型")
	}
	return workList, nil
}

// 选项评分系统
func selectMarkingSystem(text1, text2 string) float32 {
	sum := float32(0)
	for i := range text1 {
		for j := range text2 {
			if text1[i] == text2[j] {
				sum += float32(1)
			}
		}
	}
	return float32(float32(sum) / float32(len(text1)))
}

// 竞争对手之间有激烈的价格竞争
// 竞争对手之间有激烈的价格竞争
// AI回复转答案
func aiTurnYingHuaAnswer(cache *yinghuaApi.YingHuaUserCache, aiAnswer string, v entity.YingHuaExamTopic) utils.Answer {
	answer := utils.Answer{Type: v.Type}
	if v.Type == "单选" || v.Type == "判断" || v.Type == "多选" {
		var jsonStr []string
		var res []string
		json.Unmarshal([]byte(aiAnswer), &jsonStr)
		for _, item := range jsonStr {
			for _, v := range v.Selects {
				if strings.Contains(v.Text, item) ||
					strings.Contains(v.Num, item) ||
					strings.Contains(strings.ReplaceAll(strings.ReplaceAll(v.Num+v.Text, " ", ""), " ", ""), strings.ReplaceAll(item, " ", "")) ||
					strings.Contains(strings.ReplaceAll(v.Num+v.Text, " ", ""), strings.ReplaceAll(item, " ", "")) {
					res = append(res, v.Value)
				}
			}

		}
		if len(res) == 0 {
			for _, item := range jsonStr { //如果没有答案，采用第二方案，字符串评判法
				for _, v := range v.Selects {
					if selectMarkingSystem(item, v.Text) > 0.60 {
						res = append(res, v.Value)
					}
				}

			}
		}

		answer.Answers = res
	} else if v.Type == "填空" || v.Type == "简答" {
		var res []string
		json.Unmarshal([]byte(aiAnswer), &res)
		for _, item := range res {
			answer.Answers = append(answer.Answers, item)
		}
	}
	if len(answer.Answers) == 0 || answer.Answers == nil {
		if v.Type == "单选" || v.Type == "判断" {
			answer.Answers = []string{"A"}
		} else if v.Type == "多选" {
			answer.Answers = []string{"B", "C"}
		} else if v.Type == "简答" || v.Type == "填空" {
			log.Print(log.INFO, `[`, cache.Account, `] `, log.BoldRed, "\n题目类型：", v.Type, "\n题目：", v.Content, "\n\nAi回答内容无法解析，因该题为填空或简答题，所以自动留空")
			return answer
		}
		log.Print(log.INFO, `[`, cache.Account, `] `, log.BoldRed, "\n题目类型：", v.Type, "\n题目：", v.Content, "\n\nAi回答内容无法解析，使用使用随机答案策略")
	}
	return answer
}

// 将答案转换为选项
func AnswerTurnResult(answer utils.Answer, v entity.YingHuaExamTopic) utils.Answer {
	var res []string
	if v.Type == "单选" || v.Type == "判断" || v.Type == "多选" {

		for _, item := range answer.Answers {
			for _, v := range v.Selects {
				if strings.Contains(v.Text, item) ||
					strings.Contains(v.Num, item) ||
					strings.Contains(strings.ReplaceAll(strings.ReplaceAll(v.Num+v.Text, " ", ""), " ", ""), strings.ReplaceAll(item, " ", "")) ||
					strings.Contains(strings.ReplaceAll(v.Num+v.Text, " ", ""), strings.ReplaceAll(item, " ", "")) {
					res = append(res, v.Value)
				}
			}
		}
		answer.Answers = res
		return answer
	}
	return answer
}

// StartWorkAction 开始写作业
func StartWorkAction(userCache *yinghuaApi.YingHuaUserCache,
	work YingHuaWork,
	url, model, apiKey string,
	aiType ctype.AiType, isAutoSubExam int /*是否自动提交试卷，0为不自动提交，1为自动提交*/) error {
	//开始考试
	startWork, err := yinghuaApi.StartWork(*userCache, work.CourseId, work.NodeId, work.WorkId, 8, nil)
	if err != nil {
		log.Print(log.INFO, err.Error())
		return errors.New(err.Error())
	}
	//如果开始考试状态异常则直接抛错
	if int(gojsonq.New().JSONString(startWork).Find("_code").(float64)) == 9 {
		return errors.New(gojsonq.New().JSONString(startWork).Find("msg").(string))
	}
	//开始答题
	api, err := yinghuaApi.GetWorkApi(*userCache, work.NodeId, work.WorkId, 8, nil)
	if int(gojsonq.New().JSONString(startWork).Find("_code").(float64)) == 9 {
		return errors.New(gojsonq.New().JSONString(startWork).Find("msg").(string))
	}
	//html转结构体
	topic := yinghuaApi.TurnExamTopic(api)
	//fmt.Println(topic)
	//遍历题目map,并回答问题
	var lastAnswer utils.Answer
	var lastProblem string
	for k, v := range topic.YingHuaExamTopics {
		//构建统一AI消息
		aiMessage := yinghuaApi.AIProblemMessage(work.Title, entity.ExamTurn{
			YingHuaExamTopic: v,
		})
		aiAnswer, err := utils.AggregationAIApi(url, model, aiType, aiMessage, apiKey)
		answer := aiTurnYingHuaAnswer(userCache, aiAnswer, v)

		if err != nil {
			log.Print(log.INFO, `[`, userCache.Account, `] `, log.BoldRed, "Ai异常，返回信息：", err.Error())
			os.Exit(0)
		}
		//fmt.Println(aiAnswer)
		subWorkApi, err := yinghuaApi.SubmitWorkApi(*userCache, work.WorkId, k, answer, "0", 10, nil)
		if err != nil {
			log.Print(log.INFO, `[`, userCache.Account, `] `, log.BoldRed, "Ai异常，返回信息：", err.Error())
		}
		//如果提交答案服务器端返回信息异常
		if gojsonq.New().JSONString(subWorkApi).Find("msg") != "答题保存成功" {
			log.Print(log.INFO, log.BoldRed, `[`, userCache.Account, `] `, log.BoldRed, "提交答案异常，返回信息：", subWorkApi)
		}
		lastAnswer = aiTurnYingHuaAnswer(userCache, aiAnswer, v)
		lastProblem = k
	}
	//结束考试
	if isAutoSubExam == 1 {
		subWorkApi, err := yinghuaApi.SubmitWorkApi(*userCache, work.WorkId, lastProblem, lastAnswer, "1", 10, nil)
		//如果结束做题服务器端返回信息异常
		if gojsonq.New().JSONString(subWorkApi).Find("msg") != "提交作业成功" || err != nil {
			log.Print(log.INFO, log.BoldRed, `[`, userCache.Account, `] `, log.BoldRed, "提交试卷异常，返回信息：", subWorkApi, err.Error())
		}
	}

	return nil
}

// StartWorkForExternalAction 外部题库开始写作业
func StartWorkForExternalAction(userCache *yinghuaApi.YingHuaUserCache,
	queBankUrl string,
	work YingHuaWork, isAutoSubExam int /*是否自动提交试卷，0为不自动提交，1为自动提交*/) error {
	//开始考试
	startWork, err := yinghuaApi.StartWork(*userCache, work.CourseId, work.NodeId, work.WorkId, 8, nil)
	if err != nil {
		log.Print(log.INFO, err.Error())
		return errors.New(err.Error())
	}
	//如果开始考试状态异常则直接抛错
	if int(gojsonq.New().JSONString(startWork).Find("_code").(float64)) == 9 {
		return errors.New(gojsonq.New().JSONString(startWork).Find("msg").(string))
	}
	//开始答题
	api, err := yinghuaApi.GetWorkApi(*userCache, work.NodeId, work.WorkId, 8, nil)
	if int(gojsonq.New().JSONString(startWork).Find("_code").(float64)) == 9 {
		return errors.New(gojsonq.New().JSONString(startWork).Find("msg").(string))
	}
	//html转结构体
	topic := yinghuaApi.TurnExamTopic(api)
	//fmt.Println(topic)
	//遍历题目map,并回答问题
	var lastAnswer utils.Answer
	var lastProblem string
	for k, v := range topic.YingHuaExamTopics {

		standardProblem := v.TurnProblem() //转为标准问题结构体
		answer, err := standardProblem.ApiQueRequest(queBankUrl, 5, nil)
		answer = AnswerTurnResult(answer, v) //转换答案

		if err != nil {
			log.Print(log.INFO, `[`, userCache.Account, `] `, log.BoldRed, "Ai异常，返回信息：", err.Error())
			os.Exit(0)
		}
		//fmt.Println(aiAnswer)
		subWorkApi, err := yinghuaApi.SubmitWorkApi(*userCache, work.WorkId, k, answer, "0", 10, nil)
		if err != nil {
			log.Print(log.INFO, `[`, userCache.Account, `] `, log.BoldRed, "Ai异常，返回信息：", err.Error())
		}
		//如果提交答案服务器端返回信息异常
		if gojsonq.New().JSONString(subWorkApi).Find("msg") != "答题保存成功" {
			log.Print(log.INFO, log.BoldRed, `[`, userCache.Account, `] `, log.BoldRed, "提交答案异常，返回信息：", subWorkApi)
		}
		lastAnswer = answer
		lastProblem = k
	}
	//结束考试
	if isAutoSubExam == 1 {
		subWorkApi, err := yinghuaApi.SubmitWorkApi(*userCache, work.WorkId, lastProblem, lastAnswer, "1", 10, nil)
		//如果结束做题服务器端返回信息异常
		if gojsonq.New().JSONString(subWorkApi).Find("msg") != "提交作业成功" || err != nil {
			log.Print(log.INFO, log.BoldRed, `[`, userCache.Account, `] `, log.BoldRed, "提交试卷异常，返回信息：", subWorkApi, err.Error())
		}
	}

	return nil
}

// WorkedFinallyScoreAction 获取最终作业分数
func WorkedFinallyScoreAction(userCache *yinghuaApi.YingHuaUserCache, work YingHuaWork) (string, error) {
	detail, err := yinghuaApi.WorkedFinallyDetailApi(*userCache, work.CourseId, work.NodeId, work.WorkId, 8, nil)
	if err != nil {
		log.Print(log.INFO, `[`, userCache.Account, `] `, log.BoldRed, err.Error())
	}
	//fmt.Println(detail)
	scorePattern := `最高分：[^\d]*([\d]+)[^分]*分`
	scoreRegexp := regexp.MustCompile(scorePattern)
	scoreMatches := scoreRegexp.FindAllStringSubmatch(strings.ReplaceAll(detail, "&nbsp;", ""), -1)
	for _, scoreMatch := range scoreMatches {
		score := scoreMatch[1] //分数
		return score, nil
	}
	return "", errors.New("没有找到分数")
}
