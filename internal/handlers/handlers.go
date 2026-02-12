package handlers

import (
	"net/http"

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

		return c.Redirect(http.StatusFound, "/dashboard")
	}
	return c.Render(http.StatusOK, "login.html", nil)
}

// Dashboard exibe a página inicial com estatísticas
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
