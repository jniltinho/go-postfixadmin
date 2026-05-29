package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go-postfixadmin/internal/api/dto"
	"go-postfixadmin/internal/middleware"
	"go-postfixadmin/internal/models"
	"go-postfixadmin/internal/repositories"

	"github.com/labstack/echo/v5"
)

// ListAliases lists all aliases, optionally filtered by domain
func (h *Handler) ListAliases(c *echo.Context) error {
	domainFilter := c.QueryParam("domain")
	username := middleware.GetUsername(c, middleware.SessionName)
	isSuperAdmin := middleware.GetIsSuperAdmin(c)

	aliases, isSuper, err := repositories.GetAllAliases(h.DB, username, isSuperAdmin, domainFilter)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "access denied to this domain" {
			status = http.StatusForbidden
		}
		return c.Render(status, "aliases/aliases.html", map[string]interface{}{
			"Error": err.Error(),
		})
	}

	domains, _, _ := repositories.GetActiveDomains(h.DB, username, isSuper)

	domainLimitReached := false
	if h.DB != nil && domainFilter != "" {
		if domainRecord, err := repositories.GetDomainByName(h.DB, domainFilter); err == nil {
			if domainRecord.Aliases > 0 {
				if count, err := repositories.CountDomainAliases(h.DB, domainFilter); err == nil {
					domainLimitReached = count >= int64(domainRecord.Aliases)
				}
			}
		}
	}

	return c.Render(http.StatusOK, "aliases/aliases.html", map[string]interface{}{
		"Aliases":            aliases,
		"Domains":            domains,
		"DomainFilter":       domainFilter,
		"IsSuperAdmin":       isSuper,
		"SessionUser":        username,
		"DomainLimitReached": domainLimitReached,
		"Message":            GetFlash(c, "message"),
		"Error":              GetFlash(c, "error"),
	})
}

// AddAliasAPI processes the addition of a new alias via JSON API
func (h *Handler) AddAliasAPI(c *echo.Context) error {
	localPart := strings.ToLower(strings.TrimSpace(c.FormValue("local_part")))
	domain := strings.TrimSpace(c.FormValue("domain"))
	gotoRaw := c.FormValue("goto")
	active := c.FormValue("active") == "true"

	loggedInUser := middleware.GetUsername(c, middleware.SessionName)
	allowedDomains, isSuperAdmin, err := repositories.GetAllowedDomains(h.DB, loggedInUser, middleware.GetIsSuperAdmin(c))
	if err != nil {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"success": false, "error": "Permission check failed"})
	}

	if !isSuperAdmin {
		allowed := false
		for _, d := range allowedDomains {
			if d == domain {
				allowed = true
				break
			}
		}
		if !allowed {
			return c.JSON(http.StatusForbidden, map[string]interface{}{"success": false, "error": "Access denied to this domain"})
		}
	}

	if localPart == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": "Alias local part is required"})
	}
	if domain == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": "Domain is required"})
	}
	if gotoRaw == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": "To (Recipients) is required"})
	}

	address := fmt.Sprintf("%s@%s", localPart, domain)

	var recipients []string
	for _, line := range strings.Split(gotoRaw, "\n") {
		line = strings.TrimSpace(strings.Trim(line, ","))
		if line != "" {
			recipients = append(recipients, line)
		}
	}
	if len(recipients) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": "At least one valid recipient is required"})
	}

	if _, err := repositories.GetAliasByAddress(h.DB, address); err == nil {
		return c.JSON(http.StatusConflict, map[string]interface{}{"success": false, "error": "Alias already exists"})
	}
	if _, err := repositories.GetMailboxByUsername(h.DB, address); err == nil {
		return c.JSON(http.StatusConflict, map[string]interface{}{"success": false, "error": "A mailbox already exists with this address"})
	}

	domainRecord, err := repositories.GetDomainByName(h.DB, domain)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": T(c, "Domain not found")})
	}
	if domainRecord.Aliases > 0 {
		count, err := repositories.CountDomainAliases(h.DB, domain)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": T(c, "Failed to check alias limit")})
		}
		if count >= int64(domainRecord.Aliases) {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": TData(c, "Alias_LimitReached", map[string]any{
				"Count": fmt.Sprintf("%d", count),
				"Limit": fmt.Sprintf("%d", domainRecord.Aliases),
			})})
		}
	}

	now := time.Now()
	newAlias := models.Alias{
		Address:  address,
		Goto:     strings.Join(recipients, ","),
		Domain:   domain,
		Created:  now,
		Modified: now,
		Active:   active,
	}
	if err := repositories.CreateAlias(h.DB, newAlias, loggedInUser, c.RealIP()); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": "Failed to create alias"})
	}

	SetFlash(c, "message", "Alias created successfully")
	return c.JSON(http.StatusOK, map[string]interface{}{"success": true})
}

