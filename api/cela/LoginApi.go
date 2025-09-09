package cela

import (
	"bytes"
	"crypto/aes"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

type CelaUserCache struct {
	Account  string //账号
	Password string //密码
	cookie   string //cookie
	asuss    string //token
	Code     string //验证码
}

// 登录
func (cache *CelaUserCache) LoginApi() {
	// 1. 获取时间戳（假设是秒级时间戳，跟 JS 保持一致）
	//timestamp := fmt.Sprintf("%d\n", time.Now().Unix())
	timestamp := "1757425302934"
	// 2. 拼接 timestamp + code[:3]
	// 注意 Go 要检查 code 长度，避免越界
	var prefix string
	if len(cache.Code) >= 3 {
		prefix = cache.Code[:3]
	} else {
		prefix = cache.Code
	}
	keyStr := []byte(timestamp + prefix)
	fmt.Println(keyStr)
	formData := map[string]string{
		"account":          cache.Account,
		"password":         cache.Password,
		"verificationCode": cache.Code,
	}
	jsonBytes, _ := json.Marshal(formData)
	//utf8Bytes := []byte(string((jsonBytes)))
	fmt.Printf("%s\n", string(jsonBytes))
	encrypt, err := aesEncryptECB(jsonBytes, fixKey(keyStr, 16))
	if err != nil {
		fmt.Println(err)
	}
	//
	fmt.Println(base64.StdEncoding.EncodeToString(encrypt))
}

// PKCS7 填充
func pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

// AES-ECB 加密
func aesEncryptECB(src, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	bs := block.BlockSize()
	src = pkcs7Padding(src, bs)
	encrypted := make([]byte, len(src))
	for start := 0; start < len(src); start += bs {
		block.Encrypt(encrypted[start:start+bs], src[start:start+bs])
	}
	return encrypted, nil
}

func fixKey(key []byte, length int) []byte {
	if len(key) == length {
		return key
	}
	if len(key) > length {
		return key[:length] // 截断
	}
	// 不足则补 0
	newKey := make([]byte, length)
	copy(newKey, key)
	return newKey
}
