package config

type Config struct {
	MySql   MySql    `yaml:"mysql"`
	Logger  Logger   `yaml:"logger"`
	Mail    MailConf `yaml:"mail"`
	AI      AIConf   `yaml:"ai"`
	Account Account  `yaml:"account"`
}
