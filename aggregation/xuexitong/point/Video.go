package point

import (
	"fmt"
	"github.com/thedevsaddam/gojsonq"
	action "github.com/yatori-dev/yatori-go-core/aggregation/xuexitong"
	"github.com/yatori-dev/yatori-go-core/api/entity"
	api "github.com/yatori-dev/yatori-go-core/api/xuexitong"
	"log"
	"strings"
	"time"
)

// 常规刷视屏逻辑
func ExecuteVideo(cache *api.XueXiTUserCache, p *entity.PointVideoDto) {

	if state, _ := action.VideoDtoFetchAction(cache, p); state {
		log.Printf("(%s)开始模拟播放....%d:%d开始\n", p.Title, p.PlayTime, p.Duration)
		var playingTime = p.PlayTime
		var flag = 0
		for {
			if flag == 7 {
				playReport, err := cache.VideoDtoPlayReport(p, playingTime, 3, 8, nil)
				if err != nil {
					if strings.Contains(err.Error(), "failed to fetch video, status code: 403") || strings.Contains(err.Error(), "failed to fetch video, status code: 404") { //触发403立即使用人脸检测
						//uuid, qrEnc, err := cache.GetFaceQrCodeApi1(p.CourseID, p.ClassID, fmt.Sprintf("%d", p.KnowledgeID), p.Cpi)
						//if err != nil {
						//	fmt.Println(err)
						//}

						uuid, qrEnc, err := cache.GetFaceQrCodeApi2(p.CourseID, p.ClassID, p.Cpi)
						if err != nil {
							fmt.Println(err)
						}
						//获取token
						tokenJson, err := cache.GetFaceUpLoadToken()
						token := gojsonq.New().JSONString(tokenJson).Find("_token").(string)
						if err != nil {
							fmt.Println(err)
						}
						//上传人脸
						ObjectId, err := cache.UploadFaceImage(token, "C:\\Users\\Administrator\\Desktop\\img8.jpg")
						if err != nil {
							fmt.Println(err)
						}
						//uuid, qrEnc, err := cache.GetFaceQrCodeApi2(p.CourseID, p.ClassID, p.Cpi)
						//if err != nil {
						//	fmt.Println(err)
						//}
						plan1Api, err := cache.GetCourseFaceQrPlan1Api(p.CourseID, p.ClassID, uuid, ObjectId, qrEnc, "0")
						if err != nil {
							fmt.Println(err)
						}
						fmt.Println(plan1Api)
						plan2Api, err := cache.GetCourseFaceQrPlan2Api(p.ClassID, p.CourseID, fmt.Sprintf("%d", p.KnowledgeID), p.Cpi, ObjectId)
						if err != nil {
							fmt.Println(err)
						}
						fmt.Println(plan2Api)
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
				playReport, _ := cache.VideoDtoPlayReport(p, playingTime, 0, 8, nil)
				playingTime += 1
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
