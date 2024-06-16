package userService

import (
	"gorm.io/gorm"
)

type Group struct {
	Base   User
	Friend Friend
}

var GroupApp = new(Group)

func init() {
	globalTask.TransSubscribe[any](
		globalTask.TaskCreateTourist, func(db *gorm.DB, data any) error {
			_, err := GroupApp.Base.CreateTourist(db)
			return err
		},
	)
}
