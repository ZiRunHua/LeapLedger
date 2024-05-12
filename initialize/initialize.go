package initialize

import (
	"KeepAccount/global/constant"
	"KeepAccount/util"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
	"os"
)

type _config struct {
	Mode       constant.ServerMode `yaml:"Mode"`
	Redis      _redis              `yaml:"Redis"`
	Mysql      _mysql              `yaml:"Mysql"`
	Nats       _nats               `yaml:"Nats"`
	Logger     _logger             `yaml:"Logger"`
	System     _system             `yaml:"System"`
	Captcha    _captcha            `yaml:"Captcha"`
	ThirdParty _thirdParty         `yaml:"ThirdParty"`
}

var (
	Config        *_config
	Cache         util.Cache
	Db            *gorm.DB
	RequestLogger *zap.Logger
	ErrorLogger   *zap.Logger
	PanicLogger   *zap.Logger
)

func init() {
	var err error
	Config = &_config{
		Redis: _redis{}, Mysql: _mysql{}, Logger: _logger{}, System: _system{}, Captcha: _captcha{}, Nats: _nats{},
		ThirdParty: _thirdParty{WeCom: _weCom{}},
	}
	if err = initConfig(); err != nil {
		panic(err)
	}
	if err = Config.Logger.do(); err != nil {
		panic(err)
	}
	if err = Config.Mysql.do(); err != nil {
		panic(err)
	}
	if err = Config.Redis.do(); err != nil {
		print("初始化Redis错误 err: %s", err)
	}
	if err = Config.Nats.do(Config.Mode); err != nil {
		panic(err)
	}
}

const _configDirectoryPath = ""

func initConfig() error {
	configFileName := os.Getenv("CONFIG_FILE_NAME")
	if len(configFileName) == 0 {
		configFileName = "config.yaml"
	}
	configPath := _configDirectoryPath + configFileName
	yamlFile, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yamlFile, Config)
	if err != nil {
		return err
	}
	return nil
}
