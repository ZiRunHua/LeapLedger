package categoryModel

import "KeepAccount/global"

func init() {
	tables := []interface{}{
		&Category{}, &Mapping{},
		&Father{},
	}
	err := global.GvaDb.AutoMigrate(tables...)
	if err != nil {
		panic(err)
	}
}
