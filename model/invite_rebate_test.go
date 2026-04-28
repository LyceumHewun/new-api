package model

import (
	"fmt"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setInviteRebateSettingForTest(countLimit int, ratios []float64) {
	setting := operation_setting.GetInviteRebateSetting()
	setting.CountLimit = countLimit
	setting.ChainRatios = ratios
}

func insertInviteRebateUser(t *testing.T, id int, inviterID int) {
	t.Helper()
	require.NoError(t, DB.Create(&User{
		Id:        id,
		Username:  fmt.Sprintf("invite_rebate_user_%d", id),
		Status:    common.UserStatusEnabled,
		AffCode:   fmt.Sprintf("ir%d", id),
		InviterId: inviterID,
	}).Error)
}

func getInviteQuotaForTest(t *testing.T, id int) (int, int) {
	t.Helper()
	var user User
	require.NoError(t, DB.Select("aff_quota", "aff_history").Where("id = ?", id).First(&user).Error)
	return user.AffQuota, user.AffHistoryQuota
}

func TestApplyInviteRechargeRebateTx_ThreeLevelChain(t *testing.T) {
	truncateTables(t)
	setInviteRebateSettingForTest(-1, []float64{0.3, 0.2, 0.1})

	insertInviteRebateUser(t, 1, 2)
	insertInviteRebateUser(t, 2, 3)
	insertInviteRebateUser(t, 3, 4)
	insertInviteRebateUser(t, 4, 0)

	var records []InviteRebateRecord
	err := DB.Transaction(func(tx *gorm.DB) error {
		var err error
		records, err = ApplyInviteRechargeRebateTx(tx, 1, PaymentProviderEpay, "order-1", "order-1", 100000)
		return err
	})
	require.NoError(t, err)
	require.Len(t, records, 3)

	affQuota, affHistory := getInviteQuotaForTest(t, 2)
	assert.Equal(t, 30000, affQuota)
	assert.Equal(t, 30000, affHistory)
	affQuota, affHistory = getInviteQuotaForTest(t, 3)
	assert.Equal(t, 20000, affQuota)
	assert.Equal(t, 20000, affHistory)
	affQuota, affHistory = getInviteQuotaForTest(t, 4)
	assert.Equal(t, 10000, affQuota)
	assert.Equal(t, 10000, affHistory)
}

func TestApplyInviteRechargeRebateTx_CountLimitAndIdempotency(t *testing.T) {
	truncateTables(t)
	setInviteRebateSettingForTest(2, []float64{0.5})

	insertInviteRebateUser(t, 10, 11)
	insertInviteRebateUser(t, 11, 0)

	for _, sourceID := range []string{"order-1", "order-2"} {
		require.NoError(t, DB.Transaction(func(tx *gorm.DB) error {
			_, err := ApplyInviteRechargeRebateTx(tx, 10, PaymentProviderStripe, sourceID, sourceID, 1000)
			return err
		}))
	}
	require.NoError(t, DB.Transaction(func(tx *gorm.DB) error {
		_, err := ApplyInviteRechargeRebateTx(tx, 10, PaymentProviderStripe, "order-3", "order-3", 1000)
		return err
	}))
	require.NoError(t, DB.Transaction(func(tx *gorm.DB) error {
		_, err := ApplyInviteRechargeRebateTx(tx, 10, PaymentProviderStripe, "order-2", "order-2", 1000)
		return err
	}))

	affQuota, affHistory := getInviteQuotaForTest(t, 11)
	assert.Equal(t, 1000, affQuota)
	assert.Equal(t, 1000, affHistory)
	var count int64
	require.NoError(t, DB.Model(&InviteRebateRecord{}).Count(&count).Error)
	assert.EqualValues(t, 2, count)
}

func TestApplyInviteRechargeRebateTx_Disabled(t *testing.T) {
	truncateTables(t)
	setInviteRebateSettingForTest(0, []float64{0.5})

	insertInviteRebateUser(t, 20, 21)
	insertInviteRebateUser(t, 21, 0)

	require.NoError(t, DB.Transaction(func(tx *gorm.DB) error {
		_, err := ApplyInviteRechargeRebateTx(tx, 20, PaymentProviderCreem, "order-1", "order-1", 1000)
		return err
	}))

	affQuota, affHistory := getInviteQuotaForTest(t, 21)
	assert.Equal(t, 0, affQuota)
	assert.Equal(t, 0, affHistory)
}
