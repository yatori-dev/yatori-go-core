package xuexitong

import (
	"image"

	ddddocr "github.com/Changbaiqi/ddddocr-go/utils"
)

// ocrCaptchaImage 优先使用新版 common.onnx（CHARSET_BETA）识别，
// 失败时回退到 ddddocr 旧版 Go OCR。
func ocrCaptchaImage(img image.Image) (code string) {
	defer func() {
		if recover() != nil {
			code = ""
		}
	}()
	if result, err := ddddocr.BetaOCRVerification(img); err == nil && result != "" {
		return result
	}
	return ddddocr.AutoOCRVerification(img)
}
