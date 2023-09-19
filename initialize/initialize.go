package initialize

import (
	"KeepAccount/util"
	"fmt"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
	"os"
)

type _config struct {
	Redis   _redis
	Mysql   _mysql
	Logger  _logger
	System  _system
	Captcha _captcha
}

var (
	Config = &_config{
		Redis: _redis{}, Mysql: _mysql{}, Logger: _logger{}, System: _system{}, Captcha: _captcha{},
	}
	Cache         util.Cache
	Db            *gorm.DB
	RequestLogger *zap.Logger
	ErrorLogger   *zap.Logger
	PanicLogger   *zap.Logger
)

type initializer interface {
	do() error
}

func Do() {
	var err error
	if err = initConfig(); err != nil {
		print(fmt.Sprint("配置初始化失败 err: %v", err))
	}
	if err = Config.Logger.do(); err != nil {
		print("初始化logger错误 err: %v", err)
	}
	if err = Config.Mysql.do(); err != nil {
		print("初始化Mysql错误 err: %v", err)
	}
	if err = Config.Redis.do(); err != nil {
		print("初始化Redis错误 err: %v", err)
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
