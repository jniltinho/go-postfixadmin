package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"go-postfixadmin/internal/auth"
	"go-postfixadmin/internal/models"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

// Context key for storing JWT claims (used by dual-mode getters).
type contextKey string

const (
	// JWTClaimsKey is the key used to store validated *auth.Claims in echo.Context.
	JWTClaimsKey contextKey = "jwt_claims"
)

// JWTAuthMiddleware returns an Echo middleware that validates a JWT Bearer token or DB ApiKey.
// On success, it stores the parsed claims in the request context so that
// GetUsername, GetIsSuperAdmin and future helpers can read them transparently.
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
			var err error

			// 1. Try to validate as a JWT first
			claims, err = auth.ValidateToken(tokenString)
			if err != nil {
				// 2. Fall back to validating as a database API Key / BearerAuth
				var apiKey models.AdminApiKey
				dbErr := db.Where("token = ? AND active = ?", tokenString, true).First(&apiKey).Error
				if dbErr == nil {
					// Check if expired
					if apiKey.ExpiresAt == nil || apiKey.ExpiresAt.After(time.Now()) {
						// Retrieve Admin to get Superadmin status and confirm active
						var admin models.Admin
						adminErr := db.Where("username = ? AND active = ?", apiKey.Username, true).First(&admin).Error
						if adminErr == nil {
							// Determine which domains this admin can manage
							var domains []string
							if !admin.Superadmin {
								// Load domains assigned to this domain admin
								db.Model(&models.DomainAdmin{}).
									Where("username = ? AND active = ?", admin.Username, true).
									Pluck("domain", &domains)
							}

							claims = &auth.Claims{
								Username:   admin.Username,
								Superadmin: admin.Superadmin,
								Domains:    domains,
								Type:       "admin",
							}
						}
					}
				}
			}

			if claims == nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
			}

			// Store claims in context for downstream handlers and unified getters.
			ctx := context.WithValue(c.Request().Context(), JWTClaimsKey, claims)
			c.SetRequest(c.Request().WithContext(ctx))

			// Also expose via Echo's context store for convenience (echo-specific).
			c.Set(string(JWTClaimsKey), claims)

			return next(c)
		}
	}
}

// GetJWTClaims retrieves validated JWT claims from the Echo context (if present).
// Returns nil if no valid JWT was processed by JWTAuthMiddleware.
func GetJWTClaims(c *echo.Context) *auth.Claims {
	// First try Echo's fast context store (set by our middleware).
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

// getIdentityFromJWT is an internal helper used by the dual-mode getters.
// It returns username, isSuperAdmin, and whether a valid JWT identity was found.
func getIdentityFromJWT(c *echo.Context) (username string, isSuperAdmin bool, found bool) {
	claims := GetJWTClaims(c)
	if claims == nil {
		return "", false, false
	}
	return claims.Username, claims.Superadmin, true
}
