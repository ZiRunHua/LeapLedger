package userService

import (
	"KeepAccount/global/db"
	nats "KeepAccount/global/nats"
	"context"
)

type Group struct {
	User
	Friend Friend
}

var GroupApp = new(Group)

func init() {
	nats.SubscribeTaskWithPayload[any](
		nats.TaskCreateTourist, func(data any, ctx context.Context) error {
			_, err := GroupApp.CreateTourist(db.Get(ctx))
			return err
		},
	)
}
