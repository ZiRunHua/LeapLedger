package database

import (
	"fmt"

	_ "github.com/ZiRunHua/LeapLedger/model"
	"github.com/pkg/errors"
)

import (
	userModel "github.com/ZiRunHua/LeapLedger/model/user"
	"github.com/ZiRunHua/LeapLedger/script"
	"github.com/ZiRunHua/LeapLedger/service"
	"github.com/ZiRunHua/LeapLedger/util"

	_templateService "github.com/ZiRunHua/LeapLedger/service/template"

	"context"

	"github.com/ZiRunHua/LeapLedger/global/cus"
	"github.com/ZiRunHua/LeapLedger/global/db"
	testInfo "github.com/ZiRunHua/LeapLedger/test/info"
	"gorm.io/gorm"
)

var (
	userService = service.GroupApp.UserServiceGroup

	commonService = service.GroupApp.CommonServiceGroup

	templateService = service.GroupApp.TemplateServiceGroup

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
	account, _, err := templateService.CreateExampleAccount(user, ctx)
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
	fmt.Println("test user", testInfo.Data)
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
	_, accountUser, err := templateService.CreateExampleAccount(user, ctx)
	if err != nil {
		return err
	}
	err = script.User.ChangeCurrantAccount(accountUser, tx)
	if err != nil {
		return err
	}
	return err
}
