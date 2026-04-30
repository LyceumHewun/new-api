package controller

import (
	"testing"

	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/stretchr/testify/require"
)

func TestSubscriptionEpayMoneyUsesTopupPrice(t *testing.T) {
	originalRate := operation_setting.USDExchangeRate
	originalPrice := operation_setting.Price
	t.Cleanup(func() {
		operation_setting.USDExchangeRate = originalRate
		operation_setting.Price = originalPrice
	})

	operation_setting.USDExchangeRate = 99
	operation_setting.Price = 9

	require.Equal(t, "USD", normalizeSubscriptionCurrency(""))
	require.Equal(t, "CNY", normalizeSubscriptionCurrency(" cny "))

	usdPlan := &model.SubscriptionPlan{PriceAmount: 10, Currency: "USD"}
	require.InDelta(t, 90, subscriptionEpayMoney(usdPlan), 0.0001)

	cnyPlan := &model.SubscriptionPlan{PriceAmount: 10, Currency: "CNY"}
	require.InDelta(t, 10, subscriptionEpayMoney(cnyPlan), 0.0001)
}
