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
	// Security: Superadmins see all, Admins see only themselves
	username := middleware.GetUsername(c, middleware.SessionName)
	isSuper, err := repositories.IsSuperAdmin(h.DB, username)
	if err != nil {
		return c.Render(http.StatusInternalServerError, "dashboard.html", map[string]interface{}{"Error": "Permission check failed"})
	}

	var admins []models.Admin

	if h.DB == nil {
		return c.Render(http.StatusInternalServerError, "admins/admins.html", map[string]interface{}{
			"error": "Database connection unavailable",
		})
	}

	query := h.DB
	if !isSuper {
		// Non-superadmins can only see themselves
		query = query.Where("username = ?", username)
	}

	if err := query.Find(&admins).Error; err != nil {
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
			var count int64
			h.DB.Model(&models.DomainAdmin{}).Where("username = ?", admin.Username).Count(&count)
			domainCountStr = fmt.Sprintf("%d", count)
		}

		adminList = append(adminList, AdminData{
			Admin:       admin,
			DomainCount: domainCountStr,
		})
	}

	var domains []models.Domain
	if h.DB != nil && isSuper {
		h.DB.Where("domain != ?", "ALL").Where("active = ?", true).Order("domain ASC").Find(&domains)
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
	// Security: Only Superadmins
	loggedInUser := middleware.GetUsername(c, middleware.SessionName)
	isSuper, err := repositories.IsSuperAdmin(h.DB, loggedInUser)
	if err != nil || !isSuper {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"success": false, "error": "Access denied"})
	}

	// Get form values
	username := c.FormValue("username")
	password := c.FormValue("password")
	passwordConfirm := c.FormValue("password_confirm")
	active := c.FormValue("active") == "true"
	superadmin := c.FormValue("superadmin") == "true"
	domains := c.Request().Form["domains"]

	// Basic Validation
	if username == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": "Username is required"})
	}
	if validationErr := ValidatePassword(password); validationErr != "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": validationErr})
	}
	if password != passwordConfirm {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": "Passwords do not match"})
	}

	// Check if admin already exists
	var existingAdmin models.Admin
	if err := h.DB.Where("username = ?", username).First(&existingAdmin).Error; err == nil {
		return c.JSON(http.StatusConflict, map[string]interface{}{"success": false, "error": "Admin already exists"})
	}

	// Hash password
	crypted, err := utils.HashPassword(password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": "Failed to hash password"})
	}

	// Begin transaction
	tx := h.DB.Begin()

	// Create Admin
	newAdmin := models.Admin{
		Username:      username,
		Password:      crypted,
		Created:       time.Now(),
		Modified:      time.Now(),
		Active:        active,
		Superadmin:    superadmin,
		TokenValidity: time.Now().Add(3 * time.Hour),
	}

	if err := tx.Create(&newAdmin).Error; err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": "Failed to create administrator"})
	}

	// Assign Domains
	if superadmin {
		// Superadmin gets "ALL" domain
		da := models.DomainAdmin{
			Username: username,
			Domain:   "ALL",
			Created:  time.Now(),
			Active:   true,
		}
		if err := tx.Create(&da).Error; err != nil {
			tx.Rollback()
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": "Failed to assign ALL domain"})
		}
	} else if len(domains) > 0 {
		// Normal admin gets selected domains
		for _, d := range domains {
			da := models.DomainAdmin{
				Username: username,
				Domain:   d,
				Created:  time.Now(),
				Active:   true,
			}
			if err := tx.Create(&da).Error; err != nil {
				tx.Rollback()
				return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": "Failed to assign domain " + d})
			}
		}
	}

	// Log Action
	if err := utils.LogAction(tx, loggedInUser, c.RealIP(), "ALL", "create_admin", username); err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": "Failed to log action"})
	}

	tx.Commit()

	SetFlash(c, "message", "Admin created successfully")

	return c.JSON(http.StatusOK, map[string]interface{}{"success": true})
}

