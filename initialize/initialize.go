package initialize

import (
	"context"
	"os"
	"time"

	"KeepAccount/global/constant"
	"KeepAccount/util"

	"github.com/go-co-op/gocron"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

type _config struct {
	Mode       constant.ServerMode `yaml:"Mode"`
	Redis      _redis              `yaml:"Redis"`
	Mysql      _mysql              `yaml:"Mysql"`
	Nats       _nats               `yaml:"Nats"`
	Scheduler  _scheduler          `yaml:"Scheduler"`
	Logger     _logger             `yaml:"Logger"`
	System     _system             `yaml:"System"`
	Captcha    _captcha            `yaml:"Captcha"`
	ThirdParty _thirdParty         `yaml:"ThirdParty"`
}

var (
	Config        *_config
	Cache         util.Cache
	Db            *gorm.DB
	Nats          *nats.Conn
	NatsServer    *server.Server
	Scheduler     *gocron.Scheduler
	RequestLogger *zap.Logger
	ErrorLogger   *zap.Logger
	PanicLogger   *zap.Logger
	NatsLogger    *zap.Logger
	CronLogger    *zap.Logger
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
	group.Go(Config.Nats.do)
	group.Go(Config.Scheduler.do)
	if err = group.Wait(); err != nil {
		panic(err)
	}
}

const _configDirectoryPath = constant.WORK_PATH

func initConfig() error {
	configFileName := os.Getenv("CONFIG_FILE_NAME")
	if len(configFileName) == 0 {
		configFileName = "config.yaml"
	}
	configPath := _configDirectoryPath + "/" + configFileName
	yamlFile, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yamlFile, Config)
	if err != nil {
		return err
	}
	return setConfigDefault()
}
func setConfigDefault() error {
	if Config.System.LockMode == "" {
		Config.System.LockMode = "redis"
	}
	return nil
}
func reconnection[T any](connect func() (T, error), retryTimes int) (result T, err error) {
	defer func() {
		if err != nil && retryTimes > 0 {
			time.Sleep(time.Second * 3)
			result, err = reconnection[T](connect, retryTimes-1)
		}
	}()
	result, err = connect()
	return
}
