package userModel

import (
	"reflect"
	"time"
)

var DefaultConfigs = newDefaultConfigs(
	[]Config{
		&TransactionShareConfig{DisplayFlags: DISPLAY_FLAGS_DEFAULT},
		&BillImportConfig{IgnoreUnmappedCategory: false, CheckSameTransMode: CheckSameTransModeOfTip},
	},
)

type (
	Config interface {
		GetUserId() uint
		SetUserId(uint)
		TableName() string
	}
	ConfigBase struct {
		UserId    uint `gorm:"primarykey"`
		CreatedAt time.Time
		UpdatedAt time.Time
	}
	defaultConfigs struct {
		configs    []Config
		configsMap map[string]Config
	}
)

func (cb *ConfigBase) GetUserId() uint   { return cb.UserId }
func (cb *ConfigBase) SetUserId(id uint) { cb.UserId = id }

func newDefaultConfigs(configs []Config) defaultConfigs {
	configsMap := make(map[string]Config, len(configs))
	for _, config := range configs {
		configsMap[config.TableName()] = config
	}
	return defaultConfigs{configs: configs, configsMap: configsMap}
}

func (dc defaultConfigs) Iterator(userId uint) func(yield func(Config) bool) {
	return func(yield func(Config) bool) {
		for _, config := range dc.configs {
			config = deepCopyConfig(config)
			config.SetUserId(userId)
			if !yield(config) {
				return
			}
		}
	}
}

// GetConfig retrieves the default configuration associated with the table name
// A deep copy is performed using reflection to ensure the original configuration is not modified.
func (dc defaultConfigs) GetConfig(config Config) Config {
	return deepCopyConfig(dc.configsMap[config.TableName()])
}

type TransactionShareConfig struct {
	ConfigBase
	DisplayFlags Flag `gorm:"comment:'展示字段标志'"`
}

type Flag = uint

const (
	FLAG_AMOUNT Flag = 1 << iota
	FLAG_CATEGORY
	FLAG_TRADE_TIME
	FLAG_ACCOUNT
	FLAG_CREATE_TIME
	FLAG_UPDATE_TIME
	FLAG_REMARK
)

const DISPLAY_FLAGS_DEFAULT = FLAG_AMOUNT + FLAG_CATEGORY + FLAG_TRADE_TIME + FLAG_ACCOUNT + FLAG_REMARK

func (u *TransactionShareConfig) TableName() string {
	return "user_transaction_share_config"
}

func (u *TransactionShareConfig) GetFlagStatus(flag Flag) bool {
	return u.DisplayFlags&flag > 0
}

type (
	BillImportConfig struct {
		ConfigBase
		IgnoreUnmappedCategory bool
		CheckSameTransMode     CheckSameTransMode
	}
	CheckSameTransMode = uint
)

const (
	CheckSameTransModeOfIgnore = iota
	CheckSameTransModeOfTip    = iota
)

func (u *BillImportConfig) TableName() string { return "user_bill_import_config" }

func deepCopyConfig(config Config) Config {
	elem := reflect.ValueOf(config).Elem()
	value := reflect.New(elem.Type())
	value.Elem().Set(elem)
	return value.Interface().(Config)
}
