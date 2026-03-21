package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go-postfixadmin/internal/middleware"
	"go-postfixadmin/internal/models"
	"go-postfixadmin/internal/utils"

	"github.com/labstack/echo/v5"
)

// ListAliases lists all aliases, optionally filtered by domain
func (h *Handler) ListAliases(c *echo.Context) error {
	var aliases []models.Alias
	domainFilter := c.QueryParam("domain")
	var isSuperAdmin bool

	if h.DB != nil {
		// Start query with specific selection and join to filter by active domains
		// We use Table("alias") to be explicit and Select("alias.*") to avoid GORM mapping issues
		query := h.DB.Table("alias").Select("alias.*").
			Joins("JOIN domain ON alias.domain = domain.domain").
			Where("domain.active = ?", true).
			Order("alias.address ASC")

		// Security: Filter by allowed domains
		username := middleware.GetUsername(c, middleware.SessionName)
		allowedDomains, isSuper, err := utils.GetAllowedDomains(h.DB, username, middleware.GetIsSuperAdmin(c))
		if err != nil {
			return c.Render(http.StatusInternalServerError, "aliases/aliases.html", map[string]interface{}{
				"Error": "Failed to check permissions: " + err.Error(),
			})
		}
		isSuperAdmin = isSuper

		if !isSuperAdmin {
			if len(allowedDomains) == 0 {
				query = query.Where("1 = 0") // No domains allowed
			} else {
				query = query.Where("alias.domain IN ?", allowedDomains)
			}
		}

		// Apply optional domain filter
		if domainFilter != "" {
			// Ensure the requested filter is allowed
			if !isSuperAdmin {
				allowed := false
				for _, d := range allowedDomains {
					if d == domainFilter {
						allowed = true
						break
					}
				}
				if !allowed {
					// User requested a forbidden domain filter
					return c.Render(http.StatusForbidden, "aliases/aliases.html", map[string]interface{}{
						"Error": "Access denied to this domain",
					})
				}
			}
			query = query.Where("alias.domain = ?", domainFilter)
		}

		// Show only pure aliases, not mailbox aliases
		// Assuming mailbox aliases have the same address as a mailbox username
		// A common way to distinguish is if the alias points to itself (address = goto) AND there is a mailbox with that address
		// But in PostfixAdmin, a mailbox usually has an alias where address=mailbox.username and goto=mailbox.username
		// The user request "criar a listagem de alias" usually implies explicit aliases/forwarders.
		// However, standard PostfixAdmin usually shows all.
		// Let's filter out aliases that are actually mailboxes if possible, or just show all for now.
		// PostfixAdmin UI typically separates "Virtual List" (All) from specific types.
		// Let's stick to showing all for now, but maybe add a visual indicator if it's a mailbox alias?
		// Actually, let's look at domain_handlers.go - it filters aliases in count:
		// Where("address NOT IN (?)", h.DB.Table("mailbox").Select("username"))
		// We should probably apply the same logic to show "pure" aliases if that's what the user expects.
		// "Crie um novo alias (redirecionador)" implies these are forwarders.
		// Let's filter out mailbox aliases to keep the list clean for "Aliases".

		query = query.Where("address NOT IN (?)", h.DB.Table("mailbox").Select("username"))

		query.Find(&aliases)
	}

	// Fetch domains for the filter dropdown
	var domains []models.Domain
	if h.DB != nil {
		domainQuery := h.DB.Where("domain != ?", "ALL").Where("active = ?", true).Order("domain ASC")
		if !isSuperAdmin {
			// Reuse allowedDomains from earlier
			allowedDomains, _, err := utils.GetAllowedDomains(h.DB, middleware.GetUsername(c, middleware.SessionName), middleware.GetIsSuperAdmin(c))
			if err == nil {
				if len(allowedDomains) == 0 {
					domainQuery = domainQuery.Where("1 = 0")
				} else {
					domainQuery = domainQuery.Where("domain IN ?", allowedDomains)
				}
			}
		}
		domainQuery.Find(&domains)
	}

	return c.Render(http.StatusOK, "aliases/aliases.html", map[string]interface{}{
		"Aliases":      aliases,
		"Domains":      domains,
		"DomainFilter": domainFilter,
		"IsSuperAdmin": isSuperAdmin,
		"SessionUser":  middleware.GetUsername(c, middleware.SessionName),
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
	allowedDomains, isSuperAdmin, err := utils.GetAllowedDomains(h.DB, loggedInUser, middleware.GetIsSuperAdmin(c))
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
	lines := strings.Split(gotoRaw, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		line = strings.Trim(line, ",")
		if line != "" {
			recipients = append(recipients, line)
		}
	}

	if len(recipients) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": "At least one valid recipient is required"})
	}
	gotoFinal := strings.Join(recipients, ",")

	var existingAlias models.Alias
	if err := h.DB.Where("address = ?", address).First(&existingAlias).Error; err == nil {
		return c.JSON(http.StatusConflict, map[string]interface{}{"success": false, "error": "Alias already exists"})
	}

	var existingMailbox models.Mailbox
	if err := h.DB.Where("username = ?", address).First(&existingMailbox).Error; err == nil {
		return c.JSON(http.StatusConflict, map[string]interface{}{"success": false, "error": "A mailbox already exists with this address"})
	}

	now := time.Now()
	newAlias := models.Alias{
		Address:  address,
		Goto:     gotoFinal,
		Domain:   domain,
		Created:  now,
		Modified: now,
		Active:   active,
	}

	if err := h.DB.Create(&newAlias).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": "Failed to create alias"})
	}

	if err := utils.LogAction(h.DB, loggedInUser, c.RealIP(), domain, "create_alias", address); err != nil {
		fmt.Printf("Failed to log create_alias: %v\n", err)
	}

	SetFlash(c, "message", "Alias created successfully")

	return c.JSON(http.StatusOK, map[string]interface{}{"success": true})
}

