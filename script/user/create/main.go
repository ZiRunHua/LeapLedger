package main

import (
	userModel "KeepAccount/model/user"
	userService "KeepAccount/service/user"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func main() {
	create()
}
func create() {
	email := "share_account_child@gmail.com"
	password := "1999123456"
	username := "child"
	bytes := []byte(email + password)
	hash := sha256.Sum256(bytes)
	password = hex.EncodeToString(hash[:])
	addData := userModel.AddData{
		Email:    email,
		Password: password,
		Username: username,
	}
	user, err := userService.GroupApp.Register(addData, context.Background())
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
