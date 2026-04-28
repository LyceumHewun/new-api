package model

import (
	"fmt"
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/setting/operation_setting"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const (
	InviteRebateSourceRedemption = "redemption"
)

type InviteRebateRecord struct {
	Id                int     `json:"id"`
	PayerUserId       int     `json:"payer_user_id" gorm:"index;uniqueIndex:idx_invite_rebate_source_level,priority:1"`
	BeneficiaryUserId int     `json:"beneficiary_user_id" gorm:"index;uniqueIndex:idx_invite_rebate_source_level,priority:4"`
	Level             int     `json:"level" gorm:"uniqueIndex:idx_invite_rebate_source_level,priority:5"`
	SourceType        string  `json:"source_type" gorm:"type:varchar(50);index;uniqueIndex:idx_invite_rebate_source_level,priority:2"`
	SourceId          string  `json:"source_id" gorm:"type:varchar(255);index;uniqueIndex:idx_invite_rebate_source_level,priority:3"`
	SourceTradeNo     string  `json:"source_trade_no" gorm:"type:varchar(255);index"`
	BaseQuota         int     `json:"base_quota"`
	Ratio             float64 `json:"ratio"`
	RebateQuota       int     `json:"rebate_quota"`
	CreatedAt         int64   `json:"created_at" gorm:"bigint;index"`
}

func ApplyInviteRechargeRebateTx(tx *gorm.DB, payerID int, sourceType, sourceID, sourceTradeNo string, baseQuota int) ([]InviteRebateRecord, error) {
	setting := operation_setting.GetInviteRebateSetting()
	if payerID <= 0 || baseQuota <= 0 || setting.CountLimit == 0 || len(setting.ChainRatios) == 0 {
		return nil, nil
	}
	if setting.CountLimit > 0 {
		var usedCount int64
		if err := tx.Model(&InviteRebateRecord{}).
			Where("payer_user_id = ? AND level = ?", payerID, 1).
			Count(&usedCount).Error; err != nil {
			return nil, err
		}
		if usedCount >= int64(setting.CountLimit) {
			return nil, nil
		}
	}

	var payer User
	if err := tx.Select("id", "inviter_id").Where("id = ?", payerID).First(&payer).Error; err != nil {
		return nil, err
	}

	now := common.GetTimestamp()
	inviterID := payer.InviterId
	records := make([]InviteRebateRecord, 0, len(setting.ChainRatios))
	for index, ratio := range setting.ChainRatios {
		if inviterID <= 0 {
			break
		}

		var inviter User
		if err := tx.Select("id", "inviter_id").Where("id = ?", inviterID).First(&inviter).Error; err != nil {
			return nil, err
		}

		level := index + 1
		rebateQuota := int(decimal.NewFromInt(int64(baseQuota)).Mul(decimal.NewFromFloat(ratio)).IntPart())
		if rebateQuota > 0 {
			var existing int64
			if err := tx.Model(&InviteRebateRecord{}).
				Where("payer_user_id = ? AND source_type = ? AND source_id = ? AND beneficiary_user_id = ? AND level = ?",
					payerID, sourceType, sourceID, inviter.Id, level).
				Count(&existing).Error; err != nil {
				return nil, err
			}
			if existing == 0 {
				record := InviteRebateRecord{
					PayerUserId:       payerID,
					BeneficiaryUserId: inviter.Id,
					Level:             level,
					SourceType:        sourceType,
					SourceId:          sourceID,
					SourceTradeNo:     sourceTradeNo,
					BaseQuota:         baseQuota,
					Ratio:             ratio,
					RebateQuota:       rebateQuota,
					CreatedAt:         now,
				}
				if err := tx.Create(&record).Error; err != nil {
					return nil, err
				}
				if err := tx.Model(&User{}).Where("id = ?", inviter.Id).Updates(map[string]interface{}{
					"aff_quota":   gorm.Expr("aff_quota + ?", rebateQuota),
					"aff_history": gorm.Expr("aff_history + ?", rebateQuota),
				}).Error; err != nil {
					return nil, err
				}
				records = append(records, record)
			}
		}

		inviterID = inviter.InviterId
	}
	return records, nil
}

func RecordInviteRebateLogs(records []InviteRebateRecord) {
	for _, record := range records {
		other := map[string]interface{}{
			"invite_rebate":   true,
			"payer_user_id":   record.PayerUserId,
			"level":           record.Level,
			"source_type":     record.SourceType,
			"source_id":       record.SourceId,
			"source_trade_no": record.SourceTradeNo,
			"base_quota":      record.BaseQuota,
			"ratio":           record.Ratio,
		}
		content := fmt.Sprintf("邀请返现：用户 %d 充值 %s，%d 级返现 %s，已进入邀请额度",
			record.PayerUserId,
			logger.LogQuota(record.BaseQuota),
			record.Level,
			logger.LogQuota(record.RebateQuota),
		)
		RecordLogWithQuotaAndOther(record.BeneficiaryUserId, LogTypeTopup, content, record.RebateQuota, other)
	}
}

func redemptionRebateSource(id int) (string, string) {
	sourceID := strconv.Itoa(id)
	return sourceID, "redemption-" + sourceID
}
