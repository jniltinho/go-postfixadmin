package utils

import "github.com/spf13/viper"

// GetQuotaMultiplier retrieves the quota multiplier from configuration.
// If it's missing or invalid, it defaults to 1024000.
func GetQuotaMultiplier() int64 {
	unit := viper.GetInt64("quota.multiplier")
	if unit <= 0 {
		unit = 1024000
	}
	return unit
}
