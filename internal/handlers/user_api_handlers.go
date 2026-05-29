package handlers

import (
	"net/http"
	"strings"
	"time"

	"go-postfixadmin/internal/api/dto"
	"go-postfixadmin/internal/middleware"
	"go-postfixadmin/internal/models"
	"go-postfixadmin/internal/repositories"
	"go-postfixadmin/internal/utils"

	"github.com/labstack/echo/v5"
	"github.com/spf13/viper"
)

// GetUserProfile godoc
// @Summary      Get user profile info
// @Description  Returns the display name and email address of the authenticated mailbox user.
// @Tags         User Portal
// @Produce      json
// @Success      200  {object}  dto.UserProfileResponse
// @Failure      401  {object}  dto.UserErrorResponse
// @Failure      404  {object}  dto.UserErrorResponse
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

	return c.JSON(http.StatusOK, dto.UserProfileResponse{
		Username: mailbox.Username,
		Name:     mailbox.Name,
	})
}

// GetUserForwarding godoc
// @Summary      Get forwarding addresses
// @Description  Returns the current comma-separated forwarding destinations for the authenticated mailbox.
// @Tags         User Portal
// @Produce      json
// @Success      200  {object}  dto.UserForwardingResponse
// @Failure      401  {object}  dto.UserErrorResponse
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

	return c.JSON(http.StatusOK, dto.UserForwardingResponse{Goto: gotoStr})
}

// UpdateUserForwardingAPI godoc
// @Summary      Update forwarding addresses
// @Description  Sets the email forwarding destinations. Send one address per line; leave blank to reset to self-delivery.
// @Tags         User Portal
// @Accept       json
// @Produce      json
// @Param        request  body      dto.UpdateForwardingRequest  true  "Forwarding configuration"
// @Success      200      {object}  dto.UserSuccessResponse
// @Failure      400      {object}  dto.UserErrorResponse
// @Failure      401      {object}  dto.UserErrorResponse
// @Failure      500      {object}  dto.UserErrorResponse
// @Router       /user/forwarding [post]
// @Security     BearerAuth
func (h *Handler) UpdateUserForwardingAPI(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil || claims.Type != "mailbox" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	var req dto.UpdateForwardingRequest
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

	return c.JSON(http.StatusOK, dto.UserSuccessResponse{
		Success: true,
		Message: "Forwarding updated successfully",
	})
}

// UpdateUserPasswordAPI godoc
// @Summary      Change mailbox password
// @Description  Allows the authenticated mailbox user to change their own password. The current password must be provided for verification.
// @Tags         User Portal
// @Accept       json
// @Produce      json
// @Param        request  body      dto.UpdatePasswordRequest  true  "Password change payload"
// @Success      200      {object}  dto.UserSuccessResponse
// @Failure      400      {object}  dto.UserErrorResponse  "Wrong current password, validation error, or passwords do not match"
// @Failure      401      {object}  dto.UserErrorResponse
// @Failure      500      {object}  dto.UserErrorResponse
// @Router       /user/password [post]
// @Security     BearerAuth
func (h *Handler) UpdateUserPasswordAPI(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil || claims.Type != "mailbox" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	var req dto.UpdatePasswordRequest
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
		return c.JSON(http.StatusBadRequest, dto.UserErrorResponse{Success: false, Error: "Current password is incorrect"})
	}

	if validationErr := ValidatePassword(req.NewPassword); validationErr != "" {
		return c.JSON(http.StatusBadRequest, dto.UserErrorResponse{Success: false, Error: validationErr})
	}

	if req.NewPassword != req.ConfirmPassword {
		return c.JSON(http.StatusBadRequest, dto.UserErrorResponse{Success: false, Error: "Passwords do not match"})
	}

	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.UserErrorResponse{Success: false, Error: "Failed to process the new password"})
	}

	mailbox.Password = hashedPassword
	if err := repositories.SaveMailbox(h.DB, mailbox, "USER_EDIT_PASSWORD", username, c.RealIP()); err != nil {
		return c.JSON(http.StatusInternalServerError, dto.UserErrorResponse{Success: false, Error: "Failed to update password"})
	}

	return c.JSON(http.StatusOK, dto.UserSuccessResponse{
		Success: true,
		Message: "Password updated successfully",
	})
}

// GetUserVacation godoc
// @Summary      Get auto-reply / vacation settings
// @Description  Returns the current vacation (out-of-office) configuration for the authenticated mailbox. Returns defaults when no vacation is configured.
// @Tags         User Portal
// @Produce      json
// @Success      200  {object}  dto.VacationResponse
// @Failure      401  {object}  dto.UserErrorResponse
// @Router       /user/vacation [get]
// @Security     BearerAuth
func (h *Handler) GetUserVacation(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil || claims.Type != "mailbox" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	vacation, err := repositories.GetVacationByEmail(h.DB, claims.Username)
	if err != nil {
		return c.JSON(http.StatusOK, dto.VacationResponse{
			Active:       false,
			Subject:      "Out of Office",
			Body:         "",
			ActiveFrom:   time.Now().Format("2006-01-02T15:04"),
			ActiveUntil:  time.Now().AddDate(0, 0, 7).Format("2006-01-02T15:04"),
			IntervalTime: 0,
		})
	}

	return c.JSON(http.StatusOK, dto.VacationResponse{
		Active:       vacation.Active,
		Subject:      vacation.Subject,
		Body:         vacation.Body,
		ActiveFrom:   vacation.ActiveFrom.Format("2006-01-02T15:04"),
		ActiveUntil:  vacation.ActiveUntil.Format("2006-01-02T15:04"),
		IntervalTime: vacation.IntervalTime,
	})
}

// UpdateUserVacationAPI godoc
// @Summary      Save auto-reply / vacation settings
// @Description  Creates or updates the vacation (out-of-office) configuration. Date fields use format "2006-01-02T15:04". interval_time is in seconds (0 = send once per sender).
// @Tags         User Portal
// @Accept       json
// @Produce      json
// @Param        request  body      dto.UpdateVacationRequest  true  "Vacation configuration"
// @Success      200      {object}  dto.UserSuccessResponse
// @Failure      400      {object}  dto.UserErrorResponse
// @Failure      401      {object}  dto.UserErrorResponse
// @Failure      500      {object}  dto.UserErrorResponse
// @Router       /user/vacation [post]
// @Security     BearerAuth
func (h *Handler) UpdateUserVacationAPI(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil || claims.Type != "mailbox" {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	var req dto.UpdateVacationRequest
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

	return c.JSON(http.StatusOK, dto.UserSuccessResponse{
		Success: true,
		Message: "Auto-reply updated successfully",
	})
}

// DeleteUserVacationAPI godoc
// @Summary      Remove auto-reply settings
// @Description  Deletes the vacation (out-of-office) configuration for the authenticated mailbox.
// @Tags         User Portal
// @Produce      json
// @Success      200  {object}  dto.UserSuccessResponse
// @Failure      401  {object}  dto.UserErrorResponse
// @Failure      500  {object}  dto.UserErrorResponse
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

	return c.JSON(http.StatusOK, dto.UserSuccessResponse{
		Success: true,
		Message: "Auto-reply removed successfully",
	})
}
