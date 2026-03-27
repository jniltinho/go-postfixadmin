package mailbox

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	"go-postfixadmin/internal/utils"
)

func formatQuota(bytes int64) string {
	if bytes == 0 {
		return "0 B"
	}
	unit := utils.GetQuotaMultiplier()
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	val := fmt.Sprintf("%.1f", float64(bytes)/float64(div))
	return fmt.Sprintf("%s %cB", strings.TrimSuffix(val, ".0"), "KMGTPE"[exp])
}

func generateRandomPassword(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	b := make([]byte, length)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		b[i] = charset[num.Int64()]
	}
	return string(b), nil
}
