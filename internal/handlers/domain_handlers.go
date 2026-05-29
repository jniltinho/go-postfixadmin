package handlers

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go-postfixadmin/internal/api/dto"
	"go-postfixadmin/internal/middleware"
	"go-postfixadmin/internal/models"
	"go-postfixadmin/internal/repositories"

	"github.com/labstack/echo/v5"
)

// DomainDisplay representa um domínio com contadores de aliases e mailboxes
type DomainDisplay struct {
	models.Domain
	AliasCount   int64
	MailboxCount int64
}

// ListDomains lista todos os domínios com contadores de aliases e mailboxes
func (h *Handler) ListDomains(c *echo.Context) error {
	username := middleware.GetUsername(c, middleware.SessionName)
	isSuperAdmin := middleware.GetIsSuperAdmin(c)

	domains, err := repositories.GetAllDomains(h.DB, username, isSuperAdmin)
	if err != nil {
		return c.Render(http.StatusInternalServerError, "domains/domains.html", map[string]interface{}{
			"Error": "Failed to check permissions: " + err.Error(),
		})
	}

	var displayDomains []DomainDisplay
	for _, d := range domains {
		aliasCount, _ := repositories.CountDomainAliases(h.DB, d.Domain)
		mailboxCount, _ := repositories.CountDomainMailboxes(h.DB, d.Domain)
		displayDomains = append(displayDomains, DomainDisplay{
			Domain:       d,
			AliasCount:   aliasCount,
			MailboxCount: mailboxCount,
		})
	}

	transports, _ := repositories.GetAllTransports(h.DB)

	return c.Render(http.StatusOK, "domains/domains.html", map[string]interface{}{
		"Domains":      displayDomains,
		"Transports":   transports,
		"IsSuperAdmin": isSuperAdmin,
		"SessionUser":  username,
		"Message":      GetFlash(c, "message"),
		"Error":        GetFlash(c, "error"),
	})
}

