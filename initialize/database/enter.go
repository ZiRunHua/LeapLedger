package database

import (
	"KeepAccount/global/db"
	"KeepAccount/service"
)

// The starting point of database data initialization
// Trigger the init method by introducing "KeepAccount/model"
import (
	"KeepAccount/global"
	"KeepAccount/global/cus"
	_ "KeepAccount/model"
	userModel "KeepAccount/model/user"
	"KeepAccount/script"
	_templateService "KeepAccount/service/template"
	"KeepAccount/util"
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	userService = service.GroupApp.UserServiceGroup
)
var tmplUserId = _templateService.TmplUserId
var (
	templateService = _templateService.Group{}
)

const (
	tmplUserEmail    = _templateService.TmplUserEmail
	tmplUserPassword = _templateService.TmplUserPassword
	tmplUserName     = _templateService.TmplUserName
)

func init() {
	var err error
	// init tourist User
	err = db.Transaction(context.Background(), initTourist)
	if err != nil {
		panic(err)
	}
	// init test User
	err = db.Transaction(context.Background(), initTestUser)
	if err != nil {
		panic(err)
	}
}

func initTestUser(ctx *cus.TxContext) (err error) {
	tx := db.Get(ctx)
	tx = tx.Session(&gorm.Session{Logger: tx.Logger.LogMode(logger.Silent)})
	ctx = cus.WithTx(ctx, tx)
	tx = tx.Session(&gorm.Session{Logger: tx.Logger.LogMode(logger.Silent)})
	var user userModel.User
	user, err = script.User.CreateTourist(ctx)
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
	err = userService.UpdatePassword(user, util.ClientPasswordHash(user.Email, tmplUserPassword), ctx)
	if err != nil {
		return
	}
	_, _, err = script.Account.CreateExample(user, ctx)
	if err != nil {
		return
	}
	global.TestUserId = user.ID
	global.TestUserInfo = fmt.Sprintf("test user:\nemail:%s password:%s", user.Email, tmplUserPassword)
	return
}

func initTourist(ctx *cus.TxContext) error {
	tx := db.Get(ctx)
	tx = tx.Session(&gorm.Session{Logger: tx.Logger.LogMode(logger.Silent)})
	ctx = cus.WithTx(ctx, tx)
	tx = tx.Session(&gorm.Session{Logger: tx.Logger.LogMode(logger.Silent)})
	_, err := userModel.NewDao(tx).SelectByUnusedTour()
	if err == nil {
		return nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	user, err := script.User.CreateTourist(ctx)
	if err != nil {
		return err
	}
	_, accountUser, err := script.Account.CreateExample(user, ctx)
	if err != nil {
		return err
	}
	err = script.User.ChangeCurrantAccount(accountUser, tx)
	if err != nil {
		return err
	}
	return err
}
