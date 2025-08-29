package point

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/thedevsaddam/gojsonq"
	action "github.com/yatori-dev/yatori-go-core/aggregation/xuexitong"
	"github.com/yatori-dev/yatori-go-core/api/entity"
	api "github.com/yatori-dev/yatori-go-core/api/xuexitong"
	"github.com/yatori-dev/yatori-go-core/utils"
	log2 "github.com/yatori-dev/yatori-go-core/utils/log"
)

// 常规刷视屏逻辑
func ExecuteVideo(cache *api.XueXiTUserCache, p *entity.PointVideoDto, key, courseCpi int) {

	log.Println("触发人脸识别，正在进行绕过...")
	pullJson, img, err2 := cache.GetHistoryFaceImg("")
	if err2 == nil {
		log2.Print(log2.DEBUG, pullJson, err2)

		disturbImage := utils.ImageRGBDisturb(img)
		uuid, qrEnc, ObjectId, _, err := action.PassFaceAction2(cache, p.CourseID, p.ClassID, p.Cpi, fmt.Sprintf("%d", p.KnowledgeID), p.Enc, p.JobID, p.ObjectID, p.Mid, p.RandomCaptureTime, disturbImage)
		if err != nil {
			log.Println(uuid, qrEnc, ObjectId, err.Error())
		}
		//p.VideoFaceCaptureEnc = successEnc
		log.Println("绕过成功")
	}

	if state, _ := action.VideoDtoFetchAction(cache, p); state {
		log.Printf("(%s)开始模拟播放....%d:%d开始\n", p.Title, p.PlayTime, p.Duration)
		var playingTime = p.PlayTime
		var flag = 0
		stopVal := 0
		for {
			//如果到了人脸识别时间
			if fmt.Sprintf("%d", flag) == p.RandomCaptureTime {
				log.Println("到达人脸识别时间，正在进行绕过...")
				//img, err2 := utils.GetFaceBase64()
				pullJson, image, err2 := cache.GetHistoryFaceImg("")
				//image, err2 := utils.LoadImage("C:\\Users\\Administrator\\Desktop\\IMG_20250529_020045.jpg")
				if err2 != nil {
					log2.Print(log2.DEBUG, pullJson, err2)
					os.Exit(0)
				}
				originHash, _ := utils.CalculateJPEGMD5(image, 90)
				fmt.Println("原图像hash-> ", originHash)
				//disturbImage := image
				disturbImage := utils.ImageRGBDisturb(image)
				//disturbImage := utils.DisturbByShufflingBlocks(image, 1)
				utils.SaveImageAsJPEG(disturbImage, "E:\\yatori-dev\\yatori-go-core\\face\\test.jpg")
				nowHash, _ := utils.CalculateJPEGMD5(disturbImage, 90)
				fmt.Println("扰乱后Hash-> ", nowHash)
				uuid, qrEnc, ObjectId, successEnc, err := action.PassFaceAction3(cache, p.CourseID, p.ClassID, p.Cpi, fmt.Sprintf("%d", p.KnowledgeID), p.Enc, p.JobID, p.ObjectID, p.Mid, p.RandomCaptureTime, disturbImage)
				if err != nil {
					log.Println(uuid, qrEnc, ObjectId, err.Error())
				}
				p.VideoFaceCaptureEnc = successEnc
				courseId, _ := strconv.Atoi(p.CourseID)
				time.Sleep(5 * time.Second)
				card, enc, err := action.PageMobileChapterCardAction(
					cache, key, courseId, p.KnowledgeID, p.CardIndex, courseCpi)
				if err != nil {
					log.Fatal(err)
				}
				p.Enc = enc
				p.AttachmentsDetection(card)
				time.Sleep(5 * time.Second)
				playReport, err := cache.VideoSubmitStudyTime(p, playingTime, 3, 8, nil)
				if err != nil {
					log.Println(uuid, qrEnc, ObjectId, playReport, err.Error())
				}
				stopVal += 1
				log.Println("绕过成功")
			}
			if flag == 58 {
				//playReport, err := cache.VideoDtoPlayReport(p, playingTime, 3, 8, nil)
				playReport, err := cache.VideoSubmitStudyTime(p, playingTime, 0, 8, nil)
				log.Println(playReport, err)
				if err != nil {
					if strings.Contains(err.Error(), "failed to fetch video, status code: 403") || strings.Contains(err.Error(), "failed to fetch video, status code: 404") { //触发403立即使用人脸检测

						log.Println("触发人脸识别，正在进行绕过...")
						pullJson, img, err2 := cache.GetHistoryFaceImg("")
						if err2 != nil {
							log2.Print(log2.DEBUG, pullJson, err2)
							os.Exit(0)
						}
						disturbImage := utils.ImageRGBDisturb(img)
						uuid, qrEnc, ObjectId, successEnc, err := action.PassFaceAction3(cache, p.CourseID, p.ClassID, p.Cpi, fmt.Sprintf("%d", p.KnowledgeID), p.Enc, p.JobID, p.ObjectID, p.Mid, p.RandomCaptureTime, disturbImage)
						if err != nil {
							log.Println(uuid, qrEnc, ObjectId, err.Error())
						}
						p.VideoFaceCaptureEnc = successEnc
						courseId, _ := strconv.Atoi(p.CourseID)
						time.Sleep(5 * time.Second)
						card, enc, err := action.PageMobileChapterCardAction(
							cache, key, courseId, p.KnowledgeID, p.CardIndex, courseCpi)
						if err != nil {
							log.Fatal(err)
						}
						p.Enc = enc
						p.AttachmentsDetection(card)
						time.Sleep(5 * time.Second)
						playReport, err := cache.VideoSubmitStudyTime(p, playingTime, 3, 8, nil)
						if err != nil {
							log.Println(uuid, qrEnc, ObjectId, playReport, err.Error())
						}
						stopVal += 1
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
						pullJson, img, err2 := cache.GetHistoryFaceImg("")
						if err2 != nil {
							log2.Print(log2.DEBUG, pullJson, err2)
							os.Exit(0)
						}
						disturbImage := utils.ImageRGBDisturb(img)
						uuid, qrEnc, ObjectId, successEnc, err := action.PassFaceAction3(cache, p.CourseID, p.ClassID, p.Cpi, fmt.Sprintf("%d", p.KnowledgeID), p.Enc, p.JobID, p.ObjectID, p.Mid, p.RandomCaptureTime, disturbImage)
						if err != nil {
							log.Println(uuid, qrEnc, ObjectId, err.Error())
						}
						p.VideoFaceCaptureEnc = successEnc
						courseId, _ := strconv.Atoi(p.CourseID)
						time.Sleep(5 * time.Second)
						card, enc, err := action.PageMobileChapterCardAction(
							cache, key, courseId, p.KnowledgeID, p.CardIndex, courseCpi)
						if err != nil {
							log.Fatal(err)
						}
						p.Enc = enc
						p.AttachmentsDetection(card)
						time.Sleep(5 * time.Second)
						playReport, err := cache.VideoSubmitStudyTime(p, playingTime, 3, 8, nil)
						if err != nil {
							log.Println(uuid, qrEnc, ObjectId, playReport, err.Error())
						}
						stopVal += 1
						log.Println("绕过成功")
						continue
					}
				}
				stopVal = 0
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
