package transaction

import (
	_ "KeepAccount/test/initialize"
	testUtil "KeepAccount/test/util"
)
import (
	_service "KeepAccount/service"
)

var (
	get = &testUtil.Get{}

	service = _service.GroupApp.TransactionServiceGroup
)
