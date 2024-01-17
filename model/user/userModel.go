package userModel

import (
	"KeepAccount/global"
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

func (u *User) TableName() string {
	return "user"
}

func (u *User) IsEmpty() bool {
	return u.ID == 0
}

func (u *User) SelectById(id uint) error {
	return global.GvaDb.First(&u, id).Error
}

func (u *User) GetTransactionShareConfig() (*TransactionShareConfig, error) {
	data := &TransactionShareConfig{}
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
