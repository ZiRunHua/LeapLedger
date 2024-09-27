package script

import (
	accountModel "KeepAccount/model/account"
	userModel "KeepAccount/model/user"
	"KeepAccount/util"
	"KeepAccount/util/rand"
	"context"
	"gorm.io/gorm"
)

type _user struct {
}

var User = _user{}

// Create("template@gmail.com","1999123456","template")
func (u *_user) Create(email, password, username string, ctx context.Context) (userModel.User, error) {
	addData := userModel.AddData{
		Email:    email,
		Password: util.ClientPasswordHash(email, password),
		Username: username,
	}
	return userService.Register(addData, ctx, *userService.NewRegisterOption().WithSendEmail(false))
}

func (u *_user) CreateTourist(ctx context.Context) (user userModel.User, err error) {
	email := rand.String(12)
	password := rand.String(8)
	username := rand.String(8)
	addData := userModel.AddData{
		Email:    email,
		Password: util.ClientPasswordHash(email, password),
		Username: username,
	}
	option := userService.NewRegisterOption().WithTour(true)
	user, err = userService.Register(addData, ctx, *option)
	if err != nil {
		return
	}
	return
}

func (u *_user) ChangeCurrantAccount(accountUser accountModel.User, db *gorm.DB) (err error) {
	for _, client := range userModel.GetClients() {
		err = db.Model(&client).Where("user_id = ?", accountUser.UserId).Update("current_account_id", accountUser.AccountId).Error
		if err != nil {
			return
		}
	}
	return
}
