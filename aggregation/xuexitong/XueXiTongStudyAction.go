package xuexitong

import (
	"github.com/yatori-dev/yatori-go-core/api/entity"
	"github.com/yatori-dev/yatori-go-core/api/xuexitong"
	"strings"
	"time"
)

// VideoSubmitStudyTimeAction 视屏学时提交
func VideoSubmitStudyTimeAction(cache *xuexitong.XueXiTUserCache, p *entity.PointVideoDto, playingTIme int, isdrag int) (string, error) {
	playReport, err := cache.VideoSubmitStudyTime(p, playingTIme, isdrag, 8, nil)
	if err != nil {
		//预防202
		if strings.Contains(err.Error(), "failed to fetch video, status code: 202") {
			PassVerAnd202(cache) //绕过202
			time.Sleep(5 * time.Second)
			playReport, err = cache.VideoSubmitStudyTime(p, playingTIme, isdrag, 8, nil)
		}
		//预防404
		if strings.Contains(err.Error(), "failed to fetch video, status code: 404") { //触发202立即使用人脸检测
			time.Sleep(5 * time.Second)
			playReport, err = cache.VideoSubmitStudyTime(p, playingTIme, isdrag, 8, nil)
		}
	}
	return playReport, err
}
