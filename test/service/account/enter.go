package transaction

import (
	_ "github.com/ZiRunHua/LeapLedger/test/initialize"
	testUtil "github.com/ZiRunHua/LeapLedger/test/util"
)
import (
	_service "github.com/ZiRunHua/LeapLedger/service"
)

var (
	build           = &testUtil.Build{}
	query           = &testUtil.Query{}
	service         = _service.GroupApp.AccountServiceGroup
	categoryService = _service.GroupApp.CategoryServiceGroup
	transService    = _service.GroupApp.TransactionServiceGroup
)
