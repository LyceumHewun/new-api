package model

import (
	"errors"
	"fmt"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func insertRedemptionUser(t *testing.T, id int, inviterID int) {
	t.Helper()
	require.NoError(t, DB.Create(&User{
		Id:        id,
		Username:  fmt.Sprintf("redemption_user_%d", id),
		Status:    common.UserStatusEnabled,
		AffCode:   fmt.Sprintf("redemption_user_%d", id),
		InviterId: inviterID,
	}).Error)
}

func TestRedeem_ReusableCodeAllowsDifferentUsersOnceEach(t *testing.T) {
	truncateTables(t)
	setInviteRebateSettingForTest(0, []float64{0.5})
	insertRedemptionUser(t, 1, 0)
	insertRedemptionUser(t, 2, 0)

	code := &Redemption{
		UserId:      1,
		Key:         "reuse-code",
		Status:      common.RedemptionCodeStatusEnabled,
		Name:        "reuse",
		Quota:       100,
		RemainCount: 2,
	}
	require.NoError(t, code.Insert())

	quota, err := Redeem("reuse-code", 1)
	require.NoError(t, err)
	assert.Equal(t, 100, quota)

	_, err = Redeem("reuse-code", 1)
	require.ErrorIs(t, err, ErrRedemptionAlreadyRedeemed)

	quota, err = Redeem("reuse-code", 2)
	require.NoError(t, err)
	assert.Equal(t, 100, quota)

	var reloaded Redemption
	require.NoError(t, DB.First(&reloaded, code.Id).Error)
	assert.Equal(t, 0, reloaded.RemainCount)
	assert.Equal(t, common.RedemptionCodeStatusUsed, reloaded.Status)
}

func TestRedeem_UnlimitedCodeDoesNotDecreaseRemainCount(t *testing.T) {
	truncateTables(t)
	setInviteRebateSettingForTest(0, []float64{0.5})
	insertRedemptionUser(t, 1, 0)
	insertRedemptionUser(t, 2, 0)

	code := &Redemption{
		UserId:      1,
		Key:         "unlimited-code",
		Status:      common.RedemptionCodeStatusEnabled,
		Name:        "unlimited",
		Quota:       100,
		RemainCount: -1,
	}
	require.NoError(t, code.Insert())

	_, err := Redeem("unlimited-code", 1)
	require.NoError(t, err)
	_, err = Redeem("unlimited-code", 2)
	require.NoError(t, err)

	var reloaded Redemption
	require.NoError(t, DB.First(&reloaded, code.Id).Error)
	assert.Equal(t, -1, reloaded.RemainCount)
	assert.Equal(t, common.RedemptionCodeStatusEnabled, reloaded.Status)
}

func TestRedeem_ExhaustedCodeReturnsPreciseError(t *testing.T) {
	truncateTables(t)
	insertRedemptionUser(t, 1, 0)
	code := &Redemption{
		UserId:      1,
		Key:         "empty-code",
		Status:      common.RedemptionCodeStatusEnabled,
		Name:        "empty",
		Quota:       100,
		RemainCount: 0,
	}
	require.NoError(t, code.Insert())

	_, err := Redeem("empty-code", 1)
	require.True(t, errors.Is(err, ErrRedemptionExhausted))
}

func TestRedeem_CanDisableInviteRebatePerCode(t *testing.T) {
	truncateTables(t)
	setInviteRebateSettingForTest(-1, []float64{0.5})
	insertRedemptionUser(t, 1, 2)
	insertRedemptionUser(t, 2, 0)

	code := &Redemption{
		UserId:              1,
		Key:                 "no-rebate-code",
		Status:              common.RedemptionCodeStatusEnabled,
		Name:                "no rebate",
		Quota:               1000,
		RemainCount:         1,
		DisableInviteRebate: true,
	}
	require.NoError(t, code.Insert())

	_, err := Redeem("no-rebate-code", 1)
	require.NoError(t, err)

	affQuota, affHistory := getInviteQuotaForTest(t, 2)
	assert.Equal(t, 0, affQuota)
	assert.Equal(t, 0, affHistory)
}

func TestBackfillRedemptionUsageRecords_PreservesOldRedeemer(t *testing.T) {
	truncateTables(t)
	setInviteRebateSettingForTest(0, []float64{0.5})
	insertRedemptionUser(t, 1, 0)

	code := &Redemption{
		UserId:       1,
		Key:          "old-used-code",
		Status:       common.RedemptionCodeStatusUsed,
		Name:         "old used",
		Quota:        100,
		RemainCount:  0,
		UsedUserId:   1,
		RedeemedTime: 1234,
	}
	require.NoError(t, code.Insert())
	require.NoError(t, backfillRedemptionUsageRecords())
	require.NoError(t, backfillRedemptionUsageRecords())

	var usageCount int64
	require.NoError(t, DB.Model(&RedemptionUsage{}).Where("redemption_id = ? AND user_id = ?", code.Id, 1).Count(&usageCount).Error)
	assert.EqualValues(t, 1, usageCount)

	require.NoError(t, DB.Model(&Redemption{}).Where("id = ?", code.Id).Updates(map[string]interface{}{
		"status":       common.RedemptionCodeStatusEnabled,
		"remain_count": 1,
	}).Error)

	_, err := Redeem("old-used-code", 1)
	require.ErrorIs(t, err, ErrRedemptionAlreadyRedeemed)
}
