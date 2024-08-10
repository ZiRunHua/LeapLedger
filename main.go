package main

import (
	"KeepAccount/initialize"
	_ "KeepAccount/initialize/database"
)
import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	GvaTask "KeepAccount/global/task"
	"KeepAccount/router"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var httpServer *http.Server

func main() {
	engine := router.Init()
	httpServer = &http.Server{
		Addr:           fmt.Sprintf(":%d", initialize.Config.System.Addr),
		Handler:        engine,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	if global.Config.Mode == constant.Debug {
		fmt.Println(global.TestUserInfo)
	}
	err := httpServer.ListenAndServe()
	if err != nil {
		panic(err)
	}
	shutDown()
}

func shutDown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")
	GvaTask.Shutdown()

	if err := httpServer.Shutdown(context.TODO()); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
