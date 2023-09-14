package config

type Server struct {
	Redis   Redis   `json:"redis" yaml:"redis"`
	Mysql   Mysql   `json:"mysql" yaml:"mysql"`
	Logger  Logger  `json:"logger" yaml:"logger"`
	System  System  `json:"system" yaml:"system"`
	Captcha Captcha `json:"captcha" yaml:"captcha"`
}
