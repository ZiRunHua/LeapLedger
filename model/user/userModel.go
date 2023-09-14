package userModel

import (
	"KeepAccount/global"
	"crypto/sha1"
	"encoding/hex"
	"time"
)

type User struct {
	ID         uint      `gorm:"primary_key;auto_increment;comment:'主键'"`
	Username   string    `gorm:"type:varchar(128);comment:'用户名'"`
	Password   string    `gorm:"type:varchar(64);comment:'密码'"`
	CreateTime time.Time `gorm:"type:datetime;default:current_timestamp;comment:'创建时间'"`
}

func (u *User) SelectById(id uint) error {
	return global.GvaDb.First(&u, id).Error
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
