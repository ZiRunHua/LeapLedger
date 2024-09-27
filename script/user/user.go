package user

import (
	userModel "KeepAccount/model/user"
	userService "KeepAccount/service/user"
	"KeepAccount/util"
	"context"
)

// Create("template@gmail.com","1999123456","template")
func Create(email, password, username string) userModel.User {
	addData := userModel.AddData{
		Email:    email,
		Password: util.ClientPasswordHash(email, password),
		Username: username,
	}
	user, err := userService.GroupApp.Register(
		addData, context.Background(),
		*userService.GroupApp.NewRegisterOption().WithSendEmail(false))
	if err != nil {
		panic(err)
	}
	return user
}
