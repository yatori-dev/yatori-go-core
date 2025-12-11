package point

import (
	"strings"

	"github.com/thedevsaddam/gojsonq"
	xuexitong2 "github.com/yatori-dev/yatori-go-core/aggregation/xuexitong"
	"github.com/yatori-dev/yatori-go-core/api/xuexitong"
	"github.com/yatori-dev/yatori-go-core/models/ctype"
	log2 "github.com/yatori-dev/yatori-go-core/utils/log"
)

func ExecuteDocument(cache *xuexitong.XueXiTUserCache, p *xuexitong.PointDocumentDto) (string, error) {
	var report string
	var err error
	if p.Type == ctype.InsertBook {
		report, err = cache.DocumentDtoReadingBookReport(p, 3, nil)
	} else if p.Type == ctype.InsertReadV2 {
		report, err = cache.ReadV2PointPeReport(p, 3, nil)
	} else {
		report, err = cache.DocumentDtoReadingReport(p, 3, nil)
	}
	// 触发500
	if err != nil && strings.Contains(err.Error(), "status code: 500") {
		xuexitong2.ReLogin(cache) //重登
		if p.Type == ctype.InsertBook {
			report, err = cache.DocumentDtoReadingBookReport(p, 3, nil)
		} else {
			report, err = cache.DocumentDtoReadingReport(p, 3, nil)
		}
	}

	if err != nil {
		return "", err
	}
	flag := gojsonq.New().JSONString(report).Find("status").(bool)

	if flag {
		log2.Print(log2.DEBUG, "(", p.Title, ")提交成功")
	} else {
		log2.Print(log2.DEBUG, "(", p.Title, ")文档无法正常学习：返回：(", gojsonq.New().JSONString(report).Find("msg"), ")")
	}
	return report, nil
}
