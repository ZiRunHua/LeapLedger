package script

import "KeepAccount/service"

var (
	userService        = service.GroupApp.UserServiceGroup
	accountService     = service.GroupApp.AccountServiceGroup
	categoryService    = service.GroupApp.CategoryServiceGroup
	transactionService = service.GroupApp.TransactionServiceGroup
	productService     = service.GroupApp.ProductServiceGroup
	templateService    = service.GroupApp.TemplateServiceGroup
	//第三方服务
	thirdpartyService = service.GroupApp.ThirdpartyServiceGroup
)
