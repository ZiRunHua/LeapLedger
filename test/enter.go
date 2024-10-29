package test

import (
	categoryModel "github.com/ZiRunHua/LeapLedger/model/category"
	"github.com/ZiRunHua/LeapLedger/test/initialize"
)

var (
	User                = initialize.User
	Account             = initialize.Account
	ExpenseCategoryList []categoryModel.Category
)

func init() {
	ExpenseCategoryList = initialize.ExpenseCategoryList
}

var Timezones = []string{
	"UTC",
	"America/New_York",
	"Europe/London",
	"Asia/Tokyo",
	"Australia/Sydney",
	"Europe/Paris",
	"Asia/Shanghai",
	"America/Los_Angeles",
	"Europe/Berlin",
	"Asia/Kolkata",
	"America/Chicago",
	"Europe/Moscow",
	"Asia/Dubai",
	"America/Denver",
	"Europe/Madrid",
	"Asia/Singapore",
	"America/Phoenix",
	"Europe/Rome",
	"Asia/Hong_Kong",
	"America/Anchorage",
	"Europe/Athens",
	"Asia/Seoul",
	"America/Halifax",
	"Europe/Stockholm",
	"Asia/Bangkok",
	"America/St_Johns",
	"Europe/Helsinki",
	"Asia/Jakarta",
	"America/Sao_Paulo",
	"Europe/Warsaw",
	"Asia/Kuala_Lumpur",
	"America/Argentina/Buenos_Aires",
	"Europe/Istanbul",
	"Asia/Manila",
	"America/Mexico_City",
	"Europe/Brussels",
	"Asia/Taipei",
	"America/Toronto",
	"Europe/Vienna",
	"Asia/Riyadh",
	"America/Caracas",
	"Europe/Zurich",
	"Asia/Baghdad",
	"America/Lima",
	"Europe/Copenhagen",
	"Asia/Tehran",
	"America/Fort_Nelson",
	"America/Hermosillo",
	"America/Chicago",
	"America/Mexico_City",
}
