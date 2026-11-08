package xuexitong

import (
	"image"

	"github.com/yatori-dev/yatori-go-core/v3/internal/ocr"
)

// ocrCaptchaImage 优先使用新版 common.onnx（CHARSET_BETA）识别，
// 失败时回退到 ddddocr 旧版 Go OCR。
func ocrCaptchaImage(img image.Image) string {
	return ocr.Solve(img)
}
