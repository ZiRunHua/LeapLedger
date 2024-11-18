package user

import (
	"sync"
	"testing"

	"github.com/ZiRunHua/LeapLedger/global/db"
	userModel "github.com/ZiRunHua/LeapLedger/model/user"
	"github.com/ZiRunHua/LeapLedger/util/rand"
	"github.com/stretchr/testify/assert"
)

func TestUserConfig(t *testing.T) {
	dao := userModel.NewDao(db.Db)
	user := &userModel.User{Username: t.Name(), Email: t.Name() + rand.String(12)}
	err := db.Db.Create(&user).Error
	if err != nil {
		t.Fatal(err)
	}
	t.Log(user)
	config := &userModel.TransactionShareConfig{}
	config.SetUserId(user.ID)
	defaultConfig := userModel.DefaultConfigs.GetConfig(config).(*userModel.TransactionShareConfig)
	t.Run(
		"GetConfig", func(t *testing.T) {
			err = dao.GetConfig(config)
			assert.NoError(t, err)
			assert.Equal(t, config.DisplayFlags, defaultConfig.DisplayFlags)
		},
	)

	t.Run(
		"GetConfigColumns", func(t *testing.T) {
			err = dao.GetConfigColumns(config, "display_flags")
			assert.NoError(t, err)
			assert.Equal(t, config.DisplayFlags, defaultConfig.DisplayFlags)
		},
	)

	t.Run(
		"UpdateConfig", func(t *testing.T) {
			config.DisplayFlags = userModel.FLAG_ACCOUNT
			err = dao.UpdateConfig(config)
			assert.NoError(t, err)

			var updatedConfig userModel.TransactionShareConfig
			err = db.Db.First(&updatedConfig, config.UserId).Error
			assert.NoError(t, err)
			assert.Equal(t, config.DisplayFlags, updatedConfig.DisplayFlags)
		},
	)

	t.Run(
		"ConfigBinaryFlag", func(t *testing.T) {
			concurrency, group := 5, sync.WaitGroup{}
			group.Add(concurrency)
			for i := 0; i < concurrency; i++ {
				go func() {
					err = dao.DisableConfigBinaryFlag(config, "display_flags", userModel.FLAG_ACCOUNT)
					assert.NoError(t, err)

					var updatedConfig userModel.TransactionShareConfig
					err = db.Db.First(&updatedConfig, config.UserId).Error
					assert.NoError(t, err)
					assert.Equal(t, uint(0), updatedConfig.DisplayFlags&userModel.FLAG_ACCOUNT)
					group.Done()
				}()
			}
			group.Wait()
			group.Add(concurrency)
			for i := 0; i < concurrency; i++ {
				go func() {
					err = dao.EnableConfigBinaryFlag(config, "display_flags", userModel.FLAG_ACCOUNT)
					assert.NoError(t, err)

					var updatedConfig userModel.TransactionShareConfig
					err = db.Db.First(&updatedConfig, config.UserId).Error
					assert.NoError(t, err)
					assert.Equal(t, userModel.FLAG_ACCOUNT, updatedConfig.DisplayFlags&userModel.FLAG_ACCOUNT)
					group.Done()
				}()
			}
			group.Wait()
		},
	)

	t.Run(
		"UpdateConfigColumns", func(t *testing.T) {
			billConfig := &userModel.BillImportConfig{}
			billConfig.SetUserId(user.ID)
			err = dao.GetConfig(billConfig)
			assert.NoError(t, err)
			billConfig.IgnoreUnmappedCategory = !billConfig.IgnoreUnmappedCategory
			billConfig.CheckSameTransMode = 1 + billConfig.CheckSameTransMode
			err = dao.UpdateConfigColumns(billConfig, "ignore_unmapped_category", "check_same_trans_mode")
			assert.NoError(t, err)

			var updatedConfig userModel.BillImportConfig
			err = db.Db.First(&updatedConfig, billConfig.UserId).Error
			assert.NoError(t, err)
			assert.Equal(t, updatedConfig.IgnoreUnmappedCategory, billConfig.IgnoreUnmappedCategory)
			assert.Equal(t, updatedConfig.CheckSameTransMode, billConfig.CheckSameTransMode)
		},
	)
}
