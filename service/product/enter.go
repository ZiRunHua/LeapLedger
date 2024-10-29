package productService

import transactionService "github.com/ZiRunHua/LeapLedger/service/transaction"

type Group struct {
	Product
}

var GroupApp = new(Group)

var transactionServer = transactionService.GroupApp
