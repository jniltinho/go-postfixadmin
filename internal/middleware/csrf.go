package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/spf13/viper"
)

// CSRFOriginCheck is a lightweight CSRF defense for public endpoints (no JWT).
//
// Strategy:
//   - Skips safe methods (GET, HEAD, OPTIONS).
//   - Skips requests that carry a valid Authorization header — JWT Bearer tokens
//     are inherently CSRF-safe because browsers cannot set custom headers
//     cross-origin without a CORS preflight.
//   - For all other state-changing requests it validates that the Origin or
//     Referer header matches the configured server host (server.base_url in
//     config, falls back to the request Host).
func CSRFOriginCheck() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			method := c.Request().Method

			// Safe methods — no CSRF risk.
			if method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions {
				return next(c)
			}

			// JWT Bearer tokens make the request inherently CSRF-safe.
			if c.Request().Header.Get("Authorization") != "" {
				return next(c)
			}

			// application/json requires a CORS preflight for cross-origin requests,
			// so browsers cannot forge these — API clients are safe to pass through.
			ct := c.Request().Header.Get("Content-Type")
			if strings.HasPrefix(ct, "application/json") {
				return next(c)
			}

			allowed := allowedHost(c)
			origin := c.Request().Header.Get("Origin")
			referer := c.Request().Header.Get("Referer")

			if origin != "" {
				if hostMatches(origin, allowed) {
					return next(c)
				}
				return echo.NewHTTPError(http.StatusForbidden, "CSRF: origin not allowed")
			}

			if referer != "" {
				if hostMatches(referer, allowed) {
					return next(c)
				}
				return echo.NewHTTPError(http.StatusForbidden, "CSRF: referer not allowed")
			}

			// No Origin and no Referer on a state-changing public request is
			// suspicious. Browsers always include one of them for cross-site
			// requests, so blocking here is safe.
			return echo.NewHTTPError(http.StatusForbidden, "CSRF: missing origin header")
		}
	}
}

// allowedHost returns the expected host extracted from server.base_url config,
// falling back to the HTTP Host header when not configured.
func allowedHost(c *echo.Context) string {
	base := viper.GetString("server.base_url")
	if base == "" {
		base = c.Request().Host
	}
	return normalizeHost(base)
}

func normalizeHost(s string) string {
	s = strings.TrimPrefix(s, "https://")
	s = strings.TrimPrefix(s, "http://")
	s = strings.TrimSuffix(s, "/")
	return s
}

func hostMatches(headerValue, allowed string) bool {
	h := normalizeHost(headerValue)
	return strings.HasPrefix(h, allowed)
}
