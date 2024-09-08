package cus

import (
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"time"
)

type Time time.Time

// GormDataType returns the common data type in Gorm for CustomTime
func (Time) GormDataType() string {
	return "timestamp"
}

// GormDBDataType returns the specific data type for a given dialect
func (Time) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return "TIMESTAMP"
}

type DeletedAt gorm.DeletedAt

func (DeletedAt) GormDataType() string {
	return "timestamp"
}

// GormDBDataType returns the specific data type for a given dialect
func (DeletedAt) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return "TIMESTAMP"
}
