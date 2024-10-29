package router

import (
	_ "github.com/ZiRunHua/LeapLedger/docs"
	"github.com/ZiRunHua/LeapLedger/global"
	"github.com/ZiRunHua/LeapLedger/global/constant"
	routerEngine "github.com/ZiRunHua/LeapLedger/router/engine"
	"github.com/ZiRunHua/LeapLedger/router/group"
	_ "github.com/ZiRunHua/LeapLedger/router/v1"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"net/http"
)

var Engine = routerEngine.Engine

func init() {
	// health
	group.Public.GET(
		"/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, "ok")
		},
	)
	if global.Config.Mode == constant.Debug {
		group.Public.GET(
			"/swagger/*any", ginSwagger.WrapHandler(
				swaggerFiles.Handler, func(config *ginSwagger.Config) {
					config.DocExpansion = "none"
					config.DeepLinking = true
				},
			),
		)
	}
}
