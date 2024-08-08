package accountService

import (
	log "KeepAccount/service/log"
	userService "KeepAccount/service/user"
)

var ServiceGroupApp = &Group{}

type Group struct {
	Base  base
	Share share
}

var (
	logServer  = log.Log
	userServer = userService.GroupApp
)
