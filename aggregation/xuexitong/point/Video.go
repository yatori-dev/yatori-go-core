package point

import (
	"fmt"
	"github.com/thedevsaddam/gojsonq"
	action "github.com/yatori-dev/yatori-go-core/aggregation/xuexitong"
	"github.com/yatori-dev/yatori-go-core/api/entity"
	api "github.com/yatori-dev/yatori-go-core/api/xuexitong"
	lg "github.com/yatori-dev/yatori-go-core/utils/log"
	"log"
	"strconv"
	"time"
)

// 常规刷视屏逻辑
func ExecuteVideo(cache *api.XueXiTUserCache, p *entity.PointVideoDto) {
	if state, _ := action.VideoDtoFetchAction(cache, p); state {
		var playingTime = p.PlayTime
		for {
			if p.Duration-playingTime >= 58 {
				playReport, err := cache.VideoDtoPlayReport(p, playingTime, 0, 4, nil)
				if gojsonq.New().JSONString(playReport).Find("isPassed") == nil || err != nil {
					lg.Print(lg.INFO, `[`, cache.Name, `] `, lg.BoldRed, "提交学时接口访问异常，返回信息：", playReport, err.Error())
					break
				}
				if gojsonq.New().JSONString(playReport).Find("isPassed").(bool) == true { //看完了，则直接退出
					lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", " 【", p.Title, "】 >>> ", "提交状态：", lg.Green, strconv.FormatBool(gojsonq.New().JSONString(playReport).Find("isPassed").(bool)), lg.Default, " ", "观看时间：", strconv.Itoa(p.Duration)+"/"+strconv.Itoa(p.Duration), " ", "观看进度：", fmt.Sprintf("%.2f", float32(p.Duration)/float32(p.Duration)*100), "%")
					break
				}
				lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", " 【", p.Title, "】 >>> ", "提交状态：", lg.Green, lg.Green, strconv.FormatBool(gojsonq.New().JSONString(playReport).Find("isPassed").(bool)), lg.Default, " ", "观看时间：", strconv.Itoa(playingTime)+"/"+strconv.Itoa(p.Duration), " ", "观看进度：", fmt.Sprintf("%.2f", float32(playingTime)/float32(p.Duration)*100), "%")
				playingTime = playingTime + 58
				time.Sleep(58 * time.Second)
			} else if p.Duration-playingTime < 58 {
				playReport, err := cache.VideoDtoPlayReport(p, p.Duration, 2, 4, nil)
				if gojsonq.New().JSONString(playReport).Find("isPassed") == nil || err != nil {
					lg.Print(lg.INFO, `[`, cache.Name, `] `, lg.BoldRed, "提交学时接口访问异常，返回信息：", playReport)
					break
				}
				if gojsonq.New().JSONString(playReport).Find("isPassed").(bool) == true { //看完了，则直接退出
					lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", " 【", p.Title, "】 >>> ", "提交状态：", lg.Green, lg.Green, strconv.FormatBool(gojsonq.New().JSONString(playReport).Find("isPassed").(bool)), lg.Default, " ", "观看时间：", strconv.Itoa(p.Duration)+"/"+strconv.Itoa(p.Duration), " ", "观看进度：", fmt.Sprintf("%.2f", float32(p.Duration)/float32(p.Duration)*100), "%")
					break
				}
				lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", " 【", p.Title, "】 >>> ", "提交状态：", lg.Green, lg.Green, strconv.FormatBool(gojsonq.New().JSONString(playReport).Find("isPassed").(bool)), lg.Default, " ", "观看时间：", strconv.Itoa(p.Duration)+"/"+strconv.Itoa(p.Duration), " ", "观看进度：", fmt.Sprintf("%.2f", float32(p.Duration)/float32(p.Duration)*100), "%")
				time.Sleep(time.Duration(p.Duration-playingTime) * time.Second)
			}
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
