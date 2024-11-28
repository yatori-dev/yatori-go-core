package zhihuishu

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

func getConfig() map[string]map[string]string {
	// This is a placeholder for your configuration retrieval function.
	// You should implement this function to return the actual configuration.
	return map[string]map[string]string{
		"encrypt": {
			"AES_KEY": "your_aes_key_here",
			"AES_IV":  "your_aes_iv_here",
		},
	}
}

func getAESKeys(keyName string) string {
	return getConfig()["encrypt"][keyName]
}

func getEvKey(keyName string) string {
	return getConfig()["encrypt"][keyName]
}

func pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := make([]byte, padding)
	for i := range padText {
		padText[i] = byte(padding)
	}
	return append(data, padText...)
}

func AES_CBC_encrypt(key, iv, text string) (string, error) {
	// 将AES_KEY和IV转换为字节类型
	keyBytes := []byte(key)
	ivBytes := []byte(iv)

	// 初始化AES加密器
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}

	// 使用PKCS7填充
	plaintext := []byte(text)
	plaintext = pad(plaintext, aes.BlockSize)

	// 加密数据
	ciphertext := make([]byte, len(plaintext))
	mode := cipher.NewCBCEncrypter(block, ivBytes)
	mode.CryptBlocks(ciphertext, plaintext)

	// 对加密结果进行base64编码
	encryptedData := base64.StdEncoding.EncodeToString(ciphertext)
	return encryptedData, nil
}

func encryptParams(params interface{}, keyName string) (string, error) {
	var paramsStr string
	switch v := params.(type) {
	case map[string]interface{}:
		paramsBytes, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		paramsStr = string(paramsBytes)
	default:
		paramsStr = fmt.Sprintf("%v", v)
	}

	key := getAESKeys(keyName)
	iv := getConfig()["encrypt"]["AES_IV"]
	return AES_CBC_encrypt(key, iv, paramsStr)
}

func getTokenId(studiedLessonDtoId int) string {
	return base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(studiedLessonDtoId)))
}

func encryptEv(data []interface{}, key string) string {
	// 将data转换为字符串
	var dataStr string
	for _, v := range data {
		dataStr += fmt.Sprintf("%v;", v)
	}

	// 对key字符生成加密
	var ev string
	gen := func() chan int {
		ch := make(chan int)
		go func() {
			for {
				for i := 0; i < len(key); i++ {
					ch <- int(key[i])
				}
			}
		}()
		return ch
	}

	genChan := gen()
	dataStr = strings.TrimSuffix(dataStr, ";") // 去掉末尾的分号

	for _, c := range dataStr {
		tmp := fmt.Sprintf("%x", int(c)^<-genChan)
		if len(tmp) < 2 {
			tmp = "0" + tmp
		}
		ev += tmp[len(tmp)-4:]
	}
	return ev
}

func genWatchPoint(startTime, endTime int) string {
	recordInterval := 1990
	totalStudyTimeInterval := 4990
	//cacheInterval := 180000
	//databaseInterval := 300000

	var watchPoint string
	totalStudyTime := float64(startTime)
	interval := float64(endTime - startTime)

	for i := 1; i <= int(interval*1000); i++ {
		if i%totalStudyTimeInterval == 0 {
			totalStudyTime += 5
		}
		if i%recordInterval == 0 && i >= recordInterval {
			t := int(totalStudyTime/5) + 2
			if watchPoint == "" {
				watchPoint = "0,1,"
			} else {
				watchPoint += ","
			}
			watchPoint += strconv.Itoa(t)
		}
	}
	return watchPoint
}
