package service

import (
	accountService "github.com/ZiRunHua/LeapLedger/service/account"
	categoryService "github.com/ZiRunHua/LeapLedger/service/category"
	commonService "github.com/ZiRunHua/LeapLedger/service/common"
	productService "github.com/ZiRunHua/LeapLedger/service/product"
	templateService "github.com/ZiRunHua/LeapLedger/service/template"
	thirdpartyService "github.com/ZiRunHua/LeapLedger/service/thirdparty"
	transactionService "github.com/ZiRunHua/LeapLedger/service/transaction"
	userService "github.com/ZiRunHua/LeapLedger/service/user"
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
