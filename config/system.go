package config

type System struct {
	Env          string `json:"env" yaml:"env"`   // 环境值 development/production
	Addr         int    `json:"addr" yaml:"addr"` // 端口值
	RouterPrefix string `json:"router-prefix" yaml:"router-prefix"`
}
