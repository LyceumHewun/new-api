package model

import (
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManualCompletePaidOrder_CompletesSubscriptionOrder(t *testing.T) {
	truncateTables(t)
	setInviteRebateSettingForTest(-1, []float64{0.1})
	setSubscriptionRebateGlobalsForTest(t, 1000, 7.25)

	insertInviteRebateUser(t, 401, 403)
	insertInviteRebateUser(t, 403, 0)
	plan := &SubscriptionPlan{
		Id:            801,
		Title:         "Manual Complete Plan",
		PriceAmount:   9.99,
		Currency:      "USD",
		DurationUnit:  SubscriptionDurationMonth,
		DurationValue: 1,
		Enabled:       true,
		TotalAmount:   1000,
	}
	require.NoError(t, DB.Create(plan).Error)

	order := &SubscriptionOrder{
		UserId:          401,
		PlanId:          plan.Id,
		Money:           9.99,
		TradeNo:         "sub-admin-complete",
		PaymentMethod:   PaymentProviderStripe,
		PaymentProvider: PaymentProviderStripe,
		Status:          common.TopUpStatusPending,
		CreateTime:      time.Now().Unix(),
	}
	require.NoError(t, order.InsertWithBill(plan))

	topUp := GetTopUpByTradeNo("sub-admin-complete")
	require.NotNil(t, topUp)
	assert.EqualValues(t, 0, topUp.Amount)
	assert.Equal(t, common.TopUpStatusPending, topUp.Status)

	require.NoError(t, ManualCompletePaidOrder("sub-admin-complete", "127.0.0.1"))

	order = GetSubscriptionOrderByTradeNo("sub-admin-complete")
	require.NotNil(t, order)
	assert.Equal(t, common.TopUpStatusSuccess, order.Status)
	assert.NotZero(t, order.CompleteTime)

	topUp = GetTopUpByTradeNo("sub-admin-complete")
	require.NotNil(t, topUp)
	assert.Equal(t, common.TopUpStatusSuccess, topUp.Status)
	assert.EqualValues(t, 0, topUp.Amount)
	assert.EqualValues(t, 1, countUserSubscriptionsForPaymentGuardTest(t, 401))
	assert.Equal(t, 0, getUserQuotaForPaymentGuardTest(t, 401))
	affQuota, affHistory := getInviteQuotaForTest(t, 403)
	assert.Equal(t, 999, affQuota)
	assert.Equal(t, 999, affHistory)

	var rebateRecord InviteRebateRecord
	require.NoError(t, DB.Where("source_id = ?", "sub-admin-complete").First(&rebateRecord).Error)
	assert.Equal(t, PaymentProviderStripe, rebateRecord.SourceType)
	assert.Equal(t, 9990, rebateRecord.BaseQuota)

	var topupLog Log
	require.NoError(t, DB.Where("user_id = ? AND content LIKE ?", 401, "管理员补单成功，订阅套餐:%").First(&topupLog).Error)
	assert.Equal(t, "127.0.0.1", topupLog.Ip)
}

func TestManualCompletePaidOrder_FallsBackToTopUp(t *testing.T) {
	truncateTables(t)
	originalQuotaPerUnit := common.QuotaPerUnit
	t.Cleanup(func() {
		common.QuotaPerUnit = originalQuotaPerUnit
	})
	common.QuotaPerUnit = 1000

	insertUserForPaymentGuardTest(t, 402, 0)
	insertTopUpForPaymentGuardTest(t, "topup-admin-complete", 402, PaymentProviderEpay)

	require.NoError(t, ManualCompletePaidOrder("topup-admin-complete", "127.0.0.1"))

	topUp := GetTopUpByTradeNo("topup-admin-complete")
	require.NotNil(t, topUp)
	assert.Equal(t, common.TopUpStatusSuccess, topUp.Status)
	assert.Equal(t, 2000, getUserQuotaForPaymentGuardTest(t, 402))
}
