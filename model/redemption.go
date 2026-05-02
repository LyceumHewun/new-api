package model

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Redemption struct {
	Id                  int            `json:"id"`
	UserId              int            `json:"user_id"`
	Key                 string         `json:"key" gorm:"type:char(32);uniqueIndex"`
	Status              int            `json:"status" gorm:"default:1"`
	Name                string         `json:"name" gorm:"index"`
	Quota               int            `json:"quota" gorm:"default:100"`
	CreatedTime         int64          `json:"created_time" gorm:"bigint"`
	RedeemedTime        int64          `json:"redeemed_time" gorm:"bigint"`
	Count               int            `json:"count" gorm:"-:all"` // only for api request
	RemainCount         int            `json:"remain_count"`
	DisableInviteRebate bool           `json:"disable_invite_rebate" gorm:"default:false"`
	UsedUserId          int            `json:"used_user_id"`
	DeletedAt           gorm.DeletedAt `gorm:"index"`
	ExpiredTime         int64          `json:"expired_time" gorm:"bigint"` // 过期时间，0 表示不过期
}

type RedemptionUsage struct {
	Id           int   `json:"id"`
	RedemptionId int   `json:"redemption_id" gorm:"uniqueIndex:idx_redemption_usage_user,priority:1"`
	UserId       int   `json:"user_id" gorm:"uniqueIndex:idx_redemption_usage_user,priority:2"`
	CreatedTime  int64 `json:"created_time" gorm:"bigint"`
}

func GetAllRedemptions(startIdx int, num int) (redemptions []*Redemption, total int64, err error) {
	// 开始事务
	tx := DB.Begin()
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 获取总数
	err = tx.Model(&Redemption{}).Count(&total).Error
	if err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	// 获取分页数据
	err = tx.Order("id desc").Limit(num).Offset(startIdx).Find(&redemptions).Error
	if err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	// 提交事务
	if err = tx.Commit().Error; err != nil {
		return nil, 0, err
	}

	return redemptions, total, nil
}

func SearchRedemptions(keyword string, startIdx int, num int) (redemptions []*Redemption, total int64, err error) {
	tx := DB.Begin()
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Build query based on keyword type
	query := tx.Model(&Redemption{})

	// Only try to convert to ID if the string represents a valid integer
	if id, err := strconv.Atoi(keyword); err == nil {
		query = query.Where("id = ? OR name LIKE ?", id, keyword+"%")
	} else {
		query = query.Where("name LIKE ?", keyword+"%")
	}

	// Get total count
	err = query.Count(&total).Error
	if err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	// Get paginated data
	err = query.Order("id desc").Limit(num).Offset(startIdx).Find(&redemptions).Error
	if err != nil {
		tx.Rollback()
		return nil, 0, err
	}

	if err = tx.Commit().Error; err != nil {
		return nil, 0, err
	}

	return redemptions, total, nil
}

func GetRedemptionById(id int) (*Redemption, error) {
	if id == 0 {
		return nil, errors.New("id 为空！")
	}
	redemption := Redemption{Id: id}
	var err error = nil
	err = DB.First(&redemption, "id = ?", id).Error
	return &redemption, err
}

