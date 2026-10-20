package utils

import (
	"math/rand"
	"time"
)

// HumanDelay 拟人化随机等待 [min, max](含两端)。
// 用于在连续接口调用之间插入不固定的间隔，避免固定/过快的请求节奏被识别为非人类特征。
func HumanDelay(min, max time.Duration) {
	if max <= min {
		time.Sleep(min)
		return
	}
	time.Sleep(min + time.Duration(rand.Int63n(int64(max-min)+1)))
}
