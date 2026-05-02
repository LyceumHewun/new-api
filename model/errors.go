package model

import "errors"

// Common errors
var (
	ErrDatabase = errors.New("database error")
)

// User auth errors
var (
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrUserEmptyCredentials = errors.New("empty credentials")
)

// Token auth errors
var (
	ErrTokenNotProvided = errors.New("token not provided")
	ErrTokenInvalid     = errors.New("token invalid")
)

// Redemption errors
var (
	ErrRedeemFailed                 = errors.New("redeem.failed")
	ErrRedemptionNotProvided        = errors.New("redemption.not_provided")
	ErrRedemptionInvalid            = errors.New("redemption.invalid")
	ErrRedemptionUsed               = errors.New("redemption.used")
	ErrRedemptionExpired            = errors.New("redemption.expired")
	ErrRedemptionExhausted          = errors.New("redemption.exhausted")
	ErrRedemptionAlreadyRedeemed    = errors.New("redemption.already_redeemed")
	ErrRedemptionRemainCountInvalid = errors.New("redemption.remain_count_invalid")
)

// 2FA errors
var ErrTwoFANotEnabled = errors.New("2fa not enabled")
