package handlers

import (
	"net/http"
	"net/url"
	"time"

	"go-postfixadmin/internal/api/dto"
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

// =====================================================
// API v1 Alias Domain Handlers (PR 08+)
// =====================================================

// ListAliasDomainsV1 returns alias domains visible to the authenticated user
func (h *Handler) ListAliasDomainsV1(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil {
		return dto.Unauthorized(c, "not authenticated")
	}

	aliasDomains, _, err := repositories.GetAliasDomains(h.DB, claims.Username, claims.Superadmin)
	if err != nil {
		return dto.InternalError(c, "failed to fetch alias domains")
	}

	var response []dto.AliasDomainResponse
	for _, ad := range aliasDomains {
		response = append(response, dto.AliasDomainResponse{
			AliasDomain:  ad.AliasDomain,
			TargetDomain: ad.TargetDomain,
			Active:       ad.Active,
			Created:      ad.Created,
			Modified:     ad.Modified,
		})
	}

	return dto.WriteSuccess(c, response)
}

// CreateAliasDomainV1 handles POST /api/v1/alias-domains
func (h *Handler) CreateAliasDomainV1(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil {
		return dto.Unauthorized(c, "not authenticated")
	}

	var req dto.CreateAliasDomainRequest
	if err := c.Bind(&req); err != nil {
		return dto.BadRequest(c, "invalid request body")
	}

	if req.AliasDomain == "" || req.TargetDomain == "" {
		return dto.ValidationError(c, "alias_domain and target_domain are required")
	}
	if req.AliasDomain == req.TargetDomain {
		return dto.ValidationError(c, "source and target domains cannot be the same")
	}

	// Permission check
	allowedDomains, _, err := repositories.GetAllowedDomains(h.DB, claims.Username, claims.Superadmin)
	if err != nil {
		return dto.InternalError(c, "permission check failed")
	}
	if !claims.Superadmin {
		allowed := false
		for _, d := range allowedDomains {
			if d == req.TargetDomain || d == req.AliasDomain {
				allowed = true
				break
			}
		}
		if !allowed {
			return dto.Forbidden(c, "access denied to these domains")
		}
	}

	if _, err := repositories.GetAliasDomainByName(h.DB, req.AliasDomain); err == nil {
		return dto.WriteError(c, dto.ErrCodeConflict, "alias domain already exists")
	}
	if _, err := repositories.GetDomainByName(h.DB, req.TargetDomain); err != nil {
		return dto.ValidationError(c, "target domain does not exist")
	}

	now := time.Now()
	newAD := models.AliasDomain{
		AliasDomain:  req.AliasDomain,
		TargetDomain: req.TargetDomain,
		Created:      now,
		Modified:     now,
		Active:       req.Active,
	}

	if err := repositories.CreateAliasDomain(h.DB, newAD, claims.Username, c.RealIP()); err != nil {
		return dto.InternalError(c, "failed to create alias domain")
	}

	return dto.WriteSuccessWithStatus(c, http.StatusCreated, map[string]string{"alias_domain": req.AliasDomain})
}

// GetAliasDomainV1 handles GET /api/v1/alias-domains/:alias_domain
func (h *Handler) GetAliasDomainV1(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil {
		return dto.Unauthorized(c, "not authenticated")
	}

	name, _ := url.PathUnescape(c.Param("alias_domain"))

	aliasDomain, err := repositories.GetAliasDomainByName(h.DB, name)
	if err != nil {
		return dto.NotFound(c, "alias domain not found")
	}

	// Scoping
	allowedDomains, _, err := repositories.GetAllowedDomains(h.DB, claims.Username, claims.Superadmin)
	if err != nil {
		return dto.InternalError(c, "permission check failed")
	}
	if !claims.Superadmin {
		allowed := false
		for _, d := range allowedDomains {
			if d == aliasDomain.TargetDomain {
				allowed = true
				break
			}
		}
		if !allowed {
			return dto.Forbidden(c, "access denied")
		}
	}

	return dto.WriteSuccess(c, aliasDomain)
}

// UpdateAliasDomainV1 handles PUT /api/v1/alias-domains/:alias_domain
func (h *Handler) UpdateAliasDomainV1(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil {
		return dto.Unauthorized(c, "not authenticated")
	}

	name, _ := url.PathUnescape(c.Param("alias_domain"))

	aliasDomain, err := repositories.GetAliasDomainByName(h.DB, name)
	if err != nil {
		return dto.NotFound(c, "alias domain not found")
	}

	// Scoping
	allowedDomains, _, err := repositories.GetAllowedDomains(h.DB, claims.Username, claims.Superadmin)
	if err != nil {
		return dto.InternalError(c, "permission check failed")
	}
	if !claims.Superadmin {
		allowed := false
		for _, d := range allowedDomains {
			if d == aliasDomain.TargetDomain {
				allowed = true
				break
			}
		}
		if !allowed {
			return dto.Forbidden(c, "access denied")
		}
	}

	var req dto.UpdateAliasDomainRequest
	if err := c.Bind(&req); err != nil {
		return dto.BadRequest(c, "invalid request body")
	}

	if req.TargetDomain != nil {
		if *req.TargetDomain == aliasDomain.AliasDomain {
			return dto.ValidationError(c, "source and target domains cannot be the same")
		}
		if !claims.Superadmin {
			allowed := false
			for _, d := range allowedDomains {
				if d == *req.TargetDomain {
					allowed = true
					break
				}
			}
			if !allowed {
				return dto.Forbidden(c, "access denied to new target domain")
			}
		}
		if _, err := repositories.GetDomainByName(h.DB, *req.TargetDomain); err != nil {
			return dto.ValidationError(c, "target domain does not exist")
		}
		aliasDomain.TargetDomain = *req.TargetDomain
	}

	if req.Active != nil {
		aliasDomain.Active = *req.Active
	}

	if err := repositories.SaveAliasDomain(h.DB, aliasDomain, claims.Username, c.RealIP()); err != nil {
		return dto.InternalError(c, "failed to update alias domain")
	}

	return dto.WriteSuccess(c, map[string]bool{"updated": true})
}

// DeleteAliasDomainV1 handles DELETE /api/v1/alias-domains/:alias_domain
func (h *Handler) DeleteAliasDomainV1(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil {
		return dto.Unauthorized(c, "not authenticated")
	}

	name, _ := url.PathUnescape(c.Param("alias_domain"))

	aliasDomain, err := repositories.GetAliasDomainByName(h.DB, name)
	if err != nil {
		return dto.NotFound(c, "alias domain not found")
	}

	// Scoping
	allowedDomains, _, err := repositories.GetAllowedDomains(h.DB, claims.Username, claims.Superadmin)
	if err != nil {
		return dto.InternalError(c, "permission check failed")
	}
	if !claims.Superadmin {
		allowed := false
		for _, d := range allowedDomains {
			if d == aliasDomain.TargetDomain {
				allowed = true
				break
			}
		}
		if !allowed {
			return dto.Forbidden(c, "access denied")
		}
	}

	if err := repositories.DeleteAliasDomain(h.DB, name, claims.Username, c.RealIP()); err != nil {
		return dto.InternalError(c, "failed to delete alias domain")
	}

	return dto.WriteSuccess(c, map[string]string{"deleted": name})
}
