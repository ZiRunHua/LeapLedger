package accountModel

import (
	"KeepAccount/global"
	commonModel "KeepAccount/model/common"
	queryFunc "KeepAccount/model/common/query"
	userModel "KeepAccount/model/user"
	"KeepAccount/util"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Account struct {
	ID     uint   `gorm:"primarykey"`
	UserId uint   `gorm:"comment:用户id;not null"`
	Type   Type   `gorm:"default:independent;not null"`
	Name   string `gorm:"comment:名称;not null;size:128"`
	Icon   string `gorm:"comment:图标;not null;default:'payment';size:64"`
	gorm.Model
	commonModel.BaseModel
}

type AccountUpdateData struct {
	Name *string
	Icon *string
	Type Type
}

func (a *AccountUpdateData) getAccount() (result Account, err error) {
	if err = util.Data.CopyNotEmptyStringOptional(a.Name, &result.Name); err != nil {
		return result, err
	}
	if err = util.Data.CopyNotEmptyStringOptional(a.Icon, &result.Icon); err != nil {
		return result, err
	}
	return
}

type Type string

const (
	TypeIndependent Type = "independent"
	TypeShare       Type = "share"
)

func (t Type) IsIndependent() bool { return t == TypeIndependent }
func (t Type) IsShare() bool       { return t == TypeShare }

func (t Type) Handle(isIndependent, isShare func()) {
	switch t {
	case TypeIndependent:
		isIndependent()
	case TypeShare:
		isShare()
	default:
		panic("error account.Type")
	}
}

func (a *Account) GetUser(selects ...interface{}) (user userModel.User, err error) {
	err = user.SelectById(a.UserId, selects...)
	return
}

func (a *Account) ForUpdate(tx *gorm.DB) error {
	if err := tx.Model(a).Clauses(clause.Locking{Strength: "UPDATE"}).First(&a).Error; err != nil {
		return err
	}
	return nil
}

func (a *Account) ForShare(tx *gorm.DB) error {
	if err := tx.Model(a).Clauses(clause.Locking{Strength: "SHARE"}).First(&a).Error; err != nil {
		return err
	}
	return nil
}

func (a *Account) IsEmpty() bool {
	return a.ID == 0
}

func (a *Account) SelectById(id uint) error {
	return global.GvaDb.First(a, id).Error
}

func (a *Account) Exits(query interface{}, args ...interface{}) (bool, error) {
	return queryFunc.Exist[*Account](query, args)
}

func (a *Account) CheckBelongTo(user userModel.User) bool {
	return a.UserId == user.ID
}

type UserInvitationCondition struct {
	AccountId  *uint
	InviterId  *uint
	InviteeId  *uint
	Permission *UserPermission
	Limit      int
	Offset     int
}

func NewUserInvitationCondition(Limit, Offset int) *UserInvitationCondition {
	return &UserInvitationCondition{Limit: Limit, Offset: Offset}
}

func (c *UserInvitationCondition) SetAccountId(id uint) *UserInvitationCondition {
	c.AccountId = &id
	return c
}

func (c *UserInvitationCondition) SetInviterId(id uint) *UserInvitationCondition {
	c.InviterId = &id
	return c
}

func (c *UserInvitationCondition) SetInviteeId(id uint) *UserInvitationCondition {
	c.InviteeId = &id
	return c
}

func (c *UserInvitationCondition) SetPermission(up UserPermission) *UserInvitationCondition {
	c.Permission = &up
	return c
}
