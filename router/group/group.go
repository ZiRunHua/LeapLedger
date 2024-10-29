package group

import (
	"github.com/ZiRunHua/LeapLedger/global"
	"github.com/ZiRunHua/LeapLedger/global/cus"
	accountModel "github.com/ZiRunHua/LeapLedger/model/account"
	routerEngine "github.com/ZiRunHua/LeapLedger/router/engine"
	"github.com/ZiRunHua/LeapLedger/router/middleware"
	"github.com/gin-gonic/gin"
)

var engine = routerEngine.Engine
var (
	Public, Private *gin.RouterGroup
	Account         *gin.RouterGroup

	NoTourist *gin.RouterGroup

	AccountReader        *gin.RouterGroup
	AccountOwnEditor     *gin.RouterGroup
	AccountAdministrator *gin.RouterGroup
	AccountCreator       *gin.RouterGroup
)

const accountWithIdPrefixPath = "/account/:" + string(cus.AccountId)

func init() {
	Public = engine.Group(global.Config.System.RouterPrefix + "/public")
	Private = engine.Group(global.Config.System.RouterPrefix, middleware.JWTAuth())

	NoTourist = Private.Group("")
	NoTourist.Use(middleware.NoTourist())
	// account router
	Account = Private.Group(accountWithIdPrefixPath)
	// account role
	AccountReader = createAccountRoleGroup(accountModel.UserPermissionReader)
	AccountOwnEditor = createAccountRoleGroup(accountModel.UserPermissionOwnEditor)
	AccountAdministrator = createAccountRoleGroup(accountModel.UserPermissionAdministrator)
	AccountCreator = createAccountRoleGroup(accountModel.UserPermissionCreator)
}

func createAccountRoleGroup(permission accountModel.UserPermission) *gin.RouterGroup {
	group := Account.Group("", middleware.AccountAuth(permission))
	return group
}
