package categoryModel

import (
	"github.com/ZiRunHua/LeapLedger/global/db"
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
