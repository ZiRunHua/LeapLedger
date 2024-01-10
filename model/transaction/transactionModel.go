package transactionModel

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	accountModel "KeepAccount/model/account"
	categoryModel "KeepAccount/model/category"
	commonModel "KeepAccount/model/common"
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

func NewTransaction(db *gorm.DB) *Transaction {
	t := &Transaction{}
	t.SetTx(db)
	return t
}
func (t *Transaction) IsEmpty() bool {
	return t.ID == 0
}

func (t *Transaction) SelectById(id uint, forUpdate bool) error {
	return commonModel.SelectByIdOfModel(t, id, forUpdate)
}

func (t *Transaction) Exits(query interface{}, args ...interface{}) (bool, error) {
	return commonModel.ExistOfModel(t, query, args...)
}

func (t *Transaction) Update() error {
	return t.GetDb().Updates(t).Error
}

func (t *Transaction) CreateOne(transaction *Transaction) error {
	return t.GetDb().Create(transaction).Error
}

func (t *Transaction) GetCategory() (*categoryModel.Category, error) {
	var category categoryModel.Category
	err := global.GvaDb.Model(&category).First(&category, t.CategoryId).Error
	return &category, err
}

func (t *Transaction) GetUser() (*userModel.User, error) {
	var user userModel.User
	err := global.GvaDb.Model(&user).First(&user, t.UserId).Error
	return &user, err
}

func (t *Transaction) GetAccount() (*accountModel.Account, error) {
	var account accountModel.Account
	err := global.GvaDb.Model(&account).First(&account, t.AccountId).Error
	return &account, err
}