// GetAliasAPI fetches alias details via JSON
func (h *Handler) GetAliasAPI(c *echo.Context) error {
	address, _ := url.PathUnescape(c.Param("address"))

	alias, err := repositories.GetAliasByAddress(h.DB, address)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{"error": "Alias not found"})
	}

	loggedInUser := middleware.GetUsername(c, middleware.SessionName)
	allowedDomains, isSuperAdmin, err := repositories.GetAllowedDomains(h.DB, loggedInUser, middleware.GetIsSuperAdmin(c))
	if err != nil {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"error": "Permission check failed"})
	}
	if !isSuperAdmin {
		allowed := false
		for _, d := range allowedDomains {
			if d == alias.Domain {
				allowed = true
				break
			}
		}
		if !allowed {
			return c.JSON(http.StatusForbidden, map[string]interface{}{"error": "Access denied"})
		}
	}

	alias.Goto = strings.ReplaceAll(alias.Goto, ",", "\n")
	return c.JSON(http.StatusOK, map[string]interface{}{"alias": alias})
}

// EditAliasAPI processes the alias update via JSON API
func (h *Handler) EditAliasAPI(c *echo.Context) error {
	address, _ := url.PathUnescape(c.Param("address"))

	alias, err := repositories.GetAliasByAddress(h.DB, address)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{"success": false, "error": "Alias not found"})
	}

	loggedInUser := middleware.GetUsername(c, middleware.SessionName)
	allowedDomains, isSuperAdmin, err := repositories.GetAllowedDomains(h.DB, loggedInUser, middleware.GetIsSuperAdmin(c))
	if err != nil {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"success": false, "error": "Permission check failed"})
	}
	if !isSuperAdmin {
		allowed := false
		for _, d := range allowedDomains {
			if d == alias.Domain {
				allowed = true
				break
			}
		}
		if !allowed {
			return c.JSON(http.StatusForbidden, map[string]interface{}{"success": false, "error": "Access denied"})
		}
	}

	gotoRaw := c.FormValue("goto")
	if gotoRaw == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": "To (Recipients) is required"})
	}

	var recipients []string
	for _, line := range strings.Split(gotoRaw, "\n") {
		line = strings.TrimSpace(strings.Trim(line, ","))
		if line != "" {
			recipients = append(recipients, line)
		}
	}
	if len(recipients) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": "At least one valid recipient is required"})
	}

	alias.Goto = strings.Join(recipients, ",")
	alias.Active = c.FormValue("active") == "true"

	if err := repositories.SaveAlias(h.DB, alias, loggedInUser, c.RealIP()); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": "Failed to update alias"})
	}

	SetFlash(c, "message", "Alias updated successfully")
	return c.JSON(http.StatusOK, map[string]interface{}{"success": true})
}

