package xuexitong

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

const maxLoginCodeAttempts = 3

// loginCodeRejectedError 表示平台明确拒绝了验证码/密码，可以重新询问用户。
type loginCodeRejectedError struct {
	message string
}

func (e *loginCodeRejectedError) Error() string { return e.message }

// requestLoginCode 重复调用 LoginCodeProvider，直到 verify 成功或出现
// 不可重试的错误（网络错误、上下文取消等）。
func (cache *XueXiTUserCache) requestLoginCode(
	ctx context.Context,
	build func(attempt int, lastErr string) LoginCodeChallenge,
	verify func(value string) error,
) error {
	if cache.LoginCodeProvider == nil {
		return errors.New("登录需要验证码，请配置 LoginCodeProvider")
	}
	var lastErr error
	for attempt := 1; attempt <= maxLoginCodeAttempts; attempt++ {
		lastMessage := ""
		if lastErr != nil {
			lastMessage = lastErr.Error()
		}
		challenge := build(attempt, lastMessage)
		if challenge.Attempt == 0 {
			challenge.Attempt = attempt
		}
		if challenge.LastError == "" {
			challenge.LastError = lastMessage
		}
		value, err := cache.LoginCodeProvider(ctx, challenge)
		if err != nil {
			return err
		}
		value = strings.TrimSpace(value)
		if value == "" {
			return errors.New("未提供验证码")
		}
		if err := verify(value); err != nil {
			var rejected *loginCodeRejectedError
			if !errors.As(err, &rejected) {
				return err
			}
			lastErr = err
			continue
		}
		return nil
	}
	if lastErr == nil {
		lastErr = errors.New("验证码验证失败")
	}
	return fmt.Errorf("验证码重试 %d 次仍未通过: %w", maxLoginCodeAttempts, lastErr)
}
