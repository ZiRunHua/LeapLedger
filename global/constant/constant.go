package constant

import "os"

var WorkDir string

func init() {
	var err error
	WorkDir, err = os.Getwd()
	if err != nil {
		panic(err)
	}
	WorkDir += ""
}

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
