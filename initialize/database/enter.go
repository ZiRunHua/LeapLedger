package database

import (
	_ "KeepAccount/model"
)

import (
	userModel "KeepAccount/model/user"
	"KeepAccount/script"
	"KeepAccount/service"
	"KeepAccount/util"

	_templateService "KeepAccount/service/template"
	"errors"

	"KeepAccount/global/cus"
	"KeepAccount/global/db"
	testInfo "KeepAccount/test/info"
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	userService = service.GroupApp.UserServiceGroup

	commonService = service.GroupApp.CommonServiceGroup
)
var (
	templateService = _templateService.Group{}
)

const (
	testUserPassword = _templateService.TmplUserPassword
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
	err = userService.UpdatePassword(user, util.ClientPasswordHash(user.Email, testUserPassword), ctx)
	if err != nil {
		return
	}
	account, _, err := script.Account.CreateExample(user, ctx)
	if err != nil {
		return
	}
	token, err := commonService.GenerateJWT(commonService.MakeCustomClaims(user.ID))
	if err != nil {
		return
	}
	testInfo.Data = testInfo.Info{
		UserId:    user.ID,
		Email:     user.Email,
		AccountId: account.ID,
		Token:     token,
	}
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