// DeleteAlias handles alias deletion
func (h *Handler) DeleteAlias(c *echo.Context) error {
	address, _ := url.PathUnescape(c.Param("address"))
	if decoded, err := url.QueryUnescape(address); err == nil {
		address = decoded
	}

	if address == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Address required"})
	}

	loggedInUser := middleware.GetUsername(c, middleware.SessionName)
	allowedDomains, isSuperAdmin, err := repositories.GetAllowedDomains(h.DB, loggedInUser, middleware.GetIsSuperAdmin(c))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "Permission check failed"})
	}

	alias, err := repositories.GetAliasByAddress(h.DB, address)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{"error": "Alias not found"})
	}

	if !isSuperAdmin {
		allowed := false
		for _, d := range allowedDomains {
			if d == alias.Domain {
				allowed = true
				break
			}
		}
		if !allowed {
			return c.JSON(http.StatusForbidden, map[string]interface{}{"error": "Access denied"})
		}
	}

	var count int64
	if _, err := repositories.GetMailboxByUsername(h.DB, address); err == nil {
		count = 1
	}
	if count > 0 {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"error": "Cannot delete a mailbox alias via this interface. Delete the mailbox instead."})
	}

	if err := repositories.DeleteAlias(h.DB, address, loggedInUser, c.RealIP()); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "Failed to delete alias: " + err.Error()})
	}

	SetFlash(c, "message", "Alias deleted successfully")
	return c.JSON(http.StatusOK, map[string]interface{}{"success": true})
}

// =====================================================
// API v1 Alias Handlers (PR 07+)
// =====================================================

// ListAliasesV1 godoc
// @Summary      List aliases
// @Description  Returns all aliases accessible to the authenticated admin. Superadmins see all; domain admins see only their domains. Use ?domain= to filter by domain.
// @Tags         Aliases
// @Produce      json
// @Param        domain  query     string  false  "Filter by domain"
// @Success      200     {array}   dto.AliasResponse
// @Failure      401     {object}  dto.APIResponse
// @Failure      403     {object}  dto.APIResponse
// @Failure      500     {object}  dto.APIResponse
// @Router       /aliases [get]
// @Security     BearerAuth
func (h *Handler) ListAliasesV1(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil {
		return dto.Unauthorized(c, "not authenticated")
	}

	domainFilter := c.QueryParam("domain")

	aliases, _, err := repositories.GetAllAliases(h.DB, claims.Username, claims.Superadmin, domainFilter)
	if err != nil {
		if err.Error() == "access denied to this domain" {
			return dto.Forbidden(c, "access denied to this domain")
		}
		return dto.InternalError(c, "failed to fetch aliases")
	}

	var response []dto.AliasResponse
	for _, a := range aliases {
		response = append(response, dto.AliasResponse{
			Address:  a.Address,
			Goto:     a.Goto,
			Domain:   a.Domain,
			Active:   a.Active,
			Created:  a.Created,
			Modified: a.Modified,
		})
	}

	return dto.WriteSuccess(c, response)
}

