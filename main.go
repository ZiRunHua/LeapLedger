package main

import (
	"github.com/ZiRunHua/LeapLedger/initialize"
	_ "github.com/ZiRunHua/LeapLedger/initialize/database"
	"github.com/ZiRunHua/LeapLedger/router"
)
import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	// Import "time/tzdata" for loading time zone data,
	// so in order to make the binary files can be run independently,
	// or in need of extra "$GOROOT/lib/time/zoneinfo.Zip" file, see time.LoadLocation
	_ "time/tzdata"
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
		Handler:        router.Engine,
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

	if err := httpServer.Shutdown(context.TODO()); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
