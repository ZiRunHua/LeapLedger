package transactionService

type Group struct {
	Transaction
	Timing Timing
}

var (
	GroupApp = new(Group)

	server = &Transaction{}
)
