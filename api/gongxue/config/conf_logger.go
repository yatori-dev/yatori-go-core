package config

type Logger struct {
	Level        string `yaml:"level"`
	Prefix       string `yaml:"prefix"`
	Director     string `yaml:"director"`
	ShowLine     bool   `yaml:"showLine"`     //是否显示行号
	LogInConsole bool   `yaml:"logInConsole"` //是否显示打印的路径
}
