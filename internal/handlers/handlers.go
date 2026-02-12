package handlers

import (
	"net/http"

	"go-postfixadmin/internal/models"
	"go-postfixadmin/internal/utils"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

type Handler struct {
	DB *gorm.DB
}

func (h *Handler) Login(c *echo.Context) error {
	if c.Request().Method == http.MethodPost {
		username := c.FormValue("username")
		password := c.FormValue("password")

		var admin models.Admin

		if h.DB == nil {
			return c.Render(http.StatusServiceUnavailable, "login.html", map[string]interface{}{"error": "Database connection unavailable"})
		}

		if err := h.DB.Where("username = ? AND active = ?", username, true).First(&admin).Error; err != nil {
			return c.Render(http.StatusUnauthorized, "login.html", map[string]interface{}{"error": "Invalid credentials"})
		}

		match, err := utils.CheckPassword(password, admin.Password)
		if err != nil || !match {
			return c.Render(http.StatusUnauthorized, "login.html", map[string]interface{}{"error": "Invalid credentials"})
		}

		return c.Redirect(http.StatusFound, "/dashboard")
	}
	return c.Render(http.StatusOK, "login.html", nil)
}

func (h *Handler) Dashboard(c *echo.Context) error {
	var domainCount int64
	var mailboxCount int64

	if h.DB != nil {
		h.DB.Model(&models.Domain{}).Where("active = ? AND domain != ?", true, "ALL").Count(&domainCount)
		h.DB.Model(&models.Mailbox{}).Where("active = ?", true).Count(&mailboxCount)
	}

	return c.Render(http.StatusOK, "dashboard.html", map[string]interface{}{
		"DomainCount":  domainCount,
		"MailboxCount": mailboxCount,
	})
}

type DomainDisplay struct {
	models.Domain
	AliasCount   int64
	MailboxCount int64
}

func (h *Handler) ListDomains(c *echo.Context) error {
	var domains []models.Domain
	var displayDomains []DomainDisplay

	if h.DB != nil {
		h.DB.Where("domain != ?", "ALL").Find(&domains)

		for _, d := range domains {
			var aliasCount int64
			var mailboxCount int64

			// Count aliases excluding those that are mailboxes
			h.DB.Model(&models.Alias{}).
				Where("domain = ?", d.Domain).
				Where("address NOT IN (?)", h.DB.Table("mailbox").Select("username")).
				Count(&aliasCount)

			h.DB.Model(&models.Mailbox{}).Where("domain = ?", d.Domain).Count(&mailboxCount)

			displayDomains = append(displayDomains, DomainDisplay{
				Domain:       d,
				AliasCount:   aliasCount,
				MailboxCount: mailboxCount,
			})
		}
	}
	return c.Render(http.StatusOK, "domains.html", map[string]interface{}{
		"Domains": displayDomains,
	})
}
