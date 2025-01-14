package core

import (
	"BronyaBot/global"
	"crypto/tls"
	"gopkg.in/gomail.v2"
)

func InitMail() *gomail.Dialer {
	// 判断邮件配置是否为空
	if global.Config.Mail.Host == "" {
		global.Log.Warning("未配置邮箱服务器地址")
		return nil
	}
	// 设置默认值
	dialer := gomail.NewDialer(
		global.Config.Mail.Host,     // 邮件服务器地址
		global.Config.Mail.Port,     // 邮件服务器端口
		global.Config.Mail.User,     // 用户名
		global.Config.Mail.Password, // 密码
	)
	// 如果没有配置端口，设置默认值（通常 587 用于 TLS）
	if dialer.Port == 0 {
		dialer.Port = 587
	}
	// 配置 SSL 或 TLS
	if global.Config.Mail.SSL {
		// 如果启用了 SSL，使用端口 465
		dialer.Port = 465
		dialer.TLSConfig = &tls.Config{
			ServerName:         global.Config.Mail.Host,
			InsecureSkipVerify: false, // 设置是否验证 SSL/TLS 证书
		}
	} else {
		// 否则使用默认的 TLS 配置
		dialer.TLSConfig = &tls.Config{
			ServerName:         global.Config.Mail.Host,
			InsecureSkipVerify: false, // 默认不跳过验证
		}
	}

	// 配置 LocalName（一般可以不设置，除非需要指定）
	if global.Config.Mail.LocalName != "" {
		dialer.LocalName = global.Config.Mail.LocalName
	}
	// 检查是否有其他必需的参数
	if dialer.Username == "" || dialer.Password == "" {
		global.Log.Warning("未提供完整的邮箱认证信息")
		return nil
	}

	return dialer
}
