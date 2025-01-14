package core

import (
	"BronyaBot/global"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"time"
)

func InitGorm() *gorm.DB {
	if global.Config.MySql.Host == "" {
		global.Log.Warn("未配置mysql")
		return nil
	}

	dsn := global.Config.MySql.Dsn()
	var mySqlLogger logger.Interface
	if global.Config.MySql.LogLevel == "debug" {
		mySqlLogger = logger.Default.LogMode(logger.Info)
	} else if global.Config.MySql.LogLevel == "warn" {
		mySqlLogger = logger.Default.LogMode(logger.Warn)
	} else if global.Config.MySql.LogLevel == "error" {
		mySqlLogger = logger.Default.LogMode(logger.Error)
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: mySqlLogger,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
			NoLowerCase:   true,
		},
	})
	if err != nil {
		global.Log.Fatalf("[%s] mysql连接错误", dsn)
	}
	sqlDB, _ := db.DB()
	//TODO 设置mysql连接池 自定义配置 需要修改yaml与 conf—mysql
	sqlDB.SetMaxIdleConns(global.Config.MySql.MaxIdleConns)                                  //最大空闲连接数
	sqlDB.SetMaxOpenConns(global.Config.MySql.MaxOpenConns)                                  //最多容量
	sqlDB.SetConnMaxLifetime(time.Duration(global.Config.MySql.ConnMaxLifeTime) * time.Hour) //连接最大服用时间 不能超过mysql的wait_timeout

	return db
}
