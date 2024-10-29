package thirdpartyService

import (
	"context"
	"github.com/ZiRunHua/LeapLedger/global/nats"
	userModel "github.com/ZiRunHua/LeapLedger/model/user"
)

func init() {
	nats.SubscribeTaskWithPayload(
		nats.TaskSendCaptchaEmail, func(t nats.PayloadSendCaptchaEmail, ctx context.Context) error {
			return GroupApp.sendCaptchaEmail(t.Email, t.Action)
		},
	)
	nats.SubscribeTaskWithPayload(
		nats.TaskSendNotificationEmail, func(t nats.PayloadSendNotificationEmail, ctx context.Context) error {
			user, err := userModel.NewDao().SelectById(t.UserId)
			if err != nil {
				return err
			}
			return GroupApp.sendNotificationEmail(user, t.Notification)
		},
	)
}
