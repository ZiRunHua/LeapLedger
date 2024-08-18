package router

import (
	_ "KeepAccount/docs"
	"KeepAccount/global"
	"KeepAccount/global/constant"
	"KeepAccount/router/group"
	_ "KeepAccount/router/v1"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"net/http"
)

func init() {
	// health
	group.Public.GET(
		"/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, "ok")
		},
	)
	if global.Config.Mode == constant.Debug {
		group.Public.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, func(config *ginSwagger.Config) {
			config.DocExpansion = "none"
			config.DeepLinking = true
		}))
	}
}
