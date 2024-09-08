package categoryModel

import (
	"KeepAccount/global/db"
)

func init() {
	tables := []interface{}{
		Category{}, Mapping{}, Father{},
	}
	err := db.InitDb.AutoMigrate(tables...)
	if err != nil {
		panic(err)
	}
}
