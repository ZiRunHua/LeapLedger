package main

import (
	userModel "KeepAccount/model/user"
	userService "KeepAccount/service/user"
	"KeepAccount/util"
	"context"
	"fmt"
)

func main() {
	create()
}
func create() {
	email := "share_account_child@gmail.com"
	password := "1999123456"
	username := "child"
	addData := userModel.AddData{
		Email:    email,
		Password: util.ClientPasswordHash(email, password),
		Username: username,
	}
	user, err := userService.GroupApp.Register(addData, context.Background(),
		*userService.GroupApp.NewRegisterOption().WithSendEmail(false))
	if err != nil {
		panic(err)
	}
	fmt.Println(user.Email, user.Username, password)
}

func GetInput(tip string) (userInput string) {
	fmt.Println(tip)
	_, err := fmt.Scanln(&userInput)
	if err != nil {
		return ""
	}
	return
}
