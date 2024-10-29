package main

import (
	"context"
	"fmt"

	userModel "github.com/ZiRunHua/LeapLedger/model/user"
	userService "github.com/ZiRunHua/LeapLedger/service/user"
	"github.com/ZiRunHua/LeapLedger/util"
)

func main() {
	email := GetInput("email:")
	password := GetInput("password:")
	username := GetInput("username:")
	create(email, password, username)
}
func create(email, password, username string) {
	addData := userModel.AddData{
		Email:    email,
		Password: util.ClientPasswordHash(email, password),
		Username: username,
	}
	user, err := userService.GroupApp.Register(
		addData, context.Background(),
		*userService.GroupApp.NewRegisterOption().WithSendEmail(false),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println("create success:", user.Email, user.Username, password)
}

func GetInput(tip string) (userInput string) {
	fmt.Println(tip)
	_, err := fmt.Scanln(&userInput)
	if err != nil {
		return ""
	}
	return
}
