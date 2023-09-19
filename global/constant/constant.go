package constant

// IncomeExpense 收支类型
type IncomeExpense string

const (
	Income  IncomeExpense = "income"
	Expense IncomeExpense = "expense"
)

// Client 客户端
type Client string

const (
	Web     Client = "web"
	Android Client = "android"
	Ios     Client = "ios"
)

type Encoding string

const (
	GBK  Encoding = "GBK"
	UTF8 Encoding = "UTF8"
)
