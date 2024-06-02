package router

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func Init() *gin.Engine {
	engine := gin.New()
	engine.Use(
		gin.LoggerWithConfig(
			gin.LoggerConfig{
				Formatter: func(params gin.LogFormatterParams) string {
					return fmt.Sprintf(
						"[GIN] %s | %s | %s | %d | %s | %s | %s\n",
						params.TimeStamp.Format(time.RFC3339),
						params.Method,
						params.Path,
						params.StatusCode,
						params.Latency,
						params.ClientIP,
						params.ErrorMessage,
					)
				},
			},
		),
		gin.CustomRecovery(middleware.Recovery),
	)
	if global.Config.Mode == constant.Debug {
		engine.Use(middleware.RequestLogger(global.RequestLogger))
	}

	APIv1Router := RouterGroupApp.APIv1
	//公共
	PublicGroup := engine.Group(global.Config.System.RouterPrefix)
	{
		// 健康监测
		PublicGroup.GET(
			"/health", func(c *gin.Context) {
				c.JSON(http.StatusOK, "ok")
			},
		)
	}
	{
		APIv1Router.InitPublicRouter(PublicGroup)
	}
	//需要登录校验
	PrivateGroup := engine.Group(global.Config.System.RouterPrefix)
	PrivateGroup.Use(middleware.JWTAuth())
	{
		APIv1Router.InitUserRouter(PrivateGroup)
		APIv1Router.InitCategoryRouter(PrivateGroup)
		APIv1Router.InitAccountRouter(PrivateGroup)
		APIv1Router.InitTransactionImportRouter(PrivateGroup)
		APIv1Router.InitTransactionRouter(PrivateGroup)
	}
	return engine
}
