package point

import (
	"encoding/json"
	"fmt"
	"strings"

	xuexitong2 "github.com/yatori-dev/yatori-go-core/aggregation/xuexitong"
	"github.com/yatori-dev/yatori-go-core/api/xuexitong"
	"github.com/yatori-dev/yatori-go-core/models/ctype"
	log2 "github.com/yatori-dev/yatori-go-core/utils/log"
)

func ExecuteDocument(cache *xuexitong.XueXiTUserCache, p *xuexitong.PointDocumentDto) (string, error) {
	reportDocument := func() (string, error) {
		return cache.DocumentDtoReadingReport(p, 3, nil)
	}
	if p.Type == ctype.InsertBook {
		reportDocument = func() (string, error) {
			return cache.DocumentDtoReadingBookReport(p, 3, nil)
		}
	} else if p.Type == ctype.InsertReadV2 {
		reportDocument = func() (string, error) {
			return cache.ReadV2PointPeReport(p, 3, nil)
		}
	}
	return executeDocument(p, reportDocument, func() { xuexitong2.ReLogin(cache) })
}

func executeDocument(p *xuexitong.PointDocumentDto, reportDocument func() (string, error), relogin func()) (string, error) {
	report, err := reportDocument()
	// 触发500
	if err != nil && strings.Contains(err.Error(), "status code: 500") {
		relogin()
		report, err = reportDocument()
	}

	if err != nil {
		return "", err
	}
	var response struct {
		Status json.RawMessage `json:"status"`
		Msg    any             `json:"msg"`
	}
	if err := json.Unmarshal([]byte(report), &response); err != nil {
		return "", fmt.Errorf("解析文档提交响应失败: %w", err)
	}
	if len(response.Status) == 0 || string(response.Status) == "null" {
		return "", fmt.Errorf("解析文档提交响应失败: 缺少 status")
	}
	var flag bool
	if err := json.Unmarshal(response.Status, &flag); err != nil {
		return "", fmt.Errorf("解析文档提交响应失败: status 不是布尔值: %w", err)
	}

	if flag {
		log2.Print(log2.DEBUG, "(", p.Title, ")提交成功")
	} else {
		log2.Print(log2.DEBUG, "(", p.Title, ")文档无法正常学习：返回：(", response.Msg, ")")
	}
	return report, nil
}

func ExecuteDocumentStudyTime(cache *xuexitong.XueXiTUserCache, p *xuexitong.PointDocumentDto) {

}
