package templateService

import (
	"KeepAccount/model/common/query"
	userModel "KeepAccount/model/user"
	_categoryService "KeepAccount/service/category"
	"fmt"
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
		fmt.Println("查询模板用户失败")
		panic(err)
	}
}

var categoryService = _categoryService.GroupApp
