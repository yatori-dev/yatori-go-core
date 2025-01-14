package config

import (
	"fmt"
)

func BannerInit() {
	blue := "\033[34m"
	fmt.Println(blue + ` 
	 ██    ██             ██                   ██
	░░██  ██             ░██                  ░░ 
	 ░░████    ██████   ██████  ██████  ██████ ██
	  ░░██    ░░░░░░██ ░░░██░  ██░░░░██░░██░░█░██
	   ░██     ███████   ░██  ░██   ░██ ░██ ░ ░██
	   ░██    ██░░░░██   ░██  ░██   ░██ ░██   ░██
	   ░██   ░░████████  ░░██ ░░██████ ░███   ░██
	   ░░     ░░░░░░░░    ░░   ░░░░░░  ░░░    ░░	先行定制版本
	Yatori系列项目官网：https://yatori-dev.github.io/yatori-docs
	基于BronyaBot开发：https://github.com/mirai-MIC/BronyaBot
	注：该版本为定制版本请勿外传
	v1.0-release
`)
}
