package userService

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	"KeepAccount/global/cus"
	"KeepAccount/global/db"
	"KeepAccount/global/nats"
	accountModel "KeepAccount/model/account"
	userModel "KeepAccount/model/user"
	commonService "KeepAccount/service/common"
	"KeepAccount/util/rand"
	"context"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"time"
)

type User struct{}

func (userSvc *User) Login(email string, password string, clientType constant.Client, ctx context.Context) (
	user userModel.User, clientBaseInfo userModel.UserClientBaseInfo, token string, customClaims jwt.RegisteredClaims,
	err error,
) {
	password = commonService.Common.HashPassword(email, password)
	err = global.GvaDb.Where("email = ? And password = ?", email, password).First(&user).Error
	if err != nil {
		return
	}
	clientBaseInfo, err = userModel.NewDao(db.Get(ctx)).SelectUserClientBaseInfo(user.ID, clientType)
	if err != nil {
		return
	}
	customClaims = commonService.Common.MakeCustomClaims(clientBaseInfo.UserId)
	token, err = commonService.Common.GenerateJWT(customClaims)
	if err != nil {
		return
	}
	err = userSvc.updateDataAfterLogin(user, clientType, ctx)
	if err != nil {
		return
	}
	return
}

func (userSvc *User) updateDataAfterLogin(user userModel.User, clientType constant.Client, ctx context.Context) error {
	err := db.Get(ctx).Model(userModel.GetUserClientModel(clientType)).Where("user_id = ?", user.ID).Update(
		"login_time", time.Now(),
	).Error
	if err != nil {
		return err
	}
	_, err = userSvc.RecordAction(user, constant.Login, ctx)
	if err != nil {
		return err
	}
	return nil
}

type RegisterOption struct {
	tour      bool
	sendEmail bool
}

func (ro *RegisterOption) WithTour(Tour bool) *RegisterOption {
	ro.tour = Tour
	ro.sendEmail = false
	return ro
}
func (ro *RegisterOption) WithSendEmail(value bool) *RegisterOption {
	ro.sendEmail = value
	return ro
}
func (userSvc *User) NewRegisterOption() *RegisterOption {
	return &RegisterOption{sendEmail: true}
}

func (userSvc *User) Register(addData userModel.AddData, ctx context.Context, option ...RegisterOption) (
	user userModel.User, err error,
) {
	return user, db.Transaction(ctx, func(ctx *cus.TxContext) (err error) {
		addData.Password = commonService.Common.HashPassword(addData.Email, addData.Password)
		tx := db.Get(ctx)
		userDao := userModel.NewDao(tx)
		err = userDao.CheckEmail(addData.Email)
		if err != nil {
			return err
		}
		user, err = userDao.Add(addData)
		if err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return errors.New("该邮箱已注册")
			}
			return
		}
		for _, client := range userModel.GetClients() {
			if err = client.InitByUser(user, tx); err != nil {
				return
			}
		}
		_, err = userSvc.RecordAction(user, constant.Register, ctx)
		if err != nil {
			return
		}
		var isTour bool
		if len(option) != 0 {
			isTour = option[0].tour
		}
		if isTour {
			_, err = userDao.CreateTour(user)
			if err != nil {
				return
			}
			err = user.ModifyAsTourist(tx)
			if err != nil {
				return
			}
		} else {
			return db.AddCommitCallback(ctx, func() {
				nats.PublishTaskWithPayload(nats.TaskSendNotificationEmail, nats.PayloadSendNotificationEmail{
					UserId: user.ID, Notification: constant.NotificationOfRegistrationSuccess,
				})
			})
		}
		return
	})
}

func (userSvc *User) UpdatePassword(user userModel.User, newPassword string, ctx context.Context) error {
	password := commonService.Common.HashPassword(user.Email, newPassword)
	logRemark := ""
	if password == user.Password {
		logRemark = global.ErrSameAsTheOldPassword.Error()
	}
	return db.Transaction(ctx, func(ctx *cus.TxContext) error {
		tx := ctx.GetDb()
		err := tx.Model(user).Update("password", password).Error
		if err != nil {
			return err
		}
		_, err = userSvc.RecordActionAndRemark(user, constant.UpdatePassword, logRemark, ctx)
		return db.AddCommitCallback(ctx, func() {
			nats.PublishTaskWithPayload(nats.TaskSendNotificationEmail, nats.PayloadSendNotificationEmail{
				UserId:       user.ID,
				Notification: constant.NotificationOfUpdatePassword,
			})
		})
	},
	)
}

