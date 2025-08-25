package examples

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"log"
	"time"

	"testing"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	goqr "github.com/skip2/go-qrcode"
	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/api/zhihuishu"
)

// 知到扫码登录测试
func Test_ZhidaoQrcode(t *testing.T) {
	qrImgBase64, qtoken, err := zhihuishu.ZhidaoQrCode()
	raw, err := base64.StdEncoding.DecodeString(qrImgBase64)
	// --- 2) 解码图片 ---
	img, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		log.Fatalf("图片解码失败: %v", err)
	}
	// 3) 用 gozxing 识别二维码内容
	content, err := decodeQRContent(img)
	if err != nil {
		log.Fatalf("二维码识别失败: %v", err)
	}
	fmt.Println("二维码内容：", content)

	// 4) 用 go-qrcode 在控制台按模块网格重绘（可扫）
	qr, err := goqr.New(content, goqr.Medium)
	if err != nil {
		log.Fatalf("二维码生成失败: %v", err)
	}

	// ToString(false) -> 终端黑底/白底都常用；如需反色可传 true
	fmt.Println(qr.ToString(false))
	fmt.Println(qtoken)

	for {
		checkjson, err := zhihuishu.ZhidaoQrCheck(qtoken)
		if err != nil {
			log.Fatalf("检测扫码登录失败：%v", err)
		}
		if (int)(gojsonq.New().JSONString(checkjson).Find("status").(float64)) == 0 {
			fmt.Println("已扫码，等待确认登录")
		}
		if (int)(gojsonq.New().JSONString(checkjson).Find("status").(float64)) == 1 {
			fmt.Println("登录成功")
			fmt.Printf("oncePassword值：%s\n", gojsonq.New().JSONString(checkjson).Find("oncePassword"))
			fmt.Printf("uuid值：%s\n", gojsonq.New().JSONString(checkjson).Find("uuid"))
			break
		}
		time.Sleep(5 * time.Second)
	}
}
func decodeQRContent(img image.Image) (string, error) {
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return "", err
	}
	reader := qrcode.NewQRCodeReader()
	// 如果你的图像几乎是纯二维码（无透视/畸变），也可以加提示：hints := map[gozxing.DecodeHintType]interface{}{gozxing.DecodeHintType_PURE_BARCODE: true}
	result, err := reader.Decode(bmp, nil)
	if err != nil {
		return "", err
	}
	return result.String(), nil
}
