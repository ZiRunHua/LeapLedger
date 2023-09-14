package config

type Logger struct {
	Path   string `json:"path" yaml:"path"`     // 日志路径
	Level  string `json:"Level" yaml:"Level"`   // 等级
	Format string `json:"Format" yaml:"Format"` // 格式
}
