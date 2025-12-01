package xuexitong

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
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
	x := DetectSlideOffset(shapeImg, cutoutImg)
	if err != nil {
		return err
	}
	fmt.Println("x:", x)
	//runEnv参数中，web=10,android=20,ios=30,miniprogram=40
	passResult, err := cache.PassSliderApi(slider.CaptchaId, resp.Token, fmt.Sprintf("%d", x), "10", 3, nil)
	if err != nil {
		return err
	}
	fmt.Println(passResult)
	return nil
}

// image.Image → gocv.Mat
func ImageToMat(img image.Image) (gocv.Mat, error) {
	var buf bytes.Buffer
	err := png.Encode(&buf, img) // 统一编码为 PNG
	if err != nil {
		return gocv.NewMat(), err
	}
	mat, err := gocv.IMDecode(buf.Bytes(), gocv.IMReadColor)
	return mat, err
}

func DetectSlideOffset(bgImg, cutImg image.Image) int {
	// --- Convert to Mat ---
	shadeMat, err := ImageToMat(bgImg)
	if err != nil {
		return 0
	}
	defer shadeMat.Close()

	cutoutMat, err := ImageToMat(cutImg)
	if err != nil {
		return 0
	}
	defer cutoutMat.Close()

	// --- Convert cutout to gray ---
	cutoutGray := gocv.NewMat()
	defer cutoutGray.Close()
	gocv.CvtColor(cutoutMat, &cutoutGray, gocv.ColorBGRToGray)

	// --- Find contours ---
	contours := gocv.FindContours(cutoutGray, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	defer contours.Close()

	if contours.Size() == 0 {
		return 0
	}

	// 获取第一个 contour 的 boundingRect
	firstContour := contours.At(0)
	rect := gocv.BoundingRect(firstContour)
	cutoutY := rect.Min.Y

	// --- Crop cutout part: cutout_image[cutout_y+2 : cutout_y+44, 8:48] ---
	cutoutTeil := cutoutMat.Region(image.Rect(
		8, cutoutY+2,
		48, cutoutY+44,
	))
	defer cutoutTeil.Close()

	// --- Crop shade part: shade_image[cutout_y-2 : cutout_y+50] ---
	shadeTeil := shadeMat.Region(image.Rect(
		0, cutoutY-2,
		shadeMat.Cols(), cutoutY+50,
	))
	defer shadeTeil.Close()

	// ---------- 模板匹配 ----------
	result := gocv.NewMat()
	defer result.Close()

	mask := gocv.NewMat() // 空 mask
	defer mask.Close()

	err = gocv.MatchTemplate(
		shadeTeil,           // image
		cutoutTeil,          // templ
		&result,             // result
		gocv.TmCcoeffNormed, // method
		mask,                // mask
	)
	if err != nil {
		return 0
	}

	_, _, _, maxLoc := gocv.MinMaxLoc(result)

	return maxLoc.X - 5
}
