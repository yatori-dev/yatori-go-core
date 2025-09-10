package xuexitong

import (
	"strings"

	"github.com/yatori-dev/yatori-go-core/api/entity"
	"github.com/yatori-dev/yatori-go-core/api/xuexitong"
)

// VideoSubmitStudyTimeAction 视屏学时提交

func VideoSubmitStudyTimeAction(cache *xuexitong.XueXiTUserCache, p *entity.PointVideoDto, playingTime int, mode /*0为PC模式，1为PE模式*/, isdrag int /*提交模式，0代表正常视屏播放提交，2代表暂停播放状态，3代表着点击开始播放状态*/) (string, error) {
	var playReport string
	var err error
	if mode == 0 {
		playReport, err = cache.VideoSubmitStudyTimeApi(p, playingTime, isdrag, 8, nil)
	} else if mode == 1 {
		playReport, err = cache.VideoSubmitStudyTimePEApi(p, playingTime, isdrag, 8, nil)
	}
	//如果遇到500
	if err != nil && strings.Contains(err.Error(), "failed to fetch video, status code: 500") {
		ReLogin(cache) //重登
		if mode == 0 {
			playReport, err = cache.VideoSubmitStudyTimeApi(p, playingTime, isdrag, 8, nil)
		} else if mode == 1 {
			playReport, err = cache.VideoSubmitStudyTimePEApi(p, playingTime, isdrag, 8, nil)
		}
	}
	//触发202
	if err != nil && strings.Contains(err.Error(), "failed to fetch video, status code: 202") {
		ReLogin(cache) //重登
		if mode == 0 {
			playReport, err = cache.VideoSubmitStudyTimeApi(p, playingTime, isdrag, 8, nil)
		} else if mode == 1 {
			playReport, err = cache.VideoSubmitStudyTimePEApi(p, playingTime, isdrag, 8, nil)
		}
	}
	if err != nil {
		return "", err
	}

	return playReport, nil
}
