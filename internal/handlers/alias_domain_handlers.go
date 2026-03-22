package handlers

import (
	"net/http"
	"net/url"
	"time"

	"go-postfixadmin/internal/middleware"
	"go-postfixadmin/internal/models"
	"go-postfixadmin/internal/repositories"

	"github.com/labstack/echo/v5"
)

// ListAliasDomains lists all alias domains
func (h *Handler) ListAliasDomains(c *echo.Context) error {
	username := middleware.GetUsername(c, middleware.SessionName)
	isSuperAdmin := middleware.GetIsSuperAdmin(c)

	aliasDomains, isSuper, err := repositories.GetAliasDomains(h.DB, username, isSuperAdmin)
	if err != nil {
		return c.Render(http.StatusInternalServerError, "alias_domains/alias_domains.html", map[string]interface{}{
			"Error": "Failed to check permissions: " + err.Error(),
		})
	}

	domains, _, err := repositories.GetActiveDomains(h.DB, username, isSuper)
	if err != nil {
		return c.Render(http.StatusInternalServerError, "alias_domains/alias_domains.html", map[string]interface{}{
			"AliasDomains": aliasDomains,
			"Error":        "Failed to load domains: " + err.Error(),
		})
	}

	return c.Render(http.StatusOK, "alias_domains/alias_domains.html", map[string]interface{}{
		"AliasDomains": aliasDomains,
		"Domains":      domains,
		"IsSuperAdmin": isSuper,
		"SessionUser":  username,
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
	allowedDomains, isSuperAdmin, err := repositories.GetAllowedDomains(h.DB, loggedInUser, middleware.GetIsSuperAdmin(c))
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
			return c.JSON(http.StatusForbidden, map[string]interface{}{"success": false, "error": "Access denied to these domains"})
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

	if _, err := repositories.GetAliasDomainByName(h.DB, aliasDomain); err == nil {
		return c.JSON(http.StatusConflict, map[string]interface{}{"success": false, "error": "Alias Domain already exists"})
	}
	if _, err := repositories.GetDomainByName(h.DB, targetDomain); err != nil {
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
	if err := repositories.CreateAliasDomain(h.DB, newAliasDomain, loggedInUser, c.RealIP()); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": "Failed to create alias domain"})
	}

	SetFlash(c, "message", "Alias Domain created successfully")
	return c.JSON(http.StatusOK, map[string]interface{}{"success": true})
}

// DeleteAliasDomain handles alias domain deletion
func (h *Handler) DeleteAliasDomain(c *echo.Context) error {
	aliasDomainName, _ := url.PathUnescape(c.Param("alias_domain"))
	if decoded, err := url.QueryUnescape(aliasDomainName); err == nil {
		aliasDomainName = decoded
	}

	if aliasDomainName == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Alias Domain required"})
	}

	aliasDomain, err := repositories.GetAliasDomainByName(h.DB, aliasDomainName)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{"error": "Alias Domain not found"})
	}

	loggedInUser := middleware.GetUsername(c, middleware.SessionName)
	allowedDomains, isSuperAdmin, err := repositories.GetAllowedDomains(h.DB, loggedInUser, middleware.GetIsSuperAdmin(c))
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

	if err := repositories.DeleteAliasDomain(h.DB, aliasDomainName, loggedInUser, c.RealIP()); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "Failed to delete alias domain"})
	}

	SetFlash(c, "message", "Alias Domain deleted successfully")
	return c.JSON(http.StatusOK, map[string]interface{}{"success": true})
}

// GetAliasDomainAPI fetches alias domain details for the edit modal via JSON
func (h *Handler) GetAliasDomainAPI(c *echo.Context) error {
	aliasDomainName, _ := url.PathUnescape(c.Param("alias_domain"))
	if decoded, err := url.QueryUnescape(aliasDomainName); err == nil {
		aliasDomainName = decoded
	}

	aliasDomain, err := repositories.GetAliasDomainByName(h.DB, aliasDomainName)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{"error": "Alias Domain not found"})
	}

	loggedInUser := middleware.GetUsername(c, middleware.SessionName)
	allowedDomains, isSuperAdmin, err := repositories.GetAllowedDomains(h.DB, loggedInUser, middleware.GetIsSuperAdmin(c))
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

	domains, _, _ := repositories.GetActiveDomains(h.DB, loggedInUser, isSuperAdmin)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"alias_domain": aliasDomain,
		"domains":      domains,
	})
}

// EditAliasDomainAPI processes the alias domain update via JSON
func (h *Handler) EditAliasDomainAPI(c *echo.Context) error {
	aliasDomainName, _ := url.PathUnescape(c.Param("alias_domain"))
	if decoded, err := url.QueryUnescape(aliasDomainName); err == nil {
		aliasDomainName = decoded
	}

	aliasDomain, err := repositories.GetAliasDomainByName(h.DB, aliasDomainName)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{"success": false, "error": "Alias Domain not found"})
	}

	loggedInUser := middleware.GetUsername(c, middleware.SessionName)
	allowedDomains, isSuperAdmin, err := repositories.GetAllowedDomains(h.DB, loggedInUser, middleware.GetIsSuperAdmin(c))
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

	if targetDomain == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": "Target Domain is required"})
	}
	if aliasDomainName == targetDomain {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": "O domínio de origem e destino não podem ser iguais"})
	}

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
		if _, err := repositories.GetDomainByName(h.DB, targetDomain); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": "Target Domain does not exist"})
		}
	}

	aliasDomain.TargetDomain = targetDomain
	aliasDomain.Active = active

	if err := repositories.SaveAliasDomain(h.DB, aliasDomain, loggedInUser, c.RealIP()); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": "Failed to update alias domain"})
	}

	SetFlash(c, "message", "Alias Domain updated successfully")
	return c.JSON(http.StatusOK, map[string]interface{}{"success": true})
}
