package point

import (
	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/api/entity"
	"github.com/yatori-dev/yatori-go-core/api/xuexitong"
	"log"
)

func ExecuteDocument(cache *xuexitong.XueXiTUserCache, p *entity.PointDocumentDto) {
	report, err := cache.DocumentDtoReadingReport(p)
	if err != nil {
		log.Fatalln(err)
	}
	flag := gojsonq.New().JSONString(report).Find("status").(bool)
	if flag {
		log.Printf("(%s)提交成功\n", p.Title)
	}
}
