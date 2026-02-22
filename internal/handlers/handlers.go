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
		if err := middleware.SetSession(c, admin.Username, admin.Superadmin); err != nil {
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
