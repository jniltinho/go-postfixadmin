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

	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/labstack/echo/v5"
	"github.com/spf13/viper"
)

// AppVersion stores the application version to be served by the /version endpoint
var AppVersion string = "1.0.0"

// RegisterRoutes registers all the application routes
func RegisterRoutes(e *echo.Echo, h *handlers.Handler, spaFS fs.FS) {
	loginLimiter := middleware.LoginRateLimiter()
	apiLimiter := middleware.APIRateLimiter()

	e.GET("/lang/:code", h.SetLanguage)

	apiV1 := e.Group("/api/v1", apiLimiter.Middleware())

	// Public: app version
	apiV1.GET("/version", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"version": AppVersion})
	})

	// Public auth endpoints — CSRF check + strict rate limit (5 req/min per IP)
	authPublic := apiV1.Group("", middleware.CSRFOriginCheck(), loginLimiter.Middleware())
	authPublic.POST("/auth/login", h.AuthLogin)
	authPublic.POST("/auth/refresh", h.AuthRefresh)

	// All protected endpoints
	protected := apiV1.Group("", middleware.JWTAuthMiddleware(h.DB))

	protected.POST("/auth/logout", h.AuthLogout)
	protected.GET("/auth/me", h.AuthMe)

	// User portal endpoints (virtual mailbox users)
	protected.GET("/user/me", h.GetUserProfile)
	protected.GET("/user/forwarding", h.GetUserForwarding)
	protected.POST("/user/forwarding", h.UpdateUserForwardingAPI)
	protected.POST("/user/password", h.UpdateUserPasswordAPI)
	protected.GET("/user/vacation", h.GetUserVacation)
	protected.POST("/user/vacation", h.UpdateUserVacationAPI)
	protected.DELETE("/user/vacation", h.DeleteUserVacationAPI)

	protected.GET("/domains", h.ListDomainsV1)
	protected.POST("/domains", h.CreateDomainV1)
	protected.GET("/domains/:domain", h.GetDomainV1)
	protected.PUT("/domains/:domain", h.UpdateDomainV1)
	protected.DELETE("/domains/:domain", h.DeleteDomainV1)

	protected.GET("/mailboxes", h.ListMailboxesV1)
	protected.POST("/mailboxes", h.CreateMailboxV1)
	protected.GET("/mailboxes/:username", h.GetMailboxV1)
	protected.PUT("/mailboxes/:username", h.UpdateMailboxV1)
	protected.DELETE("/mailboxes/:username", h.DeleteMailboxV1)

	protected.GET("/aliases", h.ListAliasesV1)
	protected.POST("/aliases", h.CreateAliasV1)
	protected.GET("/aliases/:address", h.GetAliasV1)
	protected.PUT("/aliases/:address", h.UpdateAliasV1)
	protected.DELETE("/aliases/:address", h.DeleteAliasV1)

	protected.GET("/alias-domains", h.ListAliasDomainsV1)
	protected.POST("/alias-domains", h.CreateAliasDomainV1)
	protected.GET("/alias-domains/:alias_domain", h.GetAliasDomainV1)
	protected.PUT("/alias-domains/:alias_domain", h.UpdateAliasDomainV1)
	protected.DELETE("/alias-domains/:alias_domain", h.DeleteAliasDomainV1)

	protected.GET("/admins", h.ListAdminsV1)
	protected.POST("/admins", h.CreateAdminV1)
	protected.GET("/admins/:username", h.GetAdminV1)
	protected.PUT("/admins/:username", h.UpdateAdminV1)
	protected.DELETE("/admins/:username", h.DeleteAdminV1)

	protected.GET("/transports", h.ListTransportsV1)
	protected.POST("/transports", h.CreateTransportV1)
	protected.GET("/transports/:id", h.GetTransportV1)
	protected.PUT("/transports/:id", h.UpdateTransportV1)
	protected.DELETE("/transports/:id", h.DeleteTransportV1)

	protected.GET("/dashboard", h.DashboardStatsV1)
	protected.GET("/logs", h.LogsV1)
	protected.GET("/maillog", h.MailLogV1)

	// API Keys CRUD settings
	protected.GET("/settings/apikeys", h.ListApiKeys)
	protected.POST("/settings/apikeys", h.CreateApiKey)
	protected.PUT("/settings/apikeys/:id", h.UpdateApiKey)
	protected.DELETE("/settings/apikeys/:id", h.DeleteApiKey)

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
