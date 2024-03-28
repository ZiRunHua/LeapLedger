package userModel

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Friend struct {
	ID       uint `gorm:"primarykey"`
	UserId   uint `gorm:"uniqueIndex:idx_mapping,priority:1"`
	FriendId uint `gorm:"uniqueIndex:idx_mapping,priority:2;"`
	AddMode  AddMode
	gorm.Model
}

type AddMode string

const (
	FriendAddModeOfFriendInvitation    AddMode = "friendInvitation"
	FriendAddModeOfFriendOnInvitation  AddMode = "friendOnInvitation"
	FriendAddModeOfAccountInvitation   AddMode = "accountInvitation"
	FriendAddModeOfAccountOnInvitation AddMode = "accountOnInvitation"
)

func (f *Friend) TableName() string {
	return "user_friend"
}
func (f *Friend) GetFriend(args ...interface{}) (User, error) {
	return NewDao().SelectById(f.FriendId, args)
}

func (f *Friend) GetFriendInfo() (info UserInfo, err error) {
	return NewDao().SelectUserInfoById(f.FriendId)
}

/* 邀请 */
type FriendInvitation struct {
	ID      uint `gorm:"primarykey"`
	Inviter uint `gorm:"uniqueIndex:idx_mapping,priority:1"`
	Invitee uint `gorm:"uniqueIndex:idx_mapping,priority:2"`
	Status  FriendInvitationStatus
	gorm.Model
}

type FriendInvitationStatus int

const (
	InvitationStatsOfWaiting FriendInvitationStatus = iota
	InvitationStatsOfAccept
	InvitationStatsOfRefuse
)

func (f *FriendInvitation) TableName() string {
	return "user_friend_invitation"
}

func (f *FriendInvitation) ForUpdate(tx *gorm.DB) error {
	if err := tx.Model(f).Clauses(clause.Locking{Strength: "UPDATE"}).First(f, f.ID).Error; err != nil {
		return err
	}
	return nil
}

func (f *FriendInvitation) GetInviterInfo() (UserInfo, error) {
	return NewDao().SelectUserInfoById(f.Inviter)
}

func (f *FriendInvitation) GetInviteeInfo() (UserInfo, error) {
	return NewDao().SelectUserInfoById(f.Invitee)
}

func (f *FriendInvitation) Accept(tx *gorm.DB) (inviterFriend Friend, inviteeFriend Friend, err error) {
	err = f.ForUpdate(tx)
	if err != nil {
		return
	}
	if f.Status != InvitationStatsOfWaiting {
		err = errors.New("邀请状态异常")
		return
	}

	err = f.UpdateStatus(InvitationStatsOfAccept, tx)
	if err != nil {
		return
	}
	return f.AddFriend(tx)
}

func (f *FriendInvitation) Refuse(tx *gorm.DB) error {
	err := f.ForUpdate(tx)
	if err != nil {
		return err
	}
	if f.Status != InvitationStatsOfWaiting {
		return errors.New("邀请状态异常")
	}
	return f.UpdateStatus(InvitationStatsOfRefuse, tx)
}

func (f *FriendInvitation) UpdateStatus(status FriendInvitationStatus, tx *gorm.DB) error {
	return tx.Model(f).Update("status", status).Error
}

func (f *FriendInvitation) AddFriend(tx *gorm.DB) (inviterFriend Friend, inviteeFriend Friend, err error) {
	err = f.ForUpdate(tx)
	if err != nil {
		return
	}
	inviterFriend, err = NewDao(tx).AddFriend(f.Inviter, f.Invitee, FriendAddModeOfFriendInvitation)
	if err != nil {
		return
	}
	inviteeFriend, err = NewDao(tx).AddFriend(f.Invitee, f.Inviter, FriendAddModeOfFriendOnInvitation)
	if err != nil {
		return
	}
	return
}
