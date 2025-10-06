package icve

import (
	"errors"
	"fmt"
	"io"
	log2 "log"
	"net/http"
	"time"

	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/api/icve"
)

type IcveCourse struct {
	Id           string
	CourseId     string
	CourseName   string
	SchoolName   string
	TeacherName  string
	CourseType   string //课程类型，是资源库课程，还是职教云课程
	CourseInfoId string
	WeekStr      string //是否进行中
	Status       string //课程状态，3代表已结束（一般就不学这课程了，直接跳过）
}

// 课程任务节点
type IcveCourseNode struct {
	Id           string
	CourseId     string
	CourseInfoId string
	ParentId     string
	Name         string  //任务点名称
	FileType     string  //任务点类型
	FileUrl      string  //资源短连接
	IsLook       bool    //是否看过
	Speed        float64 //完成度，全部完成则是100%
	TotalNum     float64 //任务点需要观看的总时长，比如说视屏的总时长
}

// 拉取资源库课程
func PullZYKCourseAction(cache *icve.IcveUserCache) ([]IcveCourse, error) {
	courseList := make([]IcveCourse, 0)
	//courseResult, err := cache.PullZykCourse1Api()
	courseResult, err := cache.PullZykCourse2Api()
	if err != nil {
		log2.Fatal(err)
	}
	resultCode := gojsonq.New().JSONString(courseResult).Find("code")
	if resultCode == nil {
		return []IcveCourse{}, errors.New(courseResult)
	}
	if int(resultCode.(float64)) != 200 {
		return []IcveCourse{}, errors.New(courseResult)
	}
	rowsJson := gojsonq.New().JSONString(courseResult).Find("rows")
	if rows, ok := rowsJson.([]interface{}); ok {
		for _, row := range rows {
			if courseData, ok1 := row.(map[string]interface{}); ok1 {
				course := IcveCourse{
					CourseType: "资源库",
				}
				id := courseData["id"]
				if id != nil {
					course.Id = id.(string)
				}
				courseId := courseData["courseId"]
				if courseId != nil {
					course.CourseId = courseId.(string)
				}
				courseName := courseData["courseName"]
				if courseName != nil {
					course.CourseName = courseName.(string)
				}
				schoolName := courseData["schoolName"]
				if schoolName != nil {
					course.SchoolName = schoolName.(string)
				}
				courseInfoId := courseData["courseInfoId"]
				if courseInfoId != nil {
					course.CourseInfoId = courseInfoId.(string)
				}
				weekStr := courseData["weekStr"]
				if weekStr != nil {
					course.WeekStr = weekStr.(string)
				}
				status := courseData["status"]
				if status != nil {
					course.Status = status.(string)
				}
				courseList = append(courseList, course)
			}
		}
	}
	return courseList, nil
}

// 拉取任务节点
func PullZYKCourseNodeAction(cache *icve.IcveUserCache, course IcveCourse) ([]IcveCourseNode, error) {
	nodeList := make([]IcveCourseNode, 0)
	//拉取根目录
	rootResult, err := cache.PullRootNodeListApi(course.CourseInfoId)
	if err != nil {
		log2.Fatal(err)
	}

	rootJson := gojsonq.New().JSONString(rootResult).Get()
	if rootJson == "" {
		return []IcveCourseNode{}, errors.New(rootResult)
	}
	if root, ok := rootJson.([]interface{}); ok {
		for nodeJson := range root {
			if nodeData, ok1 := root[nodeJson].(map[string]interface{}); ok1 {
				//parentId := nodeData["id"].(string)
				nodes, err1 := pullNode(cache, nodeData, 1, course)
				if err1 != nil {
					log2.Fatal(err1)
				}
				nodeList = append(nodeList, nodes...)
			}
		}
	}
	return nodeList, nil
}

