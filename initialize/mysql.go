package initialize

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"time"
)

type _mysql struct {
	Path     string `yaml:"Path"`
	Port     string `yaml:"Port"`
	Config   string `yaml:"Config"`
	DbName   string `yaml:"DbName"`
	Username string `yaml:"Username"`
	Password string `yaml:"Password"`
}

func (m *_mysql) Dsn() string {
	return m.Username + ":" + m.Password + "@tcp(" + m.Path + ":" + m.Port + ")/" + m.DbName + "?" + m.Config
}

func (m *_mysql) do() error {
	var err error
	mysqlConfig := mysql.Config{
		DSN:                       m.Dsn(), // DSN data source name
		DefaultStringSize:         191,     // string 类型字段的默认长度
		SkipInitializeWithVersion: false,   //
	}
	db, err := m.connect(mysqlConfig, 0)
	if err != nil {
		return err
	}
	sqlDb, _ := db.DB()
	sqlDb.SetMaxIdleConns(25)
	sqlDb.SetMaxOpenConns(200)
	sqlDb.SetConnMaxLifetime(5 * time.Minute)
	db.InstanceSet("gorm:table_options", "ENGINE=InnoDB")
	db.InstanceSet("gorm:queryFields", "SET TRANSACTION ISOLATION LEVEL READ COMMITTED;")
	Db = db
	return nil
}

func (m *_mysql) connect(mysqlConfig mysql.Config, retryTimes int) (db *gorm.DB, err error) {
	defer func() {
		if err != nil && retryTimes < 3 {
			time.Sleep(time.Second * 3)
			db, err = m.connect(mysqlConfig, retryTimes+1)
		}
	}()
	db, err = gorm.Open(mysql.New(mysqlConfig), m.GormConfig())
	return
}

func (m *_mysql) GormConfig() *gorm.Config {
	config := &gorm.Config{
		SkipDefaultTransaction: true,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		DisableForeignKeyConstraintWhenMigrating: true,
		TranslateError:                           true,
		Logger:                                   logger.Default.LogMode(logger.Info),
	}
	return config
}
