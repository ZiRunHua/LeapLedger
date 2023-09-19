package main

import (
	"KeepAccount/global"
	"KeepAccount/initialize"
	"KeepAccount/router"
	"fmt"
	"net/http"
	"time"
)

func main() {
	initialize.Do()
	global.Init()
	engine := router.Init()
	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", initialize.Config.System.Addr),
		Handler:        engine,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	s.ListenAndServe()
}
