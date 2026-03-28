package handlers

import (
	"net/http"
	"strings"
	"time"

	"go-postfixadmin/internal/i18n"
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

// getLang returns the user's preferred language from cookie or Accept-Language header.
func getLang(c *echo.Context) string {
	if cookie, err := c.Cookie("lang"); err == nil {
		if v := cookie.Value; v == "en" || v == "pt" || v == "es" {
			return v
		}
	}
	if strings.HasPrefix(strings.ToLower(c.Request().Header.Get("Accept-Language")), "en") {
		return "en"
	}
	return "pt"
}

// T translates a message ID using the request's language.
func T(c *echo.Context, msgID string) string {
	return i18n.Translate(getLang(c), msgID, nil)
}

// TData translates a message ID with template data using the request's language.
func TData(c *echo.Context, msgID string, data map[string]any) string {
	return i18n.Translate(getLang(c), msgID, data)
}

// Login processa autenticação de administradores
func (h *Handler) Login(c *echo.Context) error {
	if c.Request().Method == http.MethodPost {
		username := c.FormValue("username")
		password := c.FormValue("password")

		var admin models.Admin

		if h.DB == nil {
			return c.Render(http.StatusServiceUnavailable, "auth/login.html", map[string]interface{}{"errorKey": "Login_ErrDbUnavailable"})
		}

		if err := h.DB.Where("username = ? AND active = ?", username, true).First(&admin).Error; err != nil {
			return c.Render(http.StatusUnauthorized, "auth/login.html", map[string]interface{}{"errorKey": "Login_ErrInvalidCredentials"})
		}

		match, err := utils.CheckPassword(password, admin.Password)
		if err != nil || !match {
			return c.Render(http.StatusUnauthorized, "auth/login.html", map[string]interface{}{"errorKey": "Login_ErrInvalidCredentials"})
		}

		// Set session
		if err := middleware.SetSession(c, middleware.SessionName, admin.Username, admin.Superadmin); err != nil {
			return c.Render(http.StatusInternalServerError, "auth/login.html", map[string]interface{}{"errorKey": "Login_ErrSession"})
		}

		return c.Redirect(http.StatusFound, "/dashboard")
	}
	return c.Render(http.StatusOK, "auth/login.html", nil)
}

// Logout encerra a sessão
func (h *Handler) Logout(c *echo.Context) error {
	middleware.ClearSession(c, middleware.SessionName)
	return c.Redirect(http.StatusFound, "/login")
}

// SetLanguage define o idioma da interface através de um cookie
func (h *Handler) SetLanguage(c *echo.Context) error {
	lang := c.Param("code")
	if lang != "en" && lang != "pt" && lang != "es" {
		lang = "pt"
	}

	cookie := &http.Cookie{
		Name:     "lang",
		Value:    lang,
		Path:     "/",
		Expires:  time.Now().Add(24 * 365 * time.Hour), // 1 ano
		HttpOnly: true,
	}
	c.SetCookie(cookie)

	// Redireciona de volta para a página anterior ou dashboard
	referer := c.Request().Header.Get("Referer")
	if referer == "" {
		referer = "/dashboard"
	}
	return c.Redirect(http.StatusFound, referer)
}
