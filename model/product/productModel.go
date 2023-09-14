package productModel

import (
	commonModel "KeepAccount/model/common"
	"KeepAccount/model/common/query"
)

type Product struct {
	Key    string `gorm:"primary_key;column:key"`
	Name   string `gorm:"column:name;comment:'名称'"`
	Hide   uint8  `gorm:"column:hide;default:0;comment:'隐藏标识'"`
	Weight int    `gorm:"column:weight;comment:'权重'"`
	commonModel.BaseModel
}

const index = "key"

func (p *Product) TableName() string {
	return "product"
}

func (p *Product) IsEmpty() bool {
	return p == nil || p.Key == ""
}

func (p *Product) SelectByPrimaryKey(key string) (*Product, error) {
	return query.FirstByField[*Product]("key", key)
}

func (p *Product) GetBill() (*Bill, error) {
	bill := &Bill{}
	return bill.SelectByPrimaryKey(p.Key)
}
