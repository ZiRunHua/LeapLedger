package initialize

import (
	"KeepAccount/global"
	"github.com/songzhibin97/gkit/cache/local_cache"
	"time"
)

func Other() {
	global.BlackCache = local_cache.NewCache(
		local_cache.SetDefaultExpire(time.Hour * 2),
	)
}