// CreateAliasV1 godoc
// @Summary      Create alias
// @Description  Creates a new email alias. goto accepts multiple recipients separated by newlines or commas. The caller must have admin access to the target domain.
// @Tags         Aliases
// @Accept       json
// @Produce      json
// @Param        request  body      dto.CreateAliasRequest  true  "Alias data"
// @Success      201      {object}  dto.APIResponse
// @Failure      400      {object}  dto.APIResponse
// @Failure      401      {object}  dto.APIResponse
// @Failure      403      {object}  dto.APIResponse
// @Failure      409      {object}  dto.APIResponse  "Alias already exists"
// @Failure      500      {object}  dto.APIResponse
// @Router       /aliases [post]
// @Security     BearerAuth
func (h *Handler) CreateAliasV1(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil {
		return dto.Unauthorized(c, "not authenticated")
	}

	var req dto.CreateAliasRequest
	if err := c.Bind(&req); err != nil {
		return dto.BadRequest(c, "invalid request body")
	}

	localPart := strings.ToLower(strings.TrimSpace(req.LocalPart))
	domain := strings.TrimSpace(req.Domain)
	gotoRaw := req.Goto

	if localPart == "" || domain == "" || gotoRaw == "" {
		return dto.ValidationError(c, "local_part, domain and goto are required")
	}

	// Permission scoping
	allowedDomains, _, err := repositories.GetAllowedDomains(h.DB, claims.Username, claims.Superadmin)
	if err != nil {
		return dto.InternalError(c, "permission check failed")
	}
	if !claims.Superadmin {
		allowed := false
		for _, d := range allowedDomains {
			if d == domain {
				allowed = true
				break
			}
		}
		if !allowed {
			return dto.Forbidden(c, "access denied to this domain")
		}
	}

	address := fmt.Sprintf("%s@%s", localPart, domain)

	var recipients []string
	for _, line := range strings.Split(gotoRaw, "\n") {
		line = strings.TrimSpace(strings.Trim(line, ","))
		if line != "" {
			recipients = append(recipients, line)
		}
	}
	if len(recipients) == 0 {
		return dto.ValidationError(c, "at least one valid recipient is required")
	}

	if _, err := repositories.GetAliasByAddress(h.DB, address); err == nil {
		return dto.WriteError(c, dto.ErrCodeConflict, "alias already exists")
	}
	if _, err := repositories.GetMailboxByUsername(h.DB, address); err == nil {
		return dto.WriteError(c, dto.ErrCodeConflict, "a mailbox already exists with this address")
	}

	domainRecord, err := repositories.GetDomainByName(h.DB, domain)
	if err != nil {
		return dto.ValidationError(c, "domain not found")
	}
	if domainRecord.Aliases > 0 {
		count, err := repositories.CountDomainAliases(h.DB, domain)
		if err != nil {
			return dto.InternalError(c, "failed to check alias limit")
		}
		if count >= int64(domainRecord.Aliases) {
			return dto.ValidationError(c, "alias limit reached for this domain")
		}
	}

	now := time.Now()
	newAlias := models.Alias{
		Address:  address,
		Goto:     strings.Join(recipients, ","),
		Domain:   domain,
		Created:  now,
		Modified: now,
		Active:   req.Active,
	}

	if err := repositories.CreateAlias(h.DB, newAlias, claims.Username, c.RealIP()); err != nil {
		return dto.InternalError(c, "failed to create alias")
	}

	return dto.WriteSuccessWithStatus(c, http.StatusCreated, map[string]string{"address": address})
}

// GetAliasV1 godoc
// @Summary      Get alias
// @Description  Returns details for a single alias. The goto field is returned with recipients separated by newlines. The caller must have admin access to the alias domain.
// @Tags         Aliases
// @Produce      json
// @Param        address  path      string  true  "Alias address (e.g. info@example.com)"
// @Success      200      {object}  dto.AliasResponse
// @Failure      401      {object}  dto.APIResponse
// @Failure      403      {object}  dto.APIResponse
// @Failure      404      {object}  dto.APIResponse
// @Failure      500      {object}  dto.APIResponse
// @Router       /aliases/{address} [get]
// @Security     BearerAuth
func (h *Handler) GetAliasV1(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil {
		return dto.Unauthorized(c, "not authenticated")
	}

	address, _ := url.PathUnescape(c.Param("address"))

	alias, err := repositories.GetAliasByAddress(h.DB, address)
	if err != nil {
		return dto.NotFound(c, "alias not found")
	}

	// Scoping
	allowedDomains, _, err := repositories.GetAllowedDomains(h.DB, claims.Username, claims.Superadmin)
	if err != nil {
		return dto.InternalError(c, "permission check failed")
	}
	if !claims.Superadmin {
		allowed := false
		for _, d := range allowedDomains {
			if d == alias.Domain {
				allowed = true
				break
			}
		}
		if !allowed {
			return dto.Forbidden(c, "access denied")
		}
	}

	alias.Goto = strings.ReplaceAll(alias.Goto, ",", "\n")

	return dto.WriteSuccess(c, alias)
}

