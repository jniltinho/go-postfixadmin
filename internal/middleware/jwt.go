package middleware

import (
	"context"
	"net/http"
	"strings"

	"go-postfixadmin/internal/auth"

	"github.com/labstack/echo/v5"
)

// Context key for storing JWT claims (used by dual-mode getters).
type contextKey string

const (
	// JWTClaimsKey is the key used to store validated *auth.Claims in echo.Context.
	JWTClaimsKey contextKey = "jwt_claims"
)

// JWTAuthMiddleware returns an Echo middleware that validates a JWT Bearer token.
// On success, it stores the parsed claims in the request context so that
// GetUsername, GetIsSuperAdmin and future helpers can read them transparently.
//
// This middleware is intended for new API routes (/api/v1/*). It does NOT
// perform session fallback — that is handled by the extended getters in auth.go.
func JWTAuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing Authorization header")
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid Authorization header format (expected 'Bearer <token>')")
			}

			tokenString := strings.TrimSpace(parts[1])
			if tokenString == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "empty bearer token")
			}

			claims, err := auth.ValidateToken(tokenString)
			if err != nil {
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