// AddDomainAPI processa a criação de um novo domínio via JSON API
func (h *Handler) AddDomainAPI(c *echo.Context) error {
	username := middleware.GetUsername(c, middleware.SessionName)
	isSuperAdmin := middleware.GetIsSuperAdmin(c)
	if !isSuperAdmin {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"error": "Access denied: Only Superadmins can create domains"})
	}

	domainName := strings.TrimSpace(c.FormValue("domain"))
	description := c.FormValue("description")
	active := c.FormValue("active") == "true"
	backupMX := c.FormValue("backupmx") == "true"

	aliases := 10
	if val := c.FormValue("aliases"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			aliases = parsed
		}
	}
	mailboxes := 10
	if val := c.FormValue("mailboxes"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			mailboxes = parsed
		}
	}
	quota := int64(2048)
	if val := c.FormValue("quota"); val != "" {
		if parsed, err := strconv.ParseInt(val, 10, 64); err == nil {
			quota = parsed
		}
	}
	var passwordExpiry *int
	if val := c.FormValue("password_expiry"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			passwordExpiry = &parsed
		}
	}

	transport := "virtual"
	if middleware.GetIsSuperAdmin(c) {
		if val := c.FormValue("transport"); val != "" {
			transport = val
		}
	}

	if domainName == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Domain name is required"})
	}
	domainRegex := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{0,61}[a-zA-Z0-9]?(\.[a-zA-Z0-9][a-zA-Z0-9-]{0,61}[a-zA-Z0-9]?)+$`)
	if !domainRegex.MatchString(domainName) {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Invalid domain format. Please enter a valid domain name (e.g., example.com)"})
	}

	if _, err := repositories.GetDomainByName(h.DB, domainName); err == nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Domain already exists"})
	}

	now := time.Now()
	newDomain := models.Domain{
		Domain:         domainName,
		Description:    description,
		Aliases:        aliases,
		Mailboxes:      mailboxes,
		Quota:          quota,
		BackupMX:       backupMX,
		Transport:      transport,
		Created:        now,
		Modified:       now,
		Active:         active,
		PasswordExpiry: passwordExpiry,
	}
	if err := repositories.CreateDomain(h.DB, newDomain, username, c.RealIP()); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "Failed to create domain: " + err.Error()})
	}

	SetFlash(c, "message", "Domain created successfully")
	return c.JSON(http.StatusOK, map[string]interface{}{"success": true})
}

// GetDomainAPI retorna os dados de um domínio para popular o modal de edição
func (h *Handler) GetDomainAPI(c *echo.Context) error {
	if !middleware.GetIsSuperAdmin(c) {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"error": "Access denied"})
	}

	domain, err := repositories.GetDomainByName(h.DB, c.Param("domain"))
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{"error": "Domain not found"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"success": true, "domain": domain})
}

// EditDomainAPI processa a edição de um domínio existente via JSON API
func (h *Handler) EditDomainAPI(c *echo.Context) error {
	username := middleware.GetUsername(c, middleware.SessionName)
	if !middleware.GetIsSuperAdmin(c) {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"error": "Access denied"})
	}

	domain, err := repositories.GetDomainByName(h.DB, c.Param("domain"))
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{"error": "Domain not found"})
	}

	description := c.FormValue("description")
	active := c.FormValue("active") == "true"
	backupMX := c.FormValue("backupmx") == "true"

	aliases := 10
	if val := c.FormValue("aliases"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			aliases = parsed
		}
	}
	mailboxes := 10
	if val := c.FormValue("mailboxes"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			mailboxes = parsed
		}
	}
	quota := domain.Quota
	if val := c.FormValue("quota"); val != "" {
		if parsed, err := strconv.ParseInt(val, 10, 64); err == nil {
			quota = parsed
		}
	}
	var passwordExpiry *int
	if val := c.FormValue("password_expiry"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			passwordExpiry = &parsed
		}
	}

	transport := domain.Transport // preserve existing transport
	if middleware.GetIsSuperAdmin(c) {
		if val := c.FormValue("transport"); val != "" {
			transport = val
		}
	}

	activeChanged := domain.Active != active
	domain.Description = description
	domain.Aliases = aliases
	domain.Mailboxes = mailboxes
	domain.Quota = quota
	domain.BackupMX = backupMX
	domain.Transport = transport
	domain.Active = active
	domain.PasswordExpiry = passwordExpiry

	if err := repositories.UpdateDomain(h.DB, domain, activeChanged, username, c.RealIP()); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "Failed to update domain: " + err.Error()})
	}

	SetFlash(c, "message", "Domain updated successfully")
	return c.JSON(http.StatusOK, map[string]interface{}{"success": true})
}

// DeleteDomain remove um domínio e todos os dados associados (aliases e mailboxes)
func (h *Handler) DeleteDomain(c *echo.Context) error {
	username := middleware.GetUsername(c, middleware.SessionName)
	if !middleware.GetIsSuperAdmin(c) {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"error": "Access denied: Only Superadmins can delete domains"})
	}

	domainName := c.Param("domain")
	if _, err := repositories.GetDomainByName(h.DB, domainName); err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{"success": false, "error": "Domain not found"})
	}

	if err := repositories.DeleteDomain(h.DB, domainName, username, c.RealIP()); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": "Failed to delete domain: " + err.Error()})
	}

	SetFlash(c, "message", "Domain deleted successfully")
	return c.JSON(http.StatusOK, map[string]interface{}{"success": true, "message": "Domain deleted successfully"})
}

// =====================================================
// API v1 Domain Handlers (PR 05+)
// These use JWT claims for authentication and scoping
// =====================================================

// ListDomainsV1 godoc
// @Summary      List domains
// @Description  Returns all domains the authenticated user has access to (based on JWT claims).
// @Tags         Domains
// @Produce      json
// @Success      200 {array} dto.DomainResponse
// @Failure      401 {object} dto.APIResponse
// @Router       /domains [get]
// @Security     BearerAuth
func (h *Handler) ListDomainsV1(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil {
		return dto.Unauthorized(c, "not authenticated")
	}

	// Use the claims for scoping (superadmin + domains list)
	domains, err := repositories.GetAllDomains(h.DB, claims.Username, claims.Superadmin)
	if err != nil {
		return dto.InternalError(c, "failed to fetch domains")
	}

	var response []dto.DomainResponse
	for _, d := range domains {
		aliasCount, _ := repositories.CountDomainAliases(h.DB, d.Domain)
		mailboxCount, _ := repositories.CountDomainMailboxes(h.DB, d.Domain)

		response = append(response, dto.DomainResponse{
			Domain:         d.Domain,
			Description:    d.Description,
			Aliases:        d.Aliases,
			Mailboxes:      d.Mailboxes,
			MaxQuota:       d.MaxQuota,
			Quota:          d.Quota,
			Transport:      d.Transport,
			BackupMX:       d.BackupMX,
			Active:         d.Active,
			PasswordExpiry: d.PasswordExpiry,
			Created:        d.Created,
			Modified:       d.Modified,
			AliasCount:     aliasCount,
			MailboxCount:   mailboxCount,
		})
	}

	return dto.WriteSuccess(c, response)
}

// CreateDomainV1 godoc
// @Summary      Create domain
// @Description  Creates a new domain (only superadmins can create).
// @Tags         Domains
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateDomainRequest true "Domain data"
// @Success      201 {object} map[string]string
// @Failure      400 {object} dto.APIResponse
// @Failure      401 {object} dto.APIResponse
// @Failure      403 {object} dto.APIResponse
// @Router       /domains [post]
// @Security     BearerAuth
func (h *Handler) CreateDomainV1(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil {
		return dto.Unauthorized(c, "not authenticated")
	}
	if !claims.Superadmin {
		return dto.Forbidden(c, "only superadmins can create domains")
	}

	var req dto.CreateDomainRequest
	if err := c.Bind(&req); err != nil {
		return dto.BadRequest(c, "invalid request body")
	}

	domainName := strings.TrimSpace(req.Domain)
	if domainName == "" {
		return dto.ValidationError(c, "domain name is required")
	}

	domainRegex := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{0,61}[a-zA-Z0-9]?(\.[a-zA-Z0-9][a-zA-Z0-9-]{0,61}[a-zA-Z0-9]?)+$`)
	if !domainRegex.MatchString(domainName) {
		return dto.ValidationError(c, "invalid domain format")
	}

	if _, err := repositories.GetDomainByName(h.DB, domainName); err == nil {
		return dto.WriteError(c, dto.ErrCodeConflict, "domain already exists")
	}

	now := time.Now()
	newDomain := models.Domain{
		Domain:         domainName,
		Description:    req.Description,
		Aliases:        req.Aliases,
		Mailboxes:      req.Mailboxes,
		Quota:          req.Quota,
		Transport:      req.Transport,
		BackupMX:       req.BackupMX,
		Active:         req.Active,
		PasswordExpiry: req.PasswordExpiry,
		Created:        now,
		Modified:       now,
	}

	if newDomain.Transport == "" {
		newDomain.Transport = "virtual"
	}
	if newDomain.Aliases == 0 {
		newDomain.Aliases = 10
	}
	if newDomain.Mailboxes == 0 {
		newDomain.Mailboxes = 10
	}

	if err := repositories.CreateDomain(h.DB, newDomain, claims.Username, c.RealIP()); err != nil {
		return dto.InternalError(c, "failed to create domain")
	}

	return dto.WriteSuccessWithStatus(c, http.StatusCreated, map[string]string{"domain": domainName})
}

