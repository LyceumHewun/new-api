package controller

import (
	"net/http"
	"strconv"
	"unicode/utf8"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-gonic/gin"
)

type redemptionRequest struct {
	Id                  int    `json:"id"`
	Name                string `json:"name"`
	Quota               int    `json:"quota"`
	ExpiredTime         int64  `json:"expired_time"`
	Count               int    `json:"count"`
	Status              int    `json:"status"`
	RemainCount         *int   `json:"remain_count"`
	DisableInviteRebate *bool  `json:"disable_invite_rebate"`
}

func GetAllRedemptions(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	redemptions, total, err := model.GetAllRedemptions(pageInfo.GetStartIdx(), pageInfo.GetPageSize())
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(redemptions)
	common.ApiSuccess(c, pageInfo)
	return
}

func SearchRedemptions(c *gin.Context) {
	keyword := c.Query("keyword")
	pageInfo := common.GetPageQuery(c)
	redemptions, total, err := model.SearchRedemptions(keyword, pageInfo.GetStartIdx(), pageInfo.GetPageSize())
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(redemptions)
	common.ApiSuccess(c, pageInfo)
	return
}

func GetRedemption(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	redemption, err := model.GetRedemptionById(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    redemption,
	})
	return
}

func AddRedemption(c *gin.Context) {
	req := redemptionRequest{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if utf8.RuneCountInString(req.Name) == 0 || utf8.RuneCountInString(req.Name) > 20 {
		common.ApiErrorI18n(c, i18n.MsgRedemptionNameLength)
		return
	}
	if req.Count <= 0 {
		common.ApiErrorI18n(c, i18n.MsgRedemptionCountPositive)
		return
	}
	if req.Count > 100 {
		common.ApiErrorI18n(c, i18n.MsgRedemptionCountMax)
		return
	}
	remainCount := 1
	if req.RemainCount != nil {
		remainCount = *req.RemainCount
	}
	if remainCount < -1 {
		common.ApiErrorI18n(c, i18n.MsgRedemptionRemainCountInvalid)
		return
	}
	disableInviteRebate := false
	if req.DisableInviteRebate != nil {
		disableInviteRebate = *req.DisableInviteRebate
	}
	if valid, msg := validateExpiredTime(c, req.ExpiredTime); !valid {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": msg})
		return
	}
	var keys []string
	for i := 0; i < req.Count; i++ {
		key := common.GetUUID()
		cleanRedemption := model.Redemption{
			UserId:              c.GetInt("id"),
			Name:                req.Name,
			Key:                 key,
			Status:              common.RedemptionCodeStatusEnabled,
			CreatedTime:         common.GetTimestamp(),
			Quota:               req.Quota,
			ExpiredTime:         req.ExpiredTime,
			RemainCount:         remainCount,
			DisableInviteRebate: disableInviteRebate,
		}
		if remainCount == 0 {
			cleanRedemption.Status = common.RedemptionCodeStatusUsed
		}
		err = cleanRedemption.Insert()
		if err != nil {
			common.SysError("failed to insert redemption: " + err.Error())
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": i18n.T(c, i18n.MsgRedemptionCreateFailed),
				"data":    keys,
			})
			return
		}
		keys = append(keys, key)
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    keys,
	})
	return
}

func DeleteRedemption(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	err := model.DeleteRedemptionById(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}

func UpdateRedemption(c *gin.Context) {
	statusOnly := c.Query("status_only")
	req := redemptionRequest{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	cleanRedemption, err := model.GetRedemptionById(req.Id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if statusOnly == "" {
		if valid, msg := validateExpiredTime(c, req.ExpiredTime); !valid {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": msg})
			return
		}
		// If you add more fields, please also update redemption.Update()
		cleanRedemption.Name = req.Name
		cleanRedemption.Quota = req.Quota
		cleanRedemption.ExpiredTime = req.ExpiredTime
		if req.RemainCount != nil {
			if *req.RemainCount < -1 {
				common.ApiErrorI18n(c, i18n.MsgRedemptionRemainCountInvalid)
				return
			}
			cleanRedemption.RemainCount = *req.RemainCount
			if cleanRedemption.RemainCount == 0 {
				cleanRedemption.Status = common.RedemptionCodeStatusUsed
			} else if cleanRedemption.Status == common.RedemptionCodeStatusUsed {
				cleanRedemption.Status = common.RedemptionCodeStatusEnabled
			}
		}
		if req.DisableInviteRebate != nil {
			cleanRedemption.DisableInviteRebate = *req.DisableInviteRebate
		}
	}
	if statusOnly != "" {
		cleanRedemption.Status = req.Status
	}
	err = cleanRedemption.Update()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    cleanRedemption,
	})
	return
}

func DeleteInvalidRedemption(c *gin.Context) {
	rows, err := model.DeleteInvalidRedemptions()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    rows,
	})
	return
}

func validateExpiredTime(c *gin.Context, expired int64) (bool, string) {
	if expired != 0 && expired < common.GetTimestamp() {
		return false, i18n.T(c, i18n.MsgRedemptionExpireTimeInvalid)
	}
	return true, ""
}