func (userSvc *User) UpdateInfo(user userModel.User, username string, ctx context.Context) error {
	return db.Get(ctx).Model(&user).Update("username", username).Error
}

func (userSvc *User) SetClientAccount(
	accountUser accountModel.User, client constant.Client, account accountModel.Account, ctx context.Context,
) error {
	if accountUser.AccountId != account.ID {
		return global.ErrAccountId
	}
	err := db.Get(ctx).Model(userModel.GetUserClientModel(client)).Where("user_id = ?", accountUser.UserId).Update(
		"current_account_id", account.ID,
	).Error
	if err != nil {
		return errors.Wrap(err, "")
	}
	return nil
}

func (userSvc *User) SetClientShareAccount(
	accountUser accountModel.User, client constant.Client, account accountModel.Account, ctx context.Context,
) error {
	if accountUser.AccountId != account.ID {
		return global.ErrAccountId
	}
	err := db.Get(ctx).Model(userModel.GetUserClientModel(client)).Where(
		"user_id = ?", accountUser.UserId,
	).Update("current_share_account_id", account.ID).Error
	if err != nil {
		return errors.Wrap(err, "")
	}
	return nil
}

func (userSvc *User) RecordAction(user userModel.User, action constant.UserAction, ctx context.Context) (
	*userModel.Log, error,
) {
	dao := userModel.NewLogDao(db.Get(ctx))
	log, err := dao.Add(user, &userModel.LogAddData{Action: action})
	if err != nil {
		return nil, err
	}
	return log, err
}

func (userSvc *User) RecordActionAndRemark(
	user userModel.User, action constant.UserAction, remark string, ctx context.Context,
) (*userModel.Log, error) {
	dao := userModel.NewLogDao(db.Get(ctx))
	log, err := dao.Add(user, &userModel.LogAddData{Action: action, Remark: remark})
	if err != nil {
		return nil, err
	}
	return log, err
}

func (userSvc *User) EnableTourist(
	deviceNumber string, client constant.Client, ctx context.Context,
) (user userModel.User, err error) {
	if client != constant.Android && client != constant.Ios {
		return user, global.ErrDeviceNotSupported
	}
	return user, db.Transaction(
		ctx, func(ctx *cus.TxContext) (err error) {
			tx := ctx.GetDb()
			userDao := userModel.NewDao(tx)
			userInfo, err := userDao.SelectByDeviceNumber(client, deviceNumber)
			if err == nil {
				// 设备号已存
				user, err = userDao.SelectById(userInfo.UserId)
				if err != nil {
					return
				}
				return
			} else if false == errors.Is(err, gorm.ErrRecordNotFound) {
				return
			}

			userTour, err := userDao.SelectByUnusedTour()
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					err = errors.New("访问游客过多稍后再试")
					return
				}
				return
			}
			user, err = userTour.GetUser(tx)
			if err != nil {
				return
			}
			err = tx.Model(userModel.GetUserClientModel(client)).Where("user_id = ?", user.ID).Update(
				"device_number", deviceNumber,
			).Error
			if err != nil {
				if errors.Is(err, gorm.ErrDuplicatedKey) {
					err = global.ErrOperationTooFrequent
				}
				return
			}
			err = userTour.Use(tx)
			if err != nil {
				return
			}
			user, err = userTour.GetUser(tx)
			if err != nil {
				return
			}
			return db.AddCommitCallback(
				ctx, func() {
					nats.PublishTask(nats.TaskCreateTourist)
				},
			)
		},
	)
}

func (userSvc *User) CreateTourist(ctx context.Context) (user userModel.User, err error) {
	addData := userModel.AddData{"游玩家", rand.String(8), rand.String(8)}
	option := userSvc.NewRegisterOption().WithTour(true)
	return userSvc.Register(addData, ctx, *option)
}

func (userSvc *User) ProcessAllClient(
	user userModel.User, processFunc func(userModel.Client) error, ctx context.Context,
) error {
	tx := db.Get(ctx)
	var clientInfo userModel.Client
	var err error
	for _, client := range constant.ClientList {
		clientInfo, err = userModel.NewDao(tx).SelectUserClient(user.ID, client)
		if err != nil {
			return err
		}
		err = processFunc(clientInfo)
		if err != nil {
			return err
		}
	}
	return nil
}
