package database

import (
	"KeepAccount/global/cusCtx"
	"KeepAccount/global/db"
	_ "KeepAccount/model"
	"context"
	"gorm.io/gorm/logger"
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
	err = db.Transaction(context.Background(), initTemplateUser)
	if err != nil {
		panic(err)
	}
	// init tourist User
	err = db.Transaction(context.Background(), initTourist)
	if err != nil {
		panic(err)
	}
	// init test User
	if global.Config.Mode == constant.Debug {
		err = db.Transaction(context.Background(), initTestUser)
		if err != nil {
			panic(err)
		}
	}
}

func initTemplateUser(ctx *cusCtx.TxContext) (err error) {
	tx := ctx.GetDb()
	tx = tx.Session(&gorm.Session{Logger: tx.Logger.LogMode(logger.Silent)})
	var user userModel.User
	// find user
	err = tx.First(&user, tmplUserId).Error
	if err == nil {
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return
	}
	// create user
	user, err = script.User.Create(tmplUserEmail, tmplUserPassword, tmplUserName, tx)
	if err != nil {
		return
	}
	if user.ID != tmplUserId {
		tmplUserId = user.ID

	}
	// create account
	_, _, err = script.Account.CreateExample(user, tx)
	if err != nil {
		return
	}
	_templateService.SetTmplUser(user)
	return
}

func initTestUser(ctx *cusCtx.TxContext) (err error) {
	tx := ctx.GetDb()
	tx = tx.Session(&gorm.Session{Logger: tx.Logger.LogMode(logger.Silent)})
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
	global.TestUserInfo = fmt.Sprintf("test user:\nemail:%s password:%s", user.Email, tmplUserPassword)
	return
}

func initTourist(ctx *cusCtx.TxContext) error {
	tx := ctx.GetDb()
	tx = tx.Session(&gorm.Session{Logger: tx.Logger.LogMode(logger.Silent)})
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
