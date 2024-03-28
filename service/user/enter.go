package userService

type Group struct {
	Base   User
	Friend Friend
}

var GroupApp = new(Group)
