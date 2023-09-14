package main

import (
	"KeepAccount/global"
	"KeepAccount/initialize"
	"fmt"
	"net/http"
	"time"
)

func main() {
	initialize.Config()
	initialize.Logger() // 初始化zap日志库
	initialize.Gorm()   // gorm连接数据库
	db, _ := global.GvaDb.DB()
	if db != nil {
		defer db.Close()
	}
	//initialize.Redis()
	initialize.Other()

	router := initialize.Routers()
	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", global.GvaConfig.System.Addr),
		Handler:        router,
		ReadTimeout:    20 * time.Second,
		WriteTimeout:   20 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	s.ListenAndServe()
}
