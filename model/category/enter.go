package categoryModel

import (
	"KeepAccount/global"
	"KeepAccount/util/gormFunc"
)

func CurrentInit() error {
	tables := []interface{}{
		Category{}, Mapping{},
		Father{},
	}
	err := global.GvaDb.AutoMigrate(tables...)
	if err != nil {
		return err
	}
	for _, table := range tables {
		_ = gormFunc.AlterIdToHeader(table, global.GvaDb)
	}
	return nil
}
