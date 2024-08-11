package userService

import (
	"KeepAccount/global/db"
	globalTask "KeepAccount/global/task"
	"context"
)

type Group struct {
	User
	Friend Friend
}

var GroupApp = new(Group)

func init() {
	globalTask.Subscribe[any](
		globalTask.TaskCreateTourist, func(data any, ctx context.Context) error {
			_, err := GroupApp.CreateTourist(db.Get(ctx))
			return err
		},
	)
}
