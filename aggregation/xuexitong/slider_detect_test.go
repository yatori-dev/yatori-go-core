package xuexitong

import (
	"image"
	"image/color"
	"math/rand"
	"testing"
)

// 合成一张滑块验证码底图与裁剪图:
// shade: 300x150 随机噪点底图，拼图块位于(pieceX,pieceY)大小40x42；
// cutout: 高度与底图一致、宽56的透明图，拼图块粘贴在(8,pieceY)处(与线上图结构一致)
func buildSyntheticSliderPair(seed int64, pieceX, pieceY int) (image.Image, image.Image) {
	const (
		width   = 300
		height  = 150
		pieceW  = 40
		pieceH  = 42
		cutoutW = 56
		cutoutX = 8
	)
	rng := rand.New(rand.NewSource(seed))
	shade := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			shade.SetNRGBA(x, y, color.NRGBA{
				R: uint8(rng.Intn(256)),
				G: uint8(rng.Intn(256)),
				B: uint8(rng.Intn(256)),
				A: 255,
			})
		}
	}
	cutout := image.NewNRGBA(image.Rect(0, 0, cutoutW, height))
	for y := 0; y < pieceH; y++ {
		for x := 0; x < pieceW; x++ {
			src := shade.NRGBAAt(pieceX+x, pieceY+y)
			cutout.SetNRGBA(cutoutX+x, pieceY+y, src)
		}
	}
	return shade, cutout
}

func TestDetectSlideOffsetSynthetic(t *testing.T) {
	cases := []struct {
		pieceX, pieceY int
	}{
		{100, 40},
		{180, 55},
		{30, 20},
	}
	for index, c := range cases {
		shade, cutout := buildSyntheticSliderPair(int64(index+1), c.pieceX, c.pieceY)
		want := c.pieceX - 5

		gotCV, ok := detectSlideOffsetCV(shade, cutout)
		if !ok || gotCV != want {
			t.Errorf("case%d: 轮廓模板匹配 got=%d want=%d ok=%v", index, gotCV, want, ok)
		}
		gotNCC := detectSlideOffsetNCC(shade, cutout)
		if gotNCC != want {
			t.Errorf("case%d: 掩码NCC降级 got=%d want=%d", index, gotNCC, want)
		}
		got := DetectSlideOffset(shade, cutout)
		if got != want {
			t.Errorf("case%d: DetectSlideOffset got=%d want=%d", index, got, want)
		}
	}
}
