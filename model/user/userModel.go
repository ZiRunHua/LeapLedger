package userModel

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	commonModel "KeepAccount/model/common"
	"crypto/sha1"
	"encoding/hex"
	"gorm.io/gorm"
)

type User struct {
	Username string `gorm:"type:varchar(128);comment:'用户名'"`
	Password string `gorm:"type:varchar(64);comment:'密码'"`
	Email    string `gorm:"type:varchar(64);comment:'邮箱'"`
	gorm.Model
	commonModel.BaseModel
}

type UserInfo struct {
	ID       uint
	Username string
	Email    string
}

type userDataRetriever interface {
	UserInfo | User
}

func (u *User) TableName() string {
	return "user"
}

func (u *User) IsEmpty() bool {
	return u.ID == 0
}

func (u *User) SelectById(id uint, selects ...interface{}) error {
	query := global.GvaDb.Where("id = ?", id)
	if len(selects) > 0 {
		query = query.Select(selects[0], selects[1:]...)
	}
	return query.First(u).Error
}

func (u *User) GetUserClient(client constant.Client) (clientInfo UserClientBaseInfo, err error) {
	var clientModel Client
	clientModel = GetUserClientModel(client)
	if err != nil {
		return
	}
	err = clientModel.GetByUser(*u)
	if err != nil {
		return
	}
	clientInfo = *GetUserClientBaseInfo(clientModel)
	return
}

func (u *User) GetTransactionShareConfig() (TransactionShareConfig, error) {
	data := TransactionShareConfig{}
	return data, data.SelectByUserId(u.ID)
}

func (u *User) hashPassword() error {
	data := []byte(u.Username + u.Password)
	h := sha1.Sum(data)
	u.Password = hex.EncodeToString(h[:])
	return nil
}

func (u *User) updatePassword(newPassword string) {
	u.Password = newPassword
	u.hashPassword()
}
