package initialize

import (
	"KeepAccount/global"
)

func Gorm() {
	global.GvaDb = GormMysql()
}
