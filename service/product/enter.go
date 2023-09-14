package productService

import transactionService "KeepAccount/service/transaction"

type Group struct {
	Product
}

var GroupApp = new(Group)

var transactionServer = transactionService.GroupApp
