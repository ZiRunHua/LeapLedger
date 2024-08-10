package database

import (
	_ "KeepAccount/model"
)

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	userModel "KeepAccount/model/user"
	"KeepAccount/script"
	"KeepAccount/service"
	"KeepAccount/util"
	"fmt"

	_templateService "KeepAccount/service/template"
	"errors"

	"gorm.io/gorm"
)

var (
	userService = service.GroupApp.UserServiceGroup
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
	// init test User
	if global.Config.Mode == constant.Debug {
		err = global.GvaDb.Transaction(initTestUser)
		if err != nil {
			panic(err)
		}
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
	_, _, err = script.Account.CreateExample(user, tx)
	if err != nil {
		return
	}
	_templateService.SetTmplUser(user)
	return
}

func initTestUser(tx *gorm.DB) (err error) {
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
	if err != nil {
		return
	}
	_, _, err = script.Account.CreateExample(user, tx)
	if err != nil {
		return
	}
	global.TestUserId = user.ID
	global.TestUserInfo = fmt.Sprintf("email:%s password:%s", user.Email, tmplUserPassword)
	return
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
