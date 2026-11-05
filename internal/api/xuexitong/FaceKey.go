package xuexitong

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"sync"
)

// 人脸签名私钥、上传 appid/secret 不直接以明文出现在源码/二进制里，
// 而是保存为 AES-256-GCM 密文，首次使用时再解密。这样仓库里不会
// 出现可直接检索的 PEM 私钥块或固定密钥字符串。
//
// 说明：这是静态混淆而非不可逆保护——密钥最终仍要参与本地签名，
// 只是避免明文硬编码、误提交和随手被 grep 到。
const (
	faceKeySeedA  = "yatori::cx::face::v1::"
	faceKeySeedB  = "z4ok6lu^oWp4_AES"
	faceKeyCipher = "lRw2Vaxh6mji5yWMzYeHShLRMSF0DKqXFAlEwDUTjjOWllNy5NR96iVEtPF4FDTuElkIRM7Z/QCL6IMAt/BrB65p9ba2QZwwRZx1hXQ7M0GzkNTTkHgrKe+gnpED0jECYswIIjX0glDWSlUzTjmqX4YmOiHeFg3MgL8Us7+qdJ6YdWVex3W3qlkHzNGEawTAL9+ow66OQZCrKRuHbmYEpmnLw507h9SvAvKmNBqYj+VyCY5RoCE1R3w0rdy9erLpAnhcLjHc26wCxOTe1pGjSUG+dR7bjA7XwYfYLzovYD+Wmc62w4WFzReg1syrJCoEtT2tQjoGzym8rN8hERvRYudiL1SpwRwCWDzw2dumj8oqk6rhMgQknzFBng0MzpefwJWM2tu8ilzFcYEYSIwTL2UAUwsEcdBR8tXMXA7+YYctwtGtNwX11H0x/eWEcKITbwO0CN3eyDa1CoAXqHBBS4fqnJr/ZJ1g2fzEFeDzWYLw51n/yZ0wLn7koR7UxCEhyqV7R4mar8fliPExJVmEmCI0ntBt2LtMs2BG3gkzX2+FcM/KFcSeWel+9lDCPIsj0OB7BJyfTmxm7jfafFMuNBlFcU9dGPVn0QaITlPT/5jnaqkffLFQAglBR1ASM0WclJocx75Wg9UJbEvueqFmRMqACcCUAmNHpodMmCcjemH1Gzj5xU27pxjzXWwWyO2WIurxxqTHN2eRMB4hnqD5eTXFMrajsnnlzsFUHBZYiMZVNYqOPfnIb6Jbvgwh14sJ1N9jOnpA6zLeaYGWRwQn1EthrThMV9eY3dMobBg4gGEgyUpJN94V4E0akp8TSVPFqt76dLPdI5ov0VWAdTZnR8+7G+jaJRBZQU/bdmHX8wIXO5r7gVtgMOXRRp6Ld0Bc+pn5aNhEeE5VGERVqnPxa1J083OD1ycWihz8UOERLr1LrVb1yLzISSqsyidw3htPgaFMicjLu40ubVFqxKpjR/ar8GI7DGLtaEseMYxSLFko0WteID/dhCuu1zoUvjXEruT8NbbazS7I9phdgAPRxB/RoszYsqIWnYExRhZiSl0LGR/AgyVfPXQ/kQ/6clA9H66Jlsa/SxB83OR1ohxiCL0MIhMslTPpk1JyzqWJxNEz+87kfjAxTR7xgsWpFtMQxYLB0Q/qja8CWl6UJgPCcqQHiAh7jyuqz7OZADEs0qhBElupFWRtqWbd+hN8iEW+My7H5aKKiDTnxKpJZrEb1XEl6qBO1qCdz6pzIXOCm8iFHPjHCQejQTRO0UUwx6YM"

	faceUploadAppIDSeed    = "yatori::cx::upload::appid::v1"
	faceUploadAppIDCipher  = "UzOgRwAsF42LRsumz1IaEbYL5CR8TvNDSQlLq6banTpocRPqYv1yhAe3OvL7cvkd30i+lUDqks4w9dr/RaX6yg=="
	faceUploadSecretSeed   = "yatori::cx::upload::secret::v1"
	faceUploadSecretCipher = "Q0HHqlZivNdBOyreffyscDK3LGcNoaw0kxJDFmQaws9XpfWAsmlAPhZdnWl1GwyTTl+4k4YkDjqGc1uu"
)

var (
	faceKeyOnce sync.Once
	faceKeyPEM  string
)

// DefaultFaceKeyPEM 返回学习通 App 内置的人脸签名 RSA-1024 私钥。
//
// 该私钥随 App 一起分发，所有安装包一致，不属于任何单个用户；
// 内置它是为了让纯本地人脸签名开箱即用，调用方只需要提供设备
// 的 cxcid 与当前 sc，不必再额外配置 keyPem / keyPemFile。
func DefaultFaceKeyPEM() string {
	faceKeyOnce.Do(func() {
		faceKeyPEM = decryptFaceSecret(faceKeySeedA+faceKeySeedB, faceKeyCipher)
	})
	return faceKeyPEM
}

// FaceUploadAppID 返回人脸 v2 上传接口使用的 appid。
func FaceUploadAppID() string {
	return decryptFaceSecret(faceUploadAppIDSeed, faceUploadAppIDCipher)
}

// FaceUploadSecret 返回人脸 v2 上传签名使用的 secret。
func FaceUploadSecret() string {
	return decryptFaceSecret(faceUploadSecretSeed, faceUploadSecretCipher)
}

func decryptFaceSecret(seed, cipherText string) string {
	sum := sha256.Sum256([]byte(seed))
	block, err := aes.NewCipher(sum[:])
	if err != nil {
		return ""
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return ""
	}
	data, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil || len(data) < gcm.NonceSize() {
		return ""
	}
	plain, err := gcm.Open(nil, data[:gcm.NonceSize()], data[gcm.NonceSize():], nil)
	if err != nil {
		return ""
	}
	return string(plain)
}
