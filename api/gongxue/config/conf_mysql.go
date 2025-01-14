package config

import "fmt"

type MySql struct {
	DataBase        string `yaml:"dataBase"`
	UserName        string `yaml:"userName"`
	PassWord        string `yaml:"passWord"`
	Port            int    `yaml:"port"`
	DriverName      string `yaml:"driverName"`
	Host            string `yaml:"host"`
	LogLevel        string `yaml:"log-level" mapstructure:"log-level"`
	MaxIdleConns    int    `yaml:"max-idle-conns" mapstructure:"max-idle-conns"`
	MaxOpenConns    int    `yaml:"max-open-conns" mapstructure:"max-open-conns"`
	ConnMaxLifeTime int    `yaml:"conn-max-life-time" mapstructure:"conn-max-life-time"`
}

func (m *MySql) Dsn() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True",
		m.UserName, m.PassWord, m.Host, m.Port, m.DataBase)
}
