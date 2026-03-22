package handlers

import (
	"net/http"
	"time"

	"go-postfixadmin/internal/middleware"
	"go-postfixadmin/internal/repositories"

	"github.com/labstack/echo/v5"
)

// Dashboard exibe a página inicial com estatísticas
func (h *Handler) Dashboard(c *echo.Context) error {
	username := middleware.GetUsername(c, middleware.SessionName)
	isSuperAdmin := middleware.GetIsSuperAdmin(c)

	domainCount, mailboxCount, aliasCount, err := repositories.GetDashboardCounts(h.DB, username, isSuperAdmin)
	if err != nil {
		return c.Render(http.StatusInternalServerError, "dashboard/dashboard.html", map[string]interface{}{
			"Error": "Failed to check permissions: " + err.Error(),
		})
	}

	logs, _ := repositories.GetRecentLogs(h.DB, username, isSuperAdmin, time.Now().AddDate(0, -1, 0))

	return c.Render(http.StatusOK, "dashboard/dashboard.html", map[string]interface{}{
		"DomainCount":  domainCount,
		"MailboxCount": mailboxCount,
		"AliasCount":   aliasCount,
		"IsSuperAdmin": isSuperAdmin,
		"Username":     username,
		"SessionUser":  username,
		"Logs":         logs,
	})
}
