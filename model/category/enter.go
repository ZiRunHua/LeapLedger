package categoryModel

import (
	"KeepAccount/util/gormFunc"
	"gorm.io/gorm"
)

func CurrentInit(db *gorm.DB) error {
	tables := []interface{}{
		Category{}, Mapping{}, Father{},
	}
	err := db.AutoMigrate(tables...)
	if err != nil {
		return err
	}
	for _, table := range tables {
		_ = gormFunc.AlterIdToHeader(table, db)
	}
	return nil
}