// 递归拉取节点
func pullNode(cache *icve.IcveUserCache, root map[string]interface{}, level int, course IcveCourse) ([]IcveCourseNode, error) {
	nodeList := make([]IcveCourseNode, 0)
	parentId := root["id"].(string)
	fileType := root["fileType"]
	//level检测
	switch fileType {
	case "父节点":
		nodeResult, err1 := cache.PullZykNodeListApi(1, parentId, course.CourseInfoId)
		if err1 != nil {
			log2.Fatal(err1)
		}
		//继续递归
		rootJson := gojsonq.New().JSONString(nodeResult).Get()
		if nodes, ok := rootJson.([]interface{}); ok {
			for _, nodeJson := range nodes {
				if node, ok1 := nodeJson.(map[string]interface{}); ok1 {
					result, err2 := pullNode(cache, node, level, course)
					if err2 != nil {
						log2.Fatal(err2)
					}
					nodeList = append(nodeList, result...)
				}
			}
		}
	case "子节点":
		nodeResult, err1 := cache.PullZykNodeListApi(level+1, parentId, course.CourseInfoId)
		if err1 != nil {
			log2.Fatal(err1)
		}
		//继续递归
		rootJson := gojsonq.New().JSONString(nodeResult).Get()
		if nodes, ok := rootJson.([]interface{}); ok {
			for _, nodeJson := range nodes {
				if node, ok1 := nodeJson.(map[string]interface{}); ok1 {
					result, err2 := pullNode(cache, node, level+1, course)
					if err2 != nil {
						log2.Fatal(err2)
					}
					nodeList = append(nodeList, result...)
				}
			}
		}
	case "测验":
		fallthrough
	case "mp4":
		fallthrough
	case "mp3":
		//cache.PullZykNodeInfoApi(root)
		fallthrough
	case "zip":
		fallthrough
	case "pdf":
		fallthrough
	case "doc":
		fallthrough
	case "docx":
		fallthrough
	case "ppt":
		fallthrough
	case "pptx":
		node := IcveCourseNode{
			Id:           root["id"].(string),
			CourseId:     root["courseId"].(string),
			CourseInfoId: root["courseInfoId"].(string),
			ParentId:     root["parentId"].(string),
			Name:         root["name"].(string),
			FileType:     root["fileType"].(string),
		}

		isLook := root["isLook"]
		if isLook != nil {
			node.IsLook = isLook.(bool)
		}
		//记录节点字段提取
		studentStudyRecord := root["studentStudyRecord"]
		if studentStudyRecord != nil {
			speed := root["studentStudyRecord"].(map[string]interface{})["speed"]
			if speed != nil {
				node.Speed = speed.(float64)
			}
		}

		nodeList = append(nodeList, node)
	}

	return nodeList, nil
}

// 资源库课程学时提交
func SubmitZYKStudyTimeAction(cache *icve.IcveUserCache, node IcveCourseNode) (string, error) {
	//if node.Speed >= 100 {
	//	log2.Printf("(%s)任务点已完成，已自动跳过", node.Name)
	//	return "", nil
	//}
	//参数完善-------------------
	err2 := GetNodeDurationAction(cache, &node)
	if err2 != nil {
		//log2.Fatal(err2)
		return "", err2
	}
	//学习
	studyResult, err := cache.SubmitZYKStudyTimeApi(node.CourseInfoId, "", node.ParentId, int(node.TotalNum), node.Id, cache.UserId, int(node.TotalNum), int(node.TotalNum), int(node.TotalNum))
	if err != nil {
		//log2.Fatal(err)
		return "", err
	}
	//fmt.Println(api)
	return studyResult, nil
}

// 获取任务点时长
func GetNodeDurationAction(cache *icve.IcveUserCache, node *IcveCourseNode) error {
	infoJson, err := cache.PullZykNodeInfoApi(node.Id)
	if err != nil {
		log2.Fatal(err)
	}
	resultCode := gojsonq.New().JSONString(infoJson).Find("code")
	if resultCode == nil {
		return errors.New(infoJson)
	}
	if int(resultCode.(float64)) != 200 {
		return errors.New(infoJson)
	}
	//资源链接
	urlShort := gojsonq.New().JSONString(infoJson).Find("data.urlShort")
	if urlShort != nil {
		node.FileUrl = urlShort.(string)
	}
	switch node.FileType {
	case "mp3":
		fileUrl := gojsonq.New().JSONString(infoJson).Find("data.fileUrl")
		if fileUrl == nil {
			return errors.New(infoJson)
		}
		duration, err1 := GetMP3Duration(fileUrl.(string))
		if err1 != nil {
			log2.Fatal(err1)
		}
		node.TotalNum = duration
	case "mp4":
		resultStatus, err1 := cache.PullZykNodeDurationApi(node.FileUrl)
		if err1 != nil {
			log2.Fatal(err1)
		}
		durationStr := gojsonq.New().JSONString(resultStatus).Find("args.duration")
		if durationStr == nil {
			return errors.New(resultStatus)
		}
		duration, err1 := DurationToSeconds(durationStr.(string))
		if err1 != nil {
			log2.Fatal(err1)
		}
		node.TotalNum = duration
	case "ppt":
		fallthrough
	case "pptx":
		fallthrough
	case "doc":
		fallthrough
	case "docx":
		fallthrough
	case "pdf":
		resultStatus, err1 := cache.PullZykNodeDurationApi(node.FileUrl)
		if err1 != nil {
			log2.Fatal(err1)
		}
		durationStr := gojsonq.New().JSONString(resultStatus).Find("args.page_count")
		if durationStr == nil {
			return errors.New(resultStatus)
		}
		node.TotalNum = durationStr.(float64)
	case "zip":
		node.TotalNum = 1
	}
	return nil
}

