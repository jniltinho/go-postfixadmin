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
			return c.Render(http.StatusServiceUnavailable, "users/login.html", map[string]interface{}{"errorKey": "Login_ErrDbUnavailable"})
		}

		if err := h.DB.Where("username = ? AND active = ?", username, true).First(&mailbox).Error; err != nil {
			return c.Render(http.StatusUnauthorized, "users/login.html", map[string]interface{}{"errorKey": "Login_ErrInvalidCredentials"})
		}

		match, err := utils.CheckPassword(password, mailbox.Password)
		if err != nil || !match {
			return c.Render(http.StatusUnauthorized, "users/login.html", map[string]interface{}{"errorKey": "Login_ErrInvalidCredentials"})
		}

		if err := middleware.SetSession(c, middleware.UserSessionName, mailbox.Username, false); err != nil {
			return c.Render(http.StatusInternalServerError, "users/login.html", map[string]interface{}{"errorKey": "Login_ErrSession"})
		}

		return c.Redirect(http.StatusFound, "/users/dashboard")
	}
	return c.Render(http.StatusOK, "users/login.html", nil)
}

// UserLogout clears the user session
func (h *Handler) UserLogout(c *echo.Context) error {
	middleware.ClearSession(c, middleware.UserSessionName)
	return c.Redirect(http.StatusFound, "/users/login")
}

// UserDashboard displays the user dashboard
func (h *Handler) UserDashboard(c *echo.Context) error {
	username := middleware.GetUsername(c, middleware.UserSessionName)
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
		"SessionUser": username,
		"User":        mailbox, // Still needed if dashboard body requires mailbox fields but header uses SessionUser
		"Alias":       alias,
		"Message":     middleware.GetFlash(c, "message"),
		"Error":       middleware.GetFlash(c, "error"),
	})
}

// UpdateUserPassword changes the user's password
func (h *Handler) UpdateUserPassword(c *echo.Context) error {
	username := middleware.GetUsername(c, middleware.UserSessionName)
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
	username := middleware.GetUsername(c, middleware.UserSessionName)
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

// UserVacation displays the user's vacation/auto-reply configuration form
func (h *Handler) UserVacation(c *echo.Context) error {
	username := middleware.GetUsername(c, middleware.UserSessionName)
	if username == "" {
		return c.Redirect(http.StatusFound, "/users/login")
	}

	var vacation models.Vacation
	err := h.DB.First(&vacation, "email = ?", username).Error

	templateData := map[string]interface{}{
		"SessionUser": username,
		"Message":     middleware.GetFlash(c, "message"),
		"Error":       middleware.GetFlash(c, "error"),
	}

	if err == nil {
		// Found vacation config
		templateData["Vacation"] = map[string]interface{}{
			"Subject":      vacation.Subject,
			"Body":         vacation.Body,
			"ActiveFrom":   vacation.ActiveFrom.Format("2006-01-02T15:04"),
			"ActiveUntil":  vacation.ActiveUntil.Format("2006-01-02T15:04"),
			"IntervalTime": vacation.IntervalTime,
			"Active":       vacation.Active,
		}
	}

	return c.Render(http.StatusOK, "users/vacation.html", templateData)
}

// UpdateUserVacation upserts the user's vacation configuration
func (h *Handler) UpdateUserVacation(c *echo.Context) error {
	username := middleware.GetUsername(c, middleware.UserSessionName)
	if username == "" {
		return c.Redirect(http.StatusFound, "/users/login")
	}

	parts := strings.Split(username, "@")
	if len(parts) != 2 {
		middleware.SetFlash(c, "error", "Formato de usuário inválido")
		return c.Redirect(http.StatusFound, "/users/vacation")
	}
	domain := parts[1]

	subject := c.FormValue("subject")
	body := c.FormValue("body")
	activeFromStr := c.FormValue("activefrom")
	activeUntilStr := c.FormValue("activeuntil")
	intervalTimeStr := c.FormValue("interval_time")
	activeStr := c.FormValue("active")

	activeFrom, err := time.ParseInLocation("2006-01-02T15:04", activeFromStr, time.Local)
	if err != nil {
		activeFrom = time.Now()
	}
	activeUntil, err := time.ParseInLocation("2006-01-02T15:04", activeUntilStr, time.Local)
	if err != nil {
		activeUntil = time.Now()
	}

	intervalTime := 0
	if intervalTimeStr == "1" {
		intervalTime = 1
	} else if intervalTimeStr == "7" {
		intervalTime = 7
	}

	active := false
	if activeStr == "true" || activeStr == "on" || activeStr == "1" {
		active = true
	}

	tx := h.DB.Begin()

	vacation := models.Vacation{
		Email:        username,
		Subject:      subject,
		Body:         body,
		Domain:       domain,
		Active:       active,
		ActiveFrom:   activeFrom,
		ActiveUntil:  activeUntil,
		IntervalTime: intervalTime,
		Created:      time.Now(),
		Modified:     time.Now(),
	}

	// Assuming 'Upsert' behavior or simply Save
	if err := tx.Save(&vacation).Error; err != nil {
		tx.Rollback()
		middleware.SetFlash(c, "error", "Falha ao salvar configuração da resposta automática")
		return c.Redirect(http.StatusFound, "/users/vacation")
	}

	// We also typically need an alias to route emails to the vacation script handling
	// in many PostfixAdmin implementations. However, just matching exact existing design constraint:
	// We'll trust PostfixAdmin aliases cover it or we just add the DB entry as requested.
	utils.LogAction(tx, username, c.RealIP(), domain, "USER_UPDATE_VACATION", username)

	tx.Commit()
	middleware.SetFlash(c, "message", "Resposta automática salva com sucesso")
	return c.Redirect(http.StatusFound, "/users/vacation")
}

// DeleteUserVacation removes the user's vacation configuration
func (h *Handler) DeleteUserVacation(c *echo.Context) error {
	username := middleware.GetUsername(c, middleware.UserSessionName)
	if username == "" {
		return c.Redirect(http.StatusFound, "/users/login")
	}

	parts := strings.Split(username, "@")
	domain := ""
	if len(parts) == 2 {
		domain = parts[1]
	}

	tx := h.DB.Begin()

	if err := tx.Where("email = ?", username).Delete(&models.Vacation{}).Error; err != nil {
		tx.Rollback()
		middleware.SetFlash(c, "error", "Falha ao remover resposta automática")
		return c.Redirect(http.StatusFound, "/users/vacation")
	}

	utils.LogAction(tx, username, c.RealIP(), domain, "USER_DELETE_VACATION", username)

	tx.Commit()
	middleware.SetFlash(c, "message", "Resposta automática removida com sucesso")
	return c.Redirect(http.StatusFound, "/users/dashboard")
}
