package transaction

import (
	"context"
	"time"

	"github.com/ZiRunHua/LeapLedger/global/constant"
	"github.com/ZiRunHua/LeapLedger/global/db"
	"github.com/ZiRunHua/LeapLedger/global/nats"
	accountModel "github.com/ZiRunHua/LeapLedger/model/account"
	_ "github.com/ZiRunHua/LeapLedger/test/initialize"
	"github.com/google/uuid"
)
import (
	"testing"
)

func TestTourCreate(t *testing.T) {
	t.Parallel()
	if !nats.PublishTask(nats.TaskCreateTourist) {
		t.Fail()
	}
	time.Sleep(time.Second * 10)
	user, err := service.EnableTourist(uuid.NewString(), constant.Android, context.TODO())
	if err != nil {
		t.Fatal(err)
	}
	var account accountModel.Account
	err = db.Db.Where("user_id = ?", user.ID).First(&account).Error
	if err != nil {
		t.Fatal(err)
	}
}
