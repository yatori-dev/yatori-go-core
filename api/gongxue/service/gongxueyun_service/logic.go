package gongxueyun_service

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/yatori-dev/yatori-go-core/api/gongxue/api"
	"github.com/yatori-dev/yatori-go-core/api/gongxue/entity"
	"github.com/yatori-dev/yatori-go-core/api/gongxue/global"
	data2 "github.com/yatori-dev/yatori-go-core/api/gongxue/service/gongxueyun_service/data"
	"github.com/yatori-dev/yatori-go-core/api/gongxue/utils"
	"github.com/yatori-dev/yatori-go-core/api/gongxue/utils/blockPuzzle"
	log2 "github.com/yatori-dev/yatori-go-core/utils/log"
	"strconv"
	"strings"
	"time"
	"github.com/yatori-dev/yatori-go-core/utils"
)

func (m *MoguDing) Run(runType string) {
	if err := m.HandleCaptcha(); err != nil {
		utils.SendMail(m.Email, "Block-Error", err.Error())
		global.Log.Error(err.Error())
		return
	}
	if err := m.Login(); err != nil {
		utils.SendMail(m.Email, "Login-Error-测试邮件请勿回复", err.Error())
		global.Log.Error(err.Error())
		return
	}

	if err := m.GetPlanId(); err != nil {
		global.Log.Warning(err.Error())
		return
	}
	if err := m.GetJobInfo(); err != nil {
		global.Log.Warn("Failed to get job info: %v", err)
		return
	}
	m.getWeeksTime()
	if runType == "sign" {
		m.SignIn()
	} else if runType == "week" {
		m.getSubmittedReportsInfo("week")
		m.SubmitReport("week", 1500)
	} else if runType == "month" {
		m.getSubmittedReportsInfo("month")
		m.SubmitReport("month", 1600)
	}

}

const (
	MaxRetries    = 15
	DefaultPlanID = "6686304d065db846edab7d4565065abc"
	PageSize      = 999999
)

var (
	headers = map[string][]string{
		"User-Agent":   {utils.DefaultUserAgent},
		"Content-Type": {"application/json; charset=UTF-8"},
		"host":         {"api.moguding.net:9000"},
	}
	clientUid = strings.ReplaceAll(uuid.New().String(), "-", "")
	client    = utils.NewHttpClient()
)

// 添加请求头
func addHeader(key, value string) {
	headers[key] = []string{value}
}

// 添加标准请求头
func addStandardHeaders(roleKey, userId, authorization string) {
	addHeader("rolekey", roleKey)
	addHeader("userid", userId)
	addHeader("authorization", authorization)
}

