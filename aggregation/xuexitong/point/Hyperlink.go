package point

import (
	"strings"

	"github.com/thedevsaddam/gojsonq"
	xuexitong2 "github.com/yatori-dev/yatori-go-core/aggregation/xuexitong"
	"github.com/yatori-dev/yatori-go-core/api/entity"
	"github.com/yatori-dev/yatori-go-core/api/xuexitong"
	"github.com/yatori-dev/yatori-go-core/models/ctype"
	log2 "github.com/yatori-dev/yatori-go-core/utils/log"
)

// 外链任务点学习
func ExecuteHyperlink(cache *xuexitong.XueXiTUserCache, p *entity.PointHyperlinkDto) (string, error) {
	report, err := cache.HyperlinkDtoCompleteReport(p, 3, nil)
	if err != nil {
		// 触发500
		if err != nil && strings.Contains(err.Error(), "status code: 500") {
			xuexitong2.ReLogin(cache) //重登
			if p.Type == ctype.InsertBook {
				report, err = cache.HyperlinkDtoCompleteReport(p, 3, nil)
			} else {
				report, err = cache.HyperlinkDtoCompleteReport(p, 3, nil)
			}
		}
	}
	if err != nil {
		return "", err
	}
	flag := gojsonq.New().JSONString(report).Find("status").(bool)

	if flag {
		log2.Print(log2.DEBUG, "(", p.Title, ")提交成功")
	} else {
		log2.Print(log2.DEBUG, "(", p.Title, ")外链任务点无法正常学习：返回：(", gojsonq.New().JSONString(report).Find("msg"), ")")
	}
	return report, nil
}
