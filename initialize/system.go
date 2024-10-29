package initialize

type _system struct {
	Addr         int    `yaml:"Addr"`
	RouterPrefix string `yaml:"RouterPrefix"`
	LockMode     string `yaml:"LockMode"`

	JwtKey        string `yaml:"JwtKey"`
	ClientSignKey string `yaml:"ClientSignKey"`
}
