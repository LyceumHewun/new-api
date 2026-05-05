package operation_setting

import (
	"errors"
	"fmt"

	"github.com/QuantumNous/new-api/setting/config"
)

type InviteRebateGroupSetting struct {
	CountLimit  int       `json:"count_limit"`
	ChainRatios []float64 `json:"chain_ratios"`
}

type InviteRebateSetting struct {
	CountLimit    int                                 `json:"count_limit"`
	ChainRatios   []float64                           `json:"chain_ratios"`
	GroupSettings map[string]InviteRebateGroupSetting `json:"group_settings"`
	MaxChainDepth int                                 `json:"max_chain_depth"`
}

var inviteRebateSetting = InviteRebateSetting{
	CountLimit:    0,
	ChainRatios:   []float64{},
	GroupSettings: map[string]InviteRebateGroupSetting{},
	MaxChainDepth: 0,
}

func init() {
	config.GlobalConfig.Register("invite_rebate_setting", &inviteRebateSetting)
}

func GetInviteRebateSetting() *InviteRebateSetting {
	return &inviteRebateSetting
}

func NormalizeInviteRebateSetting(setting *InviteRebateSetting) {
	if setting == nil {
		return
	}
	if setting.ChainRatios == nil {
		setting.ChainRatios = []float64{}
	}
	if setting.GroupSettings == nil {
		setting.GroupSettings = map[string]InviteRebateGroupSetting{}
	}
	setting.MaxChainDepth = CalculateInviteRebateMaxChainDepth(setting)
}

func CalculateInviteRebateMaxChainDepth(setting *InviteRebateSetting) int {
	if setting == nil {
		return 0
	}
	maxDepth := 0
	if setting.CountLimit != 0 {
		maxDepth = len(setting.ChainRatios)
	}
	for _, groupSetting := range setting.GroupSettings {
		if groupSetting.CountLimit != 0 && len(groupSetting.ChainRatios) > maxDepth {
			maxDepth = len(groupSetting.ChainRatios)
		}
	}
	return maxDepth
}

func ValidateInviteRebateRatios(ratios []float64) error {
	for _, ratio := range ratios {
		if ratio < 0 {
			return errors.New("返现比例不能为负数")
		}
	}
	return nil
}

func ValidateInviteRebateGroupSettings(groupSettings map[string]InviteRebateGroupSetting) error {
	for group, groupSetting := range groupSettings {
		if group == "" {
			return errors.New("用户分组不能为空")
		}
		if groupSetting.CountLimit < -1 {
			return fmt.Errorf("用户分组 %s 的返现次数必须为 -1、0 或正整数", group)
		}
		if err := ValidateInviteRebateRatios(groupSetting.ChainRatios); err != nil {
			return fmt.Errorf("用户分组 %s 的%s", group, err.Error())
		}
	}
	return nil
}

func SelectInviteRebateRule(setting *InviteRebateSetting, group string, level int) (countLimit int, ratio float64, ok bool) {
	if setting == nil || level <= 0 {
		return 0, 0, false
	}
	if setting.GroupSettings != nil {
		if groupSetting, found := setting.GroupSettings[group]; found {
			if groupSetting.CountLimit == 0 || len(groupSetting.ChainRatios) < level {
				return groupSetting.CountLimit, 0, false
			}
			return groupSetting.CountLimit, groupSetting.ChainRatios[level-1], true
		}
	}
	if setting.CountLimit == 0 || len(setting.ChainRatios) < level {
		return setting.CountLimit, 0, false
	}
	return setting.CountLimit, setting.ChainRatios[level-1], true
}

func EffectiveInviteRebateMaxChainDepth(setting *InviteRebateSetting) int {
	if setting == nil {
		return 0
	}
	actual := CalculateInviteRebateMaxChainDepth(setting)
	if setting.MaxChainDepth != actual {
		setting.MaxChainDepth = actual
	}
	return actual
}

func HasPositiveInviteRebateCountLimit(setting *InviteRebateSetting) bool {
	if setting == nil {
		return false
	}
	if setting.CountLimit > 0 {
		return true
	}
	for _, groupSetting := range setting.GroupSettings {
		if groupSetting.CountLimit > 0 {
			return true
		}
	}
	return false
}
