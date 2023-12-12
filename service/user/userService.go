package userService

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	accountModel "KeepAccount/model/account"
	"KeepAccount/model/common/query"
	userModel "KeepAccount/model/user"
	commonService "KeepAccount/service/common"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"time"
)

type User struct{}

func (u *User) Login(email string, password string, clientType constant.Client, tx *gorm.DB) (
	user *userModel.User, clientBaseInfo *userModel.UserClientBaseInfo, token string, err error,
) {
	password = commonService.Common.HashPassword(email, password)
	err = global.GvaDb.Where("email = ? And password = ?", email, password).First(&user).Error
	if err != nil {
		return
	}
	userClient, err := userModel.GetUserClientModel(clientType)
	if err != nil {
		return
	}
	err = userClient.GetByUser(user)
	if err != nil {
		return
	}
	clientBaseInfo = userModel.GetUserClientBaseInfo(userClient)
	customClaims := commonService.Common.MakeCustomClaims(clientBaseInfo.UserID)
	token, err = commonService.Common.GenerateJWT(customClaims)
	if err != nil {
		return
	}
	err = u.updateDataAfterLogin(user, userClient, tx)
	if err != nil {
		return
	}
	return
}

func (u *User) updateDataAfterLogin(user *userModel.User, client userModel.Client, tx *gorm.DB) error {
	var err error
	client.SetTx(tx)
	clientHandlerFunc := func(db *gorm.DB) error {
		err = db.Update("login_time", time.Now()).Error
		if err != nil {
			return err
		}
		return nil
	}
	err = userModel.HandleUserClient(client, clientHandlerFunc)
	if err != nil {
		return err
	}
	_, err = u.RecordAction(user, constant.Login, tx)
	if err != nil {
		return err
	}
	return nil
}

func (userSvc *User) Register(addData *userModel.AddData, tx *gorm.DB) (*userModel.User, error) {
	addData.Password = commonService.Common.HashPassword(addData.Email, addData.Password)
	exist, err := query.Exist[*userModel.User]("email = ?", addData.Email)
	if err != nil {
		return nil, err
	} else if exist {
		return nil, errors.New("该邮箱已注册")
	}
	userDao := userModel.Dao.NewUser(tx)
	user, err := userDao.Add(addData)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, errors.New("该邮箱已注册")
		}
		return nil, err
	}
	for _, client := range userModel.GetClients() {
		if err = client.InitByUser(user, tx); err != nil {
			return nil, err
		}
	}
	_, err = userSvc.RecordAction(user, constant.Register, tx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (userSvc *User) UpdatePassword(user *userModel.User, newPassword string, tx *gorm.DB) error {
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
	user *userModel.User, client constant.Client, account *accountModel.Account,
) error {
	if user.ID != account.UserId {
		return errors.Wrap(global.ErrInvalidParameter, "userService SetClientAccount")
	}
	userClient, err := userModel.GetUserClientModel(client)
	if err != nil {
		return errors.Wrap(err, "")
	}
	if err = userClient.GetByUser(user); err != nil {
		return errors.Wrap(err, "userClient.GetByUser")
	}

	if err = userModel.HandleUserClient(
		userClient, func(db *gorm.DB) error {
			err = db.Update("current_account_id", account.ID).Error
			if err != nil {
				return errors.Wrap(err, "update userClient:current_account_id")
			}
			return nil
		},
	); err != nil {
		return err
	}
	return nil
}

func (userSvc *User) RecordAction(user *userModel.User, action constant.UserAction, tx *gorm.DB) (
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
	user *userModel.User, action constant.UserAction, remark string, tx *gorm.DB,
) (*userModel.Log, error) {
	dao := userModel.NewLogDao(tx)
	log, err := dao.Add(user, &userModel.LogAddData{Action: action, Remark: remark})
	if err != nil {
		return nil, err
	}
	return log, err
}
