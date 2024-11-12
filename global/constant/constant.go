package constant

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type ServerMode string

const Debug, Production ServerMode = "debug", "production"

var (
	RootDir                = getRootDir()
	LogPath                = filepath.Join(RootDir, "log")
	DataPath               = filepath.Join(RootDir, "data")
	ExampleAccountJsonPath = filepath.Clean(DataPath + "/template/account/example.json")
)

// IncomeExpense 收支类型
type IncomeExpense string // @name IncomeExpense `example:"expense" enums:"income,expense" swaggertype:"string"`

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

func getRootDir() string {
	// `os.Getwd()` is avoided here because, during tests, the working directory is set to the test file’s directory.
	// This command retrieves the module's root directory instead.
	// Source of `go list` usage: https://stackoverflow.com/a/75943840/23658318
	rootDir, err := exec.Command("go", "list", "-m", "-f", "{{.Dir}}").Output()
	if err == nil {
		return strings.TrimSpace(string(rootDir))
	}
	// If `go list` fails, it may indicate the absence of a Go environment.
	// In such cases, this suggests we are not in a test environment, so fall back to `os.Getwd()` to set `RootDir`.
	workDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	// Validate that the directory exists
	_, err = os.Stat(workDir)
	if err != nil {
		if os.IsNotExist(err) {
			panic(fmt.Sprintf("Path:%s does not exists", workDir))
		}
		panic(err)
	}
	return workDir
}
