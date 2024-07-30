package userModel

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	commonModel "KeepAccount/model/common"
	"errors"
	"gorm.io/gorm"
	"strconv"
)

type User struct {
	Username string `gorm:"type:varchar(128);comment:'用户名'"`
	Password string `gorm:"type:char(64);comment:'密码'"`
	Email    string `gorm:"type:varchar(64);comment:'邮箱';unique"`
	gorm.Model
	commonModel.BaseModel
}

type UserInfo struct {
	ID       uint
	Username string
	Email    string
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
	err = clientModel.GetByUser(*u)
	if err != nil {
		return
	}
	clientInfo = *GetUserClientBaseInfo(clientModel)
	return
}

func (u *User) IsTourist(db *gorm.DB) (bool, error) {
	_, err := NewDao(db).SelectTour(u.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (u *User) ModifyAsTourist(db *gorm.DB) error {
	return db.Model(u).Updates(map[string]interface{}{
		"username": "游玩家",
		"email":    "player" + strconv.Itoa(int(u.ID)),
	}).Error
}

func (u *User) GetTransactionShareConfig() (TransactionShareConfig, error) {
	data := TransactionShareConfig{}
	return data, data.SelectByUserId(u.ID)
}

type Tour struct {
	UserId uint `gorm:"primary"`
	Status bool
	gorm.Model
	commonModel.BaseModel
}

func (u *Tour) TableName() string {
	return "user_tour"
}
func (t *Tour) GetUser(db *gorm.DB) (user User, err error) {
	err = db.First(&user, t.UserId).Error
	return user, err
}
func (t *Tour) Use(db *gorm.DB) error {
	if t.Status == true {
		return errors.New("tourist used")
	}
	return db.Model(t).Where("user_id = ?", t.UserId).Update("status", true).Error
}
