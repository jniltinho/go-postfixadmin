package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"go-postfixadmin/internal/middleware"
	"go-postfixadmin/internal/models"
	"go-postfixadmin/internal/repositories"
	"go-postfixadmin/internal/utils"

	"github.com/labstack/echo/v5"
)

type AdminData struct {
	models.Admin
	DomainCount string
}

// ListAdmins displays the list of administrators
func (h *Handler) ListAdmins(c *echo.Context) error {
	username := middleware.GetUsername(c, middleware.SessionName)
	isSuper, err := repositories.IsSuperAdmin(h.DB, username)
	if err != nil {
		return c.Render(http.StatusInternalServerError, "dashboard.html", map[string]interface{}{"Error": "Permission check failed"})
	}

	admins, err := repositories.GetAllAdmins(h.DB, username, isSuper)
	if err != nil {
		return c.Render(http.StatusInternalServerError, "admins/admins.html", map[string]interface{}{
			"error": "Failed to fetch administrators",
		})
	}

	var adminList []AdminData
	for _, admin := range admins {
		var domainCountStr string
		if admin.Superadmin {
			domainCountStr = "ALL"
		} else {
			count, _ := repositories.CountAdminDomains(h.DB, admin.Username)
			domainCountStr = fmt.Sprintf("%d", count)
		}
		adminList = append(adminList, AdminData{Admin: admin, DomainCount: domainCountStr})
	}

	var domains []models.Domain
	if isSuper {
		domains, _, _ = repositories.GetActiveDomains(h.DB, username, true)
	}

	return c.Render(http.StatusOK, "admins/admins.html", map[string]interface{}{
		"Admins":       adminList,
		"Domains":      domains,
		"IsSuperAdmin": isSuper,
		"SessionUser":  username,
		"Message":      GetFlash(c, "message"),
		"Error":        GetFlash(c, "error"),
	})
}

// AddAdminAPI processes the creation of a new administrator via JSON API
func (h *Handler) AddAdminAPI(c *echo.Context) error {
	loggedInUser := middleware.GetUsername(c, middleware.SessionName)
	isSuper, err := repositories.IsSuperAdmin(h.DB, loggedInUser)
	if err != nil || !isSuper {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"success": false, "error": "Access denied"})
	}

	username := c.FormValue("username")
	password := c.FormValue("password")
	passwordConfirm := c.FormValue("password_confirm")
	active := c.FormValue("active") == "true"
	superadmin := c.FormValue("superadmin") == "true"
	domains := c.Request().Form["domains"]

	if username == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": "Username is required"})
	}
	if validationErr := ValidatePassword(password); validationErr != "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": validationErr})
	}
	if password != passwordConfirm {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": "Passwords do not match"})
	}

	if _, err := repositories.GetAdminByUsername(h.DB, username); err == nil {
		return c.JSON(http.StatusConflict, map[string]interface{}{"success": false, "error": "Admin already exists"})
	}

	crypted, err := utils.HashPassword(password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": "Failed to hash password"})
	}

	newAdmin := models.Admin{
		Username:      username,
		Password:      crypted,
		Created:       time.Now(),
		Modified:      time.Now(),
		Active:        active,
		Superadmin:    superadmin,
		TokenValidity: time.Now().Add(3 * time.Hour),
	}

	if err := repositories.CreateAdmin(h.DB, newAdmin, domains, loggedInUser, c.RealIP()); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": "Failed to create administrator"})
	}

	SetFlash(c, "message", "Admin created successfully")
	return c.JSON(http.StatusOK, map[string]interface{}{"success": true})
}

// DeleteAdmin handles the deletion of an administrator
func (h *Handler) DeleteAdmin(c *echo.Context) error {
	loggedInUser := middleware.GetUsername(c, middleware.SessionName)
	isSuper, err := repositories.IsSuperAdmin(h.DB, loggedInUser)
	if err != nil || !isSuper {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"error": "Access denied"})
	}

	username, _ := url.PathUnescape(c.Param("username"))
	if username == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": "Username is required"})
	}

	if err := repositories.DeleteAdmin(h.DB, username, loggedInUser, c.RealIP()); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": "Failed to delete administrator"})
	}

	SetFlash(c, "message", "Admin deleted successfully")
	return c.JSON(http.StatusOK, map[string]interface{}{"success": true})
}

// GetAdminAPI fetches a single administrator details for the edit modal
func (h *Handler) GetAdminAPI(c *echo.Context) error {
	loggedInUser := middleware.GetUsername(c, middleware.SessionName)
	isSuper, err := repositories.IsSuperAdmin(h.DB, loggedInUser)
	if err != nil {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"error": "Permission check failed"})
	}

	targetUsername, _ := url.PathUnescape(c.Param("username"))
	if !isSuper && loggedInUser != targetUsername {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"error": "Access denied"})
	}
	if targetUsername == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Username required"})
	}

	admin, err := repositories.GetAdminByUsername(h.DB, targetUsername)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{"error": "Admin not found"})
	}

	allDomains, _, _ := repositories.GetActiveDomains(h.DB, loggedInUser, true)
	domainAdmins, _ := repositories.GetAdminAssignedDomains(h.DB, targetUsername)

	assignedMap := make(map[string]bool)
	for _, da := range domainAdmins {
		assignedMap[da.Domain] = true
	}

	type DomainOption struct {
		Domain   string `json:"domain"`
		Assigned bool   `json:"assigned"`
	}
	var domainOptions []DomainOption
	for _, d := range allDomains {
		domainOptions = append(domainOptions, DomainOption{Domain: d.Domain, Assigned: assignedMap[d.Domain]})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"admin":   admin,
		"domains": domainOptions,
	})
}

// EditAdminAPI processes the update of an administrator via JSON API
func (h *Handler) EditAdminAPI(c *echo.Context) error {
	loggedInUser := middleware.GetUsername(c, middleware.SessionName)
	isSuper, err := repositories.IsSuperAdmin(h.DB, loggedInUser)
	if err != nil {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"success": false, "error": "Permission check failed"})
	}

	targetUsername, _ := url.PathUnescape(c.Param("username"))
	if !isSuper && loggedInUser != targetUsername {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"success": false, "error": "Access denied"})
	}
	if targetUsername == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": "Username required"})
	}

	password := c.FormValue("password")
	passwordConfirm := c.FormValue("password_confirm")
	active := c.FormValue("active") == "true"
	superadmin := c.FormValue("superadmin") == "true"
	changePassword := c.FormValue("change_password") == "true"
	domains := c.Request().Form["domains"]

	updates := map[string]interface{}{
		"modified":       time.Now(),
		"token_validity": time.Now().Add(3 * time.Hour),
	}
	if isSuper {
		updates["active"] = active
		updates["superadmin"] = superadmin
	}

	if changePassword && password != "" {
		if validationErr := ValidatePassword(password); validationErr != "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": validationErr})
		}
		if password != passwordConfirm {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": "Passwords do not match"})
		}
		crypted, err := utils.HashPassword(password)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": "Failed to hash password"})
		}
		updates["password"] = crypted
	}

	if err := repositories.UpdateAdmin(h.DB, targetUsername, updates, domains, isSuper, loggedInUser, c.RealIP()); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": "Failed to update admin"})
	}

	SetFlash(c, "message", "Admin updated successfully")
	return c.JSON(http.StatusOK, map[string]interface{}{"success": true})
}
