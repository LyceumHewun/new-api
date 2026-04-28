package operation_setting

import "github.com/QuantumNous/new-api/setting/config"

type InviteRebateSetting struct {
	CountLimit  int       `json:"count_limit"`
	ChainRatios []float64 `json:"chain_ratios"`
}

var inviteRebateSetting = InviteRebateSetting{
	CountLimit:  0,
	ChainRatios: []float64{},
}

func init() {
	config.GlobalConfig.Register("invite_rebate_setting", &inviteRebateSetting)
}

func GetInviteRebateSetting() *InviteRebateSetting {
	return &inviteRebateSetting
}
