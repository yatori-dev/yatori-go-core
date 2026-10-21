package xuexitong

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/thedevsaddam/gojsonq"
	xuexitongApi "github.com/yatori-dev/yatori-go-core/internal/api/xuexitong"
	"github.com/yatori-dev/yatori-go-core/utils"
)

type XueXiTSlider struct {
	CaptchaId  string `json:"captchaId"`
	Referer    string `json:"referer"`
	serverTime string
}

// 过滑块验证码
func (slider *XueXiTSlider) Pass(cache *xuexitongApi.XueXiTUserCache) (string, error) {
	return slider.PassWithOffset(cache, 0)
}

// PassWithOffset 过滑块验证码(带偏移修正)
// offsetDelta: 重试时对识别出的x坐标追加的修正量
func (slider *XueXiTSlider) PassWithOffset(cache *xuexitongApi.XueXiTUserCache, offsetDelta int) (string, error) {
	//第一步:拉取关键信息-----------------------------------------
	captchaInfo, err := cache.XueXiTSliderVerificationCodeApi(slider.CaptchaId, 3, nil)
	if err != nil {
		return "", err
	}
	re := regexp.MustCompile(`"t":(\d+).*?"captchaId":"([^"]+)"`)
	matches := re.FindStringSubmatch(captchaInfo)
	if len(matches) == 3 {
		tValue := matches[1]
		//captchaId := matches[2]
		//fmt.Println("t:", tValue)
		slider.serverTime = tValue
		//fmt.Println("captchaId:", captchaId)
	}
	// 拟人化间隔: 拉取配置后短暂停顿再请求图片
	utils.HumanDelay(150*time.Millisecond, 450*time.Millisecond)
	// 动态iv: 官方前端JS iv=md5(captchaId+"slide"+当前毫秒+uuid)，
	// image接口上报后服务器记录，check时必须传同一iv，否则 verification error
	iv := xuexitongApi.NewSliderCaptchaIv(slider.CaptchaId)
	//第二步:拉取相关验证码图片等----------------------------------------
	captchaImgResult, err := cache.XueXiTSliderVerificationImgApi(slider.CaptchaId, slider.serverTime, slider.Referer, iv, 3, nil)
	if err != nil {
		return "", err
	}
	// ⭐ 正则提取括号中的 JSON 部分
	imgRe := regexp.MustCompile(`cx_captcha_function\((\{.*\})\)`)
	m := imgRe.FindStringSubmatch(captchaImgResult)
	if len(m) < 2 {
		return "", fmt.Errorf("滑块图片响应缺少 JSON: %s", responseSummary(captchaImgResult))
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
		return "", fmt.Errorf("解析滑块图片响应失败: %w", err1)
	}
	// 拟人化间隔: 拿到图片地址后稍作停顿再下载
	utils.HumanDelay(150*time.Millisecond, 400*time.Millisecond)

	//fmt.Println("Token:", resp.Token)
	//fmt.Println("Type:", resp.ImageVerificationVo.Type)
	//fmt.Println("ShadeImage:", resp.ImageVerificationVo.ShadeImage)
	//fmt.Println("CutoutImage:", resp.ImageVerificationVo.CutoutImage)
	//fmt.Println("captchaImgResult:", captchaImgResult)
	//第三步:过验证码-----------------------------------------------------
	//拉取背景图
	shapeImg, err := cache.PullSliderImgApi(resp.ImageVerificationVo.ShadeImage, 5, nil)
	if err != nil {
		return "", err
	}
	// 拟人化间隔: 两张图片下载之间略作停顿
	utils.HumanDelay(120*time.Millisecond, 350*time.Millisecond)
	//拉取裁剪图
	cutoutImg, err := cache.PullSliderImgApi(resp.ImageVerificationVo.CutoutImage, 5, nil)
	if err != nil {
		return "", err
	}
	//识别(轮廓+模板匹配优先, 失败回退掩码NCC)
	x := DetectSlideOffset(shapeImg, cutoutImg) + offsetDelta
	//fmt.Println("x:", x)
	// 拟人延时: 真人滑动从拉图到提交至少间隔数百毫秒，立即提交易触发风控
	utils.HumanDelay(800*time.Millisecond, 1800*time.Millisecond)
	//runEnv参数中，web=10,android=20,ios=30,miniprogram=40
	passResult, err := cache.PassSliderApi(slider.CaptchaId, resp.Token, fmt.Sprintf("%d", x), "10", slider.Referer, iv, 3, nil)
	if err != nil {
		return "", err
	}
	//fmt.Println(passResult)
	// ⭐ 正则提取括号中的 JSON 部分
	passResultRe := regexp.MustCompile(`cx_captcha_function\((\{.*\})\)`)
	submatchPass := passResultRe.FindStringSubmatch(passResult)
	if len(submatchPass) < 2 {
		return "", fmt.Errorf("滑块验证响应缺少 JSON: %s", responseSummary(passResult))
	}

	passJsonStr := submatchPass[1]
	if passStatus, ok := gojsonq.New().JSONString(passJsonStr).Find("result").(bool); ok {
		if passStatus == true {
			// 第一层：解析整个 JSON
			extra := gojsonq.New().JSONString(passJsonStr).Find("extraData")
			if extra == nil {
				return "", errors.New("滑块验证响应缺少 extraData")
			}
			// extra 是一个 string，需要再解析一遍
			extraJson, ok := extra.(string)
			if !ok {
				return "", fmt.Errorf("滑块验证 extraData 类型异常: %T", extra)
			}
			// 第二层：从 extraJson 里取 validate
			validate := gojsonq.New().JSONString(extraJson).Find("validate")
			if validate == nil {
				return "", errors.New("滑块验证响应缺少 validate")
			}
			validateString, ok := validate.(string)
			if !ok || validateString == "" {
				return "", fmt.Errorf("滑块验证 validate 类型异常: %T", validate)
			}
			return validateString, nil
		}
	}
	return "", errors.New(passResult)
}

