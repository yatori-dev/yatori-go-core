package gongxueyun_service

import (
	"encoding/json"
	"fmt"
	"github.com/yatori-dev/yatori-go-core/api/gongxue/api"
	"github.com/yatori-dev/yatori-go-core/api/gongxue/entity"
	"github.com/yatori-dev/yatori-go-core/api/gongxue/global"
	data2 "github.com/yatori-dev/yatori-go-core/api/gongxue/service/gongxueyun_service/data"
	"github.com/yatori-dev/yatori-go-core/api/gongxue/utils"
	"math"
	"math/rand"
	"strconv"
	"time"
)

const (
	START = iota
	END
)

// GetFormattedTime returns the current time formatted as a string.
func GetFormattedTime() (string, error) {
	now := time.Now()
	return now.Format("2006-01-02 15:04:05"), nil
}

// GetClockType determines the clock type (START or END) based on the current time.
func GetClockType() (string, error) {
	now := time.Now()
	if now.Hour() >= 12 {
		return "END", nil
	}
	return "START", nil
}

// GetCustomClockType Custom START or END
func GetCustomClockType(playType int) (string, error) {
	switch playType {
	case START:
		return "START", nil
	case END:
		return "END", nil
	default:
		return "START", nil
	}
}

// DataStructureFilling creates a data structure for signing.
func DataStructureFilling(mogu *MoguDing) map[string]interface{} {
	var typeStr string
	if !global.Config.Account.Promptly {
		resTypeStr, err := GetClockType()
		if err != nil {
			return nil
		}
		typeStr = resTypeStr
	} else {
		typeStr = global.Config.Account.PromptlyType
	}

	encryptTime, err := EncryptTimestamp(time.Now().UnixMilli())
	if err != nil {
		global.Log.Error("Failed to encrypt timestamp: ", err)
		return nil
	}

	formattedTime, err := GetFormattedTime()
	if err != nil {
		global.Log.Error("Failed to get formatted time: ", err)
		return nil
	}

	return map[string]interface{}{
		"address":    mogu.Sign.Address,
		"city":       mogu.Sign.City,
		"area":       mogu.Sign.Area,
		"country":    mogu.Sign.Country,
		"createTime": formattedTime,
		"device":     "{brand: Redmi Note 5, systemVersion: 14, Platform: Android}",
		"latitude":   mogu.Sign.Latitude,
		"longitude":  mogu.Sign.Longitude,
		"province":   mogu.Sign.Province,
		"state":      "NORMAL",
		"type":       typeStr,
		"userId":     mogu.UserId,
		"t":          encryptTime,
		"planId":     mogu.PlanID,
	}
}

// 客制化打卡上班还是下班信息
func CusDataStructureFilling(typeStr string, mogu *MoguDing) map[string]interface{} {

	encryptTime, err := EncryptTimestamp(time.Now().UnixMilli())
	if err != nil {
		global.Log.Error("Failed to encrypt timestamp: ", err)
		return nil
	}

	formattedTime, err := GetFormattedTime()
	if err != nil {
		global.Log.Error("Failed to get formatted time: ", err)
		return nil
	}

	return map[string]interface{}{
		"address":    mogu.Sign.Address,
		"city":       mogu.Sign.City,
		"area":       mogu.Sign.Area,
		"country":    mogu.Sign.Country,
		"createTime": formattedTime,
		"device":     "{brand: Redmi Note 5, systemVersion: 14, Platform: Android}",
		"latitude":   mogu.Sign.Latitude,
		"longitude":  mogu.Sign.Longitude,
		"province":   mogu.Sign.Province,
		"state":      "NORMAL",
		"type":       typeStr,
		"userId":     mogu.UserId,
		"t":          encryptTime,
		"planId":     mogu.PlanID,
	}
}

// SubmitStructureFilling creates a submission data structure.
func SubmitStructureFilling(mogu *MoguDing, content, title, retype string) map[string]interface{} {
	timestamp, err := EncryptTimestamp(time.Now().UnixMilli())
	if err != nil {
		global.Log.Error("Failed to encrypt timestamp: ", err)
		return nil
	}

	submitData := data2.SubmitData{
		Weeks:      mogu.WeekTime.Week,
		Content:    content,
		PlanId:     mogu.PlanID,
		ReportType: retype,
		Title:      title,
		JobId:      mogu.CommParameters.JobId,
		T:          timestamp,
		StartTime:  mogu.WeekTime.StartTime,
		EndTime:    mogu.WeekTime.EndTime,
	}

	return data2.SubmitDataFunc(submitData)
}

