package request

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/ZiRunHua/LeapLedger/global"
	"github.com/ZiRunHua/LeapLedger/global/constant"
	userModel "github.com/ZiRunHua/LeapLedger/model/user"
)

type UserLogin struct {
	Email    string `binding:"required"`
	Password string `binding:"required"`
	PicCaptcha
}

type UserRegister struct {
	Username string `binding:"required"`
	Password string `binding:"required"`
	Email    string `binding:"required,email"`
	Captcha  string `binding:"required"`
}

type UserForgetPassword struct {
	Email    string `binding:"required,email"`
	Password string `binding:"required"`
	Captcha  string `binding:"required"`
}

type UserUpdatePassword struct {
	Password string `binding:"required"`
	Captcha  string `binding:"required"`
}

type UserUpdateInfo struct {
	Username string `binding:"required"`
}

type UserSearch struct {
	Id       *uint  `binding:"omitempty"`
	Username string `binding:"required"`
	PageData
}

type UserSendEmail struct {
	PicCaptcha
	Type constant.UserAction `binding:"required,oneof=updatePassword"`
}

type UserHome struct {
	AccountId uint
}

type TourApply struct {
	DeviceNumber string
	Key          string
	Sign         string
}

func (t *TourApply) CheckSign() bool {
	h := hmac.New(sha256.New, []byte(global.Config.System.ClientSignKey))
	h.Write([]byte(t.DeviceNumber + t.Key))
	s := hex.EncodeToString(h.Sum(nil))
	return strings.Compare(t.Sign, s) == 0
}

type TransactionShareConfigName string

const (
	FLAG_ACCOUNT     TransactionShareConfigName = "account"
	FLAG_CREATE_TIME TransactionShareConfigName = "createTime"
	FLAG_UPDATE_TIME TransactionShareConfigName = "updateTime"
	FLAG_REMARK      TransactionShareConfigName = "remark"
)

type UserTransactionShareConfigUpdate struct {
	Flag   TransactionShareConfigName
	Status bool
}

func GetFlagByFlagName(name TransactionShareConfigName) (userModel.Flag, error) {
	switch name {
	case FLAG_ACCOUNT:
		return userModel.FLAG_ACCOUNT, nil
	case FLAG_CREATE_TIME:
		return userModel.FLAG_CREATE_TIME, nil
	case FLAG_UPDATE_TIME:
		return userModel.FLAG_UPDATE_TIME, nil
	case FLAG_REMARK:
		return userModel.FLAG_REMARK, nil
	}
	return 0, errors.New("flag参数错误")
}

type UserCreateFriendInvitation struct {
	Invitee uint
}

type UserGetFriendInvitation struct {
	IsInvite bool
}

type UserGetAccountInvitationList struct {
	PageData
}
