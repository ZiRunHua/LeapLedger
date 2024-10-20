package userModel

import (
	"errors"

	"KeepAccount/global"
	"KeepAccount/global/constant"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserDao struct {
	db *gorm.DB
}

type _baseInterface interface {
	Add(data AddData) (User, error)
	SelectById(id uint, args ...interface{}) (User, error)
	PluckNameById(id uint) (string, error)
	SelectByEmail(email string) (User, error)
	SelectUserInfoById(id uint) (result UserInfo, err error)
	SelectUserInfoByCondition(condition Condition) ([]UserInfo, error)
}

type _friendInterface interface {
	CreateFriendInvitation(uint, uint) (FriendInvitation, error)
	SelectFriendInvitation(inviter uint, invitee uint, forUpdate bool) (result FriendInvitation, err error)
	SelectFriendInvitationList(inviter *uint, invitee *uint) (result []FriendInvitation, err error)
	SelectFriend(uint, uint) (Friend, error)
	IsRealFriend(userId uint, friendId uint) (bool, error)
	AddFriend(uint, uint, AddMode) (Friend, error)
	SelectFriendList(uint) ([]Friend, error)
}

func NewDao(db ...*gorm.DB) *UserDao {
	if len(db) > 0 {
		return &UserDao{
			db: db[0],
		}
	}
	return &UserDao{global.GvaDb}
}

type AddData struct {
	Username string
	Password string
	Email    string
}

func (u *UserDao) Add(data AddData) (User, error) {
	user := User{
		Username: data.Username,
		Password: data.Password,
		Email:    data.Email,
	}
	err := u.db.Create(&user).Error
	return user, err
}

func (u *UserDao) SelectById(id uint, args ...interface{}) (User, error) {
	user := User{}
	var err error
	if len(args) > 0 {
		err = u.db.Where("Id = ?", id).Select(args).First(&user).Error
	} else {
		err = u.db.Where("Id = ?", id).First(&user).Error
	}
	return user, err
}

func (u *UserDao) CheckEmail(email string) error {
	err := u.db.Where("email = ?", email).Take(&User{}).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	} else if err == nil {
		return errors.New("邮箱已存在")
	} else {
		return err
	}
}

func (u *UserDao) SelectUserInfoById(id uint) (result UserInfo, err error) {
	err = u.db.Select("id", "username", "email").Where("id = ?", id).Model(&User{}).First(&result).Error
	return result, err
}

func (u *UserDao) PluckNameById(id uint) (string, error) {
	var name string
	err := u.db.Model(&User{}).Where("Id = ?", id).Pluck("username", &name).Error
	return name, err
}

func (u *UserDao) SelectByEmail(email string) (User, error) {
	user := User{}
	err := u.db.Where("email = ?", email).First(&user).Error
	return user, err
}

func (u *UserDao) SelectByDeviceNumber(client constant.Client, deviceNumber string) (
	clientBaseInfo UserClientBaseInfo,
	err error) {
	err = u.db.Model(GetUserClientModel(client)).Where("device_number = ?", deviceNumber).First(&clientBaseInfo).Error
	return
}

type Condition struct {
	Id                 *uint
	LikePrefixUsername *string
	Offset             int
	Limit              int
}

func (u *UserDao) SelectUserInfoByCondition(condition Condition) ([]UserInfo, error) {
	query := u.db
	if condition.Id != nil {
		query = query.Where("id = ?", *condition.Id)
	}
	if condition.LikePrefixUsername != nil {
		query = query.Where("username like ?", *condition.LikePrefixUsername+"%")
	}
	var list []UserInfo
	err := query.Model(&User{}).Select(
		"id", "username", "email",
	).Offset(condition.Offset).Limit(condition.Limit).Find(&list).Error
	return list, err
}

func (u *UserDao) SelectClientInfoByUserAndAccount(
	client constant.Client, userId, accountId uint,
) (result []UserClientBaseInfo, err error) {
	err = u.db.Where(
		"user_id = ? AND (current_account_id  = ? OR current_share_account_id  = ?)", userId, accountId, accountId,
	).Model(GetUserClientModel(client)).Find(&result).Error
	return result, err
}

