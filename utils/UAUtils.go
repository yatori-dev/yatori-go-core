package utils

import (
	"encoding/hex"
	"math/rand"
)

// 用于生成IMEI
func TokenHex(len int) string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	token := hex.EncodeToString(b)
	return token
}
