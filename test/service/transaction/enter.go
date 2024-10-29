package transaction

import (
	_ "github.com/ZiRunHua/LeapLedger/test/initialize"
	testUtil "github.com/ZiRunHua/LeapLedger/test/util"
)
import (
	_service "github.com/ZiRunHua/LeapLedger/service"
)

var (
	get = &testUtil.Get{}

	service = _service.GroupApp.TransactionServiceGroup
)
