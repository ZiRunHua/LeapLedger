package userService

import (
	userModel "KeepAccount/model/user"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type Friend struct{}

func (f *Friend) CreateInvitation(
	inviter userModel.User, invitee userModel.User, tx *gorm.DB,
) (invitation userModel.FriendInvitation, err error) {
	dao := userModel.NewDao(tx)
	invitation, err = dao.CreateFriendInvitation(inviter.ID, invitee.ID)
	if false == errors.Is(err, gorm.ErrDuplicatedKey) {
		return
	}
	// 处理重复键
	invitation, err = dao.SelectFriendInvitation(inviter.ID, invitee.ID, true)
	if err != nil {
		return
	}
	var isRealFriend bool
	isRealFriend, err = dao.IsRealFriend(inviter.ID, invitee.ID)
	if isRealFriend || err != nil {
		return
	}
	if invitation.Status == userModel.InvitationStatsOfWaiting {
		return
	}
	err = invitation.UpdateStatus(userModel.InvitationStatsOfWaiting, tx)
	if err != nil {
		return
	}
	return
}

func (f *Friend) AcceptInvitation(Invitation *userModel.FriendInvitation, tx *gorm.DB) (
	userModel.Friend, userModel.Friend, error,
) {
	return Invitation.Accept(tx)
}

func (f *Friend) RefuseInvitation(Invitation *userModel.FriendInvitation, tx *gorm.DB) error {
	err := Invitation.Refuse(tx)
	return err
}