func (u *UserDao) SelectUserClientBaseInfo(userId uint, client constant.Client) (result UserClientBaseInfo, err error) {
	err = u.db.Model(GetUserClientModel(client)).First(&result, userId).Error
	if err != nil {
		return
	}
	return
}

func (u *UserDao) SelectUserClient(userId uint, client constant.Client) (Client, error) {
	clientModel := GetUserClientModel(client)
	err := u.db.Model(clientModel).First(clientModel, userId).Error
	return clientModel, err
}

func (u *UserDao) UpdateUserClientBaseInfo(userId uint, client constant.Client) (result UserClientBaseInfo, err error) {
	err = u.db.Model(GetUserClientModel(client)).First(&result, userId).Error
	if err != nil {
		return
	}
	return
}

func (u *UserDao) CreateFriendInvitation(inviter uint, invitee uint) (FriendInvitation, error) {
	var invitation = FriendInvitation{Inviter: inviter, Invitee: invitee, Status: InvitationStatsOfWaiting}
	err := u.db.Model(&invitation).Create(&invitation).Error
	return invitation, err
}

func (u *UserDao) SelectFriendInvitation(inviter uint, invitee uint, forUpdate bool) (
	result FriendInvitation, err error,
) {
	query := u.db.Where("inviter = ? AND invitee = ?", inviter, invitee)
	if forUpdate {
		query = query.Set("gorm:query_option", "FOR UPDATE")
	}
	err = query.Find(&result).Error
	return
}

func (u *UserDao) SelectFriendInvitationList(inviter *uint, invitee *uint) (result []FriendInvitation, err error) {
	query := u.db
	if inviter != nil {
		query = query.Where("inviter = ?", inviter)
	}
	if invitee != nil {
		query = query.Where("invitee = ?", invitee)
	}
	err = query.Model(&FriendInvitation{}).Find(&result).Error
	return
}

func (u *UserDao) SelectFriend(userId uint, friendId uint) (Friend, error) {
	var friend Friend
	err := u.db.Model(&Friend{}).Where("user_id = ? AND friend_id = ?", userId, friendId).First(&friend).Error
	return friend, err
}

func (u *UserDao) IsRealFriend(userId uint, friendId uint) (bool, error) {
	var count int64
	whereSql := "user_id = ? AND friend_id = ? OR friend_id = ? AND user_id = ?"
	err := u.db.Model(&Friend{}).Where(whereSql, userId, friendId, userId, friendId).Count(&count).Error
	return count == 2, err
}

func (u *UserDao) AddFriend(userId uint, friendId uint, add AddMode) (mapping Friend, err error) {
	mapping = Friend{UserId: userId, FriendId: friendId, AddMode: add}
	err = u.db.Create(&mapping).Error
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			mapping, err = u.SelectFriend(userId, friendId)
		} else {
			return mapping, err
		}
	}
	return mapping, err
}

func (u *UserDao) SelectFriendList(userId uint) (result []Friend, err error) {
	err = u.db.Model(&Friend{}).Where("user_id = ?", userId).Find(&result).Error
	return
}

func (u *UserDao) SelectTour(userId uint) (Tour, error) {
	var tour Tour
	return tour, u.db.Where("user_id = ?", userId).First(&tour).Error
}

func (u *UserDao) CreateTour(user User) (Tour, error) {
	tour := Tour{
		UserId: user.ID,
		Status: false,
	}
	err := u.db.Create(&tour).Error
	return tour, err
}

func (u *UserDao) SelectByUnusedTour() (tour Tour, err error) {
	err = u.db.Where("status = false").Clauses(
		clause.Locking{
			Strength: clause.LockingStrengthUpdate,
			Options:  clause.LockingOptionsSkipLocked,
		},
	).First(&tour).Error
	return tour, err
}
