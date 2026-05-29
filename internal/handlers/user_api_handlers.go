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

// GetUserProfile godoc
// @Summary      Get user profile info
// @Description  Returns information about the logged-in mailbox user (name and email).
// @Tags         User Portal
// @Produce      json
// @Success      200 {object} map[string]interface{}
// @Failure      401 {object} map[string]interface{}
// @Failure      404 {object} map[string]interface{}
// @Router       /user/me [get]
// @Security     BearerAuth
func (h *Handler) GetUserProfile(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil || claims.Type != "mailbox" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	mailbox, err := repositories.GetMailboxByUsername(h.DB, claims.Username)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "mailbox not found")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"username": mailbox.Username,
		"name":     mailbox.Name,
	})
}

// GetUserForwarding godoc
// @Summary      Get user forwarding targets
// @Description  Returns current forwarding addresses for the authenticated mailbox.
// @Tags         User Portal
// @Produce      json
// @Success      200 {object} map[string]interface{}
// @Failure      401 {object} map[string]interface{}
// @Router       /user/forwarding [get]
// @Security     BearerAuth
func (h *Handler) GetUserForwarding(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil || claims.Type != "mailbox" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	alias, err := repositories.GetAliasByAddress(h.DB, claims.Username)
	gotoStr := claims.Username
	if err == nil {
		gotoStr = alias.Goto
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"goto": gotoStr,
	})
}

// UpdateUserForwardingAPI godoc
// @Summary      Update user forwarding targets
// @Description  Configures the email forwarding destinations.
// @Tags         User Portal
// @Accept       json
// @Produce      json
// @Param        request body map[string]interface{} true "Forwarding config"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]interface{}
// @Failure      401 {object} map[string]interface{}
// @Failure      500 {object} map[string]interface{}
// @Router       /user/forwarding [post]
// @Security     BearerAuth
func (h *Handler) UpdateUserForwardingAPI(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil || claims.Type != "mailbox" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	var req struct {
		Forwarding string `json:"forwarding"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	username := claims.Username
	parts := strings.Split(username, "@")
	if len(parts) != 2 {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user format")
	}
	domain := parts[1]

	forwarding := req.Forwarding
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
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update forwarding")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Forwarding updated successfully",
	})
}

// UpdateUserPasswordAPI godoc
// @Summary      Update mailbox password
// @Description  Allows the user to change their own password.
// @Tags         User Portal
// @Accept       json
// @Produce      json
// @Param        request body map[string]interface{} true "Password update info"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]interface{}
// @Failure      401 {object} map[string]interface{}
// @Failure      500 {object} map[string]interface{}
// @Router       /user/password [post]
// @Security     BearerAuth
func (h *Handler) UpdateUserPasswordAPI(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil || claims.Type != "mailbox" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
		ConfirmPassword string `json:"confirm_password"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	username := claims.Username
	mailbox, err := repositories.GetMailboxByUsername(h.DB, username)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "mailbox not found")
	}

	match, err := utils.CheckPassword(req.CurrentPassword, mailbox.Password)
	if err != nil || !match {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": "Current password is incorrect"})
	}

	if validationErr := ValidatePassword(req.NewPassword); validationErr != "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": validationErr})
	}

	if req.NewPassword != req.ConfirmPassword {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": "Passwords do not match"})
	}

	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": "Failed to process the new password"})
	}

	mailbox.Password = hashedPassword
	if err := repositories.SaveMailbox(h.DB, mailbox, "USER_EDIT_PASSWORD", username, c.RealIP()); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": "Failed to update password"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Password updated successfully",
	})
}

