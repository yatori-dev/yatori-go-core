package point

import (
	"github.com/thedevsaddam/gojsonq"
	action "github.com/yatori-dev/yatori-go-core/aggregation/xuexitong"
	"github.com/yatori-dev/yatori-go-core/api/entity"
	api "github.com/yatori-dev/yatori-go-core/api/xuexitong"
	"log"
	"time"
)

func ExecuteVideo(cache *api.XueXiTUserCache, p *entity.PointVideoDto) {

	if state, _ := action.VideoDtoFetchAction(cache, p); state {
		log.Printf("开始模拟播放....%d:%d开始\n", p.PlayTime, p.Duration)
		var playingTime = p.PlayTime
		var flag = 0
		for {
			if flag == 58 {
				playReport, _ := cache.VideoDtoPlayReport(p, playingTime)
				playingTime += flag
				flag = 0
				if gojsonq.New().JSONString(playReport).Find("isPassed").(bool) == true {
					log.Println("播放结束")
					playingTime = p.Duration
					break
				}
			}
			flag += 1
			log.Printf("播放中....%d:%d\n", playingTime, p.Duration)
			time.Sleep(time.Second * 1)
		}
	} else {
		log.Fatal("视频解析失败")
	}
}
