package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setInviteQuotaGlobalsForTest(t *testing.T, inviterQuota int, inviteeQuota int, newUserQuota int) {
	t.Helper()
	originalInviterQuota := common.QuotaForInviter
	originalInviteeQuota := common.QuotaForInvitee
	originalNewUserQuota := common.QuotaForNewUser
	t.Cleanup(func() {
		common.QuotaForInviter = originalInviterQuota
		common.QuotaForInvitee = originalInviteeQuota
		common.QuotaForNewUser = originalNewUserQuota
	})
	common.QuotaForInviter = inviterQuota
	common.QuotaForInvitee = inviteeQuota
	common.QuotaForNewUser = newUserQuota
}

func TestUserInsert_CountsInviteWhenInviterRewardIsZero(t *testing.T) {
	truncateTables(t)
	setInviteQuotaGlobalsForTest(t, 0, 0, 0)

	require.NoError(t, DB.Create(&User{
		Id:       1,
		Username: "invite_count_inviter",
		Status:   common.UserStatusEnabled,
		AffCode:  "invite_count_inviter",
	}).Error)

	invitee := &User{
		Username:  "invite_count_invitee",
		Status:    common.UserStatusEnabled,
		InviterId: 1,
	}
	require.NoError(t, invitee.Insert(1))

	var inviter User
	require.NoError(t, DB.Select("aff_count", "aff_quota", "aff_history").Where("id = ?", 1).First(&inviter).Error)
	assert.Equal(t, 1, inviter.AffCount)
	assert.Equal(t, 0, inviter.AffQuota)
	assert.Equal(t, 0, inviter.AffHistoryQuota)

	count, err := CountInvitedUsers(1)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestUserList_FillsAffCountFromInviterId(t *testing.T) {
	truncateTables(t)

	require.NoError(t, DB.Create(&User{
		Id:       10,
		Username: "stale_aff_inviter",
		Status:   common.UserStatusEnabled,
		AffCode:  "stale_aff_inviter",
		AffCount: 0,
	}).Error)
	require.NoError(t, DB.Create(&User{
		Id:        11,
		Username:  "stale_aff_invitee_1",
		Status:    common.UserStatusEnabled,
		AffCode:   "stale_aff_invitee_1",
		InviterId: 10,
	}).Error)
	require.NoError(t, DB.Create(&User{
		Id:        12,
		Username:  "stale_aff_invitee_2",
		Status:    common.UserStatusEnabled,
		AffCode:   "stale_aff_invitee_2",
		InviterId: 10,
	}).Error)

	users, _, err := GetAllUsers(&common.PageInfo{Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.Equal(t, 2, findUserAffCountForTest(t, users, 10))

	users, _, err = SearchUsers("stale_aff_inviter", "", 0, 10)
	require.NoError(t, err)
	require.Equal(t, 2, findUserAffCountForTest(t, users, 10))
}

func findUserAffCountForTest(t *testing.T, users []*User, userId int) int {
	t.Helper()
	for _, user := range users {
		if user != nil && user.Id == userId {
			return user.AffCount
		}
	}
	t.Fatalf("user %d not found", userId)
	return 0
}
