package templateService

import (
	userModel "KeepAccount/model/user"
	_categoryService "KeepAccount/service/category"
)

var GroupApp = &Group{}

type Group struct {
	Template template
}

var TmplUserId uint = 1

const (
	TmplUserEmail    = "template@gmail.com"
	TmplUserPassword = "1999123456"
	TmplUserName     = "template"
)

var (
	tmplUser userModel.User
)

func SetTmplUser(user userModel.User) {
	tmplUser = user
	TmplUserId = user.ID
}

var categoryService = _categoryService.GroupApp
