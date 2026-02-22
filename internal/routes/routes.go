package routes

import (
	"net/http"

	"go-postfixadmin/internal/handlers"
	"go-postfixadmin/internal/middleware"

	"github.com/labstack/echo/v5"
)

// RegisterRoutes registers all the application routes
func RegisterRoutes(e *echo.Echo, h *handlers.Handler) {
	// Auth Routes
	e.GET("/login", h.Login)
	e.POST("/login", h.Login)
	e.GET("/logout", h.Logout)

	// Dashboard
	e.GET("/dashboard", h.Dashboard)

	// Domains
	e.GET("/domains", h.ListDomains)
	e.GET("/domains/add", h.AddDomainForm)
	e.POST("/domains/add", h.AddDomain)
	e.GET("/domains/edit/:domain", h.EditDomainForm)
	e.POST("/domains/edit/:domain", h.EditDomain)
	e.DELETE("/domains/delete/:domain", h.DeleteDomain)

	// Mailboxes
	e.GET("/mailboxes", h.ListMailboxes)
	e.GET("/mailboxes/add", h.AddMailboxForm)
	e.POST("/mailboxes/add", h.AddMailbox)
	e.GET("/mailboxes/edit/:username", h.EditMailboxForm)
	e.POST("/mailboxes/edit/:username", h.EditMailbox)
	e.DELETE("/mailboxes/delete/:username", h.DeleteMailbox)

	// Admins
	e.GET("/admins", h.ListAdmins)
	e.GET("/admins/add", h.AddAdminForm)
	e.POST("/admins/add", h.AddAdmin)
	e.GET("/admins/edit/:username", h.EditAdminForm)
	e.POST("/admins/edit/:username", h.EditAdmin)
	e.DELETE("/admins/delete/:username", h.DeleteAdmin)

	// Aliases
	e.GET("/aliases", h.ListAliases)
	e.GET("/aliases/add", h.AddAliasForm)
	e.POST("/aliases/add", h.AddAlias)
	e.GET("/aliases/edit/:address", h.EditAliasForm)
	e.POST("/aliases/edit/:address", h.EditAlias)
	e.DELETE("/aliases/delete/:address", h.DeleteAlias)

	// Alias Domains
	e.GET("/alias-domains", h.ListAliasDomains)
	e.GET("/alias-domains/add", h.AddAliasDomainForm)
	e.POST("/alias-domains/add", h.AddAliasDomain)
	e.GET("/alias-domains/edit/:alias_domain", h.EditAliasDomainForm)
	e.POST("/alias-domains/edit/:alias_domain", h.EditAliasDomain)
	e.DELETE("/alias-domains/delete/:alias_domain", h.DeleteAliasDomain)

	// API / Utils
	e.GET("/api/generate-password", h.GeneratePassword)

	// User Portal Routes
	e.GET("/users/login", h.UserLogin)
	e.POST("/users/login", h.UserLogin)
	e.GET("/users/logout", h.UserLogout)

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
