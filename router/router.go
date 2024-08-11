package router

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	"KeepAccount/global/cusCtx"
	accountModel "KeepAccount/model/account"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

const accountWithIdPrefixPath = "/account/:" + string(cusCtx.AccountId)

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
	privateGroup := engine.Group(global.Config.System.RouterPrefix)
	privateGroup.Use(middleware.JWTAuth())

	turnAwayTouristPrivateGroup := privateGroup.Group("")
	turnAwayTouristPrivateGroup.Use(middleware.TurnAwayTourist())

	adminAuthRouter := privateGroup.Group(accountWithIdPrefixPath)
	adminAuthRouter.Use(middleware.AccountAuth(accountModel.UserPermissionAdministrator))
	{
		APIv1Router.InitUserRouter(privateGroup, turnAwayTouristPrivateGroup)
		APIv1Router.InitCategoryRouter(privateGroup)
		APIv1Router.InitAccountRouter(privateGroup)
		APIv1Router.InitTransactionImportRouter(privateGroup)
		APIv1Router.InitTransactionRouter(privateGroup, adminAuthRouter)
	}
	return engine
}
