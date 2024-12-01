package config

import "github.com/yatori-dev/yatori-go-core/models/ctype"

type AiSetting struct {
	AiType ctype.AiType `json:"aiType"`
	AiUrl  string       `json:"aiUrl"`
	Model  string       `json:"model"`
	APIKEY string       `json:"API_KEY" yaml:"API_KEY" mapstructure:"API_KEY"`
}
