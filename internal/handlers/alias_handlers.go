package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

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

	return c.Render(http.StatusOK, "aliases/aliases.html", map[string]interface{}{
		"Aliases":      aliases,
		"Domains":      domains,
		"DomainFilter": domainFilter,
		"IsSuperAdmin": isSuper,
		"SessionUser":  username,
		"Message":      GetFlash(c, "message"),
		"Error":        GetFlash(c, "error"),
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
