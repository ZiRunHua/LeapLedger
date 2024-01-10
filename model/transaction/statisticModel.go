package transactionModel

import (
	commonModel "KeepAccount/model/common"
	"time"
)

type Statistic struct {
	Date       time.Time `gorm:"column:date;primaryKey" json:"date"`
	CategoryId uint      `gorm:"column:category_id;primaryKey" json:"category_id"`
	AccountId  uint      `gorm:"column:account_id" json:"account_id"` //冗余字段
	Amount     int       `gorm:"column:amount" json:"amount"`
	Count      int       `gorm:"column:count"`
	commonModel.BaseModel
}
