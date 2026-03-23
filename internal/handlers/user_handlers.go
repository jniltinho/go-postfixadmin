package handlers

import (
	"net/http"
	"strings"
	"time"

	"go-postfixadmin/internal/middleware"
	"go-postfixadmin/internal/models"
	"go-postfixadmin/internal/repositories"
	"go-postfixadmin/internal/utils"

	"github.com/labstack/echo/v5"
	"github.com/spf13/viper"
)

// UserLogin processes user authentication (mailbox)
func (h *Handler) UserLogin(c *echo.Context) error {
	if c.Request().Method == http.MethodPost {
		username := c.FormValue("username")
		password := c.FormValue("password")

		if h.DB == nil {
			return c.Render(http.StatusServiceUnavailable, "users/login.html", map[string]interface{}{"errorKey": "Login_ErrDbUnavailable"})
		}

		mailbox, err := repositories.GetMailboxByUsername(h.DB, username)
		if err != nil || !mailbox.Active {
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
	mailbox, err := repositories.GetMailboxByUsername(h.DB, username)
	if err != nil {
		return c.Redirect(http.StatusFound, "/users/login")
	}

	alias, _ := repositories.GetAliasByAddress(h.DB, username)

	data := map[string]interface{}{
		"SessionUser": username,
		"User":        mailbox,
		"Alias":       alias,
		"Message":     message,
		"Error":       errorMsg,
	}

	vacation, err := repositories.GetVacationByEmail(h.DB, username)
	if err == nil {
		data["Vacation"] = map[string]interface{}{
			"Subject":      vacation.Subject,
			"Body":         vacation.Body,
			"ActiveFrom":   vacation.ActiveFrom.Format("2006-01-02T15:04"),
			"ActiveUntil":  vacation.ActiveUntil.Format("2006-01-02T15:04"),
			"IntervalTime": vacation.IntervalTime,
			"Active":       vacation.Active,
		}
	}

	return c.Render(http.StatusOK, "users/dashboard.html", data)
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

	mailbox, err := repositories.GetMailboxByUsername(h.DB, username)
	if err != nil {
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
	if err := repositories.SaveMailbox(h.DB, mailbox, "USER_EDIT_PASSWORD", username, c.RealIP()); err != nil {
		SetFlash(c, "error", "Failed to update the password")
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": GetFlash(c, "error")})
	}

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

	parts := strings.Split(username, "@")
	if len(parts) != 2 {
		SetFlash(c, "error", "Invalid user format")
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": GetFlash(c, "error")})
	}
	domain := parts[1]

	forwarding := c.FormValue("forwarding")
	if strings.TrimSpace(forwarding) == "" {
		forwarding = username
	}

	var addresses []string
	for _, line := range strings.Split(forwarding, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			addresses = append(addresses, line)
		}
	}
	gotoStr := strings.Join(addresses, ",")

	if err := repositories.UpdateUserForwarding(h.DB, username, gotoStr, domain, username, c.RealIP()); err != nil {
		SetFlash(c, "error", "Failed to update forwarding")
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": GetFlash(c, "error")})
	}

	SetFlash(c, "message", "Forwarding updated successfully")
	return c.JSON(http.StatusOK, map[string]interface{}{"success": true, "message": GetFlash(c, "message")})
}

// UserVacation redirects to the dashboard where the vacation form is now a modal
func (h *Handler) UserVacation(c *echo.Context) error {
	username := middleware.GetUsername(c, middleware.UserSessionName)
	if username == "" {
		return c.Redirect(http.StatusFound, "/users/login")
	}
	return c.Redirect(http.StatusFound, "/users/dashboard")
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

	activeFrom, err := time.ParseInLocation("2006-01-02T15:04", activeFromStr, time.Local)
	if err != nil {
		if activeFrom, err = time.ParseInLocation("2006-01-02T15:04:05", activeFromStr, time.Local); err != nil {
			activeFrom = time.Now()
		}
	}

	activeUntil, err := time.ParseInLocation("2006-01-02T15:04", activeUntilStr, time.Local)
	if err != nil {
		if activeUntil, err = time.ParseInLocation("2006-01-02T15:04:05", activeUntilStr, time.Local); err != nil {
			activeUntil = time.Now()
		}
	}

	intervalTime := 0
	if intervalTimeStr == "86400" {
		intervalTime = 86400
	} else if intervalTimeStr == "604800" {
		intervalTime = 604800
	}

	vacation := models.Vacation{
		Email:        username,
		Subject:      subject,
		Body:         body,
		Domain:       domain,
		Active:       activeStr == "true" || activeStr == "on" || activeStr == "1",
		ActiveFrom:   activeFrom,
		ActiveUntil:  activeUntil,
		IntervalTime: intervalTime,
		Created:      time.Now(),
		Modified:     time.Now(),
	}

	isJSON := strings.Contains(c.Request().Header.Get("Accept"), "application/json")

	if err := repositories.UpsertVacation(h.DB, vacation, username, c.RealIP()); err != nil {
		msg := T(c, "Vacation_JsFailed")
		if isJSON {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": msg})
		}
		SetFlash(c, "error", msg)
		return c.Redirect(http.StatusFound, "/users/dashboard")
	}

	if viper.GetBool("vacation.enabled") {
		_ = utils.SyncSingleVacationSieve(h.DB, username, "")
	}

	msg := T(c, "Vacation_JsSaved")
	if isJSON {
		return c.JSON(http.StatusOK, map[string]interface{}{"success": true, "message": msg})
	}
	SetFlash(c, "message", msg)
	return c.Redirect(http.StatusFound, "/users/dashboard")
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

	if err := repositories.DeleteVacation(h.DB, username, domain, username, c.RealIP()); err != nil {
		SetFlash(c, "error", "Failed to remove auto-reply")
		return c.Redirect(http.StatusFound, "/users/vacation")
	}

	if viper.GetBool("vacation.enabled") {
		_ = utils.SyncSingleVacationSieve(h.DB, username, "")
	}

	SetFlash(c, "message", "Auto-reply removed successfully")
	return c.Redirect(http.StatusFound, "/users/dashboard")
}
