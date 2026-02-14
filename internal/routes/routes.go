package routes

import (
	"net/http"

	"go-postfixadmin/internal/handlers"

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

	// API / Utils
	e.GET("/api/generate-password", h.GeneratePassword)

	// Root Redirect
	e.GET("/", func(c *echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/dashboard")
	})
}
