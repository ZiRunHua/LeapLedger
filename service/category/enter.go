package categoryService

import (
	_thirdpartyService "KeepAccount/service/thirdparty"
)

type Group struct {
	Category
}

var GroupApp = new(Group)

var thirdpartyService = _thirdpartyService.GroupApp
