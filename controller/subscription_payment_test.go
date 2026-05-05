package controller

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupSubscriptionControllerTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	gin.SetMode(gin.TestMode)

	originalUsingSQLite := common.UsingSQLite
	originalUsingMySQL := common.UsingMySQL
	originalUsingPostgreSQL := common.UsingPostgreSQL
	originalRedisEnabled := common.RedisEnabled
	originalDB := model.DB
	originalLogDB := model.LOG_DB

	common.UsingSQLite = true
	common.UsingMySQL = false
	common.UsingPostgreSQL = false
	common.RedisEnabled = false

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	model.DB = db
	model.LOG_DB = db
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.SubscriptionPlan{}, &model.SubscriptionOrder{}))

	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
		model.DB = originalDB
		model.LOG_DB = originalLogDB
		common.UsingSQLite = originalUsingSQLite
		common.UsingMySQL = originalUsingMySQL
		common.UsingPostgreSQL = originalUsingPostgreSQL
		common.RedisEnabled = originalRedisEnabled
	})

	return db
}

func insertSubscriptionControllerTestUser(t *testing.T, id int, inviterId int) {
	t.Helper()
	require.NoError(t, model.DB.Create(&model.User{
		Id:        id,
		Username:  fmt.Sprintf("subscription_user_%d", id),
		Password:  "password",
		Status:    common.UserStatusEnabled,
		Group:     "default",
		AffCode:   fmt.Sprintf("sub_aff_%d", id),
		InviterId: inviterId,
	}).Error)
}

func insertSubscriptionControllerTestPlan(t *testing.T, id int, title string, salesAgentUserId int) model.SubscriptionPlan {
	t.Helper()
	plan := model.SubscriptionPlan{
		Id:               id,
		Title:            title,
		PriceAmount:      9.99,
		Currency:         "USD",
		DurationUnit:     model.SubscriptionDurationMonth,
		DurationValue:    1,
		Enabled:          true,
		TotalAmount:      1000,
		SalesAgentUserId: salesAgentUserId,
	}
	require.NoError(t, model.DB.Create(&plan).Error)
	model.InvalidateSubscriptionPlanCache(id)
	return plan
}

func performSubscriptionControllerJSONRequest(t *testing.T, handler gin.HandlerFunc, userId int, target string, body any) *httptest.ResponseRecorder {
	t.Helper()
	bodyBytes, err := common.Marshal(body)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, target, strings.NewReader(string(bodyBytes)))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Set("id", userId)

	handler(ctx)
	return recorder
}

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

func TestGetSubscriptionPlansFiltersSalesAgentExclusivePlans(t *testing.T) {
	setupSubscriptionControllerTestDB(t)

	insertSubscriptionControllerTestUser(t, 101, 0)
	insertSubscriptionControllerTestUser(t, 201, 101)
	insertSubscriptionControllerTestUser(t, 301, 0)

	publicPlan := insertSubscriptionControllerTestPlan(t, 1001, "Public Plan", 0)
	agentPlan := insertSubscriptionControllerTestPlan(t, 1002, "Agent Plan", 101)
	otherAgentPlan := insertSubscriptionControllerTestPlan(t, 1003, "Other Agent Plan", 301)

	testCases := []struct {
		name    string
		userId  int
		wantIds []int
	}{
		{name: "agent user", userId: 101, wantIds: []int{publicPlan.Id, agentPlan.Id}},
		{name: "direct invitee", userId: 201, wantIds: []int{publicPlan.Id, agentPlan.Id}},
		{name: "unrelated user", userId: 301, wantIds: []int{publicPlan.Id, otherAgentPlan.Id}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodGet, "/api/subscription/plans", nil)
			ctx.Set("id", tc.userId)

			GetSubscriptionPlans(ctx)

			require.Equal(t, http.StatusOK, recorder.Code)
			var payload struct {
				Success bool                  `json:"success"`
				Data    []SubscriptionPlanDTO `json:"data"`
			}
			require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &payload))
			require.True(t, payload.Success)

			gotIds := make(map[int]struct{}, len(payload.Data))
			for _, item := range payload.Data {
				gotIds[item.Plan.Id] = struct{}{}
			}
			require.Len(t, gotIds, len(tc.wantIds))
			for _, id := range tc.wantIds {
				require.Contains(t, gotIds, id)
			}
		})
	}
}

func TestSubscriptionPaymentRejectsInvisibleSalesAgentPlan(t *testing.T) {
	setupSubscriptionControllerTestDB(t)

	insertSubscriptionControllerTestUser(t, 401, 0)
	insertSubscriptionControllerTestUser(t, 402, 0)
	plan := insertSubscriptionControllerTestPlan(t, 2001, "Hidden Agent Plan", 401)

	testCases := []struct {
		name    string
		target  string
		handler gin.HandlerFunc
		body    any
	}{
		{
			name:    "epay",
			target:  "/api/subscription/epay/pay",
			handler: SubscriptionRequestEpay,
			body:    gin.H{"plan_id": plan.Id, "payment_method": "alipay"},
		},
		{
			name:    "stripe",
			target:  "/api/subscription/stripe/pay",
			handler: SubscriptionRequestStripePay,
			body:    gin.H{"plan_id": plan.Id},
		},
		{
			name:    "creem",
			target:  "/api/subscription/creem/pay",
			handler: SubscriptionRequestCreemPay,
			body:    gin.H{"plan_id": plan.Id},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := performSubscriptionControllerJSONRequest(t, tc.handler, 402, tc.target, tc.body)
			require.Equal(t, http.StatusOK, recorder.Code)

			var payload struct {
				Success bool   `json:"success"`
				Message string `json:"message"`
			}
			require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &payload))
			require.False(t, payload.Success)
			require.Equal(t, "套餐不可用", payload.Message)
		})
	}
}
