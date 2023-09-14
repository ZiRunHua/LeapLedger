package router

import v1 "KeepAccount/router/v1"

type RouterGroup struct {
	APIv1 v1.RouterGroup
}

var RouterGroupApp = new(RouterGroup)