// GetDomainV1 godoc
// @Summary      Get domain
// @Description  Returns details of a specific domain (respects user scoping).
// @Tags         Domains
// @Produce      json
// @Param        domain path string true "Domain name"
// @Success      200 {object} dto.DomainResponse
// @Failure      401 {object} dto.APIResponse
// @Failure      403 {object} dto.APIResponse
// @Failure      404 {object} dto.APIResponse
// @Router       /domains/{domain} [get]
// @Security     BearerAuth
func (h *Handler) GetDomainV1(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil {
		return dto.Unauthorized(c, "not authenticated")
	}

	domainName := c.Param("domain")

	// Check scoping
	allowed, isSuper, err := repositories.GetAllowedDomains(h.DB, claims.Username, claims.Superadmin)
	if err != nil {
		return dto.InternalError(c, "failed to check permissions")
	}
	if !isSuper && !containsDomain(allowed, domainName) {
		return dto.Forbidden(c, "access denied to this domain")
	}

	domain, err := repositories.GetDomainByName(h.DB, domainName)
	if err != nil {
		return dto.NotFound(c, "domain not found")
	}

	aliasCount, _ := repositories.CountDomainAliases(h.DB, domain.Domain)
	mailboxCount, _ := repositories.CountDomainMailboxes(h.DB, domain.Domain)

	resp := dto.DomainResponse{
		Domain:         domain.Domain,
		Description:    domain.Description,
		Aliases:        domain.Aliases,
		Mailboxes:      domain.Mailboxes,
		MaxQuota:       domain.MaxQuota,
		Quota:          domain.Quota,
		Transport:      domain.Transport,
		BackupMX:       domain.BackupMX,
		Active:         domain.Active,
		PasswordExpiry: domain.PasswordExpiry,
		Created:        domain.Created,
		Modified:       domain.Modified,
		AliasCount:     aliasCount,
		MailboxCount:   mailboxCount,
	}

	return dto.WriteSuccess(c, resp)
}

