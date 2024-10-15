package transaction

import (
	"KeepAccount/global/db"
	accountModel "KeepAccount/model/account"
	userModel "KeepAccount/model/user"
	_ "KeepAccount/test/initialize"
	"context"
	"errors"
)
import (
	"testing"
)

func TestTourCreate(t *testing.T) {
	user, err := service.CreateTourist(context.TODO())
	if err != nil {
		t.Fatal(err)
	}
	t.Run(
		"Check tour data", func(t *testing.T) {
			var account accountModel.Account
			err = db.Db.Where("user_id = ?", user.ID).First(&account).Error
			if err != nil {
				t.Fatal(err)
			}
			tour, err := userModel.NewDao().SelectTour(user.ID)
			if err != nil {
				t.Fatal(err)
			}
			if tour.Status != false {
				t.Fatal(errors.New("error tour status"))
			}
		},
	)
}
