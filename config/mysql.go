package config

type Mysql struct {
	Path     string `json:"path" yaml:"path"`         // 服务器地址:端口
	Port     string `json:"port" yaml:"port"`         //:端口
	Config   string `json:"config" yaml:"config"`     // 高级配置
	Dbname   string `json:"db-name" yaml:"db-name"`   // 数据库名
	Username string `json:"username" yaml:"username"` // 数据库用户名
	Password string `json:"password" yaml:"password"` // 数据库密码
}

func (m *Mysql) Dsn() string {
	return m.Username + ":" + m.Password + "@tcp(" + m.Path + ":" + m.Port + ")/" + m.Dbname + "?" + m.Config
}