// UpdateAliasV1 godoc
// @Summary      Update alias
// @Description  Updates the goto destinations and/or active state of an existing alias. All fields are optional.
// @Tags         Aliases
// @Accept       json
// @Produce      json
// @Param        address  path      string                  true  "Alias address (e.g. info@example.com)"
// @Param        request  body      dto.UpdateAliasRequest  true  "Fields to update"
// @Success      200      {object}  dto.APIResponse
// @Failure      400      {object}  dto.APIResponse
// @Failure      401      {object}  dto.APIResponse
// @Failure      403      {object}  dto.APIResponse
// @Failure      404      {object}  dto.APIResponse
// @Failure      500      {object}  dto.APIResponse
// @Router       /aliases/{address} [put]
// @Security     BearerAuth
func (h *Handler) UpdateAliasV1(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil {
		return dto.Unauthorized(c, "not authenticated")
	}

	address, _ := url.PathUnescape(c.Param("address"))

	alias, err := repositories.GetAliasByAddress(h.DB, address)
	if err != nil {
		return dto.NotFound(c, "alias not found")
	}

	// Scoping
	allowedDomains, _, err := repositories.GetAllowedDomains(h.DB, claims.Username, claims.Superadmin)
	if err != nil {
		return dto.InternalError(c, "permission check failed")
	}
	if !claims.Superadmin {
		allowed := false
		for _, d := range allowedDomains {
			if d == alias.Domain {
				allowed = true
				break
			}
		}
		if !allowed {
			return dto.Forbidden(c, "access denied")
		}
	}

	var req dto.UpdateAliasRequest
	if err := c.Bind(&req); err != nil {
		return dto.BadRequest(c, "invalid request body")
	}

	if req.Goto != nil {
		gotoRaw := *req.Goto
		var recipients []string
		for _, line := range strings.Split(gotoRaw, "\n") {
			line = strings.TrimSpace(strings.Trim(line, ","))
			if line != "" {
				recipients = append(recipients, line)
			}
		}
		if len(recipients) == 0 {
			return dto.ValidationError(c, "at least one valid recipient is required")
		}
		alias.Goto = strings.Join(recipients, ",")
	}

	if req.Active != nil {
		alias.Active = *req.Active
	}

	if err := repositories.SaveAlias(h.DB, alias, claims.Username, c.RealIP()); err != nil {
		return dto.InternalError(c, "failed to update alias")
	}

	return dto.WriteSuccess(c, map[string]bool{"updated": true})
}

// DeleteAliasV1 godoc
// @Summary      Delete alias
// @Description  Permanently deletes an alias. The caller must have admin access to the alias domain.
// @Tags         Aliases
// @Produce      json
// @Param        address  path      string  true  "Alias address (e.g. info@example.com)"
// @Success      200      {object}  dto.APIResponse
// @Failure      401      {object}  dto.APIResponse
// @Failure      403      {object}  dto.APIResponse
// @Failure      404      {object}  dto.APIResponse
// @Failure      500      {object}  dto.APIResponse
// @Router       /aliases/{address} [delete]
// @Security     BearerAuth
func (h *Handler) DeleteAliasV1(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil {
		return dto.Unauthorized(c, "not authenticated")
	}

	address, _ := url.PathUnescape(c.Param("address"))

	alias, err := repositories.GetAliasByAddress(h.DB, address)
	if err != nil {
		return dto.NotFound(c, "alias not found")
	}

	// Scoping
	allowedDomains, _, err := repositories.GetAllowedDomains(h.DB, claims.Username, claims.Superadmin)
	if err != nil {
		return dto.InternalError(c, "permission check failed")
	}
	if !claims.Superadmin {
		allowed := false
		for _, d := range allowedDomains {
			if d == alias.Domain {
				allowed = true
				break
			}
		}
		if !allowed {
			return dto.Forbidden(c, "access denied")
		}
	}

	// Prevent deleting mailbox aliases
	if _, err := repositories.GetMailboxByUsername(h.DB, address); err == nil {
		return dto.Forbidden(c, "cannot delete a mailbox alias. Delete the mailbox instead.")
	}

	if err := repositories.DeleteAlias(h.DB, address, claims.Username, c.RealIP()); err != nil {
		return dto.InternalError(c, "failed to delete alias")
	}

	return dto.WriteSuccess(c, map[string]string{"deleted": address})
}
