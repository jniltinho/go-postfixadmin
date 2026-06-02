package middleware

import (
	"net/http"

	"go-postfixadmin/internal/rbac"

	"github.com/labstack/echo/v5"
	"github.com/spf13/viper"
)

// RequirePermission returns an Echo middleware that enforces at least one of the
// listed permission names on the authenticated request.
//
// Authorization logic (evaluated in order):
//  1. When rbac.enabled=false (default), the middleware is a no-op — all existing
//     behaviour is preserved without any role assignment.
//  2. Superadmin accounts (Superadmin=true in JWT claims) bypass all checks.
//  3. The wildcard permission "*" in Claims.Permissions also bypasses all checks.
//  4. The request is allowed if Claims.Permissions contains any of the listed perms.
//  5. Otherwise 403 Forbidden is returned.
//
// RequirePermission must be placed after [JWTAuthMiddleware] in the handler chain
// so that validated claims are already stored in the Echo context.
func RequirePermission(perms ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if !viper.GetBool("rbac.enabled") {
				return next(c)
			}

			claims := GetJWTClaims(c)
			if claims == nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
			}

			if claims.Superadmin {
				return next(c)
			}

			for _, perm := range perms {
				if rbac.HasPermission(claims.Permissions, perm) {
					return next(c)
				}
			}

			return echo.NewHTTPError(http.StatusForbidden, "insufficient permissions")
		}
	}
}

// RequireRole returns an Echo middleware that enforces at least one of the listed
// role names on the authenticated request.
//
// Like [RequirePermission], this middleware is a no-op when rbac.enabled=false,
// and superadmin accounts bypass all role checks.
func RequireRole(roles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if !viper.GetBool("rbac.enabled") {
				return next(c)
			}

			claims := GetJWTClaims(c)
			if claims == nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
			}

			if claims.Superadmin {
				return next(c)
			}

			if rbac.HasRole(claims.Roles, roles...) {
				return next(c)
			}

			return echo.NewHTTPError(http.StatusForbidden, "insufficient role")
		}
	}
}
