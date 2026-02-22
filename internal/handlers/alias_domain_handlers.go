package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"go-postfixadmin/internal/middleware"
	"go-postfixadmin/internal/models"
	"go-postfixadmin/internal/utils"

	"github.com/labstack/echo/v5"
)

// ListAliasDomains lists all alias domains
func (h *Handler) ListAliasDomains(c *echo.Context) error {
	var aliasDomains []models.AliasDomain
	var isSuperAdmin bool

	if h.DB != nil {
		query := h.DB.Table("alias_domain").Select("alias_domain.*").
			Joins("JOIN domain ON alias_domain.target_domain = domain.domain").
			Where("domain.active = ?", true).
			Order("alias_domain.alias_domain ASC")

		// Security: Filter by allowed domains
		username := middleware.GetUsername(c)
		allowedDomains, isSuper, err := utils.GetAllowedDomains(h.DB, username, middleware.GetIsSuperAdmin(c))
		if err != nil {
			return c.Render(http.StatusInternalServerError, "alias_domains.html", map[string]interface{}{
				"Error": "Failed to check permissions: " + err.Error(),
			})
		}
		isSuperAdmin = isSuper

		if !isSuperAdmin {
			if len(allowedDomains) == 0 {
				query = query.Where("1 = 0") // No domains allowed
			} else {
				query = query.Where("alias_domain.target_domain IN ?", allowedDomains)
			}
		}

		query.Find(&aliasDomains)
	}

	return c.Render(http.StatusOK, "alias_domains.html", map[string]interface{}{
		"AliasDomains": aliasDomains,
		"IsSuperAdmin": isSuperAdmin,
		"SessionUser":  middleware.GetUsername(c),
	})
}

// AddAliasDomainForm renders the add alias domain form
func (h *Handler) AddAliasDomainForm(c *echo.Context) error {
	var domains []models.Domain
	var isSuperAdmin bool

	if h.DB != nil {
		// Security: Filter domains
		username := middleware.GetUsername(c)
		allowedDomains, isSuper, err := utils.GetAllowedDomains(h.DB, username, middleware.GetIsSuperAdmin(c))
		if err != nil {
			return c.Render(http.StatusInternalServerError, "add_alias_domain.html", map[string]interface{}{"Error": "Permission check failed"})
		}
		isSuperAdmin = isSuper

		query := h.DB.Where("domain != ?", "ALL").Where("active = ?", true).Order("domain ASC")
		if !isSuperAdmin {
			if len(allowedDomains) == 0 {
				query = query.Where("1 = 0")
			} else {
				query = query.Where("domain IN ?", allowedDomains)
			}
		}
		query.Find(&domains)
	}

	return c.Render(http.StatusOK, "add_alias_domain.html", map[string]interface{}{
		"Domains":      domains,
		"IsSuperAdmin": isSuperAdmin,
		"SessionUser":  middleware.GetUsername(c),
	})
}

// AddAliasDomain processes the addition of a new alias domain
func (h *Handler) AddAliasDomain(c *echo.Context) error {
	aliasDomain := c.FormValue("alias_domain")
	targetDomain := c.FormValue("target_domain")
	active := c.FormValue("active") == "true"

	// Security: Validate target domain access
	loggedInUser := middleware.GetUsername(c)
	allowedDomains, isSuperAdmin, err := utils.GetAllowedDomains(h.DB, loggedInUser, middleware.GetIsSuperAdmin(c))
	if err != nil {
		return renderAddAliasDomainError(c, "Permission check failed", aliasDomain, targetDomain, nil, isSuperAdmin)
	}

	if !isSuperAdmin {
		allowed := false
		for _, d := range allowedDomains {
			if d == targetDomain {
				allowed = true
				break
			}
		}
		if !allowed {
			return renderAddAliasDomainError(c, "Access denied to this target domain", aliasDomain, targetDomain, nil, isSuperAdmin)
		}
	}

	// Load domains for re-rendering on error
	var domains []models.Domain
	if h.DB != nil {
		query := h.DB.Where("domain != ?", "ALL").Where("active = ?", true).Order("domain ASC")
		if !isSuperAdmin {
			if len(allowedDomains) == 0 {
				query = query.Where("1 = 0")
			} else {
				query = query.Where("domain IN ?", allowedDomains)
			}
		}
		query.Find(&domains)
	}

	if aliasDomain == "" {
		return renderAddAliasDomainError(c, "Alias Domain is required", aliasDomain, targetDomain, domains, isSuperAdmin)
	}
	if targetDomain == "" {
		return renderAddAliasDomainError(c, "Target Domain is required", aliasDomain, targetDomain, domains, isSuperAdmin)
	}

	// Validate: Alias Domain cannot be equal to Target Domain
	if aliasDomain == targetDomain {
		return renderAddAliasDomainError(c, "O domínio de origem e destino não podem ser iguais", aliasDomain, targetDomain, domains, isSuperAdmin)
	}

	// Check if already exists
	var existing models.AliasDomain
	if err := h.DB.Where("alias_domain = ?", aliasDomain).First(&existing).Error; err == nil {
		return renderAddAliasDomainError(c, "Alias Domain already exists", aliasDomain, targetDomain, domains, isSuperAdmin)
	}

	// Check if target domain exists and is active
	var target models.Domain
	if err := h.DB.Where("domain = ?", targetDomain).First(&target).Error; err != nil {
		return renderAddAliasDomainError(c, "Target Domain does not exist", aliasDomain, targetDomain, domains, isSuperAdmin)
	}

	now := time.Now()
	newAliasDomain := models.AliasDomain{
		AliasDomain:  aliasDomain,
		TargetDomain: targetDomain,
		Created:      now,
		Modified:     now,
		Active:       active,
	}

	if err := h.DB.Create(&newAliasDomain).Error; err != nil {
		return renderAddAliasDomainError(c, "Failed to create alias domain: "+err.Error(), aliasDomain, targetDomain, domains, isSuperAdmin)
	}

	// Log Action
	if err := utils.LogAction(h.DB, middleware.GetUsername(c), c.RealIP(), targetDomain, "create_alias_domain", aliasDomain); err != nil {
		fmt.Printf("Failed to log create_alias_domain: %v\n", err)
	}

	return c.Redirect(http.StatusFound, "/alias-domains")
}

