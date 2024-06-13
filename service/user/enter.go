package userService

import (
	"KeepAccount/global/nats"
	"gorm.io/gorm"
)

type Group struct {
	Base   User
	Friend Friend
}

var GroupApp = new(Group)

func init() {
	nats.TransSubscribe[any](
		nats.TaskCreateTourist, func(db *gorm.DB, data any) error {
			_, err := GroupApp.Base.CreateTourist(db)
			return err
		},
	)
}
