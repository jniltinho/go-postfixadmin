package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"go-postfixadmin/internal/auth"
	"go-postfixadmin/internal/models"
	"go-postfixadmin/internal/rbac"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

// Context key for storing JWT claims (used by dual-mode getters).
type contextKey string

const (
	// JWTClaimsKey is the key used to store validated *auth.Claims in echo.Context.
	JWTClaimsKey contextKey = "jwt_claims"
)

// JWTAuthMiddleware returns an Echo middleware that validates a JWT Bearer token
// or a database API key supplied via the Authorization or X-API-Key header.
//
// On success it stores the parsed claims in both the standard request context and
// Echo's fast context store so that [GetJWTClaims], [GetUsername], and
// [GetIsSuperAdmin] can retrieve them without a database lookup.
//
// API key authentication resolves RBAC permissions from the database so that the
// resulting claims are equivalent to those produced by a normal JWT login.
func JWTAuthMiddleware(db *gorm.DB) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				authHeader = c.Request().Header.Get("X-API-Key")
			}
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing Authorization or X-API-Key header")
			}

			var tokenString string
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				tokenString = strings.TrimSpace(parts[1])
			} else {
				tokenString = strings.TrimSpace(authHeader)
			}

			if tokenString == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "empty token")
			}

			var claims *auth.Claims

			// 1. Try JWT validation first.
			claims, _ = auth.ValidateToken(tokenString)

			// 2. Fall back to database API key lookup.
			if claims == nil {
				claims = resolveAPIKeyClaims(db, tokenString)
			}

			if claims == nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
			}

			// Store claims in both context layers for transparent retrieval.
			ctx := context.WithValue(c.Request().Context(), JWTClaimsKey, claims)
			c.SetRequest(c.Request().WithContext(ctx))
			c.Set(string(JWTClaimsKey), claims)

			return next(c)
		}
	}
}

// resolveAPIKeyClaims looks up a raw API key token in the database, verifies
// the associated admin account is active, and builds Claims with resolved RBAC
// permissions. Returns nil when the token is unknown, expired, or the admin is
// inactive.
func resolveAPIKeyClaims(db *gorm.DB, tokenString string) *auth.Claims {
	var apiKey models.AdminApiKey
	if err := db.Where("token = ? AND active = ?", tokenString, true).First(&apiKey).Error; err != nil {
		return nil
	}

	if apiKey.ExpiresAt != nil && !apiKey.ExpiresAt.After(time.Now()) {
		return nil
	}

	var admin models.Admin
	if err := db.Where("username = ? AND active = ?", apiKey.Username, true).First(&admin).Error; err != nil {
		return nil
	}

	var domains []string
	if !admin.Superadmin {
		db.Model(&models.DomainAdmin{}).
			Where("username = ? AND active = ?", admin.Username, true).
			Pluck("domain", &domains)
	}

	perms, roles, err := rbac.ResolvePermissions(db, admin.Username, admin.Superadmin)
	if err != nil {
		perms = []string{}
		roles = []string{}
	}

	return &auth.Claims{
		Username:    admin.Username,
		Superadmin:  admin.Superadmin,
		Domains:     domains,
		Type:        "admin",
		Permissions: perms,
		Roles:       roles,
	}
}

// GetJWTClaims retrieves validated JWT claims from the Echo context.
// Returns nil if JWTAuthMiddleware has not processed the request.
func GetJWTClaims(c *echo.Context) *auth.Claims {
	// Try Echo's fast context store first (set by JWTAuthMiddleware).
	if v := c.Get(string(JWTClaimsKey)); v != nil {
		if claims, ok := v.(*auth.Claims); ok {
			return claims
		}
	}

	// Fallback to standard request context.
	if claims, ok := c.Request().Context().Value(JWTClaimsKey).(*auth.Claims); ok {
		return claims
	}
	return nil
}

// getIdentityFromJWT is an internal helper used by the dual-mode session getters.
// Returns username, isSuperAdmin, and whether a valid JWT identity was found.
func getIdentityFromJWT(c *echo.Context) (username string, isSuperAdmin bool, found bool) {
	claims := GetJWTClaims(c)
	if claims == nil {
		return "", false, false
	}
	return claims.Username, claims.Superadmin, true
}
