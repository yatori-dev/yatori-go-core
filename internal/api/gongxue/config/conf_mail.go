package config

type MailConf struct {
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	User      string `yaml:"user"`
	Password  string `yaml:"password"`
	SSL       bool   `yaml:"ssl"`
	LocalName string `yaml:"local-name" mapstructure:"local-name"`
}
