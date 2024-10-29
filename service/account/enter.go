package accountService

import (
	log "github.com/ZiRunHua/LeapLedger/service/log"
	userService "github.com/ZiRunHua/LeapLedger/service/user"
)

var GroupApp = &Group{}

type Group struct {
	base
	Share share
}

var (
	logServer  = log.Log
	userServer = userService.GroupApp
)
