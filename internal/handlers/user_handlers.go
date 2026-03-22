package handlers

import (
	"net/http"
	"strings"
	"time"

	"go-postfixadmin/internal/middleware"
	"go-postfixadmin/internal/models"
	"go-postfixadmin/internal/utils"

	"github.com/labstack/echo/v5"
	"github.com/spf13/viper"
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

	return h.renderUserDashboard(c, username, GetFlash(c, "message"), GetFlash(c, "error"))
}

func (h *Handler) renderUserDashboard(c *echo.Context, username, message, errorMsg string) error {
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
		"Message":     message,
		"Error":       errorMsg,
	})
}

// UpdateUserPassword changes the user's password
func (h *Handler) UpdateUserPassword(c *echo.Context) error {
	username := middleware.GetUsername(c, middleware.UserSessionName)
	if username == "" {
		SetFlash(c, "error", "Login required")
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{"success": false, "error": GetFlash(c, "error")})
	}

	currentPassword := c.FormValue("current_password")
	newPassword := c.FormValue("new_password")
	confirmPassword := c.FormValue("confirm_password")

	var mailbox models.Mailbox
	if err := h.DB.First(&mailbox, "username = ?", username).Error; err != nil {
		SetFlash(c, "error", "Login required")
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{"success": false, "error": GetFlash(c, "error")})
	}

	match, err := utils.CheckPassword(currentPassword, mailbox.Password)
	if err != nil || !match {
		SetFlash(c, "error", "Current password is incorrect")
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": GetFlash(c, "error")})
	}

	if validationErr := ValidatePassword(newPassword); validationErr != "" {
		SetFlash(c, "error", validationErr)
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": GetFlash(c, "error")})
	}

	if newPassword != confirmPassword {
		SetFlash(c, "error", "Passwords do not match")
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": GetFlash(c, "error")})
	}

	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		SetFlash(c, "error", "Failed to process the new password")
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": GetFlash(c, "error")})
	}

	mailbox.Password = hashedPassword
	if err := h.DB.Save(&mailbox).Error; err != nil {
		SetFlash(c, "error", "Failed to update the password")
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": GetFlash(c, "error")})
	}

	// Log action
	parts := strings.Split(username, "@")
	domain := ""
	if len(parts) == 2 {
		domain = parts[1]
	}
	utils.LogAction(h.DB, username, c.RealIP(), domain, "USER_EDIT_PASSWORD", username)

	SetFlash(c, "message", "Password updated successfully")
	message := GetFlash(c, "message")
	_ = middleware.ClearSession(c, middleware.UserSessionName)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success":  true,
		"message":  message,
		"redirect": "/users/login",
	})
}

// UpdateUserForwarding updates the user's forwarding address (alias)
func (h *Handler) UpdateUserForwarding(c *echo.Context) error {
	username := middleware.GetUsername(c, middleware.UserSessionName)
	if username == "" {
		SetFlash(c, "error", "Login required")
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{"success": false, "error": GetFlash(c, "error")})
	}

	forwarding := c.FormValue("forwarding")

	tx := h.DB.Begin()

	var alias models.Alias
	if err := tx.First(&alias, "address = ?", username).Error; err != nil {
		parts := strings.Split(username, "@")
		if len(parts) != 2 {
			tx.Rollback()
			SetFlash(c, "error", "Invalid user format")
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": GetFlash(c, "error")})
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
		SetFlash(c, "error", "Failed to update forwarding")
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": GetFlash(c, "error")})
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
	SetFlash(c, "message", "Forwarding updated successfully")
	return c.JSON(http.StatusOK, map[string]interface{}{"success": true, "message": GetFlash(c, "message")})
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
		"Message":     GetFlash(c, "message"),
		"Error":       GetFlash(c, "error"),
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
		SetFlash(c, "error", "Invalid user format")
		return c.Redirect(http.StatusFound, "/users/vacation")
	}
	domain := parts[1]

	subject := c.FormValue("subject")
	body := c.FormValue("body")
	activeFromStr := c.FormValue("activefrom")
	activeUntilStr := c.FormValue("activeuntil")
	intervalTimeStr := c.FormValue("interval_time")
	activeStr := c.FormValue("active")

	// Try parsing with both minutes precision (default datetime-local) and seconds precision (fallback)
	activeFrom, err := time.ParseInLocation("2006-01-02T15:04", activeFromStr, time.Local)
	if err != nil {
		activeFrom, err = time.ParseInLocation("2006-01-02T15:04:05", activeFromStr, time.Local)
		if err != nil {
			activeFrom = time.Now()
		}
	}

	activeUntil, err := time.ParseInLocation("2006-01-02T15:04", activeUntilStr, time.Local)
	if err != nil {
		activeUntil, err = time.ParseInLocation("2006-01-02T15:04:05", activeUntilStr, time.Local)
		if err != nil {
			activeUntil = time.Now()
		}
	}

	intervalTime := 0
	if intervalTimeStr == "86400" {
		intervalTime = 86400 // 1 day
	} else if intervalTimeStr == "604800" {
		intervalTime = 604800 // 7 days
	}

	active := activeStr == "true" || activeStr == "on" || activeStr == "1"

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
		SetFlash(c, "error", "Failed to save auto-reply settings")
		return c.Redirect(http.StatusFound, "/users/vacation")
	}

	// We also typically need an alias to route emails to the vacation script handling
	// in many PostfixAdmin implementations. However, just matching exact existing design constraint:
	// We'll trust PostfixAdmin aliases cover it or we just add the DB entry as requested.
	utils.LogAction(tx, username, c.RealIP(), domain, "USER_UPDATE_VACATION", username)

	tx.Commit()

	// Update sieve script for this user if vacation is enabled globally
	if viper.GetBool("vacation.enabled") {
		_ = utils.SyncSingleVacationSieve(h.DB, username, "")
	}

	SetFlash(c, "message", "Auto-reply saved successfully")
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
		SetFlash(c, "error", "Failed to remove auto-reply")
		return c.Redirect(http.StatusFound, "/users/vacation")
	}

	utils.LogAction(tx, username, c.RealIP(), domain, "USER_DELETE_VACATION", username)

	tx.Commit()

	// Remove sieve script for this user if vacation is enabled globally
	if viper.GetBool("vacation.enabled") {
		_ = utils.SyncSingleVacationSieve(h.DB, username, "")
	}

	SetFlash(c, "message", "Auto-reply removed successfully")
	return c.Redirect(http.StatusFound, "/users/dashboard")
}
