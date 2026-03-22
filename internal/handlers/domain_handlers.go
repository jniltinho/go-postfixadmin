package handlers

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

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

	return c.Render(http.StatusOK, "domains/domains.html", map[string]interface{}{
		"Domains":      displayDomains,
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

	activeChanged := domain.Active != active
	domain.Description = description
	domain.Aliases = aliases
	domain.Mailboxes = mailboxes
	domain.Quota = quota
	domain.BackupMX = backupMX
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
