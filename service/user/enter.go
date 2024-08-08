package userService

import (
	"KeepAccount/global/contextKey"
	globalTask "KeepAccount/global/task"
	"context"
	"gorm.io/gorm"
)

type Group struct {
	User
	Friend Friend
}

var GroupApp = new(Group)

func init() {
	globalTask.Subscribe[any](
		globalTask.TaskCreateTourist, func(data any, ctx context.Context) error {
			_, err := GroupApp.CreateTourist(ctx.Value(contextKey.Tx).(*gorm.DB))
			return err
		},
	)
}
