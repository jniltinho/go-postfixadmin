// Package auth provides JWT token generation and validation for go-postfixadmin.
// It supports two token types ("admin" and "mailbox") and embeds RBAC
// permissions and roles in the access token so that middleware can enforce
// authorization without a per-request database round-trip.
package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

var (
	// ErrInvalidToken is returned when a token cannot be parsed or has expired.
	ErrInvalidToken = errors.New("invalid or expired token")
	// ErrWrongType is returned when a token's type claim does not match the expected value.
	ErrWrongType = errors.New("token type mismatch")
)

// Claims represents the custom JWT payload used for both admin and mailbox tokens.
//
// Permissions contains the permission names (e.g. "mailboxes:write") embedded at
// token-issue time so that [RequirePermission] middleware needs no DB round-trip.
// A single "*" entry in Permissions grants all permissions (superadmin fast-path).
//
// Roles lists the RBAC role names assigned to the user (e.g. "domain_admin").
// Both fields are empty for mailbox users and for admins with no RBAC roles when
// rbac.enabled=false (backward-compat mode).
type Claims struct {
	Username    string   `json:"username"`
	Superadmin  bool     `json:"superadmin"`
	Domains     []string `json:"domains"`
	Type        string   `json:"type"` // "admin" | "mailbox"
	Permissions []string `json:"permissions,omitempty"`
	Roles       []string `json:"roles,omitempty"`
	jwt.RegisteredClaims
}

// TokenParams holds all data required to sign an access or refresh token.
// Using a struct avoids a 6-parameter function signature and makes call sites
// self-documenting.
type TokenParams struct {
	Username    string
	Superadmin  bool
	Domains     []string
	Type        string   // "admin" | "mailbox"
	Permissions []string // resolved RBAC permissions; nil is safe (treated as empty)
	Roles       []string // resolved RBAC role names; nil is safe
}

// jwtConfig holds the loaded JWT configuration (read once at startup via viper).
type jwtConfig struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

var cfg jwtConfig

// initConfig loads JWT settings from viper. Called lazily on first use.
// Falls back to session_secret when jwt_secret is not explicitly set.
func initConfig() {
	secret := viper.GetString("server.jwt_secret")
	if secret == "" {
		secret = viper.GetString("server.session_secret")
	}
	if secret == "" {
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

// getSecret returns the signing secret, initializing config on first call.
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

// GenerateAccessToken creates a short-lived JWT access token from p.
// The token embeds the resolved RBAC permissions and roles so that downstream
// middleware can authorize requests without a database query.
func GenerateAccessToken(p TokenParams) (string, error) {
	now := time.Now()
	claims := Claims{
		Username:    p.Username,
		Superadmin:  p.Superadmin,
		Domains:     p.Domains,
		Type:        p.Type,
		Permissions: p.Permissions,
		Roles:       p.Roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "go-postfixadmin",
			Subject:   p.Username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(GetAccessTTL())),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getSecret())
}

// GenerateRefreshToken creates a long-lived JWT refresh token stored in an
// httpOnly cookie. The refresh token carries the same claims as the access token
// so that the refresh handler can re-derive permissions from the database and
// issue a fresh access token without requiring the user to log in again.
func GenerateRefreshToken(p TokenParams) (string, error) {
	now := time.Now()
	claims := Claims{
		Username:   p.Username,
		Superadmin: p.Superadmin,
		Domains:    p.Domains,
		Type:       p.Type,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "go-postfixadmin",
			Subject:   p.Username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(GetRefreshTTL())),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getSecret())
}

// ValidateToken parses and validates a JWT string. Returns the custom Claims
// if the token is well-formed, signed with the correct secret, and not expired.
func ValidateToken(tokenString string) (*Claims, error) {
	if len(cfg.secret) == 0 {
		initConfig()
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
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

// ValidateTokenOfType is a convenience wrapper that also enforces the expected
// token type ("admin" or "mailbox"). Returns [ErrWrongType] on a type mismatch.
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
