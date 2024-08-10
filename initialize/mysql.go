package initialize

import (
	"time"

	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type _mysql struct {
	Path     string `yaml:"Path"`
	Port     string `yaml:"Port"`
	Config   string `yaml:"Config"`
	DbName   string `yaml:"DbName"`
	Username string `yaml:"Username"`
	Password string `yaml:"Password"`
}

func (m *_mysql) dsn() string {
	return m.Username + ":" + m.Password + "@tcp(" + m.Path + ":" + m.Port + ")/" + m.DbName + "?" + m.Config
}

func (m *_mysql) do() error {
	var err error
	mysqlConfig := mysql.Config{
		DSN:                       m.dsn(), // DSN data source name
		DefaultStringSize:         191,     // string 类型字段的默认长度
		SkipInitializeWithVersion: false,   //
	}
	var db *gorm.DB
	db, err = reconnection[*gorm.DB](
		func() (*gorm.DB, error) {
			return gorm.Open(mysql.New(mysqlConfig), m.gormConfig())
		}, 10)

	if err != nil {
		return err
	}
	sqlDb, _ := db.DB()
	sqlDb.SetMaxIdleConns(50)
	sqlDb.SetMaxOpenConns(50)
	sqlDb.SetConnMaxLifetime(5 * time.Minute)
	db.InstanceSet("gorm:table_options", "ENGINE=InnoDB")
	db.InstanceSet("gorm:queryFields", "SET TRANSACTION ISOLATION LEVEL READ COMMITTED;")
	Db = db
	return nil
}

func (m *_mysql) gormConfig() *gorm.Config {
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
