package handlers

import (
	"fmt"
	"net/http"

	"go-postfixadmin/internal/models"

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
func (h *Handler) AddAdminForm(c *echo.Context) error {
	// Placeholder for now, just renders the form (which we need to create later if needed, but for now just the button exists)
	// The user request specified creating the "Create Administrator" button, but not the full creation flow yet.
	// However, to make the button work without 404, we need a route.
	// We might need a template for adding admins. For now, let's reuse a generic placeholder or create a simple one.
	return c.Render(http.StatusOK, "add_admin.html", nil)
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
	return c.Render(http.StatusOK, "edit_admin.html", nil)
}
