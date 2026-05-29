package handlers

import (
	"net/http"
	"time"

	"go-postfixadmin/internal/api/dto"
	"go-postfixadmin/internal/middleware"
	"go-postfixadmin/internal/repositories"

	"github.com/labstack/echo/v5"
)

// Dashboard exibe a página inicial com estatísticas (legacy HTML)
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

// DashboardStatsV1 godoc
// @Summary      Dashboard statistics
// @Description  Returns aggregate counts and recent activity for the authenticated user.
// @Tags         Dashboard
// @Produce      json
// @Success      200 {object} dto.DashboardStatsResponse
// @Failure      401 {object} dto.APIResponse
// @Router       /dashboard [get]
// @Security     BearerAuth
func (h *Handler) DashboardStatsV1(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil {
		return dto.Unauthorized(c, "not authenticated")
	}

	domainCount, mailboxCount, aliasCount, err := repositories.GetDashboardCounts(h.DB, claims.Username, claims.Superadmin)
	if err != nil {
		return dto.InternalError(c, "failed to fetch dashboard counts")
	}

	recentLogs, _ := repositories.GetRecentLogs(h.DB, claims.Username, claims.Superadmin, time.Now().AddDate(0, -1, 0))

	logEntries := make([]dto.LogEntryResponse, 0, len(recentLogs))
	for _, l := range recentLogs {
		logEntries = append(logEntries, dto.LogEntryResponse{
			Timestamp: l.Timestamp,
			Username:  l.Username,
			Domain:    l.Domain,
			Action:    l.Action,
			Data:      l.Data,
		})
	}

	return dto.WriteSuccess(c, dto.DashboardStatsResponse{
		Domains:    domainCount,
		Mailboxes:  mailboxCount,
		Aliases:    aliasCount,
		RecentLogs: logEntries,
	})
}
