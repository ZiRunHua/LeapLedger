package templateService

import (
	"context"
	"errors"

	"KeepAccount/global/constant"
	"KeepAccount/global/cus"
	"KeepAccount/global/db"
	"KeepAccount/global/nats"
	userModel "KeepAccount/model/user"
	_accountService "KeepAccount/service/account"
	_categoryService "KeepAccount/service/category"
	_productService "KeepAccount/service/product"
	_userService "KeepAccount/service/user"
	"KeepAccount/util"
	_log "KeepAccount/util/log"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Group struct {
	template
}

var (
	GroupApp = &Group{}
	errorLog *zap.Logger
)

var TmplUserId uint = 1

const (
	TmplUserEmail    = "template@gmail.com"
	TmplUserPassword = "1999123456"
	TmplUserName     = "template"
)

var (
	templateService = GroupApp
	userService     = _userService.GroupApp
	accountService  = _accountService.GroupApp
	categoryService = _categoryService.GroupApp
	productService  = _productService.GroupApp
)

func init() {
	var err error
	if errorLog, err = _log.GetNewZapLogger(constant.LOG_PATH + "/service/template/error.log"); err != nil {
		panic(err)
	}

	nats.SubscribeTaskWithPayloadAndProcessInTransaction(
		nats.TaskCreateTourist, func(t []byte, ctx context.Context) error {
			user, err := userService.CreateTourist(ctx)
			if err != nil {
				return err
			}
			_, _, err = templateService.CreateExampleAccount(user, ctx)
			return err
		},
	)
	ctx := cus.WithDb(context.Background(), db.InitDb)
	// init template User
	err = db.Transaction(ctx, initTemplateUser)
	if err != nil {
		panic(err)
	}
	initRank()
}

func initTemplateUser(ctx *cus.TxContext) (err error) {
	tx := db.Get(ctx)
	var user userModel.User
	// find user
	err = tx.First(&user, TmplUserId).Error
	if err == nil {
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return
	}
	// create user
	option := userService.NewRegisterOption()
	option.WithSendEmail(false)
	user, err = userService.Register(
		userModel.AddData{
			Email:    TmplUserEmail,
			Password: util.ClientPasswordHash(TmplUserEmail, TmplUserPassword),
			Username: TmplUserName,
		}, ctx,
		*option,
	)
	if err != nil {
		return
	}
	if user.ID != TmplUserId {
		TmplUserId = user.ID
	}
	// create account
	_, _, err = templateService.CreateExampleAccount(user, ctx)
	if err != nil {
		return
	}
	SetTmplUser(user)
	return
}

func SetTmplUser(user userModel.User) {
	TmplUserId = user.ID
}
