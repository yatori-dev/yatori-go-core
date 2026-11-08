package xuexitong

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/png"

	xuexitongApi "github.com/yatori-dev/yatori-go-core/v3/internal/api/xuexitong"
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
	// 先处理反爬验证码：processVerifyPng.ac + processVerify.ac(GET)。
	err := retryCaptcha(ctx, "反爬图片验证码", defaultCaptchaAttempts, func() (bool, error) {
		img, imgErr := cache.XueXiTProcessVerifyImage(2, nil)
		if imgErr != nil {
			return false, imgErr
		}
		code := solveLoginCaptcha(ctx, cache, img)
		if code == "" {
			return false, errors.New("验证码识别为空")
		}
		return cache.XueXiTPassVerificationCode(code, 2, nil)
	})
	if err == nil {
		return nil
	}

	// 反爬验证码接口不可用时，回退到章节图片验证码：
	// kaptcha-img/code + kaptcha-img/ajaxValidate2。
	return retryCaptcha(ctx, "章节图片验证码", defaultCaptchaAttempts, func() (bool, error) {
		img, imgErr := cache.XueXiTChapterVerificationCodeApi(2, nil)
		if imgErr != nil {
			return false, imgErr
		}
		code := solveLoginCaptcha(ctx, cache, img)
		if code == "" {
			return false, errors.New("验证码识别为空")
		}
		return cache.XueXiTPassCahpterVerificationCode(code, 2, nil)
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

func solveLoginCaptcha(ctx context.Context, cache *xuexitongApi.XueXiTUserCache, img image.Image) string {
	if cache.LoginCodeProvider != nil {
		var buffer bytes.Buffer
		if err := png.Encode(&buffer, img); err == nil {
			code, err := cache.LoginCodeProvider(ctx, xuexitongApi.LoginCodeChallenge{
				Kind:   "image",
				Prompt: "图片验证码",
				Image:  buffer.Bytes(),
			})
			if err == nil && code != "" {
				return code
			}
		}
	}
	return ocrCaptchaImage(img)
}
