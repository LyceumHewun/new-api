package model

import (
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setSubscriptionRebateGlobalsForTest(t *testing.T, quotaPerUnit float64, exchangeRate float64) {
	t.Helper()
	originalQuotaPerUnit := common.QuotaPerUnit
	originalUSDExchangeRate := operation_setting.USDExchangeRate
	originalPrice := operation_setting.Price
	t.Cleanup(func() {
		common.QuotaPerUnit = originalQuotaPerUnit
		operation_setting.USDExchangeRate = originalUSDExchangeRate
		operation_setting.Price = originalPrice
	})
	common.QuotaPerUnit = quotaPerUnit
	operation_setting.USDExchangeRate = exchangeRate
	operation_setting.Price = exchangeRate
}

func insertSubscriptionRebateOrder(t *testing.T, tradeNo string, userID int, plan *SubscriptionPlan, money float64, provider string) {
	t.Helper()
	order := &SubscriptionOrder{
		UserId:          userID,
		PlanId:          plan.Id,
		Money:           money,
		TradeNo:         tradeNo,
		PaymentMethod:   provider,
		PaymentProvider: provider,
		Status:          common.TopUpStatusPending,
		CreateTime:      time.Now().Unix(),
	}
	require.NoError(t, order.InsertWithBill(plan))
}

func TestSubscriptionOrderInsertWithBill_CreatesPendingBillAndSnapshot(t *testing.T) {
	truncateTables(t)
	setSubscriptionRebateGlobalsForTest(t, 1000, 7.25)

	insertInviteRebateUser(t, 1, 0)
	plan := &SubscriptionPlan{
		Id:            701,
		Title:         "Bill Plan",
		PriceAmount:   10,
		Currency:      "USD",
		DurationUnit:  SubscriptionDurationMonth,
		DurationValue: 1,
		Enabled:       true,
		TotalAmount:   1000,
	}
	require.NoError(t, DB.Create(plan).Error)
	insertSubscriptionRebateOrder(t, "sub-bill-snapshot", 1, plan, 10, PaymentProviderStripe)

	order := GetSubscriptionOrderByTradeNo("sub-bill-snapshot")
	require.NotNil(t, order)
	assert.Equal(t, 10000, order.RebateBaseQuota)

	topUp := GetTopUpByTradeNo("sub-bill-snapshot")
	require.NotNil(t, topUp)
	assert.Equal(t, common.TopUpStatusPending, topUp.Status)
	assert.Equal(t, PaymentProviderStripe, topUp.PaymentProvider)
}

func TestCompleteSubscriptionOrder_AppliesInviteRebateFromEpayCNYPayment(t *testing.T) {
	truncateTables(t)
	setInviteRebateSettingForTest(-1, []float64{0.1})
	setSubscriptionRebateGlobalsForTest(t, 1000, 7.25)

	insertInviteRebateUser(t, 1, 2)
	insertInviteRebateUser(t, 2, 0)
	plan := &SubscriptionPlan{
		Id:            702,
		Title:         "USD Plan",
		PriceAmount:   10,
		Currency:      "USD",
		DurationUnit:  SubscriptionDurationMonth,
		DurationValue: 1,
		Enabled:       true,
		TotalAmount:   1000,
	}
	require.NoError(t, DB.Create(plan).Error)
	insertSubscriptionRebateOrder(t, "sub-rebate-epay", 1, plan, 72.5, PaymentProviderEpay)

	require.NoError(t, CompleteSubscriptionOrder("sub-rebate-epay", `{"money":"72.50"}`, PaymentProviderEpay, "alipay"))

	affQuota, affHistory := getInviteQuotaForTest(t, 2)
	assert.Equal(t, 1000, affQuota)
	assert.Equal(t, 1000, affHistory)

	var record InviteRebateRecord
	require.NoError(t, DB.Where("source_id = ?", "sub-rebate-epay").First(&record).Error)
	assert.Equal(t, 10000, record.BaseQuota)
	assert.Equal(t, 1000, record.RebateQuota)
	assert.Equal(t, PaymentProviderEpay, record.SourceType)

	require.NoError(t, CompleteSubscriptionOrder("sub-rebate-epay", `{"money":"72.50"}`, PaymentProviderEpay, "alipay"))

	var count int64
	require.NoError(t, DB.Model(&InviteRebateRecord{}).Where("source_id = ?", "sub-rebate-epay").Count(&count).Error)
	assert.EqualValues(t, 1, count)
	affQuota, affHistory = getInviteQuotaForTest(t, 2)
	assert.Equal(t, 1000, affQuota)
	assert.Equal(t, 1000, affHistory)
}

func TestCompleteSubscriptionOrder_UsesOrderRebateSnapshot(t *testing.T) {
	truncateTables(t)
	setInviteRebateSettingForTest(-1, []float64{0.1})
	setSubscriptionRebateGlobalsForTest(t, 1000, 7.25)

	insertInviteRebateUser(t, 11, 12)
	insertInviteRebateUser(t, 12, 0)
	plan := &SubscriptionPlan{
		Id:            703,
		Title:         "Snapshot Plan",
		PriceAmount:   10,
		Currency:      "USD",
		DurationUnit:  SubscriptionDurationMonth,
		DurationValue: 1,
		Enabled:       true,
		TotalAmount:   1000,
	}
	require.NoError(t, DB.Create(plan).Error)
	insertSubscriptionRebateOrder(t, "sub-rebate-snapshot", 11, plan, 10, PaymentProviderStripe)

	order := GetSubscriptionOrderByTradeNo("sub-rebate-snapshot")
	require.NotNil(t, order)
	require.Equal(t, 10000, order.RebateBaseQuota)

	setSubscriptionRebateGlobalsForTest(t, 1000, 10)
	require.NoError(t, DB.Model(&SubscriptionPlan{}).Where("id = ?", plan.Id).Updates(map[string]interface{}{
		"currency": "CNY",
	}).Error)
	InvalidateSubscriptionPlanCache(plan.Id)

	require.NoError(t, CompleteSubscriptionOrder("sub-rebate-snapshot", `{"provider":"stripe"}`, PaymentProviderStripe, "stripe"))

	var record InviteRebateRecord
	require.NoError(t, DB.Where("source_id = ?", "sub-rebate-snapshot").First(&record).Error)
	assert.Equal(t, 10000, record.BaseQuota)
	assert.Equal(t, 1000, record.RebateQuota)

	topUp := GetTopUpByTradeNo("sub-rebate-snapshot")
	require.NotNil(t, topUp)
	assert.Equal(t, common.TopUpStatusSuccess, topUp.Status)
}
