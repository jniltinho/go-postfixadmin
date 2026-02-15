package handlers

import (
	"net/http"

	"go-postfixadmin/internal/middleware"
	"go-postfixadmin/internal/models"
	"go-postfixadmin/internal/utils"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

// Handler é o controlador principal da aplicação
type Handler struct {
	DB *gorm.DB
}

// Login processa autenticação de administradores
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

		// Set session
		if err := middleware.SetSession(c, admin.Username); err != nil {
			return c.Render(http.StatusInternalServerError, "login.html", map[string]interface{}{"error": "Failed to create session"})
		}

		return c.Redirect(http.StatusFound, "/dashboard")
	}
	return c.Render(http.StatusOK, "login.html", nil)
}

// Logout encerra a sessão
func (h *Handler) Logout(c *echo.Context) error {
	middleware.ClearSession(c)
	return c.Redirect(http.StatusFound, "/login")
}

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

	return c.Render(http.StatusOK, "dashboard.html", map[string]interface{}{
		"DomainCount":  domainCount,
		"MailboxCount": mailboxCount,
		"IsSuperAdmin": isSuperAdmin,
		"Username":     username,
		"SessionUser":  username,
	})
}
