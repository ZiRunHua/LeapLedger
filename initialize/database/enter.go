package database

import (
	_ "KeepAccount/model"
	"github.com/pkg/errors"
)

import (
	userModel "KeepAccount/model/user"
	"KeepAccount/script"
	"KeepAccount/service"
	"KeepAccount/util"

	_templateService "KeepAccount/service/template"

	"context"

	"KeepAccount/global/cus"
	"KeepAccount/global/db"
	testInfo "KeepAccount/test/info"
	"gorm.io/gorm"
)

var (
	userService = service.GroupApp.UserServiceGroup

	commonService = service.GroupApp.CommonServiceGroup
)

const (
	testUserPassword = _templateService.TmplUserPassword
)

func init() {
	var err error
	ctx := cus.WithDb(context.Background(), db.InitDb)
	// init tourist User
	err = db.Transaction(ctx, initTourist)
	if err != nil {
		panic(err)
	}
	// init test User
	err = db.Transaction(ctx, initTestUser)
	if err != nil {
		panic(err)
	}
}

func initTestUser(ctx *cus.TxContext) (err error) {
	tx := db.Get(ctx)
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
