package routes

import (
	"net/http"

	"go-postfixadmin/internal/handlers"
	"go-postfixadmin/internal/middleware"

	"github.com/labstack/echo/v5"
)

// RegisterRoutes registers all the application routes
func RegisterRoutes(e *echo.Echo, h *handlers.Handler) {
	// Public Auth Routes (no middleware needed)
	e.GET("/login", h.Login)
	e.POST("/login", h.Login)
	e.GET("/logout", h.Logout)

	// Static files and utils (public)
	e.GET("/lang/:code", h.SetLanguage)

	// Protected Admin Routes
	adminGroup := e.Group("")
	adminGroup.Use(middleware.AuthMiddleware)

	// Dashboard
	adminGroup.GET("/dashboard", h.Dashboard)

	// Logs
	adminGroup.GET("/logs", h.Logs)
	adminGroup.GET("/api/logs", h.LogsData)

	// Maillog
	adminGroup.GET("/maillog", h.MailLog)
	adminGroup.GET("/api/maillog", h.MailLogData)

	// Domains
	adminGroup.GET("/domains", h.ListDomains)
	adminGroup.POST("/api/domains", h.AddDomainAPI)
	adminGroup.GET("/api/domains/:domain", h.GetDomainAPI)
	adminGroup.POST("/api/domains/:domain", h.EditDomainAPI)
	adminGroup.DELETE("/domains/delete/:domain", h.DeleteDomain)

	// Mailboxes
	adminGroup.GET("/mailboxes", h.ListMailboxes)
	adminGroup.POST("/api/mailboxes", h.AddMailboxAPI)
	adminGroup.GET("/api/mailboxes/:username", h.GetMailboxAPI)
	adminGroup.POST("/api/mailboxes/:username", h.EditMailboxAPI)
	adminGroup.DELETE("/mailboxes/delete/:username", h.DeleteMailbox)

	// Admins
	adminGroup.GET("/admins", h.ListAdmins)
	adminGroup.POST("/api/admins", h.AddAdminAPI)
	adminGroup.GET("/api/admins/:username", h.GetAdminAPI)
	adminGroup.POST("/api/admins/:username", h.EditAdminAPI)
	adminGroup.DELETE("/admins/delete/:username", h.DeleteAdmin)

	// Aliases
	adminGroup.GET("/aliases", h.ListAliases)
	adminGroup.POST("/api/aliases", h.AddAliasAPI)
	adminGroup.GET("/api/aliases/:address", h.GetAliasAPI)
	adminGroup.POST("/api/aliases/:address", h.EditAliasAPI)
	adminGroup.DELETE("/aliases/delete/:address", h.DeleteAlias)

	// Alias Domains
	adminGroup.GET("/alias-domains", h.ListAliasDomains)
	adminGroup.POST("/api/alias-domains", h.AddAliasDomainAPI)
	adminGroup.GET("/api/alias-domains/:alias_domain", h.GetAliasDomainAPI)
	adminGroup.POST("/api/alias-domains/:alias_domain", h.EditAliasDomainAPI)
	adminGroup.DELETE("/alias-domains/delete/:alias_domain", h.DeleteAliasDomain)

	// Fetchmail
	adminGroup.GET("/fetchmail/add", h.AddFetchmailGET)
	adminGroup.POST("/fetchmail/add", h.AddFetchmailPOST)

	// User Portal Routes (public)
	e.GET("/users/login", h.UserLogin)
	e.POST("/users/login", h.UserLogin)
	e.GET("/users/logout", h.UserLogout)

	// Protected User Portal Routes
	userGroup := e.Group("/users")
	userGroup.Use(middleware.UserAuthMiddleware)
	userGroup.GET("/dashboard", h.UserDashboard)
	userGroup.POST("/password", h.UpdateUserPassword)
	userGroup.POST("/forwarding", h.UpdateUserForwarding)
	userGroup.GET("/vacation", h.UserVacation)
	userGroup.POST("/vacation", h.UpdateUserVacation)
	userGroup.POST("/vacation/delete", h.DeleteUserVacation)

	// Root Redirect
	e.GET("/", func(c *echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/dashboard")
	})

	// Catch-all route for unknown pages (404)
	e.Any("/*", func(c *echo.Context) error {
		// If User is logged in
		if middleware.GetUsername(c, middleware.UserSessionName) != "" {
			return c.Redirect(http.StatusFound, "/users/dashboard")
		}
		// If Admin is logged in
		if middleware.GetUsername(c, middleware.SessionName) != "" {
			return c.Redirect(http.StatusFound, "/dashboard")
		}
		// Otherwise standard 404
		return echo.ErrNotFound
	})
}