func Redeem(key string, userId int) (quota int, err error) {
	if key == "" {
		return 0, ErrRedemptionNotProvided
	}
	if userId == 0 {
		return 0, errors.New("无效的 user id")
	}
	redemption := &Redemption{}
	var rebateRecords []InviteRebateRecord

	keyCol := "`key`"
	if common.UsingPostgreSQL {
		keyCol = `"key"`
	}
	common.RandomSleep()
	err = DB.Transaction(func(tx *gorm.DB) error {
		err := tx.Set("gorm:query_option", "FOR UPDATE").Where(keyCol+" = ?", key).First(redemption).Error
		if err != nil {
			return ErrRedemptionInvalid
		}
		if redemption.Status != common.RedemptionCodeStatusEnabled {
			if redemption.Status == common.RedemptionCodeStatusUsed {
				return ErrRedemptionExhausted
			}
			return ErrRedemptionUsed
		}
		if redemption.ExpiredTime != 0 && redemption.ExpiredTime < common.GetTimestamp() {
			return ErrRedemptionExpired
		}
		if redemption.RemainCount == 0 {
			return ErrRedemptionExhausted
		}
		now := common.GetTimestamp()
		usageResult := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&RedemptionUsage{
			RedemptionId: redemption.Id,
			UserId:       userId,
			CreatedTime:  now,
		})
		if usageResult.Error != nil {
			return usageResult.Error
		}
		if usageResult.RowsAffected == 0 {
			return ErrRedemptionAlreadyRedeemed
		}
		if redemption.RemainCount > 0 {
			result := tx.Model(&Redemption{}).
				Where("id = ? AND status = ? AND remain_count > 0", redemption.Id, common.RedemptionCodeStatusEnabled).
				Updates(map[string]interface{}{
					"remain_count":  gorm.Expr("remain_count - ?", 1),
					"redeemed_time": now,
					"used_user_id":  userId,
				})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return ErrRedemptionExhausted
			}
			if err := tx.Model(&Redemption{}).Where("id = ? AND remain_count = 0", redemption.Id).Update("status", common.RedemptionCodeStatusUsed).Error; err != nil {
				return err
			}
			redemption.RemainCount--
			if redemption.RemainCount == 0 {
				redemption.Status = common.RedemptionCodeStatusUsed
			}
		} else {
			result := tx.Model(&Redemption{}).
				Where("id = ? AND status = ? AND remain_count = -1", redemption.Id, common.RedemptionCodeStatusEnabled).
				Updates(map[string]interface{}{
					"redeemed_time": now,
					"used_user_id":  userId,
				})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return ErrRedemptionUsed
			}
		}
		redemption.RedeemedTime = now
		redemption.UsedUserId = userId
		err = tx.Model(&User{}).Where("id = ?", userId).Update("quota", gorm.Expr("quota + ?", redemption.Quota)).Error
		if err != nil {
			return err
		}
		if !redemption.DisableInviteRebate {
			sourceID, sourceTradeNo := redemptionRebateSource(redemption.Id)
			rebateRecords, err = ApplyInviteRechargeRebateTx(tx, userId, InviteRebateSourceRedemption, sourceID, sourceTradeNo, redemption.Quota)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		if isRedemptionUserError(err) {
			return 0, err
		}
		common.SysError("redemption failed: " + err.Error())
		return 0, ErrRedeemFailed
	}
	RecordLog(userId, LogTypeTopup, fmt.Sprintf("通过兑换码充值 %s，兑换码ID %d", logger.LogQuota(redemption.Quota), redemption.Id))
	RecordInviteRebateLogs(rebateRecords)
	return redemption.Quota, nil
}

func (redemption *Redemption) Insert() error {
	var err error
	err = DB.Select("*").Create(redemption).Error
	return err
}

func (redemption *Redemption) SelectUpdate() error {
	// This can update zero values
	return DB.Model(redemption).Select("redeemed_time", "status", "remain_count").Updates(redemption).Error
}

// Update Make sure your token's fields is completed, because this will update non-zero values
func (redemption *Redemption) Update() error {
	var err error
	err = DB.Model(redemption).Select("name", "status", "quota", "redeemed_time", "expired_time", "remain_count", "disable_invite_rebate").Updates(redemption).Error
	return err
}

func (redemption *Redemption) Delete() error {
	var err error
	err = DB.Delete(redemption).Error
	return err
}

func DeleteRedemptionById(id int) (err error) {
	if id == 0 {
		return errors.New("id 为空！")
	}
	redemption := Redemption{Id: id}
	err = DB.Where(redemption).First(&redemption).Error
	if err != nil {
		return err
	}
	return redemption.Delete()
}

func DeleteInvalidRedemptions() (int64, error) {
	now := common.GetTimestamp()
	result := DB.Where("status IN ? OR (status = ? AND expired_time != 0 AND expired_time < ?)", []int{common.RedemptionCodeStatusUsed, common.RedemptionCodeStatusDisabled}, common.RedemptionCodeStatusEnabled, now).Delete(&Redemption{})
	return result.RowsAffected, result.Error
}

func isRedemptionUserError(err error) bool {
	return errors.Is(err, ErrRedemptionNotProvided) ||
		errors.Is(err, ErrRedemptionInvalid) ||
		errors.Is(err, ErrRedemptionUsed) ||
		errors.Is(err, ErrRedemptionExpired) ||
		errors.Is(err, ErrRedemptionExhausted) ||
		errors.Is(err, ErrRedemptionAlreadyRedeemed)
}
