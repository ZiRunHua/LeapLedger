package main

import (
	"KeepAccount/global"
	userModel "KeepAccount/model/user"
	userService "KeepAccount/service/user"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"gorm.io/gorm"
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
	var user userModel.User
	err := global.GvaDb.Transaction(
		func(tx *gorm.DB) error {
			var err error
			user, err = userService.GroupApp.Base.Register(addData, tx)
			return err
		},
	)
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
