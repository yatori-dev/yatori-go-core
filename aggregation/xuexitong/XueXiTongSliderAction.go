package xuexitong

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"math"
	"regexp"

	"github.com/thedevsaddam/gojsonq"
	xuexitongApi "github.com/yatori-dev/yatori-go-core/api/xuexitong"
)

type XueXiTSlider struct {
	CaptchaId  string `json:"captchaId"`
	Referer    string `json:"referer"`
	serverTime string
}

// 过滑块验证码
func (slider *XueXiTSlider) Pass(cache *xuexitongApi.XueXiTUserCache) (string, error) {
	//第一步:拉取关键信息-----------------------------------------
	captchaInfo, err := cache.XueXiTSliderVerificationCodeApi(slider.CaptchaId, 3, nil)
	if err != nil {
		return "", err
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
		return "", err
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

	//fmt.Println("Token:", resp.Token)
	//fmt.Println("Type:", resp.ImageVerificationVo.Type)
	//fmt.Println("ShadeImage:", resp.ImageVerificationVo.ShadeImage)
	//fmt.Println("CutoutImage:", resp.ImageVerificationVo.CutoutImage)
	//fmt.Println("captchaImgResult:", captchaImgResult)
	//第三步:过验证码-----------------------------------------------------
	//拉取背景图
	shapeImg, err := cache.PullSliderImgApi(resp.ImageVerificationVo.ShadeImage)
	if err != nil {
		return "", err
	}
	//拉取裁剪图
	cutoutImg, err := cache.PullSliderImgApi(resp.ImageVerificationVo.CutoutImage)
	if err != nil {
		return "", err
	}
	//识别
	x := DetectSlideOffset(shapeImg, cutoutImg)
	if err != nil {
		return "", err
	}
	//fmt.Println("x:", x)
	//runEnv参数中，web=10,android=20,ios=30,miniprogram=40
	passResult, err := cache.PassSliderApi(slider.CaptchaId, resp.Token, fmt.Sprintf("%d", x), "10", 3, nil)
	if err != nil {
		return "", err
	}
	//fmt.Println(passResult)
	// ⭐ 正则提取括号中的 JSON 部分
	passResultRe := regexp.MustCompile(`cx_captcha_function\((\{.*\})\)`)
	submatchPass := passResultRe.FindStringSubmatch(passResult)
	if len(submatchPass) < 2 {
		panic("JSON not found")
	}

	passJsonStr := submatchPass[1]
	if passStatus, ok := gojsonq.New().JSONString(passJsonStr).Find("result").(bool); ok {
		if passStatus == true {
			// 第一层：解析整个 JSON
			extra := gojsonq.New().JSONString(passJsonStr).Find("extraData")
			if extra == nil {
				panic("extraData 为空")
			}
			// extra 是一个 string，需要再解析一遍
			extraJson := extra.(string)
			// 第二层：从 extraJson 里取 validate
			validate := gojsonq.New().JSONString(extraJson).Find("validate")
			if validate == nil {
				panic("validate 为空")
			}
			return validate.(string), nil
		}
	}
	return "", errors.New(passResult)
}

// 灰度化
func toGray(img image.Image) [][]float64 {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	gray := make([][]float64, h)
	for y := 0; y < h; y++ {
		gray[y] = make([]float64, w)
		for x := 0; x < w; x++ {
			r, g, b, _ := img.At(x+bounds.Min.X, y+bounds.Min.Y).RGBA()
			gray[y][x] = 0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(b>>8)
		}
	}
	return gray
}

// 验证码模板匹配，归一化互相关(NCC)
func normCrossCorrelation(src, tpl [][]float64) (bestX int, bestScore float64) {

	h1 := len(src)
	w1 := len(src[0])
	h2 := len(tpl)
	w2 := len(tpl[0])

	bestScore = -2

	// 遍历所有可能的 x 偏移
	for y := 0; y <= h1-h2; y++ {
		for x := 0; x <= w1-w2; x++ {

			var sumSrc, sumTpl float64
			var sumSrc2, sumTpl2 float64
			var sumMul float64
			num := float64(w2 * h2)

			for j := 0; j < h2; j++ {
				for i := 0; i < w2; i++ {
					a := src[y+j][x+i]
					b := tpl[j][i]
					sumSrc += a
					sumTpl += b
					sumSrc2 += a * a
					sumTpl2 += b * b
					sumMul += a * b
				}
			}

			meanA := sumSrc / num
			meanB := sumTpl / num

			// NCC 计算
			var numerator float64
			var denom float64

			numerator = sumMul - num*meanA*meanB
			denom = math.Sqrt((sumSrc2-num*meanA*meanA)*(sumTpl2-num*meanB*meanB) + 1e-9)

			score := numerator / denom

			// 找最大值
			if score > bestScore {
				bestScore = score
				bestX = x
			}
		}
	}
	return bestX, bestScore
}

// 计算滑块偏移量
func DetectSlideOffset(bgImg, cutImg image.Image) int {

	//全图转灰度矩阵
	bg := toGray(bgImg)
	cut := toGray(cutImg)

	offsetX, _ := normCrossCorrelation(bg, cut)

	return offsetX - 5
}
