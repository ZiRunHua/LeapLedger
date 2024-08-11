package userService

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	"KeepAccount/global/db"
	globalTask "KeepAccount/global/task"
	accountModel "KeepAccount/model/account"
	"KeepAccount/model/common/query"
	userModel "KeepAccount/model/user"
	commonService "KeepAccount/service/common"
	"KeepAccount/util"
	"KeepAccount/util/rand"
	"context"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"time"
)

type User struct{}

func (userSvc *User) Login(email string, password string, clientType constant.Client, tx *gorm.DB) (
	user userModel.User, clientBaseInfo userModel.UserClientBaseInfo, token string, customClaims util.CustomClaims, err error,
) {
	password = commonService.Common.HashPassword(email, password)
	err = global.GvaDb.Where("email = ? And password = ?", email, password).First(&user).Error
	if err != nil {
		return
	}
	clientBaseInfo, err = userModel.NewDao(tx).SelectUserClientBaseInfo(user.ID, clientType)
	if err != nil {
		return
	}
	customClaims = commonService.Common.MakeCustomClaims(clientBaseInfo.UserId)
	token, err = commonService.Common.GenerateJWT(customClaims)
	if err != nil {
		return
	}
	err = userSvc.updateDataAfterLogin(user, clientType, tx)
	if err != nil {
		return
	}
	return
}

func (userSvc *User) updateDataAfterLogin(user userModel.User, clientType constant.Client, tx *gorm.DB) error {
	err := tx.Model(userModel.GetUserClientModel(clientType)).Where("user_id = ?", user.ID).Update(
		"login_time", time.Now(),
	).Error
	if err != nil {
		return err
	}
	_, err = userSvc.RecordAction(user, constant.Login, tx)
	if err != nil {
		return err
	}
	return nil
}

type RegisterOption struct {
	Tour bool
}

func (ro *RegisterOption) WithTour(Tour bool) *RegisterOption {
	ro.Tour = Tour
	return ro
}

func (userSvc *User) NewRegisterOption() *RegisterOption { return &RegisterOption{} }

func (userSvc *User) Register(addData userModel.AddData, tx *gorm.DB, option ...RegisterOption) (user userModel.User, err error) {
	addData.Password = commonService.Common.HashPassword(addData.Email, addData.Password)
	exist, err := query.Exist[*userModel.User]("email = ?", addData.Email)
	if err != nil {
		return
	} else if exist {
		return user, errors.New("该邮箱已注册")
	}
	userDao := userModel.NewDao(tx)
	user, err = userDao.Add(addData)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return user, errors.New("该邮箱已注册")
		}
		return
	}
	for _, client := range userModel.GetClients() {
		if err = client.InitByUser(user, tx); err != nil {
			return
		}
	}
	_, err = userSvc.RecordAction(user, constant.Register, tx)

	if err != nil {
		return
	}
	if len(option) == 0 {
		return
	}
	if option[0].Tour {
		_, err = userDao.CreateTour(user)
		if err != nil {
			return
		}
		err = user.ModifyAsTourist(tx)
		if err != nil {
			return
		}
	}
	return
}

func (userSvc *User) UpdatePassword(user userModel.User, newPassword string, tx *gorm.DB) error {
	password := commonService.Common.HashPassword(user.Email, newPassword)
	logRemark := ""
	if password == user.Password {
		logRemark = global.ErrSameAsTheOldPassword.Error()
	}
	err := tx.Model(user).Update("password", password).Error
	if err != nil {
		return err
	}
	_, err = userSvc.RecordActionAndRemark(user, constant.UpdatePassword, logRemark, tx)
	if err != nil {
		return err
	}
	return nil
}

func (userSvc *User) UpdateInfo(user *userModel.User, username string, tx *gorm.DB) error {
	err := tx.Model(user).Update("username", username).Error
	if err != nil {
		return err
	}
	return nil
}

func (userSvc *User) SetClientAccount(
	accountUser accountModel.User, client constant.Client, account accountModel.Account, tx *gorm.DB,
) error {
	if accountUser.AccountId != account.ID {
		return global.ErrAccountId
	}
	err := tx.Model(userModel.GetUserClientModel(client)).Where("user_id = ?", accountUser.UserId).Update(
		"current_account_id", account.ID,
	).Error
	if err != nil {
		return errors.Wrap(err, "")
	}
	return nil
}

func (userSvc *User) SetClientShareAccount(
	accountUser accountModel.User, client constant.Client, account accountModel.Account, tx *gorm.DB,
) error {
	if accountUser.AccountId != account.ID {
		return global.ErrAccountId
	}
	err := tx.Model(userModel.GetUserClientModel(client)).Where(
		"user_id = ?", accountUser.UserId,
	).Update("current_share_account_id", account.ID).Error
	if err != nil {
		return errors.Wrap(err, "")
	}
	return nil
}

func (userSvc *User) RecordAction(user userModel.User, action constant.UserAction, tx *gorm.DB) (
	*userModel.Log, error,
) {
	dao := userModel.NewLogDao(tx)
	log, err := dao.Add(user, &userModel.LogAddData{Action: action})
	if err != nil {
		return nil, err
	}
	return log, err
}

func (userSvc *User) RecordActionAndRemark(
	user userModel.User, action constant.UserAction, remark string, tx *gorm.DB,
) (*userModel.Log, error) {
	dao := userModel.NewLogDao(tx)
	log, err := dao.Add(user, &userModel.LogAddData{Action: action, Remark: remark})
	if err != nil {
		return nil, err
	}
	return log, err
}

func (userSvc *User) EnableTourist(
	deviceNumber string, client constant.Client, tx *gorm.DB,
) (user userModel.User, err error) {
	if client != constant.Android && client != constant.Ios {
		return user, global.ErrDeviceNotSupported
	}
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
	err = tx.Model(userModel.GetUserClientModel(client)).Where("user_id = ?", user.ID).Update("device_number", deviceNumber).Error
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
	globalTask.Publish[any](globalTask.TaskCreateTourist, struct{}{})
	return
}

func (userSvc *User) CreateTourist(tx *gorm.DB) (user userModel.User, err error) {
	addData := userModel.AddData{"游玩家", rand.String(8), rand.String(8)}
	option := userSvc.NewRegisterOption().WithTour(true)
	return userSvc.Register(addData, tx, *option)
}

func (userSvc *User) ProcessAllClient(user userModel.User, processFunc func(userModel.Client) error, ctx context.Context) error {
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