// 滑块验证码逻辑
func (mo *MoguDing) HandleCaptcha() error {
	for attempt := 1; attempt <= MaxRetries; attempt++ {
		if err := mo.processBlock(); err == nil {
			return nil
		}
		//global.Log.Warn(fmt.Sprintf("Retry captcha (%d/%d)", attempt, MaxRetries))
		log2.Print(log2.INFO, fmt.Sprintf("Retry captcha (%d/%d)", attempt, MaxRetries))
		time.Sleep(10 * time.Second)
	}
	return fmt.Errorf("captcha failed after %d attempts", MaxRetries)
}
func (mo *MoguDing) processBlock() error {
	// 获取验证码数据
	requestData := map[string]interface{}{
		"clientUid":   clientUid,
		"captchaType": "blockPuzzle",
	}
	body, _, err := client.SendRequest("POST", api.BaseApi+api.BlockPuzzle, requestData, headers)
	if err != nil {
		return fmt.Errorf("failed to fetch block puzzle: %v", err)
	}
	// 解析响应数据
	blockData := &data2.BlockRes{}
	if err := json.Unmarshal(body, &blockData); err != nil {
		return fmt.Errorf("failed to parse block puzzle response: %v", err)
	}
	// 初始化滑块验证码
	captcha, err := blockPuzzle.NewSliderCaptcha(blockData.Data.JigsawImageBase64, blockData.Data.OriginalImageBase64)
	if err != nil {
		return fmt.Errorf("failed to initialize captcha: %v", err)
	}
	x, _ := captcha.FindBestMatch()

	// 加密并验证
	xY := map[string]string{"x": strconv.FormatFloat(GenerateRandomFloat(x), 'f', -1, 64), "y": strconv.Itoa(5)}
	//global.Log.Info(fmt.Sprintf("Captcha matched at: xY=%s", xY))
	log2.Print(log2.INFO, fmt.Sprintf("Captcha matched at: xY=%s", xY))

	marshal, err := json.Marshal(xY)

	mo.CommParameters.xY = string(marshal)
	mo.CommParameters.token = blockData.Data.Token
	mo.CommParameters.secretKey = blockData.Data.SecretKey
	cipher, _ := utils.NewAESECBPKCS5Padding(mo.CommParameters.secretKey, "base64")
	encrypt, _ := cipher.Encrypt(mo.CommParameters.xY)
	requestData = map[string]interface{}{
		"pointJson":   encrypt,
		"token":       blockData.Data.Token,
		"captchaType": "blockPuzzle",
	}
	body, _, err = client.SendRequest("POST", api.BaseApi+api.CHECK, requestData, headers)
	if err != nil {
		return fmt.Errorf("failed to verify captcha: %v", err)
	}

	// 解析验证结果
	jsonContent := &data2.CheckData{}
	if err := json.Unmarshal(body, &jsonContent); err != nil {
		return fmt.Errorf("failed to parse check response: %v", err)
	}
	if jsonContent.Code == 6111 {
		return fmt.Errorf("captcha verification failed, retry needed")
	}
	//global.Log.Info("Captcha verification successful")
	log2.Print(log2.INFO, "Captcha verification successful")
	padding, _ := utils.NewAESECBPKCS5Padding(blockData.Data.SecretKey, "base64")
	encrypt, err = padding.Encrypt(jsonContent.Data.Token + "---" + mo.CommParameters.xY)
	if err != nil {
		//global.Log.Info(fmt.Sprintf("Failed to encrypt captcha: %v", err))
		log2.Print(log2.INFO, fmt.Sprintf("Failed to encrypt captcha: %v", err))
	}
	mo.CommParameters.captcha = encrypt
	return nil
}
func (mogu *MoguDing) Login() error {
	padding, _ := utils.NewAESECBPKCS5Padding(utils.MoGuKEY, "hex")
	encryptPhone, _ := padding.Encrypt(mogu.PhoneNumber)
	encryptPassword, _ := padding.Encrypt(mogu.Password)
	timestamp, _ := EncryptTimestamp(time.Now().UnixMilli())
	requestData := map[string]interface{}{
		"phone":     encryptPhone,
		"password":  encryptPassword,
		"captcha":   mogu.CommParameters.captcha,
		"loginType": "android",
		"uuid":      clientUid,
		"device":    "android",
		"version":   "5.15.0",
		"t":         timestamp,
	}
	var login = &data2.Login{}
	var loginData = &data2.LoginData{}
	body, _, err := client.SendRequest("POST", api.BaseApi+api.LoginAPI, requestData, headers)
	if err != nil {
		global.Log.Info(fmt.Sprintf("Failed to send request: %v", err))
	}
	json.Unmarshal(body, &login)
	if login.Code != 200 {
		return fmt.Errorf(login.Msg)

	}
	decrypt, err := padding.Decrypt(login.Data)
	json.Unmarshal([]byte(decrypt), &loginData)
	if err != nil {
		global.Log.Info(fmt.Sprintf("Failed to decrypt data: %v", err))
	}
	mogu.RoleKey = loginData.RoleKey
	mogu.UserId = loginData.UserId
	mogu.Authorization = loginData.Token
	log2.Print(log2.INFO, "================")
	log2.Print(log2.INFO, loginData.NikeName)
	log2.Print(log2.INFO, loginData.Phone)
	log2.Print(log2.INFO, "================")
	log2.Print(log2.INFO, "Login successful")
	return nil
}
func (mogu *MoguDing) GetPlanId() error {
	defaultId := "6686304d065db846edab7d4565065abc"
	planData := &data2.PlanByStuData{}
	timestamp, _ := EncryptTimestamp(time.Now().UnixMilli())
	sign := utils.CreateSign(mogu.UserId, mogu.RoleKey)
	addHeader("sign", sign)
	addStandardHeaders(mogu.RoleKey, mogu.UserId, mogu.Authorization)
	body := map[string]interface{}{
		"pageSize": strconv.Itoa(999999),
		"t":        timestamp,
	}
	request, _, err := client.SendRequest("POST", api.BaseApi+api.GetPlanIDAPI, body, headers)
	if err != nil {
		global.Log.Info(fmt.Sprintf("Failed to send request: %v", err))
	}
	json.Unmarshal(request, &planData)
	for i := range planData.Data {
		mogu.PlanID = planData.Data[i].PlanId
		mogu.PlanName = planData.Data[i].PlanName
	}
	if strings.EqualFold(mogu.PlanID, defaultId) {
		return fmt.Errorf(mogu.PlanName)
	}
	//global.Log.Info("================")
	log2.Print(log2.INFO, "================")
	//global.Log.Info(mogu.PlanID)
	log2.Print(log2.INFO, mogu.PlanID)
	//global.Log.Info(mogu.PlanName)
	log2.Print(log2.INFO, mogu.PlanName)
	//global.Log.Info("================")
	log2.Print(log2.INFO, "================")
	return nil
}
func (mogu *MoguDing) GetJobInfo() error {
	job := &data2.JobInfoData{}
	addStandardHeaders(mogu.RoleKey, mogu.UserId, mogu.Authorization)
	timestamp, _ := EncryptTimestamp(time.Now().UnixMilli())
	body := map[string]interface{}{
		"planId": mogu.PlanID,
		"t":      timestamp,
	}
	request, _, err := client.SendRequest("POST", api.BaseApi+api.GetJobInfoAPI, body, headers)
	if err != nil {
		global.Log.Info(fmt.Sprintf("Failed to send request: %v", err))
	}
	json.Unmarshal(request, &job)
	if job.Data.JobId == "" {
		return fmt.Errorf("job info not found")
	} else {
		mogu.JobInfo.JobName = job.Data.JobName
		mogu.JobInfo.Address = job.Data.Address
		mogu.JobInfo.CompanyName = job.Data.CompanyName
	}
	return nil
}
func (mogu *MoguDing) SignIn() {
	resdata := &data2.SaveData{}
	filling := DataStructureFilling(mogu)
	sign := utils.CreateSign(filling["device"].(string), filling["type"].(string), mogu.PlanID, mogu.UserId, filling["address"].(string))
	addStandardHeaders(mogu.RoleKey, mogu.UserId, mogu.Authorization)
	addHeader("sign", sign)
	request, _, err := client.SendRequest("POST", api.BaseApi+api.SignAPI, filling, headers)
	if err != nil {
		global.Log.Info(fmt.Sprintf("Failed to send request: %v", err))
	}

	json.Unmarshal(request, &resdata)
	global.Log.Info("================")
	global.Log.Info(resdata.Msg)
	global.Log.Info("================")
	if resdata.Msg == "success" {
		//log2.Print(log2.INFO, "打卡成功：", resdata)
	} else {
		log2.Print(log2.INFO, "打卡失败：", resdata)
	}
	utils.SendMail(mogu.Email, "检查是否打卡完成", resdata.Msg+"\n如果未成功请联系管理员")

}

