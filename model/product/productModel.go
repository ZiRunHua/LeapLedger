package productModel

import (
	"KeepAccount/global"
	commonModel "KeepAccount/model/common"
)

type Product struct {
	Key    Key    `gorm:"primary_key"`
	Name   string `gorm:"comment:'名称'"`
	Hide   uint8  `gorm:"default:0;comment:'隐藏标识'"`
	Weight int    `gorm:"default:0;comment:'权重'"`
	commonModel.BaseModel
}

type Key string

const AliPay, WeChatPay Key = "AliPay", "WeChatPay"

func (p *Product) TableName() string {
	return "product"
}

func (p *Product) IsEmpty() bool {
	return p == nil || p.Key == ""
}

func (p *Product) SelectByKey(key Key) (result Product, err error) {
	err = global.GvaDb.Where("key = ?", key).First(&result).Error
	return
}

func (p *Product) GetBill() (*Bill, error) {
	bill := &Bill{}
	return bill.SelectByPrimaryKey(string(p.Key))
}
