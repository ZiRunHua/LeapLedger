package transaction

import (
	_ "KeepAccount/test/initialize"
	testUtil "KeepAccount/test/util"
)
import (
	_service "KeepAccount/service"
)

var (
	build           = &testUtil.Build{}
	query           = &testUtil.Query{}
	service         = _service.GroupApp.AccountServiceGroup
	categoryService = _service.GroupApp.CategoryServiceGroup
	transService    = _service.GroupApp.TransactionServiceGroup
)