func (mogu *MoguDing) CusSignIn(signType string) {
	resdata := &data2.SaveData{}
	filling := CusDataStructureFilling(signType, mogu)
	sign := utils.CreateSign(filling["device"].(string), filling["type"].(string), mogu.PlanID, mogu.UserId, filling["address"].(string))
	addStandardHeaders(mogu.RoleKey, mogu.UserId, mogu.Authorization)
	addHeader("sign", sign)
	request, _, err := client.SendRequest("POST", api.BaseApi+api.SignAPI, filling, headers)
	if err != nil {
		global.Log.Info(fmt.Sprintf("Failed to send request: %v", err))
	}

	json.Unmarshal(request, &resdata)
	log2.Print(log2.INFO, "================")
	log2.Print(log2.INFO, resdata.Msg)
	log2.Print(log2.INFO, "================")
	if resdata.Msg == "success" {
		//log2.Print(log2.INFO, "打卡成功：", resdata)
	} else {
		log2.Print(log2.INFO, "打卡失败：", resdata)
	}
	//utils.SendMail(mogu.Email, "检查是否打卡完成", resdata.Msg+"\n如果未成功请联系管理员")

}

func (mogu *MoguDing) updateSignState(state int) {
	// 更新数据库表中的 state 字段
	if mogu.ID != -1 {
		err := global.DB.Model(&entity.SignEntity{}).Where("username = ?", mogu.PhoneNumber).Update("state", state).Error
		if err != nil {
			global.Log.Error(fmt.Sprintf("Failed to update state for user %s: %v", mogu.PhoneNumber, err))
		} else {
			global.Log.Info(fmt.Sprintf("Successfully updated state for user %s to %d", mogu.PhoneNumber, state))
		}
	}
}

