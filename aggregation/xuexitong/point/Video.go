package point

import (
	"fmt"
	"github.com/thedevsaddam/gojsonq"
	action "github.com/yatori-dev/yatori-go-core/aggregation/xuexitong"
	"github.com/yatori-dev/yatori-go-core/api/entity"
	api "github.com/yatori-dev/yatori-go-core/api/xuexitong"
	"github.com/yatori-dev/yatori-go-core/utils"
	"log"
	"strings"
	"time"
)

// 常规刷视屏逻辑
func ExecuteVideo(cache *api.XueXiTUserCache, p *entity.PointVideoDto) {
	log.Println("自主人脸识别")
	//image, _ := utils.LoadImage("E:\\Yatori-Dev\\yatori-go-core\\face\\test2.jpg")
	image, err1 := utils.GetFaceBase64()
	if err1 != nil {
		log.Println(err1)
	}
	disturbImage := utils.ImageRGBDisturb(image)

	uuid, qrEnc, ObjectId, err := action.PassFaceAction(cache, p.CourseID, p.ClassID, p.Cpi, fmt.Sprintf("%d", p.KnowledgeID), p.Enc, p.JobID, p.ObjectID, disturbImage)
	if err != nil {
		log.Println(uuid, qrEnc, ObjectId, err.Error())
	}
	p.VideoFaceCaptureEnc = qrEnc
	log.Println("识别成功")
	if state, _ := action.VideoDtoFetchAction(cache, p); state {
		log.Printf("(%s)开始模拟播放....%d:%d开始\n", p.Title, p.PlayTime, p.Duration)
		var playingTime = p.PlayTime
		var flag = 0
		for {
			if flag == 30 {
				//monitorApi, _ := cache.MonitorApi()
				//fmt.Println(monitorApi)
			}
			if flag == 58 {
				//playReport, err := cache.VideoDtoPlayReport(p, playingTime, 3, 8, nil)
				playReport, err := cache.VideoSubmitStudyTime(p, playingTime, 0, 8, nil)
				log.Println(playReport, err)
				if err != nil {
					if strings.Contains(err.Error(), "failed to fetch video, status code: 403") || strings.Contains(err.Error(), "failed to fetch video, status code: 404") { //触发403立即使用人脸检测

						log.Println("触发人脸识别，正在进行绕过...")
						//image, _ := utils.LoadImage("E:\\Yatori-Dev\\yatori-go-core\\face\\test2.jpg")
						image, err1 := utils.GetFaceBase64()
						if err1 != nil {
							log.Println(err)
						}
						disturbImage := utils.ImageRGBDisturb(image)

						uuid, qrEnc, ObjectId, err := action.PassFaceAction(cache, p.CourseID, p.ClassID, p.Cpi, fmt.Sprintf("%d", p.KnowledgeID), p.Enc, p.JobID, p.ObjectID, disturbImage)
						if err != nil {
							log.Println(uuid, qrEnc, ObjectId, err.Error())
						}
						p.VideoFaceCaptureEnc = qrEnc
						log.Println("绕过成功")
						continue
					}
				}
				//playReport, _ := cache.VideoSubmitStudyTime(p, playingTime, 3, 8, nil)
				playingTime += flag
				flag = 0
				if gojsonq.New().JSONString(playReport).Find("isPassed").(bool) == true {
					log.Println("播放结束")
					playingTime = p.Duration
					break
				}
				log.Printf("播放中....%d:%d\n", playingTime, p.Duration)
			} else if playingTime >= p.Duration {
				playReport, err := cache.VideoSubmitStudyTime(p, playingTime, 0, 8, nil)
				//playReport, err := cache.VideoDtoPlayReport(p, playingTime, 0, 8, nil)
				playingTime += 1
				log.Println(playReport, err)
				if err != nil {
					if strings.Contains(err.Error(), "failed to fetch video, status code: 403") || strings.Contains(err.Error(), "failed to fetch video, status code: 404") { //触发403立即使用人脸检测

						log.Println("触发人脸识别，正在进行绕过...")
						//image, _ := utils.LoadImage("E:\\Yatori-Dev\\yatori-go-core\\face\\test2.jpg")
						image, err1 := utils.GetFaceBase64()
						if err1 != nil {
							log.Println(err)
						}
						disturbImage := utils.ImageRGBDisturb(image)

						uuid, qrEnc, ObjectId, err := action.PassFaceAction(cache, p.CourseID, p.ClassID, p.Cpi, fmt.Sprintf("%d", p.KnowledgeID), p.Enc, p.JobID, p.ObjectID, disturbImage)
						if err != nil {
							log.Println(uuid, qrEnc, ObjectId, err.Error())
						}
						p.VideoFaceCaptureEnc = qrEnc
						log.Println("绕过成功")
						continue
					}
				}
				if gojsonq.New().JSONString(playReport).Find("isPassed").(bool) == true {
					log.Println("播放结束")
					playingTime = p.Duration
					break
				}
				log.Printf("播放中....%d:%d\n", playingTime, p.Duration)
			}
			flag += 1
			time.Sleep(time.Second * 1)
		}
	} else {
		log.Fatal("视频解析失败")
	}
}

// 第二套方案
func ExecuteVideo2(cache *api.XueXiTUserCache, p *entity.PointVideoDto) {

}

// 秒刷视屏逻辑
func ExecuteFastVideo(cache *api.XueXiTUserCache, p *entity.PointVideoDto) {
	if state, _ := action.VideoDtoFetchAction(cache, p); state {
		log.Printf("(%s)开始模拟播放....%d:%d开始\n", p.Title, p.PlayTime, p.Duration)
		var playingTime = p.PlayTime
		for {
			playReport, _ := cache.VideoSubmitStudyTime(p, playingTime, 3, 8, nil)
			if gojsonq.New().JSONString(playReport).Find("isPassed").(bool) == true {
				log.Println("播放结束")
				break
			}
			playingTime += 16
			log.Printf("播放中....%d:%d\n", playingTime, p.Duration)
			time.Sleep(time.Second)
		}
	} else {
		log.Fatal("视频解析失败")
	}
}