// EncryptTimestamp encrypts a given timestamp.
func EncryptTimestamp(timestamp int64) (string, error) {
	padding, err := utils.NewAESECBPKCS5Padding(utils.MoGuKEY, "hex")
	if err != nil {
		return "", fmt.Errorf("failed to initialize padding: %v", err)
	}

	encryptTime, err := padding.Encrypt(strconv.FormatInt(timestamp, 10))
	if err != nil {
		return "", fmt.Errorf("failed to encrypt timestamp: %v", err)
	}
	return encryptTime, nil
}

// GenerateRandomFloat generates a random float based on a base integer part.
func GenerateRandomFloat(baseIntegerPart int) float64 {
	rand.Seed(time.Now().UnixNano())

	adjustment := rand.Intn(3) - 1 // Generates -1, 0, or 1
	integerPart := baseIntegerPart + adjustment

	intPartLength := len(fmt.Sprintf("%d", integerPart))
	totalLength := rand.Intn(10) + 10 // Total length between 10 and 19
	decimalPlaces := totalLength - intPartLength

	if decimalPlaces <= 0 {
		decimalPlaces = 1
	}

	decimalPart := rand.Float64() * math.Pow(10, float64(decimalPlaces))
	decimalPart = math.Trunc(decimalPart) / math.Pow(10, float64(decimalPlaces))

	return float64(integerPart) + decimalPart
}

// GenerateReportAI generates a report using AI based on user input.
func GenerateReportAI(userInput string, wordLimit int) string {
	resData := &data2.AIData{}

	headers := map[string][]string{
		"Content-Type":  {"application/json"},
		"Authorization": {"Bearer " + global.Config.AI.Token},
	}
	dat := map[string]interface{}{
		"max_tokens":  4096,
		"top_k":       4,
		"temperature": 0.5,
		"messages": []map[string]string{
			{"role": "user", "content": fmt.Sprintf("According to the information provided by the user, write an article strictly according to the template. Do not use Markdown syntax, HTML tags, or any special formatting. The output must be plain text and match the job description. The content should be fluent, conform to Chinese grammatical conventions, Repeat: DO NOT USE MARKDOWN ,and have more than %d characters.", wordLimit)},
			{"role": "user", "content": "模板：实习地点：xxxx\n\n工作内容：\n\nxzzzx\n\n工作总结：\n\nxxxxxx\n\n遇到问题：\n\nxzzzx\n\n自我评价：\n\nxxxxxx"},
			{"role": "user", "content": "不能使用markdown格式输出注意！！！这是重要的"},
			{"role": "user", "content": userInput},
		},
		"model":  "4.0Ultra",
		"stream": false,
	}

	request, err, _ := utils.NewHttpClient().SendRequest("POST", api.XUNFEIAPI, dat, headers)
	if err != nil {
		global.Log.Error("Failed to send request: ", err)
		return ""
	}

	if err := json.Unmarshal(request, resData); err != nil {
		global.Log.Error("Failed to parse AI response: ", err)
		return ""
	}

	global.Log.Info("Generate successful!")
	global.Log.Infof("Input token usage: %d\tOutput token usage: %d\tTotal token usage: %d", resData.Usage.PromptTokens, resData.Usage.CompletionTokens, resData.Usage.TotalTokens)
	return resData.Choices[0].Message.Content
}

// LoadUsers loads user data either from configuration or database.
func LoadUsers() []entity.SignEntity {
	if global.Config.Account.Off {
		global.Log.Info("Local single-user mode enabled.")
		var entityList []entity.SignEntity
		for k, v := range global.Config.Account.Gongxueyun {
			entityList = append(entityList, entity.SignEntity{
				ID:        k + 1,
				Username:  v.Phone,
				Password:  v.Password,
				Country:   v.Country,
				Province:  v.Province,
				City:      v.City,
				Area:      v.Area,
				Address:   v.Address,
				Latitude:  v.Latitude,
				Longitude: v.Longitude,
				Email:     v.Email,
			})
		}
		return entityList
	}

	var users []entity.SignEntity
	global.DB.Find(&users)
	if len(users) == 0 {
		global.Log.Warn("No users found in the database.")
	} else {
		global.Log.Info("Users loaded successfully.")
	}
	return users
}
