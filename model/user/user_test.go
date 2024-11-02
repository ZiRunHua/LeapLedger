package userModel

import (
	"testing"

	"github.com/ZiRunHua/LeapLedger/global/db"
)

func TestName(t *testing.T) {
	err := db.Db.AutoMigrate(&Config[bool]{}, Config[int]{}, Config[string]{})
	if err != nil {
		t.Fatal(err)
	}
	err = db.Db.Create(
		Config[bool]{
			UserId: 1, Key: "pass", Value: false,
		},
	).Error
	if err != nil {
		t.Fatal(err)
	}
	err = db.Db.Create(
		Config[int]{
			UserId: 1, Key: "pass", Value: 45,
		},
	).Error
	if err != nil {
		t.Fatal(err)
	}
	err = db.Db.Create(
		Config[string]{
			UserId: 1, Key: "pass", Value: "okok",
		},
	).Error
	if err != nil {
		t.Fatal(err)
	}
}
