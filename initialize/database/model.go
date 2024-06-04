package database

import (
	"KeepAccount/global"
	userModel "KeepAccount/model/user"
	userService "KeepAccount/service/user"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"gorm.io/gorm"
)

type table interface {
	initData()
}

func initUser() {
	email := "template@gmail.com"
	password := "1999123456"
	username := "template"
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

func createUser(email, password, username string) {

}
