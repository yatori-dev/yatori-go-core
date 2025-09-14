package point

import (
	"fmt"
	"strings"
	"time"

	"github.com/thedevsaddam/gojsonq"
	xuexitong2 "github.com/yatori-dev/yatori-go-core/aggregation/xuexitong"
	"github.com/yatori-dev/yatori-go-core/api/entity"
	"github.com/yatori-dev/yatori-go-core/api/xuexitong"
	log2 "github.com/yatori-dev/yatori-go-core/utils/log"
)

// 外链任务点学习
func ExecuteLive(cache *xuexitong.XueXiTUserCache, p *entity.PointLiveDto, Uparam string, watchMoment float64, isSaveTime bool) (string, error) {
	report, err := cache.LiveWatchMomentReport(p, Uparam, watchMoment, 3, nil)
	if err != nil {
		// 触发500
		if err != nil && strings.Contains(err.Error(), "status code: 500") {
			xuexitong2.ReLogin(cache) //重登
			report, err = cache.LiveWatchMomentReport(p, Uparam, watchMoment, 3, nil)
		} else {
			report, err = cache.LiveWatchMomentReport(p, Uparam, watchMoment, 3, nil)
		}
	}
	if err != nil {
		return "", err
	}
	flag := int(gojsonq.New().JSONString(report).Find("result").(float64))

	if flag == 1 {
		log2.Print(log2.DEBUG, "(", p.Title, ")提交成功")
	} else {
		log2.Print(log2.DEBUG, "(", p.Title, ")外链任务点无法正常学习：返回：(", gojsonq.New().JSONString(report).Find("msg"), ")")
	}

	if isSaveTime {
		report1, err1 := cache.LiveSaveTimePcReport(p, 3, nil)

		if err1 != nil {
			// 触发500
			if err1 != nil && strings.Contains(err1.Error(), "status code: 500") {
				xuexitong2.ReLogin(cache) //重登
				report1, err1 = cache.LiveSaveTimePcReport(p, 3, nil)
			} else {
				report1, err1 = cache.LiveSaveTimePcReport(p, 3, nil)
			}
		}
		if err1 != nil {
			return "", err1
		}

		if strings.Contains(report1, "@success") {
			log2.Print(log2.DEBUG, "(", p.Title, ")提交成功")
		} else {
			log2.Print(log2.DEBUG, "(", p.Title, ")外链任务点无法正常学习：返回：(", gojsonq.New().JSONString(report).Find("msg"), ")")
		}
		return report1, nil
	}
	return report, nil
}

// 测试用的Live直播学时函数
func ExecuteLiveTest(cache *xuexitong.XueXiTUserCache, p *entity.PointLiveDto) {
	report, err := cache.LiveRelationReport(p)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(report)
	param, s := cache.PullLiveUParam(p.LiveId)
	var playTime float64 = float64(s)
	submitTotal := 0
	for {
		var live string
		var err error
		if submitTotal >= 5 {
			live, err = ExecuteLive(cache, p, param, playTime, true)
			submitTotal = 0
		} else {
			live, err = ExecuteLive(cache, p, param, playTime, false)
		}

		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(live)
		playTime += 10
		submitTotal++
		time.Sleep(10 * time.Second)
	}
}