// DeleteAdmin handles the deletion of an administrator
func (h *Handler) DeleteAdmin(c *echo.Context) error {
	// Security: Only Superadmins
	loggedInUser := middleware.GetUsername(c, middleware.SessionName)
	isSuper, err := repositories.IsSuperAdmin(h.DB, loggedInUser)
	if err != nil || !isSuper {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"error": "Access denied"})
	}

	username, _ := url.PathUnescape(c.Param("username"))
	if username == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Username is required",
		})
	}

	// Prevent deleting yourself? Maybe later.

	if h.DB == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"success": false,
			"error":   "Database connection unavailable",
		})
	}

	// Start transaction to delete from admin and domain_admins
	tx := h.DB.Begin()

	if err := tx.Where("username = ?", username).Delete(&models.DomainAdmin{}).Error; err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "Failed to delete associated domain permissions",
		})
	}

	if err := tx.Where("username = ?", username).Delete(&models.Admin{}).Error; err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "Failed to delete administrator",
		})
	}

	// Log Action
	// For delete_admin, we use "ALL" as domain context
	if err := utils.LogAction(tx, loggedInUser, c.RealIP(), "ALL", "delete_admin", username); err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "Failed to log action",
		})
	}

	tx.Commit()

	SetFlash(c, "message", "Admin deleted successfully")

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
	})
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

	var admin models.Admin
	if err := h.DB.First(&admin, "username = ?", targetUsername).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{"error": "Admin not found"})
	}

	// Fetch all domains
	var allDomains []models.Domain
	h.DB.Where("domain != ? AND active = ?", "ALL", true).Find(&allDomains)

	// Fetch assigned domains for this admin
	var domainAdmins []models.DomainAdmin
	h.DB.Where("username = ?", targetUsername).Find(&domainAdmins)

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
		domainOptions = append(domainOptions, DomainOption{
			Domain:   d.Domain,
			Assigned: assignedMap[d.Domain],
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"admin":   admin,
		"domains": domainOptions,
	})
}

// EditAdminAPI processes the update of an administrator via JSON API
func (h *Handler) EditAdminAPI(c *echo.Context) error {
	// Security: Superadmins OR Self
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

	// Get form values
	password := c.FormValue("password")
	passwordConfirm := c.FormValue("password_confirm")
	active := c.FormValue("active") == "true"
	superadmin := c.FormValue("superadmin") == "true"
	changePassword := c.FormValue("change_password") == "true"
	domains := c.Request().Form["domains"]

	// Using a transaction
	tx := h.DB.Begin()

	// 1. Update Admin fields
	updates := map[string]interface{}{
		"modified":       time.Now(),
		"token_validity": time.Now().Add(3 * time.Hour),
	}

	// Only Superadmins can change Active status and Superadmin role
	if isSuper {
		updates["active"] = active
		updates["superadmin"] = superadmin
	}

	if changePassword && password != "" {
		if validationErr := ValidatePassword(password); validationErr != "" {
			tx.Rollback()
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": validationErr})
		}
		if password != passwordConfirm {
			tx.Rollback()
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": "Passwords do not match"})
		}
		crypted, err := utils.HashPassword(password)
		if err != nil {
			tx.Rollback()
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": "Failed to hash password"})
		}
		updates["password"] = crypted
	}

	if err := tx.Model(&models.Admin{}).Where("username = ?", targetUsername).Updates(updates).Error; err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": "Failed to update admin"})
	}

	// 2. Update Domain Assignments - Only for Superadmins
	if isSuper {
		// First, remove all existing assignments
		if err := tx.Where("username = ?", targetUsername).Delete(&models.DomainAdmin{}).Error; err != nil {
			tx.Rollback()
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": "Failed to clear domain permissions"})
		}

		// Assign Domains
		if superadmin {
			da := models.DomainAdmin{
				Username: targetUsername,
				Domain:   "ALL",
				Created:  time.Now(),
				Active:   true,
			}
			if err := tx.Create(&da).Error; err != nil {
				tx.Rollback()
				return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": "Failed to assign ALL domain"})
			}
		} else if len(domains) > 0 {
			for _, d := range domains {
				da := models.DomainAdmin{
					Username: targetUsername,
					Domain:   d,
					Created:  time.Now(),
					Active:   true,
				}
				if err := tx.Create(&da).Error; err != nil {
					tx.Rollback()
					return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": "Failed to assign domain " + d})
				}
			}
		}
	}

	// Log Action
	if err := utils.LogAction(tx, loggedInUser, c.RealIP(), "ALL", "edit_admin", targetUsername); err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": "Failed to log action"})
	}

	tx.Commit()

	SetFlash(c, "message", "Admin updated successfully")

	return c.JSON(http.StatusOK, map[string]interface{}{"success": true})
}
