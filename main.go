package main

import (
	_ "KeepAccount/global"
	_ "KeepAccount/global/constant"
	"KeepAccount/initialize"
	_ "KeepAccount/initialize/database"
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
	fmt.Println()
	engine := router.Init()
	httpServer = &http.Server{
		Addr:           fmt.Sprintf(":%d", initialize.Config.System.Addr),
		Handler:        engine,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxHeaderBytes: 1 << 20,
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
