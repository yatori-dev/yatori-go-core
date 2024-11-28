package zhihuishu

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

func GetValidate() string {
	return ""
}

// 通过用户账号和密码获取SecretStr值
func GenerateSecretStr(username, password string) (string, error) {
	validate := GetValidate()
	if validate == "" {
		return "", fmt.Errorf("validate is empty")
	}
	// 创建参数
	params := map[string]string{
		"account":  username,
		"password": password,
		"validate": validate,
	}

	// 将参数转换为JSON
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return "", err
	}
	// URL编码
	encodedParams := url.QueryEscape(string(paramsJSON))

	// 替换特定字符
	encodedParams = strings.ReplaceAll(encodedParams, "%3A", ":")
	encodedParams = strings.ReplaceAll(encodedParams, "%2C", ",")

	// Base64编码
	encodedStr := base64.StdEncoding.EncodeToString([]byte(encodedParams))
	return encodedStr, nil
}
