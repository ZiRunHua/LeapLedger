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

func (userSvc *User) Login(email string, password string, clientType constant.Client, tx *gorm.DB) (
	user userModel.User, clientBaseInfo userModel.UserClientBaseInfo, token string, err error,
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
	customClaims := commonService.Common.MakeCustomClaims(clientBaseInfo.UserId)
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

func (userSvc *User) Register(addData userModel.AddData, tx *gorm.DB) (user userModel.User, err error) {
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
