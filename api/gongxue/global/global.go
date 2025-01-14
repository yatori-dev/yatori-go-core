package global

import (
	"github.com/sirupsen/logrus"
	"github.com/yatori-dev/yatori-go-core/api/gongxue/config"
	"gopkg.in/gomail.v2"
	"gorm.io/gorm"
)

var (
	Config *config.Config
	DB     *gorm.DB
	Log    *logrus.Logger
	Mail   *gomail.Dialer
)