// GetUserVacation godoc
// @Summary      Get auto-reply / vacation settings
// @Description  Returns current vacation configuration for the mailbox user.
// @Tags         User Portal
// @Produce      json
// @Success      200 {object} map[string]interface{}
// @Failure      401 {object} map[string]interface{}
// @Router       /user/vacation [get]
// @Security     BearerAuth
func (h *Handler) GetUserVacation(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil || claims.Type != "mailbox" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	vacation, err := repositories.GetVacationByEmail(h.DB, claims.Username)
	if err != nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"active":        false,
			"subject":       "Out of Office",
			"body":          "",
			"activefrom":    time.Now().Format("2006-01-02T15:04"),
			"activeuntil":   time.Now().AddDate(0, 0, 7).Format("2006-01-02T15:04"),
			"interval_time": 0,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"active":        vacation.Active,
		"subject":       vacation.Subject,
		"body":          vacation.Body,
		"activefrom":    vacation.ActiveFrom.Format("2006-01-02T15:04"),
		"activeuntil":   vacation.ActiveUntil.Format("2006-01-02T15:04"),
		"interval_time": vacation.IntervalTime,
	})
}

// UpdateUserVacationAPI godoc
// @Summary      Upsert auto-reply / vacation settings
// @Description  Configures the vacation active state, message subject, body, interval and date range.
// @Tags         User Portal
// @Accept       json
// @Produce      json
// @Param        request body map[string]interface{} true "Vacation configuration data"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]interface{}
// @Failure      401 {object} map[string]interface{}
// @Failure      500 {object} map[string]interface{}
// @Router       /user/vacation [post]
// @Security     BearerAuth
func (h *Handler) UpdateUserVacationAPI(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil || claims.Type != "mailbox" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	var req struct {
		Active       bool   `json:"active"`
		Subject      string `json:"subject"`
		Body         string `json:"body"`
		ActiveFrom   string `json:"activefrom"`
		ActiveUntil  string `json:"activeuntil"`
		IntervalTime int    `json:"interval_time"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	username := claims.Username
	parts := strings.Split(username, "@")
	if len(parts) != 2 {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user format")
	}
	domain := parts[1]

	activeFrom, err := time.ParseInLocation("2006-01-02T15:04", req.ActiveFrom, time.Local)
	if err != nil {
		if activeFrom, err = time.ParseInLocation("2006-01-02T15:04:05", req.ActiveFrom, time.Local); err != nil {
			activeFrom = time.Now()
		}
	}

	activeUntil, err := time.ParseInLocation("2006-01-02T15:04", req.ActiveUntil, time.Local)
	if err != nil {
		if activeUntil, err = time.ParseInLocation("2006-01-02T15:04:05", req.ActiveUntil, time.Local); err != nil {
			activeUntil = time.Now().AddDate(0, 0, 7)
		}
	}

	vacation := models.Vacation{
		Email:        username,
		Subject:      req.Subject,
		Body:         req.Body,
		Domain:       domain,
		Active:       req.Active,
		ActiveFrom:   activeFrom,
		ActiveUntil:  activeUntil,
		IntervalTime: req.IntervalTime,
		Created:      time.Now(),
		Modified:     time.Now(),
	}

	if err := repositories.UpsertVacation(h.DB, vacation, username, c.RealIP()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update auto-reply")
	}

	if viper.GetBool("vacation.enabled") {
		_ = utils.SyncSingleVacationSieve(h.DB, username, "")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Auto-reply updated successfully",
	})
}

// DeleteUserVacationAPI godoc
// @Summary      Deactivate / remove auto-reply
// @Description  Removes vacation configuration for the mailbox user.
// @Tags         User Portal
// @Produce      json
// @Success      200 {object} map[string]interface{}
// @Failure      401 {object} map[string]interface{}
// @Failure      500 {object} map[string]interface{}
// @Router       /user/vacation [delete]
// @Security     BearerAuth
func (h *Handler) DeleteUserVacationAPI(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil || claims.Type != "mailbox" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	username := claims.Username
	parts := strings.Split(username, "@")
	domain := ""
	if len(parts) == 2 {
		domain = parts[1]
	}

	if err := repositories.DeleteVacation(h.DB, username, domain, username, c.RealIP()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete vacation settings")
	}

	if viper.GetBool("vacation.enabled") {
		_ = utils.SyncSingleVacationSieve(h.DB, username, "")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Auto-reply removed successfully",
	})
}
