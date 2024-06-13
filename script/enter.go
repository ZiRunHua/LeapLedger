package script

import "KeepAccount/service"

var (
	userService        = service.GroupApp.UserServiceGroup.Base
	accountService     = service.GroupApp.AccountServiceGroup.Base
	categoryService    = service.GroupApp.CategoryServiceGroup.Category
	transactionService = service.GroupApp.TransactionServiceGroup.Transaction
	productService     = service.GroupApp.ProductServiceGroup.Product
	templateService    = service.GroupApp.TemplateService.Template
	//第三方服务
	thirdpartyService = service.GroupApp.ThirdpartyServiceGroup
)
