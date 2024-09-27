package thirdpartyService

import (
	"KeepAccount/global/nats"
	"context"
)

func init() {
	nats.SubscribeTaskWithPayload(
		nats.TaskSendCaptchaEmail, func(t nats.PayloadSendCaptchaEmail, ctx context.Context) error {
			return nil
			// return GroupApp.sendCaptchaEmail(t.Email, t.Action)
		},
	)
	nats.SubscribeTaskWithPayload(
		nats.TaskSendNotificationEmail, func(t nats.PayloadSendNotificationEmail, ctx context.Context) error {
			// user, err := userModel.NewDao().SelectById(t.UserId)
			// if err != nil {
			// 	return err
			// }
			return nil
			// return GroupApp.sendNotificationEmail(user, t.Notification)
		},
	)
}
