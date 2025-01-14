package core

import (
	"BronyaBot/config"
	"BronyaBot/global"
	"github.com/spf13/viper"
	"log"
)

const ConfigFile = "./configuration/application.yaml"

func InitConf() {
	viper.SetConfigFile(ConfigFile) // 指定配置文件路径
	viper.SetConfigType("yaml")     // 指定配置文件类型
	viper.WatchConfig()
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("读取失败: %v", err)
	}
	if err := viper.Unmarshal(&global.Config); err != nil {
		log.Fatalf("配置解析失败: %v", err)
	}
	config.BannerInit()
	log.Printf("load %s success...\n", ConfigFile)
}
