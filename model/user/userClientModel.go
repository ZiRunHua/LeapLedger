package userModel

import (
	"KeepAccount/global"
	commonModel "KeepAccount/model/common"
	"errors"
	"gorm.io/gorm"
	"strings"
	"time"
)

type ClientMap map[global.Client]Client

func GetClients() ClientMap {
	return ClientMap{
		global.Web:     new(UserClientWeb),
		global.Android: new(UserClientAndroid),
		global.Ios:     new(UserClientIos),
	}
}

type Client interface {
	GetByUser(*User) error
	CheckUserAgent(userAgent string) bool
}

type UserClientBaseInfo struct {
	UserID           uint `gorm:"primaryKey"`
	CurrentAccountID uint
	LoginTime        time.Time
}

func (u *UserClientBaseInfo) GetByUser(user *User) error {
	panic("implement me")
}

func (u *UserClientBaseInfo) CheckUserAgent(userAgent string) bool {
	panic("implement me")
}

type UserClientWeb struct {
	UserClientBaseInfo
	WebName string
	commonModel.BaseModel
}
type UserClientAndroid struct {
	UserClientBaseInfo
	Version string
	commonModel.BaseModel
}
type UserClientIos struct {
	UserClientBaseInfo
	Version string
	commonModel.BaseModel
}

func (w *UserClientWeb) GetByUser(user *User) error {
	return global.GvaDb.Where("user_id = ?", user.ID).First(&w).Error
}
func (a *UserClientAndroid) GetByUser(user *User) error {
	return global.GvaDb.Where("user_id = ?", user.ID).First(&a).Error
}
func (i *UserClientIos) GetByUser(user *User) error {
	return global.GvaDb.Where("user_id = ?", user.ID).First(&i).Error
}

func (w *UserClientWeb) CheckUserAgent(userAgent string) bool {
	return strings.Contains(userAgent, "web")
}
func (a *UserClientAndroid) CheckUserAgent(userAgent string) bool {
	return strings.Contains(userAgent, "android")
}
func (i *UserClientIos) CheckUserAgent(userAgent string) bool {
	return strings.Contains(userAgent, "iPhone") || strings.Contains(userAgent, "iPad")
}
func GetUserClientModel(client global.Client) (Client, error) {
	switch client {
	case global.Web:
		return new(UserClientWeb), nil
	case global.Android:
		return new(UserClientAndroid), nil
	case global.Ios:
		return new(UserClientIos), nil
	default:
		return nil, errors.New("unknown client")
	}
}

var ErrClientNotFound = errors.New("client not found")

func GetUserClientBaseInfo(client Client) *UserClientBaseInfo {
	switch client.(type) {
	case *UserClientWeb:
		clientWeb := client.(*UserClientWeb)
		return &UserClientBaseInfo{
			UserID:           clientWeb.UserID,
			CurrentAccountID: clientWeb.CurrentAccountID,
			LoginTime:        clientWeb.LoginTime,
		}
	case *UserClientAndroid:
		clientAndroid := client.(*UserClientAndroid)
		return &UserClientBaseInfo{
			UserID:           clientAndroid.UserID,
			CurrentAccountID: clientAndroid.CurrentAccountID,
			LoginTime:        clientAndroid.LoginTime,
		}
	case *UserClientIos:
		clientIos := client.(*UserClientIos)
		return &UserClientBaseInfo{
			UserID:           clientIos.UserID,
			CurrentAccountID: clientIos.CurrentAccountID,
			LoginTime:        clientIos.LoginTime,
		}
	}
	panic(ErrClientNotFound)
}

type UserClientDbFunc func(db *gorm.DB) error

func HandleUserClient(client Client, handleFunc UserClientDbFunc) error {
	switch client.(type) {
	case *UserClientWeb:
		clientWeb := client.(*UserClientWeb)
		return handleFunc(clientWeb.GetDb().Model(clientWeb))
	case *UserClientAndroid:
		clientAndroid := client.(*UserClientAndroid)
		return handleFunc(clientAndroid.GetDb().Model(clientAndroid))
	case *UserClientIos:
		clientIos := client.(*UserClientIos)
		return handleFunc(clientIos.GetDb().Model(clientIos))
	}
	return nil
}
