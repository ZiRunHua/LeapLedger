package gormFunc

import (
	"gorm.io/gorm"
	"reflect"
	"strings"
)

func AlterFiledToHeader(table interface{}, filed string, db *gorm.DB) error {
	return db.Exec("ALTER TABLE ? MODIFY COLUMN ? INT FIRST", GetTableName(table), filed).Error
}

func AlterIdToHeader(table interface{}, db *gorm.DB) error {
	reflectType := reflect.TypeOf(table)
	for i := 0; i < reflectType.NumField(); i++ {
		field := reflectType.Field(i)
		if field.Name == "ID" {
			return AlterFiledToHeader(table, "id", db)
		}
	}
	return nil
}

func GetTableName(table interface{}) string {
	val := reflect.ValueOf(table)
	method := val.MethodByName("TableName")
	if method.IsValid() {
		results := method.Call(nil)
		if len(results) > 0 {
			if tableName, ok := results[0].Interface().(string); ok {
				return tableName
			}
		}
	}
	tableName := reflect.TypeOf(table).Name()
	return strings.ToLower(tableName[:1]) + tableName[1:]
}
