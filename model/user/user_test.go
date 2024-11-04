package userModel

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/ZiRunHua/LeapLedger/global/db"
)

func TestExampleConfig(t *testing.T) {
	type TestConfig struct {
		ConfigBase
		Flags  int
		Number int
		Type   string
	}
	userId := uint(rand.Intn(100000))
	err := db.Db.AutoMigrate(TestConfig{})
	if err != nil {
		t.Fatal(err)
	}
	dao, config := NewDao(), &TestConfig{ConfigBase: ConfigBase{UserId: userId}}
	optionList := []func() error{
		func() error {
			return dao.CreateConfig(config)
		},
		func() error {
			config.Type = "test"
			return dao.UpdateConfigByKey(config, "type")
		},
		func() error {
			return dao.OpenConfigBinaryField(config, "flags", 8)
		},
		func() error {
			return dao.ClosedConfigBinaryField(config, "flags", 8)
		},
		func() error {
			config = &TestConfig{
				ConfigBase: ConfigBase{UserId: userId},
				Flags:      16 + 8 + 4,
				Type:       "update_config",
				Number:     65,
			}
			return dao.UpdateConfig(config)
		},
	}
	for _, f := range optionList {
		err := f()
		if err != nil {
			t.Fatal(err)
		}
		err = dao.ReadConfig(config)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(fmt.Sprintln(config))
	}
	_ = db.Db.Delete(config).Error
}
