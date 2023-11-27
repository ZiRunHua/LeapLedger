package accountModel

import (
	commonModel "KeepAccount/model/common"
	"KeepAccount/model/common/query"
	userModel "KeepAccount/model/user"
	"gorm.io/gorm"
)

type Account struct {
	gorm.Model
	UserId uint   `gorm:"comment:用户id;not null"`
	Name   string `gorm:"comment:名称;size:128"`
	Icon   string `gorm:"comment:图标;size:64"`
	commonModel.BaseModel
}

func (a *Account) GetUser() (*userModel.User, error) {
	return query.FirstByPrimaryKey[*userModel.User](a.UserId)
}

func (a *Account) IsEmpty() bool {
	return a.ID == 0
}

func (a *Account) SelectByPrimaryKey(id uint) (*Account, error) {
	return query.FirstByPrimaryKey[*Account](id)
}

func (a *Account) Exits(query interface{}, args ...interface{}) (bool, error) {
	return commonModel.ExistOfModel(a, query, args)
}

func (a *Account) CheckBelongTo(user userModel.User) bool {
	return a.UserId == user.ID
}

func (a *Account) CreateOne() error {
	return a.GetDb().Create(a).Error
}
