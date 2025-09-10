package point

import (
	"log"

	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/api/entity"
	"github.com/yatori-dev/yatori-go-core/api/xuexitong"
	"github.com/yatori-dev/yatori-go-core/models/ctype"
)

func ExecuteDocument(cache *xuexitong.XueXiTUserCache, p *entity.PointDocumentDto) {
	var report string
	var err error
	if p.Type == ctype.Document || p.Type == ctype.Insertdoc {
		report, err = cache.DocumentDtoReadingReport(p, 3, nil)
	} else {
		report, err = cache.DocumentDtoReadingBookReport(p, 3, nil)
	}

	if err != nil {
		log.Fatalln(err)
	}
	flag := gojsonq.New().JSONString(report).Find("status").(bool)

	if flag {
		log.Printf("(%s)提交成功\n", p.Title)
	} else {
		log.Printf("(%s)文档无法正常学习：返回：(%s)\n", p.Title, gojsonq.New().JSONString(report).Find("msg"))
	}
}