// DeleteAliasDomain handles alias domain deletion
func (h *Handler) DeleteAliasDomain(c *echo.Context) error {
	aliasDomainName := c.Param("alias_domain")

	decodedName, err := url.QueryUnescape(aliasDomainName)
	if err == nil {
		aliasDomainName = decodedName
	}

	if aliasDomainName == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Alias Domain required"})
	}

	if h.DB == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{"error": "Database unavailable"})
	}

	// Fetch to check permission
	var aliasDomain models.AliasDomain
	if err := h.DB.Where("alias_domain = ?", aliasDomainName).First(&aliasDomain).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{"error": "Alias Domain not found"})
	}

	loggedInUser := middleware.GetUsername(c)
	allowedDomains, isSuperAdmin, err := utils.GetAllowedDomains(h.DB, loggedInUser, middleware.GetIsSuperAdmin(c))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "Permission check failed"})
	}

	if !isSuperAdmin {
		allowed := false
		for _, d := range allowedDomains {
			if d == aliasDomain.TargetDomain {
				allowed = true
				break
			}
		}
		if !allowed {
			return c.JSON(http.StatusForbidden, map[string]interface{}{"error": "Access denied"})
		}
	}

	if err := h.DB.Delete(&aliasDomain).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "Failed to delete alias domain"})
	}

	// Log Action
	if err := utils.LogAction(h.DB, middleware.GetUsername(c), c.RealIP(), aliasDomain.TargetDomain, "delete_alias_domain", aliasDomainName); err != nil {
		fmt.Printf("Failed to log delete_alias_domain: %v\n", err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"success": true})
}

func renderAddAliasDomainError(c *echo.Context, errorMsg, aliasDomain, targetDomain string, domains []models.Domain, isSuperAdmin bool) error {
	return c.Render(http.StatusBadRequest, "add_alias_domain.html", map[string]interface{}{
		"Error":        errorMsg,
		"AliasDomain":  aliasDomain,
		"TargetDomain": targetDomain,
		"Domains":      domains,
		"IsSuperAdmin": isSuperAdmin,
		"SessionUser":  middleware.GetUsername(c),
	})
}

// EditAliasDomainForm renders the edit alias domain form
func (h *Handler) EditAliasDomainForm(c *echo.Context) error {
	aliasDomainName := c.Param("alias_domain")

	decodedName, err := url.QueryUnescape(aliasDomainName)
	if err == nil {
		aliasDomainName = decodedName
	}

	var aliasDomain models.AliasDomain
	if err := h.DB.Where("alias_domain = ?", aliasDomainName).First(&aliasDomain).Error; err != nil {
		return c.Render(http.StatusNotFound, "alias_domains.html", map[string]interface{}{
			"Error": "Alias Domain not found",
		})
	}

	var domains []models.Domain
	var isSuperAdmin bool

	if h.DB != nil {
		username := middleware.GetUsername(c)
		allowedDomains, isSuper, err := utils.GetAllowedDomains(h.DB, username, middleware.GetIsSuperAdmin(c))
		if err != nil {
			return c.Render(http.StatusInternalServerError, "alias_domains.html", map[string]interface{}{"Error": "Permission check failed"})
		}
		isSuperAdmin = isSuper

		// Check if user has access to this alias domain (via target domain)
		if !isSuperAdmin {
			allowed := false
			for _, d := range allowedDomains {
				if d == aliasDomain.TargetDomain {
					allowed = true
					break
				}
			}
			if !allowed {
				return c.Render(http.StatusForbidden, "alias_domains.html", map[string]interface{}{"Error": "Access denied"})
			}
		}

		query := h.DB.Where("domain != ?", "ALL").Where("active = ?", true).Order("domain ASC")
		if !isSuperAdmin {
			if len(allowedDomains) == 0 {
				query = query.Where("1 = 0")
			} else {
				query = query.Where("domain IN ?", allowedDomains)
			}
		}
		query.Find(&domains)
	}

	return c.Render(http.StatusOK, "edit_alias_domain.html", map[string]interface{}{
		"AliasDomain":  aliasDomain,
		"Domains":      domains,
		"IsSuperAdmin": isSuperAdmin,
		"SessionUser":  middleware.GetUsername(c),
	})
}

