package templateService

import (
	"KeepAccount/model/common/query"
	userModel "KeepAccount/model/user"
	_categoryService "KeepAccount/service/category"
)

var GroupApp = &Group{}

type Group struct {
	Template template
}

const templateUserId = 1

var (
	tempUser = &userModel.User{}
)

func init() {
	var err error
	tempUser, err = query.FirstByPrimaryKey[*userModel.User](templateUserId)
	if err != nil {
		panic(err)
	}
}

var categoryService = _categoryService.GroupApp
