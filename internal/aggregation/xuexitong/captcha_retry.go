package xuexitong

import (
	"context"
	"errors"
	"fmt"
	"image"
	"time"

	xuexitongApi "github.com/yatori-dev/yatori-go-core/v3/internal/api/xuexitong"
)

const defaultCaptchaAttempts = 5

// imageCaptchaRetryDelay 是图片验证码识别失败后重新拉图的间隔。
// 测试里会临时改成 0。
var imageCaptchaRetryDelay = 2 * time.Second

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
	err := retryImageCaptcha(ctx, "反爬图片验证码", func() (bool, error) {
		img, imgErr := cache.XueXiTProcessVerifyImage(2, nil)
		if imgErr != nil {
			return false, imgErr
		}
		code := solveLoginCaptcha(img)
		if code == "" {
			return false, nil
		}
		return cache.XueXiTPassVerificationCode(code, 2, nil)
	})
	if err == nil {
		return nil
	}

	// 反爬验证码接口不可用时，回退到章节图片验证码：
	// kaptcha-img/code + kaptcha-img/ajaxValidate2。
	return retryImageCaptcha(ctx, "章节图片验证码", func() (bool, error) {
		img, imgErr := cache.XueXiTChapterVerificationCodeApi(2, nil)
		if imgErr != nil {
			return false, imgErr
		}
		code := solveLoginCaptcha(img)
		if code == "" {
			return false, nil
		}
		return cache.XueXiTPassCahpterVerificationCode(code, 2, nil)
	})
}

// retryImageCaptcha 对图片验证码持续“换图 + OCR + 提交”，直到通过或 ctx 结束。
// 识别为空或验证码被拒绝都不转人工，只换一张继续识别；只有连续 3 次连
// 验证码图片都拿不到时，才返回错误让上层切换到其它验证码接口。
func retryImageCaptcha(ctx context.Context, kind string, attempt func() (bool, error)) error {
	var (
		lastErr       error
		networkErrors int
	)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		passed, err := attempt()
		if passed {
			return nil
		}
		if err != nil {
			lastErr = err
			networkErrors++
			if networkErrors >= 3 {
				return fmt.Errorf("%s接口连续失败: %w", kind, lastErr)
			}
		} else {
			networkErrors = 0
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(imageCaptchaRetryDelay):
		}
	}
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

// solveLoginCaptcha 始终使用内置 ddddocr 识别，不再回调调用方转人工。
func solveLoginCaptcha(img image.Image) string {
	return ocrCaptchaImage(img)
}
