package xuexitong

import (
	"encoding/json"
	"fmt"
	"image"
	"regexp"

	xuexitongApi "github.com/yatori-dev/yatori-go-core/api/xuexitong"
	"gocv.io/x/gocv"
)

type XueXiTSlider struct {
	CaptchaId  string `json:"captchaId"`
	Referer    string `json:"referer"`
	serverTime string
}

// 过验证码
func (slider *XueXiTSlider) Pass(cache *xuexitongApi.XueXiTUserCache) error {
	//第一步:拉取关键信息-----------------------------------------
	captchaInfo, err := cache.XueXiTSliderVerificationCodeApi(slider.CaptchaId, 3, nil)
	if err != nil {
		return err
	}
	re := regexp.MustCompile(`"t":(\d+).*?"captchaId":"([^"]+)"`)
	matches := re.FindStringSubmatch(captchaInfo)
	if len(matches) == 3 {
		tValue := matches[1]
		captchaId := matches[2]
		fmt.Println("t:", tValue)
		slider.serverTime = tValue
		fmt.Println("captchaId:", captchaId)
	}
	//第二步:拉取相关验证码图片等----------------------------------------
	captchaImgResult, err := cache.XueXiTSliderVerificationImgApi(slider.CaptchaId, slider.serverTime, slider.Referer, 3, nil)
	if err != nil {
		return err
	}
	// ⭐ 正则提取括号中的 JSON 部分
	imgRe := regexp.MustCompile(`cx_captcha_function\((\{.*\})\)`)
	m := imgRe.FindStringSubmatch(captchaImgResult)
	if len(m) < 2 {
		panic("JSON not found")
	}

	jsonStr := m[1]

	// ⭐ 解析 JSON
	type CaptchaResponse struct {
		Token               string `json:"token"`
		ImageVerificationVo struct {
			Type        string `json:"type"`
			ShadeImage  string `json:"shadeImage"`
			CutoutImage string `json:"cutoutImage"`
		} `json:"imageVerificationVo"`
	}
	var resp CaptchaResponse
	err1 := json.Unmarshal([]byte(jsonStr), &resp)
	if err1 != nil {
		panic(err)
	}

	fmt.Println("Token:", resp.Token)
	fmt.Println("Type:", resp.ImageVerificationVo.Type)
	fmt.Println("ShadeImage:", resp.ImageVerificationVo.ShadeImage)
	fmt.Println("CutoutImage:", resp.ImageVerificationVo.CutoutImage)
	fmt.Println("captchaImgResult:", captchaImgResult)
	//第三步:过验证码-----------------------------------------------------
	//拉取背景图
	shapeImg, err := cache.PullSliderImgApi(resp.ImageVerificationVo.ShadeImage)
	if err != nil {
		return err
	}
	//拉取裁剪图
	cutoutImg, err := cache.PullSliderImgApi(resp.ImageVerificationVo.CutoutImage)
	if err != nil {
		return err
	}
	//识别
	offset, f, err := DetectSlideOffset(shapeImg, cutoutImg)
	if err != nil {
		return err
	}
	fmt.Println("offset:", offset)
	fmt.Println("f:", f)
	return nil
}

// image.Image → gocv.Mat
func ImageToMatSafe(img image.Image) gocv.Mat {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	// BGR 3 通道 Mat
	mat := gocv.NewMatWithSize(h, w, gocv.MatTypeCV8UC3)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {

			// Safe 获取像素
			r32, g32, b32, _ := img.At(x+bounds.Min.X, y+bounds.Min.Y).RGBA()

			r := uint8(r32 >> 8)
			g := uint8(g32 >> 8)
			b := uint8(b32 >> 8)

			// ⭐ 使用正确的 SetUCharAt (1D)，避免 SetUCharAt3 崩溃
			mat.SetUCharAt(y, x*3+0, b) // B
			mat.SetUCharAt(y, x*3+1, g) // G
			mat.SetUCharAt(y, x*3+2, r) // R
		}
	}

	return mat
}

func DetectSlideOffset(bgImg, cutImg image.Image) (int, float32, error) {
	bg := ImageToMatSafe(bgImg)
	defer bg.Close()

	cut := ImageToMatSafe(cutImg)
	defer cut.Close()

	if bg.Empty() || cut.Empty() {
		return 0, 0, fmt.Errorf("图片为空")
	}

	result := gocv.NewMatWithSize(
		bg.Rows()-cut.Rows()+1,
		bg.Cols()-cut.Cols()+1,
		gocv.MatTypeCV32F)
	defer result.Close()

	// 你提供的签名：必须 5 个参数
	err := gocv.MatchTemplate(bg, cut, &result, gocv.TmCcoeffNormed, gocv.NewMat())
	if err != nil {
		return 0, 0, err
	}

	_, maxVal, _, maxLoc := gocv.MinMaxLoc(result)
	return maxLoc.X, maxVal, nil
}
