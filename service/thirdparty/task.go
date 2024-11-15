package thirdpartyService

import (
	"context"
	"errors"

	"github.com/ZiRunHua/LeapLedger/global"
	"github.com/ZiRunHua/LeapLedger/global/nats"
	userModel "github.com/ZiRunHua/LeapLedger/model/user"
)

func init() {
	nats.SubscribeTaskWithPayload(
		nats.TaskSendCaptchaEmail, func(t nats.PayloadSendCaptchaEmail, ctx context.Context) error {
			err := GroupApp.sendCaptchaEmail(t.Email, t.Action)
			if errors.Is(err, global.ErrServiceClosed) {
				return nil
			}
			return err
		},
	)
	nats.SubscribeTaskWithPayload(
		nats.TaskSendNotificationEmail, func(t nats.PayloadSendNotificationEmail, ctx context.Context) error {
			user, err := userModel.NewDao().SelectById(t.UserId)
			if err != nil {
				return err
			}
			err = GroupApp.sendNotificationEmail(user, t.Notification)
			if errors.Is(err, global.ErrServiceClosed) {
				return nil
			}
			return err
		},
	)
}
