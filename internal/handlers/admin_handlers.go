package handlers

import (
	"fmt"
	"net/http"
	"time"

	"go-postfixadmin/internal/models"
	"go-postfixadmin/internal/utils"

	"github.com/labstack/echo/v5"
)

type AdminData struct {
	models.Admin
	DomainCount string
}

// ListAdmins displays the list of administrators
func (h *Handler) ListAdmins(c *echo.Context) error {
	var admins []models.Admin

	if h.DB == nil {
		return c.Render(http.StatusInternalServerError, "admins.html", map[string]interface{}{
			"error": "Database connection unavailable",
		})
	}

	if err := h.DB.Find(&admins).Error; err != nil {
		return c.Render(http.StatusInternalServerError, "admins.html", map[string]interface{}{
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

	return c.Render(http.StatusOK, "admins.html", map[string]interface{}{
		"Admins": adminList,
	})
}

// AddAdminForm displays the form to add a new administrator
// AddAdminForm displays the form to add a new administrator
func (h *Handler) AddAdminForm(c *echo.Context) error {
	var domains []models.Domain

	if h.DB != nil {
		h.DB.Where("domain != ?", "ALL").Order("domain ASC").Find(&domains)
	}

	return c.Render(http.StatusOK, "add_admin.html", map[string]interface{}{
		"Domains": domains,
	})
}

// AddAdmin processes the creation of a new administrator
func (h *Handler) AddAdmin(c *echo.Context) error {
	// Get form values
	username := c.FormValue("username")
	password := c.FormValue("password")
	passwordConfirm := c.FormValue("password_confirm")
	active := c.FormValue("active") == "true"
	superadmin := c.FormValue("superadmin") == "true"
	domains := c.Request().Form["domains"]

	// Basic Validation
	if username == "" {
		return h.renderAddAdminError(c, "O nome de usuário é obrigatório", username)
	}
	if len(password) < 8 {
		return h.renderAddAdminError(c, "A senha deve ter no mínimo 8 caracteres", username)
	}
	if password != passwordConfirm {
		return h.renderAddAdminError(c, "As senhas não conferem", username)
	}

	// Check if admin already exists
	var existingAdmin models.Admin
	if err := h.DB.Where("username = ?", username).First(&existingAdmin).Error; err == nil {
		return h.renderAddAdminError(c, "O administrador já existe", username)
	}

	// Hash password
	crypted, err := utils.HashPassword(password)
	if err != nil {
		return h.renderAddAdminError(c, "Falha ao gerar hash da senha: "+err.Error(), username)
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
		TokenValidity: time.Now(),
	}

	if err := tx.Create(&newAdmin).Error; err != nil {
		tx.Rollback()
		return h.renderAddAdminError(c, "Falha ao criar administrador: "+err.Error(), username)
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
			return h.renderAddAdminError(c, "Falha ao atribuir domínio ALL: "+err.Error(), username)
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
				return h.renderAddAdminError(c, "Falha ao atribuir domínio "+d+": "+err.Error(), username)
			}
		}
	}

	tx.Commit()

	return c.Redirect(http.StatusFound, "/admins")
}

// renderAddAdminError helper to render the form with error message
func (h *Handler) renderAddAdminError(c *echo.Context, errorMsg, username string) error {
	var domains []models.Domain
	if h.DB != nil {
		h.DB.Where("domain != ?", "ALL").Order("domain ASC").Find(&domains)
	}

	return c.Render(http.StatusBadRequest, "add_admin.html", map[string]interface{}{
		"Error":    errorMsg,
		"Username": username,
		"Domains":  domains,
	})
}

// DeleteAdmin handles the deletion of an administrator
func (h *Handler) DeleteAdmin(c *echo.Context) error {
	username := c.Param("username")
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

	tx.Commit()

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
	})
}

// EditAdminForm displays the form to edit an administrator
func (h *Handler) EditAdminForm(c *echo.Context) error {
	username := c.Param("username")
	if username == "" {
		return c.Redirect(http.StatusFound, "/admins")
	}

	var admin models.Admin
	if err := h.DB.First(&admin, "username = ?", username).Error; err != nil {
		return c.Redirect(http.StatusFound, "/admins")
	}

	// Fetch all domains
	var allDomains []models.Domain
	h.DB.Find(&allDomains)

	// Fetch assigned domains for this admin
	var domainAdmins []models.DomainAdmin
	h.DB.Where("username = ?", username).Find(&domainAdmins)

	assignedMap := make(map[string]bool)
	for _, da := range domainAdmins {
		assignedMap[da.Domain] = true
	}

	type DomainOption struct {
		Domain   string
		Assigned bool
	}

	var domainOptions []DomainOption
	for _, d := range allDomains {
		if d.Domain == "ALL" {
			continue
		}
		domainOptions = append(domainOptions, DomainOption{
			Domain:   d.Domain,
			Assigned: assignedMap[d.Domain],
		})
	}

	return c.Render(http.StatusOK, "edit_admin.html", map[string]interface{}{
		"Admin":   admin,
		"Domains": domainOptions,
	})
}

// EditAdmin processes the update of an administrator
func (h *Handler) EditAdmin(c *echo.Context) error {
	username := c.Param("username")
	if username == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Username required"})
	}

	// Get form values
	password := c.FormValue("password")
	active := c.FormValue("active") == "true"
	superadmin := c.FormValue("superadmin") == "true"
	domains := c.Request().Form["domains"] // Helper to get multiple values for checkbox array

	// Using a transaction
	tx := h.DB.Begin()

	// 1. Update Admin fields
	updates := map[string]interface{}{
		"active":         active,
		"superadmin":     superadmin,
		"modified":       time.Now(),
		"token_validity": time.Now(),
	}

	if password != "" {
		crypted, err := utils.HashPassword(password)
		if err != nil {
			tx.Rollback()
			return c.Render(http.StatusOK, "edit_admin.html", map[string]interface{}{
				"Admin": models.Admin{Username: username}, // minimal data to re-render? ideally we re-fetch
				"Error": "Failed to hash password",
			})
		}
		updates["password"] = crypted
	}

	if err := tx.Model(&models.Admin{}).Where("username = ?", username).Updates(updates).Error; err != nil {
		tx.Rollback()
		return c.Render(http.StatusOK, "edit_admin.html", map[string]interface{}{
			"Error": "Failed to update admin: " + err.Error(),
		})
	}

	// 2. Update Domain Assignments
	// First, remove all existing assignments
	if err := tx.Where("username = ?", username).Delete(&models.DomainAdmin{}).Error; err != nil {
		tx.Rollback()
		return c.Render(http.StatusOK, "edit_admin.html", map[string]interface{}{
			"Error": "Failed to update domain permissions: " + err.Error(),
		})
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
			return c.Render(http.StatusOK, "edit_admin.html", map[string]interface{}{
				"Error": "Failed to assign domain ALL: " + err.Error(),
			})
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
				return c.Render(http.StatusOK, "edit_admin.html", map[string]interface{}{
					"Error": "Failed to assign domain " + d + ": " + err.Error(),
				})
			}
		}
	}

	tx.Commit()

	return c.Redirect(http.StatusFound, "/admins")
}
