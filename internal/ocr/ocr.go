// Package ocr wraps the embedded ddddocr models used for login captchas.
package ocr

import (
	"image"
	"sync"

	ddddocr "github.com/Changbaiqi/ddddocr-go/utils"
)

var once sync.Once

// Ensure loads the bundled ONNX runtime and models at most once.
func Ensure() {
	once.Do(func() { ddddocr.DDDDOcrCoreInit() })
}

// Solve recognizes a login captcha, preferring the newer common.onnx model
// and falling back to the legacy ddddocr recognizer.
func Solve(img image.Image) (code string) {
	Ensure()
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
