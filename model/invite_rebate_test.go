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
	setting.GroupSettings = map[string]operation_setting.InviteRebateGroupSetting{}
	operation_setting.NormalizeInviteRebateSetting(setting)
}

func setInviteRebateGroupSettingForTest(groupSettings map[string]operation_setting.InviteRebateGroupSetting) {
	setting := operation_setting.GetInviteRebateSetting()
	setting.GroupSettings = groupSettings
	operation_setting.NormalizeInviteRebateSetting(setting)
}

func insertInviteRebateUser(t *testing.T, id int, inviterID int) {
	t.Helper()
	require.NoError(t, DB.Create(&User{
		Id:        id,
		Username:  fmt.Sprintf("invite_rebate_user_%d", id),
		Status:    common.UserStatusEnabled,
		Group:     "default",
		AffCode:   fmt.Sprintf("ir%d", id),
		InviterId: inviterID,
	}).Error)
}

func insertInviteRebateUserWithGroup(t *testing.T, id int, inviterID int, group string) {
	t.Helper()
	require.NoError(t, DB.Create(&User{
		Id:        id,
		Username:  fmt.Sprintf("invite_rebate_user_%d", id),
		Status:    common.UserStatusEnabled,
		Group:     group,
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

func TestApplyInviteRechargeRebateTx_CountLimitUsesDistinctSources(t *testing.T) {
	truncateTables(t)
	setInviteRebateSettingForTest(2, []float64{0, 0.5})

	insertInviteRebateUser(t, 30, 31)
	insertInviteRebateUser(t, 31, 32)
	insertInviteRebateUser(t, 32, 0)

	for _, sourceID := range []string{"order-1", "order-2", "order-3", "order-2"} {
		require.NoError(t, DB.Transaction(func(tx *gorm.DB) error {
			_, err := ApplyInviteRechargeRebateTx(tx, 30, PaymentProviderStripe, sourceID, sourceID, 1000)
			return err
		}))
	}

	affQuota, affHistory := getInviteQuotaForTest(t, 32)
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

func TestApplyInviteRechargeRebateTx_MissingInviterStopsChain(t *testing.T) {
	truncateTables(t)
	setInviteRebateSettingForTest(-1, []float64{0.5})

	insertInviteRebateUser(t, 30, 999)

	var records []InviteRebateRecord
	require.NoError(t, DB.Transaction(func(tx *gorm.DB) error {
		var err error
		records, err = ApplyInviteRechargeRebateTx(tx, 30, PaymentProviderEpay, "order-1", "order-1", 1000)
		return err
	}))
	require.Empty(t, records)

	var count int64
	require.NoError(t, DB.Model(&InviteRebateRecord{}).Count(&count).Error)
	assert.EqualValues(t, 0, count)
}

func TestApplyInviteRechargeRebateTx_GroupSettingUsesBeneficiaryGroupAndLevel(t *testing.T) {
	truncateTables(t)
	setInviteRebateSettingForTest(-1, []float64{0.25, 0.1})
	setInviteRebateGroupSettingForTest(map[string]operation_setting.InviteRebateGroupSetting{
		"vip": {
			CountLimit:  -1,
			ChainRatios: []float64{0.5, 0.2},
		},
	})

	insertInviteRebateUser(t, 100, 101)
	insertInviteRebateUser(t, 101, 102)
	insertInviteRebateUserWithGroup(t, 102, 0, "vip")

	var records []InviteRebateRecord
	require.NoError(t, DB.Transaction(func(tx *gorm.DB) error {
		var err error
		records, err = ApplyInviteRechargeRebateTx(tx, 100, PaymentProviderEpay, "order-group", "order-group", 1000)
		return err
	}))
	require.Len(t, records, 2)

	affQuota, _ := getInviteQuotaForTest(t, 101)
	assert.Equal(t, 250, affQuota)
	affQuota, _ = getInviteQuotaForTest(t, 102)
	assert.Equal(t, 200, affQuota)
}

func TestApplyInviteRechargeRebateTx_GroupSettingMissingLevelDoesNotFallback(t *testing.T) {
	truncateTables(t)
	setInviteRebateSettingForTest(-1, []float64{0.25, 0.1})
	setInviteRebateGroupSettingForTest(map[string]operation_setting.InviteRebateGroupSetting{
		"vip": {
			CountLimit:  -1,
			ChainRatios: []float64{0.5},
		},
	})

	insertInviteRebateUser(t, 110, 111)
	insertInviteRebateUser(t, 111, 112)
	insertInviteRebateUserWithGroup(t, 112, 0, "vip")

	var records []InviteRebateRecord
	require.NoError(t, DB.Transaction(func(tx *gorm.DB) error {
		var err error
		records, err = ApplyInviteRechargeRebateTx(tx, 110, PaymentProviderEpay, "order-missing-level", "order-missing-level", 1000)
		return err
	}))
	require.Len(t, records, 1)

	affQuota, _ := getInviteQuotaForTest(t, 111)
	assert.Equal(t, 250, affQuota)
	affQuota, _ = getInviteQuotaForTest(t, 112)
	assert.Equal(t, 0, affQuota)
}

func TestApplyInviteRechargeRebateTx_GroupCountLimitSkipsOnlyThatBeneficiary(t *testing.T) {
	truncateTables(t)
	setInviteRebateSettingForTest(-1, []float64{0.25, 0.1})
	setInviteRebateGroupSettingForTest(map[string]operation_setting.InviteRebateGroupSetting{
		"limited": {
			CountLimit:  1,
			ChainRatios: []float64{0.5, 0.2},
		},
	})

	insertInviteRebateUser(t, 120, 121)
	insertInviteRebateUserWithGroup(t, 121, 122, "limited")
	insertInviteRebateUser(t, 122, 0)

	for _, sourceID := range []string{"order-limit-1", "order-limit-2"} {
		require.NoError(t, DB.Transaction(func(tx *gorm.DB) error {
			_, err := ApplyInviteRechargeRebateTx(tx, 120, PaymentProviderEpay, sourceID, sourceID, 1000)
			return err
		}))
	}

	affQuota, _ := getInviteQuotaForTest(t, 121)
	assert.Equal(t, 500, affQuota)
	affQuota, _ = getInviteQuotaForTest(t, 122)
	assert.Equal(t, 200, affQuota)
}