// EditAliasDomain processes the alias domain update
func (h *Handler) EditAliasDomain(c *echo.Context) error {
	aliasDomainName := c.Param("alias_domain")
	decodedName, err := url.QueryUnescape(aliasDomainName)
	if err == nil {
		aliasDomainName = decodedName
	}

	var aliasDomain models.AliasDomain
	if err := h.DB.Where("alias_domain = ?", aliasDomainName).First(&aliasDomain).Error; err != nil {
		return c.Render(http.StatusNotFound, "alias_domains.html", map[string]interface{}{
			"Error": "Alias Domain not found",
		})
	}

	// Security Check
	loggedInUser := middleware.GetUsername(c)
	allowedDomains, isSuperAdmin, err := utils.GetAllowedDomains(h.DB, loggedInUser, middleware.GetIsSuperAdmin(c))
	if err != nil {
		return c.Render(http.StatusInternalServerError, "alias_domains.html", map[string]interface{}{"Error": "Permission check failed"})
	}

	if !isSuperAdmin {
		allowed := false
		for _, d := range allowedDomains {
			if d == aliasDomain.TargetDomain {
				allowed = true
				break
			}
		}
		if !allowed {
			return c.Render(http.StatusForbidden, "alias_domains.html", map[string]interface{}{"Error": "Access denied"})
		}
	}

	targetDomain := c.FormValue("target_domain")
	active := c.FormValue("active") == "true"

	// Validate target domain access (if changed)
	if targetDomain != aliasDomain.TargetDomain {
		if !isSuperAdmin {
			allowed := false
			for _, d := range allowedDomains {
				if d == targetDomain {
					allowed = true
					break
				}
			}
			if !allowed {
				return renderEditAliasDomainError(c, "Access denied to new target domain", aliasDomain, nil, isSuperAdmin) // pass domains if possible, or fetch again
			}
		}
	}

	// Re-fetch domains for error rendering
	var domains []models.Domain
	if h.DB != nil {
		query := h.DB.Where("domain != ?", "ALL").Where("active = ?", true).Order("domain ASC")
		if !isSuperAdmin {
			if len(allowedDomains) == 0 {
				query = query.Where("1 = 0")
			} else {
				query = query.Where("domain IN ?", allowedDomains)
			}
		}
		query.Find(&domains)
	}

	if targetDomain == "" {
		return renderEditAliasDomainError(c, "Target Domain is required", aliasDomain, domains, isSuperAdmin)
	}

	if aliasDomainName == targetDomain {
		return renderEditAliasDomainError(c, "O domínio de origem e destino não podem ser iguais", aliasDomain, domains, isSuperAdmin)
	}

	// Check if target domain exists
	var target models.Domain
	if err := h.DB.Where("domain = ?", targetDomain).First(&target).Error; err != nil {
		return renderEditAliasDomainError(c, "Target Domain does not exist", aliasDomain, domains, isSuperAdmin)
	}

	// Update
	aliasDomain.TargetDomain = targetDomain
	aliasDomain.Active = active
	aliasDomain.Modified = time.Now()

	if err := h.DB.Save(&aliasDomain).Error; err != nil {
		return renderEditAliasDomainError(c, "Failed to update alias domain: "+err.Error(), aliasDomain, domains, isSuperAdmin)
	}

	// Log Action
	if err := utils.LogAction(h.DB, middleware.GetUsername(c), c.RealIP(), targetDomain, "edit_alias_domain", aliasDomainName); err != nil {
		fmt.Printf("Failed to log edit_alias_domain: %v\n", err)
	}

	return c.Redirect(http.StatusFound, "/alias-domains")
}

func renderEditAliasDomainError(c *echo.Context, errorMsg string, aliasDomain models.AliasDomain, domains []models.Domain, isSuperAdmin bool) error {
	return c.Render(http.StatusBadRequest, "edit_alias_domain.html", map[string]interface{}{
		"Error":        errorMsg,
		"AliasDomain":  aliasDomain,
		"Domains":      domains,
		"IsSuperAdmin": isSuperAdmin,
		"SessionUser":  middleware.GetUsername(c),
	})
}
