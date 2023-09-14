package config

type Redis struct {
	DB       int    `json:"db" yaml:"db"`             // redis的哪个数据库
	Addr     string `json:"addr" yaml:"addr"`         // 服务器地址:端口
	Password string `json:"password" yaml:"password"` // 密码
}
