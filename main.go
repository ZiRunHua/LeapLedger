package main

import (
	"KeepAccount/initialize"
	_ "KeepAccount/initialize/database"
	"KeepAccount/router/engine"
)
import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	GvaTask "KeepAccount/global/task"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)
import (
	_ "KeepAccount/router"
)

var httpServer *http.Server

//	@title		LeapLedger API
//	@version	1.0

//	@contact.name	ZiRunHua

//	@license.name	AGPL 3.0
//	@license.url	https://www.gnu.org/licenses/agpl-3.0.html

//	@host	localhost:8080

// @securityDefinitions.jwt	Bearer
// @in							header
// @name						Authorization
func main() {
	httpServer = &http.Server{
		Addr:           fmt.Sprintf(":%d", initialize.Config.System.Addr),
		Handler:        engine.Engine,
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
