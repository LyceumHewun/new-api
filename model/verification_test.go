package model

import (
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerificationCodeUsesSharedStore(t *testing.T) {
	truncateTables(t)
	originalValidMinutes := common.VerificationValidMinutes
	common.VerificationValidMinutes = 10
	defer func() {
		common.VerificationValidMinutes = originalValidMinutes
	}()

	require.NoError(t, RegisterVerificationCodeWithKey("user@example.com", "123456", common.EmailVerificationPurpose))
	assert.True(t, VerifyCodeWithKey("user@example.com", "123456", common.EmailVerificationPurpose))
	assert.False(t, VerifyCodeWithKey("user@example.com", "000000", common.EmailVerificationPurpose))

	require.NoError(t, RegisterVerificationCodeWithKey("user@example.com", "654321", common.EmailVerificationPurpose))
	assert.False(t, VerifyCodeWithKey("user@example.com", "123456", common.EmailVerificationPurpose))
	assert.True(t, VerifyCodeWithKey("user@example.com", "654321", common.EmailVerificationPurpose))

	require.NoError(t, DeleteVerificationCodeWithKey("user@example.com", common.EmailVerificationPurpose))
	assert.False(t, VerifyCodeWithKey("user@example.com", "654321", common.EmailVerificationPurpose))
}

func TestVerificationCodeExpiresFromSharedStore(t *testing.T) {
	truncateTables(t)
	originalValidMinutes := common.VerificationValidMinutes
	common.VerificationValidMinutes = 10
	defer func() {
		common.VerificationValidMinutes = originalValidMinutes
	}()

	require.NoError(t, DB.Create(&VerificationCode{
		VerificationKey: "expired@example.com",
		Purpose:         common.EmailVerificationPurpose,
		Code:            "123456",
		CreatedTime:     time.Now().Unix() - int64(common.VerificationValidMinutes*60) - 1,
	}).Error)

	assert.False(t, VerifyCodeWithKey("expired@example.com", "123456", common.EmailVerificationPurpose))

	var count int64
	require.NoError(t, DB.Model(&VerificationCode{}).Where("verification_key = ?", "expired@example.com").Count(&count).Error)
	assert.Equal(t, int64(0), count)
}
