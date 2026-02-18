package handlers

import (
	"net/http"
	"strings"
	"time"

	"go-postfixadmin/internal/middleware"
	"go-postfixadmin/internal/models"
	"go-postfixadmin/internal/utils"

	"github.com/labstack/echo/v5"
)

// UserLogin processes user authentication (mailbox)
func (h *Handler) UserLogin(c *echo.Context) error {
	if c.Request().Method == http.MethodPost {
		username := c.FormValue("username")
		password := c.FormValue("password")

		var mailbox models.Mailbox

		if h.DB == nil {
			return c.Render(http.StatusServiceUnavailable, "users/login.html", map[string]interface{}{"error": "Database connection unavailable"})
		}

		if err := h.DB.Where("username = ? AND active = ?", username, true).First(&mailbox).Error; err != nil {
			return c.Render(http.StatusUnauthorized, "users/login.html", map[string]interface{}{"error": "Invalid credentials"})
		}

		match, err := utils.CheckPassword(password, mailbox.Password)
		if err != nil || !match {
			return c.Render(http.StatusUnauthorized, "users/login.html", map[string]interface{}{"error": "Invalid credentials"})
		}

		if err := middleware.SetUserSession(c, mailbox.Username); err != nil {
			return c.Render(http.StatusInternalServerError, "users/login.html", map[string]interface{}{"error": "Failed to create session"})
		}

		return c.Redirect(http.StatusFound, "/users/dashboard")
	}
	return c.Render(http.StatusOK, "users/login.html", nil)
}

// UserLogout clears the user session
func (h *Handler) UserLogout(c *echo.Context) error {
	middleware.ClearUserSession(c)
	return c.Redirect(http.StatusFound, "/users/login")
}

// UserDashboard displays the user dashboard
func (h *Handler) UserDashboard(c *echo.Context) error {
	username := middleware.GetUser(c)
	if username == "" {
		return c.Redirect(http.StatusFound, "/users/login")
	}

	var mailbox models.Mailbox
	if err := h.DB.First(&mailbox, "username = ?", username).Error; err != nil {
		return c.Redirect(http.StatusFound, "/users/login")
	}

	var alias models.Alias
	h.DB.First(&alias, "address = ?", username)

	return c.Render(http.StatusOK, "users/dashboard.html", map[string]interface{}{
		"User":    mailbox,
		"Alias":   alias,
		"Message": middleware.GetFlash(c, "message"),
		"Error":   middleware.GetFlash(c, "error"),
	})
}

// UpdateUserPassword changes the user's password
func (h *Handler) UpdateUserPassword(c *echo.Context) error {
	username := middleware.GetUser(c)
	currentPassword := c.FormValue("current_password")
	newPassword := c.FormValue("new_password")
	confirmPassword := c.FormValue("confirm_password")

	if newPassword != confirmPassword {
		middleware.SetFlash(c, "error", "As senhas não conferem")
		return c.Redirect(http.StatusFound, "/users/dashboard")
	}

	var mailbox models.Mailbox
	if err := h.DB.First(&mailbox, "username = ?", username).Error; err != nil {
		return c.Redirect(http.StatusFound, "/users/login")
	}

	match, err := utils.CheckPassword(currentPassword, mailbox.Password)
	if err != nil || !match {
		middleware.SetFlash(c, "error", "Senha atual incorreta")
		return c.Redirect(http.StatusFound, "/users/dashboard")
	}

	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		middleware.SetFlash(c, "error", "Falha ao processar a nova senha")
		return c.Redirect(http.StatusFound, "/users/dashboard")
	}

	mailbox.Password = hashedPassword
	if err := h.DB.Save(&mailbox).Error; err != nil {
		middleware.SetFlash(c, "error", "Falha ao atualizar a senha")
		return c.Redirect(http.StatusFound, "/users/dashboard")
	}

	// Log action
	parts := strings.Split(username, "@")
	domain := ""
	if len(parts) == 2 {
		domain = parts[1]
	}
	utils.LogAction(h.DB, username, c.RealIP(), domain, "USER_EDIT_PASSWORD", username)

	middleware.SetFlash(c, "message", "Senha atualizada com sucesso")
	return c.Redirect(http.StatusFound, "/users/dashboard")
}

// UpdateUserForwarding updates the user's forwarding address (alias)
func (h *Handler) UpdateUserForwarding(c *echo.Context) error {
	username := middleware.GetUser(c)
	forwarding := c.FormValue("forwarding")

	tx := h.DB.Begin()

	var alias models.Alias
	if err := tx.First(&alias, "address = ?", username).Error; err != nil {
		parts := strings.Split(username, "@")
		if len(parts) != 2 {
			tx.Rollback()
			middleware.SetFlash(c, "error", "Formato de usuário inválido")
			return c.Redirect(http.StatusFound, "/users/dashboard")
		}
		domain := parts[1]

		alias = models.Alias{
			Address:  username,
			Goto:     username,
			Domain:   domain,
			Created:  time.Now(),
			Modified: time.Now(),
			Active:   true,
		}
	}

	if strings.TrimSpace(forwarding) == "" {
		forwarding = username
	}

	// Convert newlines to comma-separated for DB storage
	lines := strings.Split(forwarding, "\n")
	var addresses []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			addresses = append(addresses, line)
		}
	}
	alias.Goto = strings.Join(addresses, ",")

	if err := tx.Save(&alias).Error; err != nil {
		tx.Rollback()
		middleware.SetFlash(c, "error", "Falha ao atualizar o redirecionamento")
		return c.Redirect(http.StatusFound, "/users/dashboard")
	}

	// Log action
	parts := strings.Split(username, "@")
	domain := ""
	if len(parts) == 2 {
		domain = parts[1]
	}
	if err := utils.LogAction(tx, username, c.RealIP(), domain, "USER_EDIT_ALIAS", alias.Goto); err != nil {
		// Log error but don't fail transaction? Or should we?
		// PostfixAdmin logs are usually best-effort.
	}

	tx.Commit()
	middleware.SetFlash(c, "message", "Redirecionamento atualizado com sucesso")
	return c.Redirect(http.StatusFound, "/users/dashboard")
}
