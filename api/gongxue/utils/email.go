package utils

import (
	"fmt"
	"github.com/yatori-dev/yatori-go-core/api/gongxue/global"
	"gopkg.in/gomail.v2"
	"time"
)

// SendMail 发送邮件的通用函数
// 参数：to 目标收件人，subject 邮件主题，body 邮件正文内容
func SendMail(to, subject, body string) {
	// 如果 Mail 配置为空，输出错误并返回
	if global.Mail == nil {
		global.Log.Warning("邮件服务未初始化")
		return
	}

	// 获取当前时间
	currentTime := time.Now().Format("2006-01-02 15:04:05") // 格式化时间

	// 创建新的邮件内容
	msg := gomail.NewMessage()

	// 设置发件人（这里假设发件人是配置中的邮箱）
	msg.SetHeader("From", global.Config.Mail.User)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", fmt.Sprintf("%s - %s", subject, currentTime))         // 将时间添加到主题
	msg.SetBody("text/plain", fmt.Sprintf("%s\n\nSent at: %s", body, currentTime)) // 将时间添加到邮件正文

	// 调用 gomail.Dialer 发送邮件
	if err := global.Mail.DialAndSend(msg); err != nil {
		global.Log.Warningf("发送邮件失败: %v\n", err)
	} else {
		global.Log.Info("邮件发送成功")
	}

}
