package initialize

import (
	"KeepAccount/global/constant"
	"KeepAccount/util"
	"context"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
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
	NatsLogger    *zap.Logger
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
	group, _ := errgroup.WithContext(context.TODO())
	group.Go(Config.Logger.do)
	group.Go(Config.Mysql.do)
	group.Go(Config.Redis.do)
	group.Go(func() error { return Config.Nats.do(Config.Mode) })
	if err = group.Wait(); err != nil {
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
