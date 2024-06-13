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
	userService        = service.GroupApp.UserServiceGroup.Base
	accountService     = service.GroupApp.AccountServiceGroup.Base
	categoryService    = service.GroupApp.CategoryServiceGroup.Category
	transactionService = service.GroupApp.TransactionServiceGroup.Transaction
	productService     = service.GroupApp.ProductServiceGroup.Product
	templateService    = service.GroupApp.TemplateService.Template
	//第三方服务
	thirdpartyService = service.GroupApp.ThirdpartyServiceGroup
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
	// init tourist
	err = global.GvaDb.Transaction(initTourist)
	if err != nil {
		panic(err)
	}
}

func initTemplateUser(tx *gorm.DB) (err error) {
	var user userModel.User
	defer func() {
		if err == nil {
			_templateService.SetTmplUser(user)
		}
	}()
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
	return
}

func initTourist(tx *gorm.DB) error {
	var user userModel.User
	var err error
	deferFunc := func() error {
		err = userService.UpdatePassword(user, util.ClientPasswordHash(user.Email, tmplUserPassword), tx)
		if err != nil {
			return err
		}
		fmt.Println("email:", user.Email, " password:", tmplUserPassword)
		return nil
	}
	defer func() {
		if err == nil {
			err = deferFunc()
		}
	}()

	tourist, err := userModel.NewDao(tx).SelectByUnusedTour()
	if err == nil {
		err = tourist.Use(tx)
		if err != nil {
			return err
		}
		user, err = tourist.GetUser(tx)
		if err != nil {
			return err
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	user, err = script.User.CreateTourist(tx)
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
