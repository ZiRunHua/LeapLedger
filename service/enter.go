package service

import (
	accountService "KeepAccount/service/account"
	categoryService "KeepAccount/service/category"
	commonService "KeepAccount/service/common"
	productService "KeepAccount/service/product"
	templateService "KeepAccount/service/template"
	thirdpartyService "KeepAccount/service/thirdparty"
	transactionService "KeepAccount/service/transaction"
	userService "KeepAccount/service/user"
)

var GroupApp = new(Group)

type Group struct {
	CommonServiceGroup      commonService.Group
	CategoryServiceGroup    categoryService.Group
	AccountServiceGroup     accountService.Group
	TransactionServiceGroup transactionService.Group
	UserServiceGroup        userService.Group
	ProductServiceGroup     productService.Group
	TemplateServiceGroup    templateService.Group
	ThirdpartyServiceGroup  thirdpartyService.Group
}
