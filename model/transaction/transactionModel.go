package transactionModel

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	accountModel "KeepAccount/model/account"
	categoryModel "KeepAccount/model/category"
	commonModel "KeepAccount/model/common"
	queryFunc "KeepAccount/model/common/query"
	userModel "KeepAccount/model/user"
	"gorm.io/gorm"
	"time"
)

type Transaction struct {
	gorm.Model
	UserId        uint `gorm:"column:user_id"`
	AccountId     uint `gorm:"column:account_id"`
	CategoryId    uint `gorm:"column:category_id"`
	IncomeExpense constant.IncomeExpense
	Amount        int
	Remark        string
	TradeTime     time.Time
	commonModel.BaseModel
}

func (t *Transaction) IsEmpty() bool {
	return t.ID == 0
}

func (t *Transaction) SelectById(id uint) error {
	return global.GvaDb.First(t, id).Error
}

func (t *Transaction) Exits(query interface{}, args ...interface{}) (bool, error) {
	return queryFunc.Exist[*Transaction](query, args)
}

func (t *Transaction) GetCategory() (category categoryModel.Category, err error) {
	err = global.GvaDb.First(&category, t.CategoryId).Error
	return
}

func (t *Transaction) GetUser(selects ...interface{}) (user userModel.User, err error) {
	if len(selects) > 0 {
		err = global.GvaDb.Select(selects[0], selects[1:]...).First(&user, t.UserId).Error
	} else {
		err = global.GvaDb.First(&user, t.UserId).Error
	}
	return
}

func (t *Transaction) GetAccount() (accountModel.Account, error) {
	var account accountModel.Account
	err := global.GvaDb.Model(&account).First(&account, t.AccountId).Error
	return account, err
}
