package productModel

import (
	"github.com/ZiRunHua/LeapLedger/global/constant"
	commonModel "github.com/ZiRunHua/LeapLedger/model/common"
	queryFunc "github.com/ZiRunHua/LeapLedger/model/common/query"
)

type Bill struct {
	ProductKey Key `gorm:"primary_key;"`
	Encoding   constant.Encoding
	StartRow   int
	DateFormat string `gorm:"default:2006-01-02 15:04:05;"`
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
