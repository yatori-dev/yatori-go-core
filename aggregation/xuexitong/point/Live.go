package point

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/thedevsaddam/gojsonq"
	xuexitong2 "github.com/yatori-dev/yatori-go-core/aggregation/xuexitong"
	"github.com/yatori-dev/yatori-go-core/api/xuexitong"
	log2 "github.com/yatori-dev/yatori-go-core/utils/log"
)

// 外链任务点学习
func ExecuteLive(cache *xuexitong.XueXiTUserCache, p *xuexitong.PointLiveDto) (string, error) {

	report1, err1 := cache.LiveSaveTimePcReport(p, 3, nil)

	// 触发500
	if err1 != nil && strings.Contains(err1.Error(), "status code: 500") {
		xuexitong2.ReLogin(cache) //重登
		report1, err1 = cache.LiveSaveTimePcReport(p, 3, nil)
	}

	if err1 != nil {
		return "", err1
	}

	if strings.Contains(report1, "@success") {
		log2.Print(log2.DEBUG, "(", p.Title, ")提交成功")
	} else {
		log2.Print(log2.DEBUG, "(", p.Title, ")外链任务点无法正常学习：返回：(", gojsonq.New().JSONString(report1).Find("msg"), ")")
	}
	return report1, nil
}

// {
// "msg": "关联建立成功",
// "status": true
// }
// 建立直播联系
func LiveCreateRelationAction(cache *xuexitong.XueXiTUserCache, p *xuexitong.PointLiveDto) (string, error) {
	report, err := cache.LiveRelationReport(p, 3, nil)

	// 触发500
	if err != nil && strings.Contains(err.Error(), "status code: 500") {
		xuexitong2.ReLogin(cache) //重登
		report, err = cache.LiveRelationReport(p, 3, nil)
	}

	if err != nil {
		return "", err
	}
	if !gojsonq.New().JSONString(report).Find("status").(bool) {
		return "", errors.New(report)
	}
	return report, nil
}

// 获取直播信息
func PullLiveInfoAction(cache *xuexitong.XueXiTUserCache, p *xuexitong.PointLiveDto) error {
	liveData, err1 := cache.PullLiveInfoApi(p, 3, nil)

	// 触发500
	if err1 != nil && strings.Contains(err1.Error(), "status code: 500") {
		xuexitong2.ReLogin(cache) //重登
		liveData, err1 = cache.PullLiveInfoApi(p, 3, nil)
	}

	if err1 != nil {
		return err1
	}

	//如果获取成功
	if gojsonq.New().JSONString(liveData).Find("status").(bool) {
		//获取观看进度
		percentValue := gojsonq.New().JSONString(liveData).Find("temp.data.percentValue")
		if percentValue != nil {
			p.VideoCompletePercent = percentValue.(float64)
		}
		videoDuration := gojsonq.New().JSONString(liveData).Find("temp.data.duration")
		if videoDuration != nil {
			p.VideoDuration = int(videoDuration.(float64))
		}
		liveStatusCode := gojsonq.New().JSONString(liveData).Find("temp.data.liveStatus")
		if liveStatusCode != nil {
			p.LiveStatusCode = int(liveStatusCode.(float64))
		}
	}
	return nil
}

// 测试用的Live直播学时函数
func ExecuteLiveTest(cache *xuexitong.XueXiTUserCache, p *xuexitong.PointLiveDto) {
	PullLiveInfoAction(cache, p)
	var passValue float64 = 90
	if p.LiveStatusCode == 0 {
		fmt.Println(p.Title, "该直播还未开始，已自动跳过：")
		return
	}
	if p.VideoCompletePercent >= passValue {
		return //防止学了的继续学
	}
	relation, err := LiveCreateRelationAction(cache, p)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(relation)
	for {
		live, err1 := ExecuteLive(cache, p)
		PullLiveInfoAction(cache, p) //实时更新直播结构体信息
		if err1 != nil {
			fmt.Println(err1)
		}
		fmt.Println(p.Title, "观看状态："+live, "当前观看进度：", fmt.Sprintf("%.2f", p.VideoCompletePercent), "%")
		if p.VideoCompletePercent >= passValue {
			fmt.Println(p.Title, "已完成直播任务点")
			break
		}
		time.Sleep(30 * time.Second)
	}
}
