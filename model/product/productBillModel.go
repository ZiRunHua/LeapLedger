package productModel

import (
	"KeepAccount/global"
	commonModel "KeepAccount/model/common"
	"KeepAccount/model/common/query"
)

type Bill struct {
	ProductKey string `gorm:"primary_key;column:product_key"`
	Encoding   global.Encoding
	StartRow   int
	DateFormat string
	commonModel.BaseModel
}

func (b *Bill) TableName() string {
	return "product_bill"
}

func (b *Bill) IsEmpty() bool {
	return b.ProductKey == ""
}

func (b *Bill) SelectByPrimaryKey(key string) (*Bill, error) {
	return query.FirstByField[*Bill]("product_key", key)
}

func (b *Bill) Exits(query interface{}, args ...interface{}) (bool, error) {
	return commonModel.ExistOfModel(b, query, args)
}
