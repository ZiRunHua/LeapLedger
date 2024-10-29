package engine

import (
	"fmt"
	"github.com/ZiRunHua/LeapLedger/global"
	"github.com/ZiRunHua/LeapLedger/global/constant"
	"github.com/ZiRunHua/LeapLedger/router/middleware"
	"github.com/gin-gonic/gin"
	"time"
)

var Engine *gin.Engine

func init() {
	Engine = gin.New()
	Engine.Use(
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
		Engine.Use(middleware.RequestLogger(global.RequestLogger))
	}
}