func responseSummary(value string) string {
	const limit = 256
	summary := strings.Join(strings.Fields(value), " ")
	if len(summary) > limit {
		return summary[:limit] + "..."
	}
	return summary
}

// 灰度化
func toGray(img image.Image) [][]float64 {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	gray := make([][]float64, h)
	for y := 0; y < h; y++ {
		gray[y] = make([]float64, w)
		for x := 0; x < w; x++ {
			// 使用非预乘RGB，避免透明像素被预乘为黑
			c := color.NRGBAModel.Convert(img.At(bounds.Min.X+x, bounds.Min.Y+y)).(color.NRGBA)
			gray[y][x] = 0.299*float64(c.R) + 0.587*float64(c.G) + 0.114*float64(c.B)
		}
	}
	return gray
}

// 灰度化并提取alpha通道(无alpha通道时alpha全255)
func toGrayAlpha(img image.Image) (gray [][]float64, alpha [][]uint8) {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	gray = make([][]float64, h)
	alpha = make([][]uint8, h)
	for y := 0; y < h; y++ {
		gray[y] = make([]float64, w)
		alpha[y] = make([]uint8, w)
		for x := 0; x < w; x++ {
			c := color.NRGBAModel.Convert(img.At(bounds.Min.X+x, bounds.Min.Y+y)).(color.NRGBA)
			gray[y][x] = 0.299*float64(c.R) + 0.587*float64(c.G) + 0.114*float64(c.B)
			alpha[y][x] = c.A
		}
	}
	return gray, alpha
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

// cropMatrix 按numpy切片语义裁剪灰度矩阵[y1:y2, x1:x2]，越界自动收敛；空区域返回nil
func cropMatrix(matrix [][]float64, x1, y1, x2, y2 int) [][]float64 {
	if len(matrix) == 0 || len(matrix[0]) == 0 {
		return nil
	}
	h, w := len(matrix), len(matrix[0])
	if x1 < 0 {
		x1 = 0
	}
	if y1 < 0 {
		y1 = 0
	}
	if x2 > w {
		x2 = w
	}
	if y2 > h {
		y2 = h
	}
	if x1 >= x2 || y1 >= y2 {
		return nil
	}
	out := make([][]float64, 0, y2-y1)
	for y := y1; y < y2; y++ {
		row := make([]float64, x2-x1)
		copy(row, matrix[y][x1:x2])
		out = append(out, row)
	}
	return out
}

// foregroundTop 返回非黑(灰度>0)像素区域最上方的y坐标(取轮廓外接矩形顶部)
func foregroundTop(gray [][]float64) (int, bool) {
	for y := range gray {
		for _, v := range gray[y] {
			if v > 0 {
				return y, true
			}
		}
	}
	return 0, false
}

// 裁剪图模板裁剪的横向起点与宽度(两条识别路径共用同一坐标约定)
const (
	sliderCutoutCropLeft  = 8
	sliderCutoutCropWidth = 40
)

// detectSlideOffsetCV 轮廓定位+模板匹配偏移识别
// 流程: cutout灰度→非黑区域顶部(cutout_y)→裁上边缘作模板([cutout_y+2,cutout_y+44) x [8,48))→
// 在shade同y行带([cutout_y-2,cutout_y+50))内做归一化互相关匹配→max_loc[0]-5
func detectSlideOffsetCV(bgImg, cutImg image.Image) (int, bool) {
	cutGray := toGray(cutImg)
	top, ok := foregroundTop(cutGray)
	if !ok {
		return 0, false
	}
	tpl := cropMatrix(cutGray, sliderCutoutCropLeft, top+2, sliderCutoutCropLeft+sliderCutoutCropWidth, top+44)
	if len(tpl) == 0 || len(tpl[0]) == 0 {
		return 0, false
	}
	bgGray := toGray(bgImg)
	if len(bgGray) == 0 || len(bgGray[0]) == 0 {
		return 0, false
	}
	band := cropMatrix(bgGray, 0, top-2, len(bgGray[0]), top+50)
	if len(band) == 0 || len(band[0]) == 0 {
		return 0, false
	}
	if len(band) < len(tpl) || len(band[0]) < len(tpl[0]) {
		return 0, false
	}
	bestX, _ := normCrossCorrelation(band, tpl)
	return bestX - 5, true
}

// buildSliderMask 构建裁剪图有效像素掩码:
// 有真实alpha通道(PNG)时用alpha>=128；无alpha(如JPEG)时以非黑像素作掩码
func buildSliderMask(gray [][]float64, alpha [][]uint8) [][]bool {
	hasAlpha := false
	for y := range alpha {
		for _, a := range alpha[y] {
			if a < 255 {
				hasAlpha = true
				break
			}
		}
		if hasAlpha {
			break
		}
	}
	mask := make([][]bool, len(gray))
	for y := range gray {
		mask[y] = make([]bool, len(gray[y]))
		for x := range gray[y] {
			if hasAlpha {
				mask[y][x] = alpha[y][x] >= 128
			} else {
				mask[y][x] = gray[y][x] > 0
			}
		}
	}
	return mask
}

// nccMatchMasked 带掩码的归一化互相关匹配，仅模板中不透明像素参与计算
// 关键: 裁剪图透明/纯黑区域必须从NCC中剔除 —— 若参与计算，深色背景
// (如暗色海水)会与透明黑块高相关，产生错误峰值，导致服务器报 verification error
// fixedY >= 0 时仅在该y位置搜索(同高裁剪图只需横向滑动)
func nccMatchMasked(src, tpl [][]float64, mask [][]bool, fixedY int) (bestX int, bestScore float64) {
	h1, w1 := len(src), len(src[0])
	h2, w2 := len(tpl), len(tpl[0])

	num := 0
	tplSum := 0.0
	tplSum2 := 0.0
	for j := 0; j < h2; j++ {
		for i := 0; i < w2; i++ {
			if mask[j][i] {
				num++
				v := tpl[j][i]
				tplSum += v
				tplSum2 += v * v
			}
		}
	}
	if num == 0 {
		return 0, -2
	}
	meanB := tplSum / float64(num)
	tplVarTerm := tplSum2 - float64(num)*meanB*meanB

	bestScore = -2.0
	yStart, yEnd := 0, h1-h2
	if fixedY >= 0 {
		if fixedY > h1-h2 {
			fixedY = h1 - h2
		}
		yStart, yEnd = fixedY, fixedY
	}
	for y := yStart; y <= yEnd; y++ {
		for x := 0; x <= w1-w2; x++ {
			sumSrc := 0.0
			sumSrc2 := 0.0
			sumMul := 0.0
			for j := 0; j < h2; j++ {
				srow := src[y+j]
				trow := tpl[j]
				mrow := mask[j]
				for i := 0; i < w2; i++ {
					if mrow[i] {
						a := srow[x+i]
						b := trow[i]
						sumSrc += a
						sumSrc2 += a * a
						sumMul += a * b
					}
				}
			}
			meanA := sumSrc / float64(num)
			numerator := sumMul - float64(num)*meanA*meanB
			denomV := (sumSrc2 - float64(num)*meanA*meanA) * tplVarTerm
			if denomV <= 0 {
				continue
			}
			score := numerator / math.Sqrt(denomV+1e-9)
			if score > bestScore {
				bestScore = score
				bestX = x
			}
		}
	}
	return bestX, bestScore
}

// detectSlideOffsetNCC 掩码NCC(全分辨率)，偏移换算回提交坐标
// 结果统一到主识别路径的坐标约定: 提交x = shade中与裁剪图模板列对齐的位置 - 5
func detectSlideOffsetNCC(bgImg, cutImg image.Image) int {
	bg, _ := toGrayAlpha(bgImg)
	cut, cutAlpha := toGrayAlpha(cutImg)
	mask := buildSliderMask(cut, cutAlpha)
	// 裁剪图与底图同高时拼图块y坐标一致，只需横向滑动匹配
	fixedY := -1
	if len(bg) == len(cut) && len(cut) > 0 {
		fixedY = 0
	}
	bestX, _ := nccMatchMasked(bg, cut, mask, fixedY)
	// bestX为裁剪图第0列在shade中的对齐位置，加模板列起点后与主路径坐标约定一致
	return bestX + sliderCutoutCropLeft - 5
}

// 计算滑块偏移量: 轮廓+模板匹配优先，失败回退掩码NCC
func DetectSlideOffset(bgImg, cutImg image.Image) int {
	if x, ok := detectSlideOffsetCV(bgImg, cutImg); ok {
		return x
	}
	return detectSlideOffsetNCC(bgImg, cutImg)
}