// 获取MP3时长
func GetMP3Duration(mp3URL string) (float64, error) {
	// 发起 HTTP 请求
	resp, err := http.Get(mp3URL)
	if err != nil {
		return 0, fmt.Errorf("下载失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, errors.New("无法下载 MP3 文件")
	}

	// 读取前 100KB 数据
	buf := make([]byte, 100000)
	n, err := io.ReadFull(resp.Body, buf)
	if err != nil && err != io.ErrUnexpectedEOF {
		return 0, fmt.Errorf("读取数据失败: %v", err)
	}
	data := buf[:n]

	// MPEG 版本映射表
	mpegVersions := map[uint8]string{
		0b00: "MPEG-2.5",
		0b01: "", // 保留
		0b10: "MPEG-2",
		0b11: "MPEG-1",
	}

	// 采样率表
	sampleRates := map[string][]int{
		"MPEG-1":   {44100, 48000, 32000, 0},
		"MPEG-2":   {22050, 24000, 16000, 0},
		"MPEG-2.5": {11025, 12000, 8000, 0},
	}

	// 比特率表
	bitrates := map[string][]int{
		"MPEG-1": {0, 32, 40, 48, 56, 64, 80, 96, 112, 128, 160, 192, 224, 256, 320},
		"MPEG-2": {0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160},
	}

	// 获取文件大小
	contentLength := resp.ContentLength
	if contentLength <= 0 {
		return 0, errors.New("无法获取文件大小")
	}

	var frameCount int
	var totalBitrate float64

	i := 0
	for i < len(data)-4 {
		// 查找帧同步字节 0xFFEx
		if data[i] == 0xFF && (data[i+1]&0xE0) == 0xE0 {
			// MPEG 版本
			mpegVersionID := (data[i+1] >> 3) & 0x03
			mpegVersion := mpegVersions[mpegVersionID]
			if mpegVersion == "" {
				i++
				continue
			}

			// 比特率索引
			bitrateIndex := (data[i+2] >> 4) & 0x0F
			if bitrateIndex == 0 || bitrateIndex >= uint8(len(bitrates[mpegVersion])) {
				i++
				continue
			}
			bitrate := float64(bitrates[mpegVersion][bitrateIndex] * 1000)

			// 采样率索引
			sampleRateIndex := (data[i+2] >> 2) & 0x03
			if sampleRateIndex >= uint8(len(sampleRates[mpegVersion])) {
				i++
				continue
			}
			sampleRate := float64(sampleRates[mpegVersion][sampleRateIndex])
			if sampleRate == 0 {
				i++
				continue
			}

			// padding
			padding := (data[i+2] >> 1) & 0x01

			// 计算帧大小
			frameSize := int((144*bitrate)/sampleRate + float64(padding))

			totalBitrate += bitrate
			frameCount++

			i += frameSize
		} else {
			i++
		}
	}

	if frameCount == 0 {
		return 0.0, nil
	}

	// 平均比特率
	avgBitrate := totalBitrate / float64(frameCount)

	// 时长（秒）
	duration := (float64(contentLength) * 8) / avgBitrate
	return duration, nil
}

// 时间转换
func DurationToSeconds(s string) (float64, error) {
	// 解析为时间格式（支持带微秒或纳秒）
	t, err := time.Parse("15:04:05.9999999", s)
	if err != nil {
		return 0, err
	}

	// 计算总秒数
	seconds := float64(t.Hour()*3600 + t.Minute()*60 + t.Second())
	seconds += float64(t.Nanosecond()) / 1e9

	return seconds, nil
}
