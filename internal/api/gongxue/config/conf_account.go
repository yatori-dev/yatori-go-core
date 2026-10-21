package config

type Account struct {
	Off          bool               `yaml:"off"`
	Promptly     bool               `yaml:"promptly"`
	PromptlyType string             `yaml:"promptlyType"`
	Gongxueyun   []GongXueYunStruct `yaml:"gongxueyun"`
	Cx           CxStruct           `yaml:"cx"`
}

type GongXueYunStruct struct {
	Phone     string `yaml:"phone"`
	Password  string `yaml:"password"`
	Country   string `yaml:"country"`
	Province  string `yaml:"province"`
	City      string `yaml:"city"`
	Area      string `yaml:"area"`
	Latitude  string `yaml:"latitude"`
	Longitude string `yaml:"longitude"`
	Email     string `yaml:"email"`
	Address   string `yaml:"address"`
}
type CxStruct struct {
	Phone    string `yaml:"phone"`
	Password string `yaml:"password"`
}