// 获取已经提交的日报、周报或月报的数量。
func (mogu *MoguDing) getSubmittedReportsInfo(reportType string) {
	report := &data2.ReportsInfo{}
	sign := utils.CreateSign(mogu.UserId, mogu.RoleKey, reportType)
	addStandardHeaders(mogu.RoleKey, mogu.UserId, mogu.Authorization)
	addHeader("sign", sign)
	timestamp, _ := EncryptTimestamp(time.Now().UnixMilli())
	body := map[string]interface{}{
		"currPage":   1,
		"pageSize":   10,
		"reportType": reportType,
		"planId":     mogu.PlanID,
		"t":          timestamp,
	}
	request, _, err := client.SendRequest("POST", api.BaseApi+api.GetWeekCountAPI, body, headers)
	if err != nil {
		global.Log.Info(fmt.Sprintf("Failed to send request: %v", err))
	}
	json.Unmarshal(request, &report)
	if report.Flag == 0 {
		global.Log.Warning("未发现之前存在报告，初始化报告为0")
		mogu.ReportStruct.CreateTime = ""
		mogu.ReportStruct.ReportId = ""
		mogu.ReportStruct.ReportType = ""
		mogu.ReportStruct.Flag = 0
		return
	} else {
		mogu.ReportStruct.CreateTime = report.Data[0].CreateTime
		mogu.ReportStruct.ReportId = report.Data[0].ReportId
		mogu.ReportStruct.ReportType = report.Data[0].ReportType
		mogu.ReportStruct.Flag = report.Flag
	}
}

// 获取提交周时间
func (mogu *MoguDing) getWeeksTime() {
	week := &data2.WeeksData{}
	addStandardHeaders(mogu.RoleKey, mogu.UserId, mogu.Authorization)
	timestamp, _ := EncryptTimestamp(time.Now().UnixMilli())
	body := map[string]interface{}{
		"t": timestamp,
	}
	request, _, err := client.SendRequest("POST", api.BaseApi+api.GetWeeks, body, headers)
	if err != nil {
		global.Log.Info(fmt.Sprintf("Failed to send request: %v", err))
	}
	json.Unmarshal(request, &week)
	if len(week.Data) > 0 {
		mogu.WeekTime.Week = week.Data[0].Weeks
		mogu.WeekTime.StartTime = week.Data[0].StartTime
		mogu.WeekTime.EndTime = week.Data[0].EndTime
		mogu.WeekTime.IsDefault = week.Data[0].IsDefault
		mogu.WeekTime.Flag = week.Flag
	}
}

// SubmitReport
// 提交定时报告
func (mogu *MoguDing) SubmitReport(reportType string, limit int) {
	res := &data2.RepResData{}
	var _t string
	switch reportType {
	case "week":
		_t = "周报"
	case "month":
		_t = "月报"
	case "day":
		_t = "日报"
	}
	input := fmt.Sprintf("报告类型: %s 工作地点: %s 公司名: %s 岗位职责: %s", _t, mogu.JobInfo.Address, mogu.JobInfo.CompanyName, mogu.JobInfo.JobName)
	ai := GenerateReportAI(input, limit)
	addStandardHeaders(mogu.RoleKey, mogu.UserId, mogu.Authorization)
	filling := SubmitStructureFilling(mogu, ai, "报告", reportType)
	sign := utils.CreateSign(mogu.UserId, reportType, mogu.PlanID, "报告")
	addHeader("sign", sign)
	request, _, _ := client.SendRequest("POST", api.BaseApi+api.SubmitAReport, filling, headers)
	json.Unmarshal(request, &res)
	global.Log.Info(fmt.Sprintf("Submit report: %v", res))
	utils.SendMail(mogu.Email, strconv.Itoa(res.Code), res.Msg+"\n如果未成功请联系管理员")
}
