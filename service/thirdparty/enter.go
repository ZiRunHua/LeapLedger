package thirdpartyService

import (
	"github.com/ZiRunHua/LeapLedger/service/thirdparty/email"
)

type Group struct {
	Ai aiServer
}

var (
	GroupApp    = new(Group)
	emailServer = email.Service
)
