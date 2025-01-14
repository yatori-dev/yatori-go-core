package blockPuzzle

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
)

// SliderCaptcha 用于封装滑块验证码的处理逻辑
type SliderCaptcha struct {
	Slider     image.Image // 滑块图片
	Background image.Image // 背景图片
}

// NewSliderCaptcha 初始化 SliderCaptcha
func NewSliderCaptcha(sliderBase64, backgroundBase64 string) (*SliderCaptcha, error) {
	slider, err := decodeBase64Image(sliderBase64)
	if err != nil {
		return nil, err
	}
	background, err := decodeBase64Image(backgroundBase64)
	if err != nil {
		return nil, err
	}
	return &SliderCaptcha{Slider: slider, Background: background}, nil
}

// decodeBase64Image 将 Base64 字符串解码为图像
func decodeBase64Image(data string) (image.Image, error) {
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(bytes.NewReader(decoded))
	return img, err
}

func (sc *SliderCaptcha) FindBestMatch() (int, int) {
	bestX, bestY := 0, 0
	bestScore := 0.0
	for y := 0; y <= sc.Background.Bounds().Dy()-sc.Slider.Bounds().Dy(); y++ {
		for x := 0; x <= sc.Background.Bounds().Dx()-sc.Slider.Bounds().Dx(); x++ {
			score := sc.computeMatch(x, y)
			if score > bestScore {
				bestScore = score
				bestX, bestY = x, y
			}
		}
	}
	//global.Log.Info(fmt.Sprintf("Best match found at: (%d, %d) with score: %.4f", bestX, bestY, bestScore))
	return bestX, bestY
}

// computeMatch 计算滑块在背景指定位置的匹配度
func (sc *SliderCaptcha) computeMatch(offsetX, offsetY int) float64 {
	var total, match float64
	for y := 0; y < sc.Slider.Bounds().Dy(); y++ {
		for x := 0; x < sc.Slider.Bounds().Dx(); x++ {
			bgX, bgY := x+offsetX, y+offsetY
			if bgX >= sc.Background.Bounds().Dx() || bgY >= sc.Background.Bounds().Dy() {
				continue
			}

			sliderPixel := color.RGBAModel.Convert(sc.Slider.At(x, y)).(color.RGBA)
			bgPixel := color.RGBAModel.Convert(sc.Background.At(bgX, bgY)).(color.RGBA)

			// 跳过透明像素
			if sliderPixel.A > 0 {
				total++
				if math.Abs(float64(sliderPixel.R)-float64(bgPixel.R)) < 30 &&
					math.Abs(float64(sliderPixel.G)-float64(bgPixel.G)) < 30 &&
					math.Abs(float64(sliderPixel.B)-float64(bgPixel.B)) < 30 {
					match++
				}
			}
		}
	}
	if total == 0 {
		return 0
	}
	return match / total
}

// SaveImage 保存图像到文件，便于调试
func SaveImage(img image.Image, filename string) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer file.Close()
	if err := png.Encode(file, img); err != nil {
		log.Fatalf("Failed to encode image: %v", err)
	}
}
