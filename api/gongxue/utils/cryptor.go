package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
)

// 常量定义
const (
	SignSuffix = "3478cbbc33f84bd00d75d7dfa69e0daa"
	MoGuKEY    = "23DbtQHR2UMbH6mJ"
)

var CXKEY = []byte("u2oh6Vu^HWe4_AES")

// PKCS7Padding 对数据进行填充，使其长度对齐到块大小的倍数
func PKCS7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	return append(data, bytes.Repeat([]byte{byte(padding)}, padding)...)
}

// PKCS7Unpadding 去除 PKCS7 填充
func PKCS7Unpadding(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("数据长度为 0，无法去除填充")
	}
	padding := int(data[length-1])
	if padding > length || padding == 0 {
		return nil, errors.New("填充长度不合法")
	}
	for _, pad := range data[length-padding:] {
		if pad != byte(padding) {
			return nil, errors.New("填充内容不合法")
		}
	}
	return data[:length-padding], nil
}

// AESCommonEncrypt 使用 AES 加密（通用方法）
func AESCommonEncrypt(plaintext, key []byte, blockSize int, mode cipher.BlockMode) (string, error) {
	// 填充明文
	paddedText := PKCS7Padding(plaintext, blockSize)

	// 加密
	ciphertext := make([]byte, len(paddedText))
	mode.CryptBlocks(ciphertext, paddedText)

	// 返回 Base64 编码的密文
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// AESCommonDecrypt 使用 AES 解密（通用方法）
func AESCommonDecrypt(ciphertext string, key []byte, blockSize int, mode cipher.BlockMode) (string, error) {
	// 解码密文
	decodedCiphertext, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("解码密文失败: %v", err)
	}

	// 解密
	plaintext := make([]byte, len(decodedCiphertext))
	mode.CryptBlocks(plaintext, decodedCiphertext)

	// 去除填充
	unpaddedText, err := PKCS7Unpadding(plaintext)
	if err != nil {
		return "", fmt.Errorf("去除填充失败: %v", err)
	}
	return string(unpaddedText), nil
}

// AESCBCEncrypt 使用 AES CBC 模式加密
func AESCBCEncrypt(plaintext []byte) (string, error) {
	block, err := aes.NewCipher(CXKEY)
	if err != nil {
		return "", fmt.Errorf("创建 AES 加密器失败: %v", err)
	}

	mode := cipher.NewCBCEncrypter(block, CXKEY[:block.BlockSize()])
	return AESCommonEncrypt(plaintext, CXKEY, block.BlockSize(), mode)
}

// CreateSign 生成签名
func CreateSign(args ...string) string {
	// 拼接字符串
	signStr := ""
	for _, arg := range args {
		signStr += arg
	}
	signStr += SignSuffix
	// 生成 MD5 签名
	hash := md5.Sum([]byte(signStr))
	return hex.EncodeToString(hash[:])
}

// AESECBPKCS5Padding 封装的 AES-ECB 加密器
type AESECBPKCS5Padding struct {
	key       []byte
	outFormat string
}

// NewAESECBPKCS5Padding 创建一个 AES-ECB 加密器实例
func NewAESECBPKCS5Padding(key string, outFormat string) (*AESECBPKCS5Padding, error) {
	if len(key) != 16 {
		return nil, errors.New("密钥长度必须为 16 字节")
	}
	if outFormat != "hex" && outFormat != "base64" {
		return nil, errors.New("输出格式必须是 'hex' 或 'base64'")
	}
	return &AESECBPKCS5Padding{
		key:       []byte(key),
		outFormat: outFormat,
	}, nil
}

// Encrypt 实现 AES-ECB 加密
func (a *AESECBPKCS5Padding) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return "", fmt.Errorf("创建 AES 加密器失败: %v", err)
	}

	// PKCS5 填充
	paddedText := PKCS7Padding([]byte(plaintext), aes.BlockSize)

	// 分块加密
	ciphertext := make([]byte, len(paddedText))
	for i := 0; i < len(paddedText); i += aes.BlockSize {
		block.Encrypt(ciphertext[i:i+aes.BlockSize], paddedText[i:i+aes.BlockSize])
	}

	// 返回指定格式的密文
	return formatOutput(ciphertext, a.outFormat)
}

// Decrypt 实现 AES-ECB 解密
func (a *AESECBPKCS5Padding) Decrypt(ciphertext string) (string, error) {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return "", fmt.Errorf("创建 AES 解密器失败: %v", err)
	}

	// 解码密文
	decodedCiphertext, err := parseInput(ciphertext, a.outFormat)
	if err != nil {
		return "", fmt.Errorf("解码密文失败: %v", err)
	}

	// 分块解密
	plaintext := make([]byte, len(decodedCiphertext))
	for i := 0; i < len(decodedCiphertext); i += aes.BlockSize {
		block.Decrypt(plaintext[i:i+aes.BlockSize], decodedCiphertext[i:i+aes.BlockSize])
	}

	// 去除填充
	unpaddedPlaintext, err := PKCS7Unpadding(plaintext)
	if err != nil {
		return "", fmt.Errorf("去除填充失败: %v", err)
	}

	return string(unpaddedPlaintext), nil
}

// 通用工具函数

// parseInput 解码密文
func parseInput(input, format string) ([]byte, error) {
	switch format {
	case "hex":
		return hex.DecodeString(input)
	case "base64":
		return base64.StdEncoding.DecodeString(input)
	default:
		return nil, errors.New("未知的输入格式")
	}
}

// formatOutput 格式化密文
func formatOutput(data []byte, format string) (string, error) {
	switch format {
	case "hex":
		return hex.EncodeToString(data), nil
	case "base64":
		return base64.StdEncoding.EncodeToString(data), nil
	default:
		return "", errors.New("未知的输出格式")
	}
}
