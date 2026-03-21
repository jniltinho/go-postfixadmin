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
		username := middleware.GetUsername(c, middleware.SessionName)
		allowedDomains, isSuper, err := utils.GetAllowedDomains(h.DB, username, middleware.GetIsSuperAdmin(c))
		if err != nil {
			return c.Render(http.StatusInternalServerError, "alias_domains/alias_domains.html", map[string]interface{}{
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

	return c.Render(http.StatusOK, "alias_domains/alias_domains.html", map[string]interface{}{
		"AliasDomains": aliasDomains,
		"IsSuperAdmin": isSuperAdmin,
		"SessionUser":  middleware.GetUsername(c, middleware.SessionName),
		"Message":      GetFlash(c, "message"),
		"Error":        GetFlash(c, "error"),
	})
}

// AddAliasDomainAPI processes the addition of a new alias domain via JSON API
func (h *Handler) AddAliasDomainAPI(c *echo.Context) error {
	aliasDomain := c.FormValue("alias_domain")
	targetDomain := c.FormValue("target_domain")
	active := c.FormValue("active") == "true"

	loggedInUser := middleware.GetUsername(c, middleware.SessionName)
	allowedDomains, isSuperAdmin, err := utils.GetAllowedDomains(h.DB, loggedInUser, middleware.GetIsSuperAdmin(c))
	if err != nil {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"success": false, "error": "Permission check failed"})
	}

	if !isSuperAdmin {
		allowed := false
		for _, d := range allowedDomains {
			if d == targetDomain || d == aliasDomain {
				allowed = true
				break
			}
		}
		if !allowed {
			return c.JSON(http.StatusForbidden, map[string]interface{}{"success": false, "error": "Access denied to these domains"}) // Needs access to target at least according to Postfixadmin logic usually, but let's just assert target domain access
		}
	}

	if aliasDomain == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": "Alias Domain is required"})
	}
	if targetDomain == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": "Target Domain is required"})
	}
	if aliasDomain == targetDomain {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": "O domínio de origem e destino não podem ser iguais"})
	}

	var existing models.AliasDomain
	if err := h.DB.Where("alias_domain = ?", aliasDomain).First(&existing).Error; err == nil {
		return c.JSON(http.StatusConflict, map[string]interface{}{"success": false, "error": "Alias Domain already exists"})
	}

	var target models.Domain
	if err := h.DB.Where("domain = ?", targetDomain).First(&target).Error; err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": "Target Domain does not exist"})
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
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": "Failed to create alias domain"})
	}

	if err := utils.LogAction(h.DB, loggedInUser, c.RealIP(), targetDomain, "create_alias_domain", aliasDomain); err != nil {
		fmt.Printf("Failed to log create_alias_domain: %v\n", err)
	}

	SetFlash(c, "message", "Alias Domain created successfully")
	return c.JSON(http.StatusOK, map[string]interface{}{"success": true})
}

// DeleteAliasDomain handles alias domain deletion
func (h *Handler) DeleteAliasDomain(c *echo.Context) error {
	aliasDomainName, _ := url.PathUnescape(c.Param("alias_domain"))

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

	loggedInUser := middleware.GetUsername(c, middleware.SessionName)
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
	if err := utils.LogAction(h.DB, middleware.GetUsername(c, middleware.SessionName), c.RealIP(), aliasDomain.TargetDomain, "delete_alias_domain", aliasDomainName); err != nil {
		fmt.Printf("Failed to log delete_alias_domain: %v\n", err)
	}

	SetFlash(c, "message", "Alias Domain deleted successfully")

	return c.JSON(http.StatusOK, map[string]interface{}{"success": true})
}


// GetAliasDomainAPI fetches alias domain details for the edit modal via JSON
func (h *Handler) GetAliasDomainAPI(c *echo.Context) error {
	aliasDomainName, _ := url.PathUnescape(c.Param("alias_domain"))
	decodedName, err := url.QueryUnescape(aliasDomainName)
	if err == nil {
		aliasDomainName = decodedName
	}

	var aliasDomain models.AliasDomain
	if err := h.DB.Where("alias_domain = ?", aliasDomainName).First(&aliasDomain).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{"error": "Alias Domain not found"})
	}

	var domains []models.Domain
	if h.DB != nil {
		loggedInUser := middleware.GetUsername(c, middleware.SessionName)
		allowedDomains, isSuperAdmin, err := utils.GetAllowedDomains(h.DB, loggedInUser, middleware.GetIsSuperAdmin(c))
		if err != nil {
			return c.JSON(http.StatusForbidden, map[string]interface{}{"error": "Permission check failed"})
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

	return c.JSON(http.StatusOK, map[string]interface{}{
		"alias_domain": aliasDomain,
		"domains":      domains,
	})
}

// EditAliasDomainAPI processes the alias domain update via JSON
func (h *Handler) EditAliasDomainAPI(c *echo.Context) error {
	aliasDomainName, _ := url.PathUnescape(c.Param("alias_domain"))
	decodedName, err := url.QueryUnescape(aliasDomainName)
	if err == nil {
		aliasDomainName = decodedName
	}

	var aliasDomain models.AliasDomain
	if err := h.DB.Where("alias_domain = ?", aliasDomainName).First(&aliasDomain).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{"success": false, "error": "Alias Domain not found"})
	}

	loggedInUser := middleware.GetUsername(c, middleware.SessionName)
	allowedDomains, isSuperAdmin, err := utils.GetAllowedDomains(h.DB, loggedInUser, middleware.GetIsSuperAdmin(c))
	if err != nil {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"success": false, "error": "Permission check failed"})
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
			return c.JSON(http.StatusForbidden, map[string]interface{}{"success": false, "error": "Access denied"})
		}
	}

	targetDomain := c.FormValue("target_domain")
	active := c.FormValue("active") == "true"

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
				return c.JSON(http.StatusForbidden, map[string]interface{}{"success": false, "error": "Access denied to new target domain"})
			}
		}
	}

	if targetDomain == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": "Target Domain is required"})
	}

	if aliasDomainName == targetDomain {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": "O domínio de origem e destino não podem ser iguais"})
	}

	var target models.Domain
	if err := h.DB.Where("domain = ?", targetDomain).First(&target).Error; err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": "Target Domain does not exist"})
	}

	aliasDomain.TargetDomain = targetDomain
	aliasDomain.Active = active
	aliasDomain.Modified = time.Now()

	if err := h.DB.Save(&aliasDomain).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": "Failed to update alias domain"})
	}

	if err := utils.LogAction(h.DB, loggedInUser, c.RealIP(), targetDomain, "edit_alias_domain", aliasDomainName); err != nil {
		fmt.Printf("Failed to log edit_alias_domain: %v\n", err)
	}

	SetFlash(c, "message", "Alias Domain updated successfully")
	return c.JSON(http.StatusOK, map[string]interface{}{"success": true})
}
