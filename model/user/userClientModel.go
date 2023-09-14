package userModel

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	commonModel "KeepAccount/model/common"
	"errors"
	"gorm.io/gorm"
	"strings"
	"time"
)

type ClientMap map[constant.Client]Client

func GetClients() ClientMap {
	return ClientMap{
		constant.Web:     new(UserClientWeb),
		constant.Android: new(UserClientAndroid),
		constant.Ios:     new(UserClientIos),
	}
}

type Client interface {
	commonModel.Model
	GetByUser(User) error
	CheckUserAgent(userAgent string) bool
	InitByUser(User, *gorm.DB) error
}

type UserClientBaseInfo struct {
	UserId                uint `gorm:"primaryKey"`
	CurrentAccountId      uint
	CurrentShareAccountId uint
	LoginTime             time.Time
}

func (u *UserClientBaseInfo) GetByUser(user User) error {
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

func (w *UserClientWeb) IsEmpty() bool {
	return w.UserId == 0
}
func (a *UserClientAndroid) IsEmpty() bool {
	return a.UserId == 0
}
func (i *UserClientIos) IsEmpty() bool {
	return i.UserId == 0
}

func (w *UserClientWeb) GetByUser(user User) error {
	return global.GvaDb.Where("user_id = ?", user.ID).First(&w).Error
}
func (a *UserClientAndroid) GetByUser(user User) error {
	return global.GvaDb.Where("user_id = ?", user.ID).First(&a).Error
}
func (i *UserClientIos) GetByUser(user User) error {
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

func (w *UserClientWeb) InitByUser(user User, tx *gorm.DB) error {
	w.UserId = user.ID
	w.CurrentAccountId = 0
	w.LoginTime = time.Now()
	return tx.Create(w).Error
}
func (a *UserClientAndroid) InitByUser(user User, tx *gorm.DB) error {
	a.UserId = user.ID
	a.CurrentAccountId = 0
	a.LoginTime = time.Now()
	return tx.Create(a).Error
}
func (i *UserClientIos) InitByUser(user User, tx *gorm.DB) error {
	i.UserId = user.ID
	i.CurrentAccountId = 0
	i.LoginTime = time.Now()
	return tx.Create(i).Error
}

func GetUserClientModel(client constant.Client) Client {
	switch client {
	case constant.Web:
		return &UserClientWeb{}
	case constant.Android:
		return &UserClientAndroid{}
	case constant.Ios:
		return &UserClientIos{}
	default:
		panic("unknown client")
	}
}

var ErrClientNotFound = errors.New("client not found")

func GetUserClientBaseInfo(client Client) *UserClientBaseInfo {
	switch client.(type) {
	case *UserClientWeb:
		clientWeb := client.(*UserClientWeb)
		return &UserClientBaseInfo{
			UserId:           clientWeb.UserId,
			CurrentAccountId: clientWeb.CurrentAccountId,
			LoginTime:        clientWeb.LoginTime,
		}
	case *UserClientAndroid:
		clientAndroid := client.(*UserClientAndroid)
		return &UserClientBaseInfo{
			UserId:           clientAndroid.UserId,
			CurrentAccountId: clientAndroid.CurrentAccountId,
			LoginTime:        clientAndroid.LoginTime,
		}
	case *UserClientIos:
		clientIos := client.(*UserClientIos)
		return &UserClientBaseInfo{
			UserId:           clientIos.UserId,
			CurrentAccountId: clientIos.CurrentAccountId,
			LoginTime:        clientIos.LoginTime,
		}
	}
	panic(ErrClientNotFound)
}

type UserClientDbFunc func(db *gorm.DB) error

func HandleUserClient(client Client, handleFunc UserClientDbFunc, tx *gorm.DB) error {
	switch client.(type) {
	case *UserClientWeb:
		clientWeb := client.(*UserClientWeb)
		return handleFunc(tx.Model(clientWeb))
	case *UserClientAndroid:
		clientAndroid := client.(*UserClientAndroid)
		return handleFunc(tx.Model(clientAndroid))
	case *UserClientIos:
		clientIos := client.(*UserClientIos)
		return handleFunc(tx.Model(clientIos))
	}
	return nil
}
