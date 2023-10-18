package service

import (
	accountService "KeepAccount/service/account"
	categoryService "KeepAccount/service/category"
	productService "KeepAccount/service/product"
	templateService "KeepAccount/service/template"
	transactionService "KeepAccount/service/transaction"
	userService "KeepAccount/service/user"
)

var GroupApp = new(Group)

type Group struct {
	CategoryServiceGroup    categoryService.Group
	AccountServiceGroup     accountService.Group
	TransactionServiceGroup transactionService.Group
	UserServiceGroup        userService.Group
	ProductServiceGroup     productService.Group
	TemplateService         templateService.Group
}
