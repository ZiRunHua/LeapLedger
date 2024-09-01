package userService

import (
	nats "KeepAccount/global/nats"
	"context"
)

type Group struct {
	User
	Friend Friend
}

var GroupApp = new(Group)

func init() {
	nats.SubscribeTask(
		nats.TaskCreateTourist, func(ctx context.Context) error {
			_, err := GroupApp.CreateTourist(ctx)
			return err
		},
	)
}
