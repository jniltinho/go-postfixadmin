package handlers

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/nilton/Go-Postfixadmin/internal/models"
	"github.com/nilton/Go-Postfixadmin/internal/utils"
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
		if h.DB != nil {
			if err := h.DB.Where("username = ? AND active = ?", username, true).First(&admin).Error; err != nil {
				return c.Render(http.StatusUnauthorized, "login.html", map[string]interface{}{"error": "Invalid credentials"})
			}

			match, err := utils.CheckPassword(password, admin.Password)
			if err != nil || !match {
				return c.Render(http.StatusUnauthorized, "login.html", map[string]interface{}{"error": "Invalid credentials"})
			}
		}

		return c.Redirect(http.StatusFound, "/dashboard")
	}
	return c.Render(http.StatusOK, "login.html", nil)
}

func (h *Handler) Dashboard(c *echo.Context) error {
	var domainCount int64
	var mailboxCount int64

	if h.DB != nil {
		h.DB.Model(&models.Domain{}).Where("active = ?", true).Count(&domainCount)
		h.DB.Model(&models.Mailbox{}).Where("active = ?", true).Count(&mailboxCount)
	}

	return c.Render(http.StatusOK, "dashboard.html", map[string]interface{}{
		"DomainCount":  domainCount,
		"MailboxCount": mailboxCount,
	})
}

func (h *Handler) ListDomains(c *echo.Context) error {
	var domains []models.Domain
	if h.DB != nil {
		h.DB.Find(&domains)
	}
	return c.Render(http.StatusOK, "domains.html", map[string]interface{}{
		"Domains": domains,
	})
}
