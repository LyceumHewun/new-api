package model

import (
	"time"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type VerificationCode struct {
	Id              int    `json:"id" gorm:"primaryKey"`
	VerificationKey string `json:"verification_key" gorm:"type:varchar(191);uniqueIndex:idx_verification_key_purpose"`
	Purpose         string `json:"purpose" gorm:"type:varchar(32);uniqueIndex:idx_verification_key_purpose"`
	Code            string `json:"code" gorm:"type:varchar(128)"`
	CreatedTime     int64  `json:"created_time" gorm:"index"`
}

func RegisterVerificationCodeWithKey(key string, code string, purpose string) error {
	removeExpiredVerificationCodes()
	record := VerificationCode{
		VerificationKey: key,
		Purpose:         purpose,
		Code:            code,
		CreatedTime:     time.Now().Unix(),
	}
	return DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "verification_key"},
			{Name: "purpose"},
		},
		DoUpdates: clause.AssignmentColumns([]string{"code", "created_time"}),
	}).Create(&record).Error
}

func VerifyCodeWithKey(key string, code string, purpose string) bool {
	var record VerificationCode
	err := DB.Where("verification_key = ? AND purpose = ?", key, purpose).First(&record).Error
	if err != nil {
		return false
	}
	if time.Now().Unix()-record.CreatedTime >= int64(common.VerificationValidMinutes*60) {
		_ = DB.Delete(&record).Error
		return false
	}
	return record.Code == code
}

func DeleteVerificationCodeWithKey(key string, purpose string) error {
	return DB.Where("verification_key = ? AND purpose = ?", key, purpose).Delete(&VerificationCode{}).Error
}

func removeExpiredVerificationCodes() {
	if DB == nil {
		return
	}
	expiredBefore := time.Now().Unix() - int64(common.VerificationValidMinutes*60)
	err := DB.Where("created_time <= ?", expiredBefore).Delete(&VerificationCode{}).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		common.SysLog("failed to delete expired verification codes: " + err.Error())
	}
}
