package userService

type Group struct {
	User
	Friend Friend
}

var GroupApp = new(Group)
