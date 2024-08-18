package constant

type ServerMode string

var Debug, Production ServerMode = "debug", "production"

const WORK_PATH = "/go/LeapLedger"
const RUNTIME_DATA_PATH = WORK_PATH + "/runtime/data"

const LOG_PATH = WORK_PATH + "/log"
const DATA_PATH = WORK_PATH + "/data"

var ExampleAccountJsonPath = DATA_PATH + "/template/account/example.json"

// IncomeExpense 收支类型
type IncomeExpense string //@name IncomeExpense `example:"expense" enums:"income,expense" swaggertype:"string"`

const (
	Income  IncomeExpense = "income"
	Expense IncomeExpense = "expense"
)

func (ie *IncomeExpense) QueryIncome() bool {
	return ie == nil || *ie == Income
}

func (ie *IncomeExpense) QueryExpense() bool {
	return ie == nil || *ie == Expense
}

// Client 客户端
type Client string

var ClientList = []Client{Web, Android, Ios}

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

type UserAction string

const (
	Login          UserAction = "login"
	Register       UserAction = "register"
	ForgetPassword UserAction = "forgetPassword"
	UpdatePassword UserAction = "updatePassword"
)

type CacheTab string

const (
	LoginFailCount         CacheTab = "loginFailCount"
	EmailCaptcha           CacheTab = "emailCaptcha"
	CaptchaEmailErrorCount CacheTab = "captchaEmailErrorCount"
)

type Notification int

const (
	NotificationOfCaptcha             Notification = iota
	NotificationOfRegistrationSuccess Notification = iota
	NotificationOfUpdatePassword      Notification = iota
)

type LogOperation string

const (
	LogOperationOfAdd    LogOperation = "add"
	LogOperationOfUpdate LogOperation = "update"
	LogOperationOfDelete LogOperation = "delete"
)

// nats

type Subject string
