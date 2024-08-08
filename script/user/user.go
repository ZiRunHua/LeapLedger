package user

import (
	"KeepAccount/global"
	userModel "KeepAccount/model/user"
	userService "KeepAccount/service/user"
	"KeepAccount/util"
	"gorm.io/gorm"
)

// Create("template@gmail.com","1999123456","template")
func Create(email, password, username string) userModel.User {
	addData := userModel.AddData{
		Email:    email,
		Password: util.ClientPasswordHash(email, password),
		Username: username,
	}
	var user userModel.User
	err := global.GvaDb.Transaction(
		func(tx *gorm.DB) error {
			var err error
			user, err = userService.GroupApp.Register(addData, tx)
			return err
		},
	)
	if err != nil {
		panic(err)
	}
	return user
}
