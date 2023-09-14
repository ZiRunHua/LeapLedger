package accountService

import (
	log "KeepAccount/service/log"
)

var ServiceGroupApp = &Group{}

type Group struct {
	Base  base
	Share share
}

var logServer = log.Log
