package initialize

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"time"
)

type _mysql struct {
	Path     string
	Port     string
	Config   string
	Dbname   string
	Username string
	Password string
}

func (m *_mysql) Dsn() string {
	return m.Username + ":" + m.Password + "@tcp(" + m.Path + ":" + m.Port + ")/" + m.Dbname + "?" + m.Config
}
func (m *_mysql) do() error {
	mysqlConfig := mysql.Config{
		DSN:                       m.Dsn(), // DSN data source name
		DefaultStringSize:         191,     // string 类型字段的默认长度
		SkipInitializeWithVersion: false,   //
	}
	db, err := gorm.Open(mysql.New(mysqlConfig), m.GormConfig())
	if err != nil {
		return err
	}
	sqlDb, _ := db.DB()
	sqlDb.SetMaxIdleConns(25)
	sqlDb.SetMaxOpenConns(200)
	sqlDb.SetConnMaxLifetime(5 * time.Minute)
	db.InstanceSet("gorm:table_options", "ENGINE=InnoDB")
	Db = db
	return nil
}
func (m *_mysql) GormConfig() *gorm.Config {
	config := &gorm.Config{
		SkipDefaultTransaction: true,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   logger.Default.LogMode(logger.Info),
	}
	return config
}
