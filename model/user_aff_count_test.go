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

func TestUserEditWithInviter_UpdatesInviterAndCounts(t *testing.T) {
	truncateTables(t)

	require.NoError(t, DB.Create(&User{
		Id:        20,
		Username:  "edit_inviter_target",
		Status:    common.UserStatusEnabled,
		AffCode:   "edit_inviter_target",
		InviterId: 21,
	}).Error)
	require.NoError(t, DB.Create(&User{
		Id:       21,
		Username: "edit_inviter_old",
		Status:   common.UserStatusEnabled,
		AffCode:  "edit_inviter_old",
		AffCount: 1,
	}).Error)
	require.NoError(t, DB.Create(&User{
		Id:       22,
		Username: "edit_inviter_new",
		Status:   common.UserStatusEnabled,
		AffCode:  "edit_inviter_new",
	}).Error)

	user := &User{
		Id:          20,
		Username:    "edit_inviter_target",
		DisplayName: "edit_inviter_target",
		Group:       "default",
		InviterId:   22,
	}
	require.NoError(t, user.EditWithInviter(false, true))

	var target User
	require.NoError(t, DB.Select("inviter_id").Where("id = ?", 20).First(&target).Error)
	assert.Equal(t, 22, target.InviterId)

	oldCount, err := CountInvitedUsers(21)
	require.NoError(t, err)
	assert.Equal(t, 0, oldCount)
	newCount, err := CountInvitedUsers(22)
	require.NoError(t, err)
	assert.Equal(t, 1, newCount)

	var oldInviter User
	require.NoError(t, DB.Select("aff_count").Where("id = ?", 21).First(&oldInviter).Error)
	assert.Equal(t, 0, oldInviter.AffCount)
	var newInviter User
	require.NoError(t, DB.Select("aff_count").Where("id = ?", 22).First(&newInviter).Error)
	assert.Equal(t, 1, newInviter.AffCount)
}

func TestUserEditWithInviter_RejectsSelfAndCycle(t *testing.T) {
	truncateTables(t)

	require.NoError(t, DB.Create(&User{
		Id:        30,
		Username:  "cycle_user_30",
		Status:    common.UserStatusEnabled,
		AffCode:   "cycle_user_30",
		InviterId: 31,
	}).Error)
	require.NoError(t, DB.Create(&User{
		Id:        31,
		Username:  "cycle_user_31",
		Status:    common.UserStatusEnabled,
		AffCode:   "cycle_user_31",
		InviterId: 32,
	}).Error)
	require.NoError(t, DB.Create(&User{
		Id:       32,
		Username: "cycle_user_32",
		Status:   common.UserStatusEnabled,
		AffCode:  "cycle_user_32",
	}).Error)

	self := &User{
		Id:          30,
		Username:    "cycle_user_30",
		DisplayName: "cycle_user_30",
		Group:       "default",
		InviterId:   30,
	}
	require.Error(t, self.EditWithInviter(false, true))

	cycle := &User{
		Id:          32,
		Username:    "cycle_user_32",
		DisplayName: "cycle_user_32",
		Group:       "default",
		InviterId:   30,
	}
	require.Error(t, cycle.EditWithInviter(false, true))
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
