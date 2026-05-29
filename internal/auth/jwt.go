package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

var (
	ErrInvalidToken = errors.New("invalid or expired token")
	ErrWrongType    = errors.New("token type mismatch")
)

// Claims represents the custom claims used for both admin and mailbox users.
// This structure is the foundation for the dual-portal JWT strategy described
// in the Migration Plan (DOCUMENTS/MIGRATION_PLAN_VUE3_QUASAR_JWT.md).
type Claims struct {
	Username   string   `json:"username"`
	Superadmin bool     `json:"superadmin"`
	Domains    []string `json:"domains"` // domains this identity can manage (empty for superadmin or mailbox users)
	Type       string   `json:"type"`    // "admin" | "mailbox"
	jwt.RegisteredClaims
}

// jwtConfig holds the loaded JWT configuration (read once at startup via viper).
type jwtConfig struct {
	secret      []byte
	accessTTL   time.Duration
	refreshTTL  time.Duration
}

var cfg jwtConfig

// initConfig loads JWT settings from viper. Called lazily on first use.
// Falls back to session_secret when jwt_secret is not explicitly set (smooth transition).
func initConfig() {
	secret := viper.GetString("server.jwt_secret")
	if secret == "" {
		secret = viper.GetString("server.session_secret")
	}
	if secret == "" {
		// Last resort (should never happen in production)
		secret = "insecure-dev-only-change-me"
	}
	cfg.secret = []byte(secret)

	access := viper.GetString("server.jwt_access_ttl")
	if access == "" {
		access = "15m"
	}
	refresh := viper.GetString("server.jwt_refresh_ttl")
	if refresh == "" {
		refresh = "168h"
	}

	var err error
	cfg.accessTTL, err = time.ParseDuration(access)
	if err != nil {
		cfg.accessTTL = 15 * time.Minute
	}
	cfg.refreshTTL, err = time.ParseDuration(refresh)
	if err != nil {
		cfg.refreshTTL = 7 * 24 * time.Hour
	}
}

// getSecret returns the signing secret (initializes config on first call).
func getSecret() []byte {
	if len(cfg.secret) == 0 {
		initConfig()
	}
	return cfg.secret
}

// GetAccessTTL returns the configured access token lifetime.
func GetAccessTTL() time.Duration {
	if cfg.accessTTL == 0 {
		initConfig()
	}
	return cfg.accessTTL
}

// GetRefreshTTL returns the configured refresh token lifetime.
func GetRefreshTTL() time.Duration {
	if cfg.refreshTTL == 0 {
		initConfig()
	}
	return cfg.refreshTTL
}

// GenerateAccessToken creates a short-lived JWT access token.
func GenerateAccessToken(username string, superadmin bool, domains []string, tokenType string) (string, error) {
	now := time.Now()
	claims := Claims{
		Username:   username,
		Superadmin: superadmin,
		Domains:    domains,
		Type:       tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "go-postfixadmin",
			Subject:   username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(GetAccessTTL())),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getSecret())
}

// GenerateRefreshToken creates a longer-lived JWT refresh token (for httpOnly cookie).
func GenerateRefreshToken(username string, superadmin bool, domains []string, tokenType string) (string, error) {
	now := time.Now()
	claims := Claims{
		Username:   username,
		Superadmin: superadmin,
		Domains:    domains,
		Type:       tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "go-postfixadmin",
			Subject:   username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(GetRefreshTTL())),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getSecret())
}

// ValidateToken parses and validates a JWT. Returns the custom Claims if valid.
func ValidateToken(tokenString string) (*Claims, error) {
	if len(cfg.secret) == 0 {
		initConfig()
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return cfg.secret, nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

// ValidateTokenOfType is a convenience wrapper that also enforces the expected token type ("admin" or "mailbox").
func ValidateTokenOfType(tokenString, expectedType string) (*Claims, error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.Type != expectedType {
		return nil, ErrWrongType
	}
	return claims, nil
}
