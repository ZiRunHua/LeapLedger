package accountService

import (
	log "KeepAccount/service/log"
	userService "KeepAccount/service/user"
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
