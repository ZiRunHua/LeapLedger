package accountModel

import (
	commonModel "KeepAccount/model/common"
	"KeepAccount/model/common/query"
	userModel "KeepAccount/model/user"
	"gorm.io/gorm"
)

type Account struct {
	gorm.Model
	UserId uint
	Name   string
	User   userModel.User
	commonModel.BaseModel
}

func (a *Account) IsEmpty() bool {
	return a.ID == 0
}
func (b *Account) SelectByPrimaryKey(id uint) (*Account, error) {
	return query.FirstByPrimaryKey[*Account](id)
}

func (c *Account) Exits(query interface{}, args ...interface{}) (bool, error) {
	return commonModel.ExistOfModel(c, query, args)
}

func (a *Account) CreateOne() error {
	return a.GetDb().Create(a).Error
}

func (a *Account) GetUser() (*userModel.User, error) {
	return query.FirstByPrimaryKey[*userModel.User](a.UserId)
}
