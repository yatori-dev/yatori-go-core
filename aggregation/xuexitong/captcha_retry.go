package xuexitong

import (
	"context"
	"errors"
	"fmt"

	ddddocr "github.com/Changbaiqi/ddddocr-go/utils"
	ort "github.com/yalue/onnxruntime_go"
	xuexitongApi "github.com/yatori-dev/yatori-go-core/api/xuexitong"
	"github.com/yatori-dev/yatori-go-core/utils"
)

const defaultCaptchaAttempts = 5

type CaptchaRetryError struct {
	Kind     string
	Attempts int
	LastErr  error
}

func (e *CaptchaRetryError) Error() string {
	if e.LastErr == nil {
		return fmt.Sprintf("%s重试 %d 次后仍未通过", e.Kind, e.Attempts)
	}
	return fmt.Sprintf("%s重试 %d 次后仍未通过: %v", e.Kind, e.Attempts, e.LastErr)
}

func (e *CaptchaRetryError) Unwrap() error {
	return e.LastErr
}

func retryCaptcha(ctx context.Context, kind string, attempts int, attempt func() (bool, error)) error {
	if attempts <= 0 {
		return &CaptchaRetryError{Kind: kind, Attempts: attempts, LastErr: errors.New("未配置重试次数")}
	}
	var lastErr error
	for index := 0; index < attempts; index++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		passed, err := attempt()
		if passed {
			return nil
		}
		if err != nil {
			lastErr = err
		} else {
			lastErr = errors.New("验证码被拒绝")
		}
	}
	return &CaptchaRetryError{Kind: kind, Attempts: attempts, LastErr: lastErr}
}

func passImageCaptcha(ctx context.Context, cache *xuexitongApi.XueXiTUserCache) error {
	return retryCaptcha(ctx, "图片验证码", defaultCaptchaAttempts, func() (bool, error) {
		img, err := cache.XueXiTVerificationCodeApi(7, nil)
		if err != nil {
			return false, err
		}
		_, width, _ := utils.GetImageShape(img)
		shape := ort.NewShape(1, 30)
		if width == 140 {
			shape = ort.NewShape(1, 23)
		}
		codeResult := ddddocr.SemiOCRVerification(img, shape)
		return cache.XueXiTPassVerificationCode(codeResult, 7, nil)
	})
}

// 偏移修正探测表: 第N次重试使用 delta[(N-1)%len]，x_submit = x_detect + delta。
// 超星滑块容差有限，通过多候选探测确定服务器接受的修正量
var sliderOffsetDeltas = []int{0, -2, 2, -5, 5}

func passSliderCaptcha(ctx context.Context, cache *xuexitongApi.XueXiTUserCache, slider *XueXiTSlider) (string, error) {
	var validate string
	attemptIndex := 0
	err := retryCaptcha(ctx, "滑块验证码", defaultCaptchaAttempts, func() (bool, error) {
		delta := sliderOffsetDeltas[attemptIndex%len(sliderOffsetDeltas)]
		attemptIndex++
		value, err := slider.PassWithOffset(cache, delta)
		if err != nil {
			return false, err
		}
		if value == "" {
			return false, errors.New("滑块验证成功响应缺少 validate")
		}
		validate = value
		return true, nil
	})
	return validate, err
}
