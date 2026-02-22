package handlers

import (
	"net/http"
	"time"

	"go-postfixadmin/internal/middleware"
	"go-postfixadmin/internal/models"
	"go-postfixadmin/internal/utils"

	"github.com/labstack/echo/v5"
)

// Dashboard exibe a página inicial com estatísticas
func (h *Handler) Dashboard(c *echo.Context) error {
	username := middleware.GetUsername(c)
	allowedDomains, isSuperAdmin, err := utils.GetAllowedDomains(h.DB, username)
	if err != nil {
		return c.Render(http.StatusInternalServerError, "dashboard.html", map[string]interface{}{
			"Error": "Failed to check permissions: " + err.Error(),
		})
	}

	var domainCount int64
	var mailboxCount int64

	if h.DB != nil {
		domainQuery := h.DB.Model(&models.Domain{}).Where("active = ? AND domain != ?", true, "ALL")
		mailboxQuery := h.DB.Model(&models.Mailbox{}).Where("active = ?", true)

		if !isSuperAdmin {
			if len(allowedDomains) == 0 {
				// No domains allowed, counts are 0
				domainCount = 0
				mailboxCount = 0
			} else {
				domainQuery = domainQuery.Where("domain IN ?", allowedDomains)
				mailboxQuery = mailboxQuery.Where("domain IN ?", allowedDomains)
				domainQuery.Count(&domainCount)
				mailboxQuery.Count(&mailboxCount)
			}
		} else {
			domainQuery.Count(&domainCount)
			mailboxQuery.Count(&mailboxCount)
		}
	}

	var logs []models.Log
	if h.DB != nil {
		oneMonthAgo := time.Now().AddDate(0, -1, 0)
		logQuery := h.DB.Order("timestamp desc").Where("timestamp >= ?", oneMonthAgo).Limit(500)

		if !isSuperAdmin {
			if len(allowedDomains) == 0 {
				// No domains allowed, no logs
				logQuery = logQuery.Where("1 = 0")
			} else {
				// Filter logs by allowed domains
				logQuery = logQuery.Where("domain IN ?", allowedDomains)
			}
		}

		logQuery.Find(&logs)
	}

	return c.Render(http.StatusOK, "dashboard.html", map[string]interface{}{
		"DomainCount":  domainCount,
		"MailboxCount": mailboxCount,
		"IsSuperAdmin": isSuperAdmin,
		"Username":     username,
		"SessionUser":  username,
		"Logs":         logs,
	})
}