// GetAliasAPI fetches alias details via JSON
func (h *Handler) GetAliasAPI(c *echo.Context) error {
	address, _ := url.PathUnescape(c.Param("address"))

	var alias models.Alias
	if err := h.DB.Where("address = ?", address).First(&alias).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{"error": "Alias not found"})
	}

	loggedInUser := middleware.GetUsername(c, middleware.SessionName)
	allowedDomains, isSuperAdmin, err := utils.GetAllowedDomains(h.DB, loggedInUser, middleware.GetIsSuperAdmin(c))
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

	var alias models.Alias
	if err := h.DB.Where("address = ?", address).First(&alias).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{"success": false, "error": "Alias not found"})
	}

	loggedInUser := middleware.GetUsername(c, middleware.SessionName)
	allowedDomains, isSuperAdmin, err := utils.GetAllowedDomains(h.DB, loggedInUser, middleware.GetIsSuperAdmin(c))
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
	active := c.FormValue("active") == "true"

	if gotoRaw == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": "To (Recipients) is required"})
	}

	var recipients []string
	lines := strings.Split(gotoRaw, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		line = strings.Trim(line, ",")
		if line != "" {
			recipients = append(recipients, line)
		}
	}

	if len(recipients) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": "At least one valid recipient is required"})
	}

	gotoFinal := strings.Join(recipients, ",")

	alias.Goto = gotoFinal
	alias.Active = active
	alias.Modified = time.Now()

	if err := h.DB.Save(&alias).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": "Failed to update alias"})
	}

	if err := utils.LogAction(h.DB, loggedInUser, c.RealIP(), alias.Domain, "edit_alias", address); err != nil {
		fmt.Printf("Failed to log edit_alias: %v\n", err)
	}

	SetFlash(c, "message", "Alias updated successfully")

	return c.JSON(http.StatusOK, map[string]interface{}{"success": true})
}

// DeleteAlias handles alias deletion
func (h *Handler) DeleteAlias(c *echo.Context) error {
	address, _ := url.PathUnescape(c.Param("address"))

	// URL Decode the address to ensure we have the correct string
	decodedAddress, err := url.QueryUnescape(address)
	if err == nil {
		address = decodedAddress
	}

	if address == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Address required"})
	}

	// Security: Check permission (Pre-check)
	loggedInUser := middleware.GetUsername(c, middleware.SessionName)
	allowedDomains, isSuperAdmin, err := utils.GetAllowedDomains(h.DB, loggedInUser, middleware.GetIsSuperAdmin(c))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "Permission check failed"})
	}

	if h.DB == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{"error": "Database unavailable"})
	}

	// Fetch alias first to check domain
	var alias models.Alias
	if err := h.DB.Select("domain").Where("address = ?", address).First(&alias).Error; err != nil {
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

	// Prevent deleting mailbox aliases via this endpoint
	// Use count instead of First to avoid "record not found" error log
	var count int64
	h.DB.Model(&models.Mailbox{}).Where("username = ?", address).Count(&count)
	if count > 0 {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"error": "Cannot delete a mailbox alias via this interface. Delete the mailbox instead."})
	}

	// Delete
	result := h.DB.Where("address = ?", address).Delete(&models.Alias{})
	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "Failed to delete alias: " + result.Error.Error()})
	}

	if result.RowsAffected == 0 {
		return c.JSON(http.StatusNotFound, map[string]interface{}{"error": "Alias not found"})
	}

	// Log Action
	// For delete, we need the domain. We fetched 'alias' earlier which has the domain.
	if err := utils.LogAction(h.DB, middleware.GetUsername(c, middleware.SessionName), c.RealIP(), alias.Domain, "delete_alias", address); err != nil {
		// Just log error
		fmt.Printf("Failed to log delete_alias: %v\n", err)
	}

	SetFlash(c, "message", "Alias deleted successfully")

	return c.JSON(http.StatusOK, map[string]interface{}{"success": true})
}
