package productService

import _transactionService "github.com/ZiRunHua/LeapLedger/service/transaction"

type Group struct {
	Product
}

var GroupApp = new(Group)

var transactionService = _transactionService.GroupApp
