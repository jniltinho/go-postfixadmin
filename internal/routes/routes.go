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

	// API Routes
	adminGroup.GET("/api/generate-password", h.GeneratePassword)

	// Domains
	adminGroup.GET("/domains", h.ListDomains)
	adminGroup.GET("/domains/add", h.AddDomainForm)
	adminGroup.POST("/domains/add", h.AddDomain)
	adminGroup.GET("/domains/edit/:domain", h.EditDomainForm)
	adminGroup.POST("/domains/edit/:domain", h.EditDomain)
	adminGroup.DELETE("/domains/delete/:domain", h.DeleteDomain)

	// Mailboxes
	adminGroup.GET("/mailboxes", h.ListMailboxes)
	adminGroup.GET("/mailboxes/add", h.AddMailboxForm)
	adminGroup.POST("/mailboxes/add", h.AddMailbox)
	adminGroup.GET("/mailboxes/edit/:username", h.EditMailboxForm)
	adminGroup.POST("/mailboxes/edit/:username", h.EditMailbox)
	adminGroup.DELETE("/mailboxes/delete/:username", h.DeleteMailbox)

	// Admins
	adminGroup.GET("/admins", h.ListAdmins)
	adminGroup.GET("/admins/add", h.AddAdminForm)
	adminGroup.POST("/admins/add", h.AddAdmin)
	adminGroup.GET("/admins/edit/:username", h.EditAdminForm)
	adminGroup.POST("/admins/edit/:username", h.EditAdmin)
	adminGroup.DELETE("/admins/delete/:username", h.DeleteAdmin)

	// Aliases
	adminGroup.GET("/aliases", h.ListAliases)
	adminGroup.GET("/aliases/add", h.AddAliasForm)
	adminGroup.POST("/aliases/add", h.AddAlias)
	adminGroup.GET("/aliases/edit/:address", h.EditAliasForm)
	adminGroup.POST("/aliases/edit/:address", h.EditAlias)
	adminGroup.DELETE("/aliases/delete/:address", h.DeleteAlias)

	// Alias Domains
	adminGroup.GET("/alias-domains", h.ListAliasDomains)
	adminGroup.GET("/alias-domains/add", h.AddAliasDomainForm)
	adminGroup.POST("/alias-domains/add", h.AddAliasDomain)
	adminGroup.GET("/alias-domains/edit/:alias_domain", h.EditAliasDomainForm)
	adminGroup.POST("/alias-domains/edit/:alias_domain", h.EditAliasDomain)
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
		if middleware.GetUser(c) != "" {
			return c.Redirect(http.StatusFound, "/users/dashboard")
		}
		// If Admin is logged in
		if middleware.GetUsername(c) != "" {
			return c.Redirect(http.StatusFound, "/dashboard")
		}
		// Otherwise standard 404
		return echo.ErrNotFound
	})
}
