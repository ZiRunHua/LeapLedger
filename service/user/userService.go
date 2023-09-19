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
)

type User struct{}

func (u *User) Login(username string, password string, clientType constant.Client) (
	*accountModel.Account, string, error,
) {
	var user userModel.User
	password = commonService.Common.HashPassword(username, password)
	err := global.GvaDb.Where("username = ? And password = ?", username, password).First(&user).Error
	if err != nil {
		return nil, "", errors.New("账号或密码错误")
	}
	userClient, err := userModel.GetUserClientModel(clientType)
	if err != nil {
		return nil, "", err
	}
	err = userClient.GetByUser(&user)
	if err != nil {
		return nil, "", err
	}
	userClientInfo := userModel.GetUserClientBaseInfo(userClient)
	customClaims := commonService.Common.MakeCustomClaims(userClientInfo)
	token, err := commonService.Common.GenerateJWT(customClaims)
	if err != nil {
		return nil, "", err
	}
	account, err := query.FirstByPrimaryKey[*accountModel.Account](userClientInfo.CurrentAccountID)
	return account, token, err
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

type AddUserData struct {
	username string
	password string
}

func (userSvc *User) AddUser(addData *AddUserData, tx gorm.DB) error {

	addData.password = commonService.Common.HashPassword(addData.username, addData.password)

	if exist, err := query.Exist[*userModel.User]("username = ?", addData.username); err != nil {
		return err
	} else if exist {
		return errors.New("该用户已存在")
	}
	user := &userModel.User{
		Username: addData.username,
		Password: addData.password,
	}
	return tx.Create(&user).Error
}
