package productModel

import (
	"KeepAccount/global/constant"
	commonModel "KeepAccount/model/common"
	queryFunc "KeepAccount/model/common/query"
)

type Bill struct {
	ProductKey string `gorm:"primary_key;column:product_key"`
	Encoding   constant.Encoding
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
	return queryFunc.FirstByField[*Bill]("product_key", key)
}

func (b *Bill) Exits(query interface{}, args ...interface{}) (bool, error) {
	return queryFunc.Exist[*Bill](query, args)
}
