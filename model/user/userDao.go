package userModel

import (
	"KeepAccount/global"
	"gorm.io/gorm"
)

type UserDao struct {
	db *gorm.DB
}

func (*dao) NewUser(db *gorm.DB) *UserDao {
	if db == nil {
		db = global.GvaDb
	}
	return &UserDao{db}
}

type AddData struct {
	Username string
	Password string
	Email    string
}

func (u *UserDao) Add(data *AddData) (*User, error) {
	user := &User{
		Username: data.Username,
		Password: data.Password,
		Email:    data.Email,
	}
	err := u.db.Create(user).Error
	return user, err
}

func (u *UserDao) SelectByEmail(email string) (*User, error) {
	user := &User{}
	err := u.db.Where("email = ?", email).First(user).Error
	return user, err
}
