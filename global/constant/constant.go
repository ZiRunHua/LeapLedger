package constant

import (
	"os"
)

var WORK_PATH string

var LOG_PAYH = WORK_PATH + "/log"
var DATA_PATH = WORK_PATH + "/data"

func init() {
	var err error
	WORK_PATH, err = os.Getwd()
	if err != nil {
		panic(err)
	}
	LOG_PAYH = WORK_PATH + "/log"
	DATA_PATH = WORK_PATH + "/data"
}

// IncomeExpense 收支类型
type IncomeExpense string

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