// UpdateDomainV1 handles PUT /api/v1/domains/:domain
func (h *Handler) UpdateDomainV1(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil {
		return dto.Unauthorized(c, "not authenticated")
	}
	if !claims.Superadmin {
		return dto.Forbidden(c, "only superadmins can update domains")
	}

	domainName := c.Param("domain")
	domain, err := repositories.GetDomainByName(h.DB, domainName)
	if err != nil {
		return dto.NotFound(c, "domain not found")
	}

	var req dto.UpdateDomainRequest
	if err := c.Bind(&req); err != nil {
		return dto.BadRequest(c, "invalid request body")
	}

	// Apply updates only for provided fields
	if req.Description != nil {
		domain.Description = *req.Description
	}
	if req.Aliases != nil {
		domain.Aliases = *req.Aliases
	}
	if req.Mailboxes != nil {
		domain.Mailboxes = *req.Mailboxes
	}
	if req.Quota != nil {
		domain.Quota = *req.Quota
	}
	if req.Transport != nil {
		domain.Transport = *req.Transport
	}
	if req.BackupMX != nil {
		domain.BackupMX = *req.BackupMX
	}
	activeChanged := false
	if req.Active != nil {
		activeChanged = domain.Active != *req.Active
		domain.Active = *req.Active
	}
	if req.PasswordExpiry != nil {
		domain.PasswordExpiry = req.PasswordExpiry
	}

	domain.Modified = time.Now()

	if err := repositories.UpdateDomain(h.DB, domain, activeChanged, claims.Username, c.RealIP()); err != nil {
		return dto.InternalError(c, "failed to update domain")
	}

	return dto.WriteSuccess(c, map[string]bool{"updated": true})
}

// DeleteDomainV1 godoc
// @Summary      Delete domain
// @Description  Deletes a domain and all related data (only superadmins).
// @Tags         Domains
// @Produce      json
// @Param        domain path string true "Domain name"
// @Success      200 {object} map[string]string
// @Failure      401 {object} dto.APIResponse
// @Failure      403 {object} dto.APIResponse
// @Failure      404 {object} dto.APIResponse
// @Router       /domains/{domain} [delete]
// @Security     BearerAuth
func (h *Handler) DeleteDomainV1(c *echo.Context) error {
	claims := middleware.GetJWTClaims(c)
	if claims == nil {
		return dto.Unauthorized(c, "not authenticated")
	}
	if !claims.Superadmin {
		return dto.Forbidden(c, "only superadmins can delete domains")
	}

	domainName := c.Param("domain")
	if _, err := repositories.GetDomainByName(h.DB, domainName); err != nil {
		return dto.NotFound(c, "domain not found")
	}

	if err := repositories.DeleteDomain(h.DB, domainName, claims.Username, c.RealIP()); err != nil {
		return dto.InternalError(c, "failed to delete domain")
	}

	return dto.WriteSuccess(c, map[string]string{"deleted": domainName})
}

// containsDomain is a small helper
func containsDomain(domains []string, target string) bool {
	for _, d := range domains {
		if d == target {
			return true
		}
	}
	return false
}
