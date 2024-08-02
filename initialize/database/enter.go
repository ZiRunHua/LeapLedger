package database

import (
	"KeepAccount/global"
	_ "KeepAccount/model"
	userModel "KeepAccount/model/user"
	"KeepAccount/script"
	"KeepAccount/service"
	_templateService "KeepAccount/service/template"
	"KeepAccount/util"
	"errors"
	"fmt"
	"gorm.io/gorm"
)

var (
	userService = service.GroupApp.UserServiceGroup.Base
)
var tmplUserId = _templateService.TmplUserId

const (
	tmplUserEmail    = _templateService.TmplUserEmail
	tmplUserPassword = _templateService.TmplUserPassword
	tmplUserName     = _templateService.TmplUserName
)

func init() {
	var err error
	// init template User
	err = global.GvaDb.Transaction(initTemplateUser)
	if err != nil {
		panic(err)
	}
	// init tourist User
	err = global.GvaDb.Transaction(initTourist)
	if err != nil {
		panic(err)
	}
}

func initTemplateUser(tx *gorm.DB) (err error) {
	var user userModel.User
	//find user
	err = global.GvaDb.First(&user, tmplUserId).Error
	if err == nil {
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return
	}
	//create user
	user, err = script.User.Create(tmplUserEmail, tmplUserPassword, tmplUserName, tx)
	if err != nil {
		return
	}
	if user.ID != tmplUserId {
		tmplUserId = user.ID
		if err != nil {
			return
		}
	}
	//create account
	_, accountUser, err := script.Account.CreateExample(user, tx)
	if err != nil {
		return
	}
	//update user client
	err = script.User.ChangeCurrantAccount(accountUser, tx)
	if err != nil {
		return
	}
	_templateService.SetTmplUser(user)
	return
}
func NewTestUser() (result string) {
	var err error
	err = global.GvaDb.Transaction(func(tx *gorm.DB) error {
		result, err = newTestUser(tx)
		return err
	})
	if err != nil {
		result = err.Error()
	}
	return result
}

func newTestUser(tx *gorm.DB) (tip string, err error) {
	var user userModel.User
	user, err = script.User.CreateTourist(tx)
	if err != nil {
		return
	}
	var tourist userModel.Tour
	tourist, err = userModel.NewDao(tx).SelectTour(user.ID)
	if err != nil {
		return
	}
	err = tourist.Use(tx)
	if err != nil {
		return
	}
	err = userService.UpdatePassword(user, util.ClientPasswordHash(user.Email, tmplUserPassword), tx)
	return fmt.Sprintf("email:%s password:%s", user.Email, tmplUserPassword), err
}

func initTourist(tx *gorm.DB) error {
	_, err := userModel.NewDao(tx).SelectByUnusedTour()
	if err == nil {
		return nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	user, err := script.User.CreateTourist(tx)
	if err != nil {
		return err
	}
	_, accountUser, err := script.Account.CreateExample(user, tx)
	if err != nil {
		return err
	}
	err = script.User.ChangeCurrantAccount(accountUser, tx)
	if err != nil {
		return err
	}
	return err
}
