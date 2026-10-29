package point

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/thedevsaddam/gojsonq"
	action "github.com/yatori-dev/yatori-go-core/internal/aggregation/xuexitong"
	api "github.com/yatori-dev/yatori-go-core/internal/api/xuexitong"
	"github.com/yatori-dev/yatori-go-core/models/ctype"
	"github.com/yatori-dev/yatori-go-core/utils"
	log2 "github.com/yatori-dev/yatori-go-core/utils/log"
)

// playReportPassed 安全读取播放上报结果。
// 响应异常(空响应、验证码页、字段缺失)时 Find 返回 nil，
// 直接 .(bool) 会 panic 并让整个程序退出。
func playReportPassed(report string) bool {
	passed, ok := gojsonq.New().JSONString(report).Find("isPassed").(bool)
	return ok && passed
}

// 常规刷视屏逻辑
func ExecuteVideoTest(cache *api.XueXiTUserCache, p *api.PointVideoDto, key, courseCpi int) {
	if p.Type == ctype.Video {
		if state, _ := action.VideoDtoFetchAction(cache, p); state {
			log.Printf("(%s)开始模拟播放....%d:%d开始\n", p.Title, p.PlayTime, p.Duration)
			var playingTime = p.PlayTime
			var flag = 0
			stopVal := 0
			for {
				if flag == 58 {
					playReport, err := action.VideoSubmitStudyTimeAction(cache, p, playingTime, 1, 0)
					log.Println(playReport, err)
					if err != nil {
						if strings.Contains(err.Error(), "failed to fetch video, status code: 403") || strings.Contains(err.Error(), "failed to fetch video, status code: 404") { //触发403立即使用人脸检测

							log.Println("触发人脸识别，正在进行绕过...")
							pullJson, img, err2 := cache.GetHistoryFaceImg("")
							if err2 == nil {
								log2.Print(log2.DEBUG, pullJson, err2)

								disturbImage := utils.ImageRGBDisturb(img)
								uuid, qrEnc, ObjectId, _, err := action.PassFacePCAction(cache, p.CourseID, p.ClassID, p.Cpi, fmt.Sprintf("%d", p.KnowledgeID), p.Enc, p.JobID, p.ObjectID, p.Mid, p.RandomCaptureTime, disturbImage)
								if err != nil {
									log.Printf("人脸识别失败: %v (uuid=%s qrEnc=%s objectId=%s)", err, uuid, qrEnc, ObjectId)
								} else {
									log.Printf("人脸识别通过 objectId=%s", ObjectId)
								}
								//p.VideoFaceCaptureEnc = successEnc
								time.Sleep(5 * time.Second)
							}
							stopVal += 1

							continue
						}

					}
					//playReport, _ := cache.VideoSubmitStudyTime(p, playingTime, 3, 8, nil)
					playingTime += flag
					flag = 0
					if playReportPassed(playReport) {
						log.Println("播放结束")
						playingTime = p.Duration
						break
					}
					log.Printf("播放中....%d:%d\n", playingTime, p.Duration)
				} else if playingTime >= p.Duration {
					playReport, err := action.VideoSubmitStudyTimeAction(cache, p, playingTime, 1, 0)
					//playReport, err := cache.VideoDtoPlayReport(p, playingTime, 0, 8, nil)
					playingTime += 1
					log.Println(playReport, err)
					if err != nil {
						if strings.Contains(err.Error(), "failed to fetch video, status code: 403") || strings.Contains(err.Error(), "failed to fetch video, status code: 404") { //触发403立即使用人脸检测

							log.Println("触发人脸识别，正在进行绕过...")
							//image, _ := utils.LoadImage("E:\\Yatori-Dev\\yatori-go-core\\face\\test2.jpg")
							//pullJson, img, err2 := cache.GetHistoryFaceImg("")
							//if err2 != nil {
							//	log2.Print(log2.DEBUG, pullJson, err2)
							//	os.Exit(0)
							//}
							//disturbImage := utils.ImageRGBDisturb(img)
							//uuid, qrEnc, ObjectId, successEnc, err := action.PassFaceAction3(cache, p.CourseID, p.ClassID, p.Cpi, fmt.Sprintf("%d", p.KnowledgeID), p.Enc, p.JobID, p.ObjectID, p.Mid, p.RandomCaptureTime, disturbImage)
							//if err != nil {
							//	log.Println(uuid, qrEnc, ObjectId, err.Error())
							//}
							//p.VideoFaceCaptureEnc = successEnc
							//courseId, _ := strconv.Atoi(p.CourseID)
							//time.Sleep(5 * time.Second)
							//card, enc, err := action.PageMobileChapterCardAction(
							//	cache, key, courseId, p.KnowledgeID, p.CardIndex, courseCpi)
							//if err != nil {
							//	log.Fatal(err)
							//}
							//p.Enc = enc
							//p.AttachmentsDetection(card)
							//time.Sleep(5 * time.Second)
							//						playReport, err := cache.VideoSubmitStudyTime(p, playingTime, 3, 8, nil)
							//						if err != nil {
							//							log.Println(uuid, qrEnc, ObjectId, playReport, err.Error())
							//						}
							pullJson, img, err2 := cache.GetHistoryFaceImg("")
							if err2 == nil {
								log2.Print(log2.DEBUG, pullJson, err2)

								disturbImage := utils.ImageRGBDisturb(img)
								uuid, qrEnc, ObjectId, _, err := action.PassFaceAction1(cache, p.CourseID, p.ClassID, p.Cpi, fmt.Sprintf("%d", p.KnowledgeID), p.Enc, p.JobID, p.ObjectID, p.Mid, p.RandomCaptureTime, disturbImage)
								if err != nil {
									log.Printf("人脸识别失败: %v (uuid=%s qrEnc=%s objectId=%s)", err, uuid, qrEnc, ObjectId)
								} else {
									log.Printf("人脸识别通过 objectId=%s", ObjectId)
								}
								//p.VideoFaceCaptureEnc = successEnc
								time.Sleep(5 * time.Second)
							}

							stopVal += 1
							continue
						}
					}
					stopVal = 0
					if playReportPassed(playReport) {
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
	if p.Type == ctype.InsertAudio {
		if state, _ := action.VideoDtoFetchAction(cache, p); state {
			log.Printf("(%s)开始模拟播放....%d:%d开始\n", p.Title, p.PlayTime, p.Duration)
			var playingTime = p.PlayTime
			var flag = 0
			stopVal := 0
			for {
				if flag == 58 {
					playReport, err := action.AudioSubmitStudyTimeAction(cache, p, playingTime, 1, 0)
					log.Println(playReport, err)
					if err != nil {
						if strings.Contains(err.Error(), "failed to fetch video, status code: 403") || strings.Contains(err.Error(), "failed to fetch video, status code: 404") { //触发403立即使用人脸检测

							log.Println("触发人脸识别，正在进行绕过...")
							pullJson, img, err2 := cache.GetHistoryFaceImg("")
							if err2 == nil {
								log2.Print(log2.DEBUG, pullJson, err2)

								disturbImage := utils.ImageRGBDisturb(img)
								uuid, qrEnc, ObjectId, _, err := action.PassFacePCAction(cache, p.CourseID, p.ClassID, p.Cpi, fmt.Sprintf("%d", p.KnowledgeID), p.Enc, p.JobID, p.ObjectID, p.Mid, p.RandomCaptureTime, disturbImage)
								if err != nil {
									log.Printf("人脸识别失败: %v (uuid=%s qrEnc=%s objectId=%s)", err, uuid, qrEnc, ObjectId)
								} else {
									log.Printf("人脸识别通过 objectId=%s", ObjectId)
								}
								//p.VideoFaceCaptureEnc = successEnc
								time.Sleep(5 * time.Second)
							}
							stopVal += 1

							continue
						}

					}
					//playReport, _ := cache.VideoSubmitStudyTime(p, playingTime, 3, 8, nil)
					playingTime += flag
					flag = 0
					if playReportPassed(playReport) {
						log.Println("播放结束")
						playingTime = p.Duration
						break
					}
					log.Printf("播放中....%d:%d\n", playingTime, p.Duration)
				} else if playingTime >= p.Duration {
					playReport, err := action.AudioSubmitStudyTimeAction(cache, p, playingTime, 1, 0)
					//playReport, err := cache.VideoDtoPlayReport(p, playingTime, 0, 8, nil)
					playingTime += 1
					log.Println(playReport, err)
					if err != nil {
						if strings.Contains(err.Error(), "failed to fetch video, status code: 403") || strings.Contains(err.Error(), "failed to fetch video, status code: 404") { //触发403立即使用人脸检测

							log.Println("触发人脸识别，正在进行绕过...")
							pullJson, img, err2 := cache.GetHistoryFaceImg("")
							if err2 == nil {
								log2.Print(log2.DEBUG, pullJson, err2)

								disturbImage := utils.ImageRGBDisturb(img)
								uuid, qrEnc, ObjectId, _, err := action.PassFaceAction1(cache, p.CourseID, p.ClassID, p.Cpi, fmt.Sprintf("%d", p.KnowledgeID), p.Enc, p.JobID, p.ObjectID, p.Mid, p.RandomCaptureTime, disturbImage)
								if err != nil {
									log.Printf("人脸识别失败: %v (uuid=%s qrEnc=%s objectId=%s)", err, uuid, qrEnc, ObjectId)
								} else {
									log.Printf("人脸识别通过 objectId=%s", ObjectId)
								}
								//p.VideoFaceCaptureEnc = successEnc
								time.Sleep(5 * time.Second)
							}

							stopVal += 1
							continue
						}
					}
					stopVal = 0
					if playReportPassed(playReport) {
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

}
