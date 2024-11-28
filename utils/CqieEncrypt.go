package utils

import (
	"github.com/tjfoc/gmsm/sm4"
	"log"
)

// fo 数组，假设它是一个密钥或初始化向量（IV）
var fo = []byte{8, 5, 4, 1, 2, 7, 6, 8, 9, 4, 5, 1, 2, 4, 1, 4}

// 加密函数：如果输入的 o 有效，则使用 SM4 加密
func CqieEncrypt(str string) []byte {
	o := []byte(str)
	if o == nil || len(o) == 0 {
		// 如果输入为空或无效，返回空切片
		return nil
	}

	// 使用 SM4 算法加密
	encrypted, err := sm4Encrypt(o, fo)
	if err != nil {
		log.Fatal("Encryption failed:", err)
	}

	return encrypted
}

// SM4 加密函数
func sm4Encrypt(data, key []byte) ([]byte, error) {
	// 创建一个 SM4 密码实例
	block, err := sm4.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 加密数据，SM4 是块加密算法，要求数据是16字节的倍数
	// 如果数据不足 16 字节，我们需要填充
	// 这里的填充方式与 JavaScript 中类似，可以使用 PKCS#7 填充
	data = pkcs7Padding(data, sm4.BlockSize)

	// 加密数据
	encrypted := make([]byte, len(data))
	block.Encrypt(encrypted, data)

	return encrypted, nil
}

// PKCS#7 填充函数，确保数据长度是 16 的倍数
func pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := make([]byte, padding)
	for i := 0; i < padding; i++ {
		padText[i] = byte(padding)
	}
	return append(data, padText...)
}
