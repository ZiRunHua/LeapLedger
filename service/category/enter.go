package categoryService

import (
	_thirdpartyService "github.com/ZiRunHua/LeapLedger/service/thirdparty"
)

type Group struct {
	Category
	Task _task
}

var GroupApp = new(Group)

var aiService = _thirdpartyService.GroupApp.Ai
