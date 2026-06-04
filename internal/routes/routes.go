// Package routes wires all HTTP routes for the go-postfixadmin API server.
// Every protected endpoint is guarded by [middleware.JWTAuthMiddleware] and,
// when rbac.enabled=true, by a [middleware.RequirePermission] call that enforces
// the appropriate permission from the RBAC catalog.
//
// When rbac.enabled=false (the default), RequirePermission is a no-op and all
// existing authorization behaviour is preserved unchanged.
package routes

import (
	"io/fs"
	"log/slog"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"go-postfixadmin/internal/handlers"
	"go-postfixadmin/internal/middleware"
	"go-postfixadmin/internal/rbac"

	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/labstack/echo/v5"
	"github.com/spf13/viper"
)

// AppVersion stores the application version served by the /version endpoint.
var AppVersion string = "1.0.0"

// RegisterRoutes registers all application routes on e.
// spaFS is the embedded Vue SPA filesystem; pass nil to disable SPA serving.
func RegisterRoutes(e *echo.Echo, h *handlers.Handler, spaFS fs.FS) {
	loginLimiter := middleware.LoginRateLimiter()
	apiLimiter := middleware.APIRateLimiter()

	e.GET("/lang/:code", h.SetLanguage)

	apiV1 := e.Group("/api/v1", apiLimiter.Middleware())

	// Public: app version (no auth required)
	apiV1.GET("/version", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"version": AppVersion})
	})

	// Public auth endpoints — CSRF check + strict rate limit (5 req/min per IP)
	authPublic := apiV1.Group("", middleware.CSRFOriginCheck(), loginLimiter.Middleware())
	authPublic.POST("/auth/login", h.AuthLogin)
	authPublic.POST("/auth/user-login", h.UserPortalLogin)
	authPublic.POST("/auth/refresh", h.AuthRefresh)

	// All protected endpoints require a valid JWT or API key.
	protected := apiV1.Group("", middleware.JWTAuthMiddleware(h.DB))

	protected.POST("/auth/logout", h.AuthLogout)
	protected.GET("/auth/me", h.AuthMe)

	// User portal endpoints (virtual mailbox users — no RBAC permission required)
	protected.GET("/user/me", h.GetUserProfile)
	protected.GET("/user/forwarding", h.GetUserForwarding)
	protected.POST("/user/forwarding", h.UpdateUserForwardingAPI)
	protected.POST("/user/password", h.UpdateUserPasswordAPI)
	protected.GET("/user/vacation", h.GetUserVacation)
	protected.POST("/user/vacation", h.UpdateUserVacationAPI)
	protected.DELETE("/user/vacation", h.DeleteUserVacationAPI)

	// Domains
	protected.GET("/domains", h.ListDomainsV1,
		middleware.RequirePermission(rbac.PermDomainsRead))
	protected.POST("/domains", h.CreateDomainV1,
		middleware.RequirePermission(rbac.PermDomainAddWrite))
	protected.GET("/domains/:domain", h.GetDomainV1,
		middleware.RequirePermission(rbac.PermDomainsRead))
	protected.PUT("/domains/:domain", h.UpdateDomainV1,
		middleware.RequirePermission(rbac.PermDomainsWrite))
	protected.DELETE("/domains/:domain", h.DeleteDomainV1,
		middleware.RequirePermission(rbac.PermDomainsDelete))

	// Mailboxes
	protected.GET("/mailboxes", h.ListMailboxesV1,
		middleware.RequirePermission(rbac.PermMailboxesRead))
	protected.POST("/mailboxes", h.CreateMailboxV1,
		middleware.RequirePermission(rbac.PermMailboxesWrite))
	protected.GET("/mailboxes/:username", h.GetMailboxV1,
		middleware.RequirePermission(rbac.PermMailboxesRead))
	protected.PUT("/mailboxes/:username", h.UpdateMailboxV1,
		middleware.RequirePermission(rbac.PermMailboxesWrite))
	protected.DELETE("/mailboxes/:username", h.DeleteMailboxV1,
		middleware.RequirePermission(rbac.PermMailboxesDelete))

	// Aliases
	protected.GET("/aliases", h.ListAliasesV1,
		middleware.RequirePermission(rbac.PermAliasesRead))
	protected.POST("/aliases", h.CreateAliasV1,
		middleware.RequirePermission(rbac.PermAliasesWrite))
	protected.GET("/aliases/:address", h.GetAliasV1,
		middleware.RequirePermission(rbac.PermAliasesRead))
	protected.PUT("/aliases/:address", h.UpdateAliasV1,
		middleware.RequirePermission(rbac.PermAliasesWrite))
	protected.DELETE("/aliases/:address", h.DeleteAliasV1,
		middleware.RequirePermission(rbac.PermAliasesDelete))

	// Alias domains
	protected.GET("/alias-domains", h.ListAliasDomainsV1,
		middleware.RequirePermission(rbac.PermAliasDomainsRead))
	protected.POST("/alias-domains", h.CreateAliasDomainV1,
		middleware.RequirePermission(rbac.PermAliasDomainsWrite))
	protected.GET("/alias-domains/:alias_domain", h.GetAliasDomainV1,
		middleware.RequirePermission(rbac.PermAliasDomainsRead))
	protected.PUT("/alias-domains/:alias_domain", h.UpdateAliasDomainV1,
		middleware.RequirePermission(rbac.PermAliasDomainsWrite))
	protected.DELETE("/alias-domains/:alias_domain", h.DeleteAliasDomainV1,
		middleware.RequirePermission(rbac.PermAliasDomainsDelete))

	// Admins
	// List and create require full admins:read/write (superadmin-level).
	// Get and update also accept profile:read/write so that a regular admin
	// can view and edit their own account — the handler enforces self-access only.
	protected.GET("/admins", h.ListAdminsV1,
		middleware.RequirePermission(rbac.PermAdminsRead))
	protected.POST("/admins", h.CreateAdminV1,
		middleware.RequirePermission(rbac.PermAdminsWrite))
	protected.GET("/admins/:username", h.GetAdminV1,
		middleware.RequirePermission(rbac.PermAdminsRead, rbac.PermProfileRead))
	protected.PUT("/admins/:username", h.UpdateAdminV1,
		middleware.RequirePermission(rbac.PermAdminsWrite, rbac.PermProfileWrite))
	protected.DELETE("/admins/:username", h.DeleteAdminV1,
		middleware.RequirePermission(rbac.PermAdminsDelete))

	// Transports
	protected.GET("/transports", h.ListTransportsV1,
		middleware.RequirePermission(rbac.PermTransportsRead))
	protected.POST("/transports", h.CreateTransportV1,
		middleware.RequirePermission(rbac.PermTransportsWrite))
	protected.GET("/transports/:id", h.GetTransportV1,
		middleware.RequirePermission(rbac.PermTransportsRead))
	protected.PUT("/transports/:id", h.UpdateTransportV1,
		middleware.RequirePermission(rbac.PermTransportsWrite))
	protected.DELETE("/transports/:id", h.DeleteTransportV1,
		middleware.RequirePermission(rbac.PermTransportsDelete))

	// Dashboard and logs
	protected.GET("/dashboard", h.DashboardStatsV1,
		middleware.RequirePermission(rbac.PermDashboardRead))
	protected.GET("/logs", h.LogsV1,
		middleware.RequirePermission(rbac.PermLogsRead))
	protected.GET("/maillog", h.MailLogV1,
		middleware.RequirePermission(rbac.PermLogsRead))

	// API key management
	protected.GET("/settings/apikeys", h.ListApiKeys,
		middleware.RequirePermission(rbac.PermAPIKeysRead))
	protected.POST("/settings/apikeys", h.CreateApiKey,
		middleware.RequirePermission(rbac.PermAPIKeysWrite))
	protected.PUT("/settings/apikeys/:id", h.UpdateApiKey,
		middleware.RequirePermission(rbac.PermAPIKeysWrite))
	protected.DELETE("/settings/apikeys/:id", h.DeleteApiKey,
		middleware.RequirePermission(rbac.PermAPIKeysWrite))

	// RBAC management — superadmin or settings:write required (enforced inside handlers)
	protected.GET("/rbac/roles", h.ListRBACRoles)
	protected.POST("/rbac/roles", h.CreateRBACRole)
	protected.GET("/rbac/roles/:id", h.GetRBACRole)
	protected.PUT("/rbac/roles/:id", h.UpdateRBACRole)
	protected.DELETE("/rbac/roles/:id", h.DeleteRBACRole)
	protected.GET("/rbac/permissions", h.ListRBACPermissions)
	protected.GET("/rbac/admins/:username/roles", h.ListAdminRoles)
	protected.POST("/rbac/admins/:username/roles", h.AssignAdminRole)
	protected.DELETE("/rbac/admins/:username/roles/:assignmentId", h.RemoveAdminRole)

	if viper.GetBool("server.swagger_enable") {
		e.GET("/swagger/*", echo.WrapHandler(httpSwagger.WrapHandler))
	}

	if spaFS != nil {
		e.GET("/*", spaHandler(spaFS))
		slog.Info("Vue SPA frontend active at /")
	}
}

func spaHandler(spaFS fs.FS) echo.HandlerFunc {
	return func(c *echo.Context) error {
		urlPath := c.Request().URL.Path
		ext := strings.ToLower(filepath.Ext(urlPath))
		if ext != "" {
			fsPath := strings.TrimPrefix(urlPath, "/")
			data, err := fs.ReadFile(spaFS, fsPath)
			if err != nil {
				return echo.ErrNotFound
			}
			ct := mime.TypeByExtension(ext)
			if ct == "" {
				ct = "application/octet-stream"
			}
			if ext == ".js" || ext == ".mjs" {
				ct = "application/javascript; charset=utf-8"
			}
			if strings.HasPrefix(urlPath, "/assets/") {
				c.Response().Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			} else {
				c.Response().Header().Set("Cache-Control", "public, max-age=3600")
			}
			c.Response().Header().Set("Content-Type", ct)
			_, _ = c.Response().Write(data)
			return nil
		}
		indexHTML, err := fs.ReadFile(spaFS, "index.html")
		if err != nil {
			return echo.ErrNotFound
		}
		c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
		c.Response().Header().Set("Cache-Control", "no-cache")
		_, _ = c.Response().Write(indexHTML)
		return nil
	}
}
