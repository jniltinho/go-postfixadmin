package handlers

import "regexp"

var (
	passwordUpperRegex   = regexp.MustCompile(`[A-Z]`)
	passwordLowerRegex   = regexp.MustCompile(`[a-z]`)
	passwordDigitRegex   = regexp.MustCompile(`[0-9]`)
	passwordSpecialRegex = regexp.MustCompile(`[^A-Za-z0-9]`)
)

// ValidatePassword validates the backend password policy and returns an
// English error message when invalid. It returns an empty string when valid.
func ValidatePassword(password string) string {
	if len(password) < 8 {
		return "Password must be at least 8 characters long"
	}
	if !passwordUpperRegex.MatchString(password) {
		return "Password must contain at least one uppercase letter"
	}
	if !passwordLowerRegex.MatchString(password) {
		return "Password must contain at least one lowercase letter"
	}
	if !passwordDigitRegex.MatchString(password) {
		return "Password must contain at least one number"
	}
	if !passwordSpecialRegex.MatchString(password) {
		return "Password must contain at least one special character"
	}

	return ""
}
