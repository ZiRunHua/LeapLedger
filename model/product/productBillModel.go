package productModel

import (
	"github.com/ZiRunHua/LeapLedger/global/constant"
)

type Bill struct {
	ProductKey Key `gorm:"primary_key;"`
	Encoding   constant.Encoding
	StartRow   int
	DateFormat string `gorm:"default:2006-01-02 15:04:05;"`
}

func (b *Bill) TableName() string {
	return "product_bill"
}
