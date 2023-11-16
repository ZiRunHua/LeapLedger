package response

import (
	userModel "KeepAccount/model/user"
)

type Login struct {
	Token          string
	CurrentAccount *AccountOne
	User           *UserOne
}

type Register struct {
	Token string
}

type UserOne struct {
	Username   string
	Email      string
	CreateTime int64
}

func UserModelToResponse(user *userModel.User) *UserOne {
	if user == nil {
		return &UserOne{}
	}
	return &UserOne{
		Email:      user.Email,
		Username:   user.Username,
		CreateTime: user.CreatedAt.Unix(),
	}
}
