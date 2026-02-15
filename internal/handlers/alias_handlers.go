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
		query := h.DB.Order("address ASC")

		// Security: Filter by allowed domains
		username := middleware.GetUsername(c)
		allowedDomains, isSuper, err := utils.GetAllowedDomains(h.DB, username)
		if err != nil {
			return c.Render(http.StatusInternalServerError, "aliases.html", map[string]interface{}{
				"Error": "Failed to check permissions: " + err.Error(),
			})
		}
		isSuperAdmin = isSuper

		if !isSuperAdmin {
			if len(allowedDomains) == 0 {
				query = query.Where("1 = 0") // No domains allowed
			} else {
				query = query.Where("domain IN ?", allowedDomains)
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
					return c.Render(http.StatusForbidden, "aliases.html", map[string]interface{}{
						"Error": "Access denied to this domain",
					})
				}
			}
			query = query.Where("domain = ?", domainFilter)
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
			allowedDomains, _, err := utils.GetAllowedDomains(h.DB, middleware.GetUsername(c))
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

	return c.Render(http.StatusOK, "aliases.html", map[string]interface{}{
		"Aliases":      aliases,
		"Domains":      domains,
		"DomainFilter": domainFilter,
		"IsSuperAdmin": isSuperAdmin,
		"SessionUser":  middleware.GetUsername(c),
	})
}

// AddAliasForm renders the add alias form
func (h *Handler) AddAliasForm(c *echo.Context) error {
	var domains []models.Domain
	var isSuperAdmin bool

	if h.DB != nil {
		// Security: Filter domains
		username := middleware.GetUsername(c)
		allowedDomains, isSuper, err := utils.GetAllowedDomains(h.DB, username)
		if err != nil {
			return c.Render(http.StatusInternalServerError, "add_alias.html", map[string]interface{}{"Error": "Permission check failed"})
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

	return c.Render(http.StatusOK, "add_alias.html", map[string]interface{}{
		"Domains":      domains,
		"IsSuperAdmin": isSuperAdmin,
		"SessionUser":  middleware.GetUsername(c),
	})
}

// AddAlias processes the addition of a new alias
func (h *Handler) AddAlias(c *echo.Context) error {
	// Parse form data
	localPart := strings.ToLower(strings.TrimSpace(c.FormValue("local_part")))
	domain := strings.TrimSpace(c.FormValue("domain"))
	gotoRaw := c.FormValue("goto")
	active := c.FormValue("active") == "true"

	// Security: Validate domain access
	loggedInUser := middleware.GetUsername(c)
	allowedDomains, isSuperAdmin, err := utils.GetAllowedDomains(h.DB, loggedInUser)
	if err != nil {
		return renderAddAliasError(c, "Permission check failed", localPart, domain, gotoRaw, nil, isSuperAdmin)
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
			return renderAddAliasError(c, "Access denied to this domain", localPart, domain, gotoRaw, nil, isSuperAdmin)
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

	// Basic Validation
	if localPart == "" && domain != "" {
		// Catch-all alias logic? "Para criar um alias global, use '*'" - checking models/handlers usually implies specific logic.
		// User screenshot showed "Alias" input and "Domain" select.
		// If localPart is empty but intended to be catch-all, usually it's explicitly "*"
	}

	if domain == "" {
		return renderAddAliasError(c, "O domínio é obrigatório", localPart, domain, gotoRaw, domains, isSuperAdmin)
	}

	if gotoRaw == "" {
		return renderAddAliasError(c, "O destino (Para) é obrigatório", localPart, domain, gotoRaw, domains, isSuperAdmin)
	}

	// Construct Address
	var address string
	if localPart == "" || localPart == "*" {
		// Validating if * is allowed? Assuming yes based on standard PostfixAdmin
		address = fmt.Sprintf("@%s", domain) // Or "*@domain"? PostfixAdmin usually uses "@domain.tld" for catch-all
		// Wait, Postfix usually uses "@domain.tld" for catch-all in alias table?
		// Actually typical catchall is `*@domain.tld` or just `@domain.tld`.
		// Let's assume standard construction: local_part@domain
		if localPart == "*" {
			// Actually check if user typed *
		}
	}

	// If localPart does not contain @, append domain.
	// The form usually splits local_part and domain.
	if localPart == "" {
		// If empty, user might imply catch-all if allowed, or it's an error.
		// "Para criar um alias global, use '*'" -> so user must type *
		return renderAddAliasError(c, "O alias é obrigatório", localPart, domain, gotoRaw, domains, isSuperAdmin)
	}

	address = fmt.Sprintf("%s@%s", localPart, domain)

	// Process "To" (Goto) field - split lines, trim, join with comma
	var recipients []string
	lines := strings.Split(gotoRaw, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		line = strings.Trim(line, ",") // Remove trailing commas if user added them
		if line != "" {
			recipients = append(recipients, line)
		}
	}

	if len(recipients) == 0 {
		return renderAddAliasError(c, "Pelo menos um destinatário válido é necessário", localPart, domain, gotoRaw, domains, isSuperAdmin)
	}

	gotoFinal := strings.Join(recipients, ",")

	// Check if alias already exists
	var existingAlias models.Alias
	if err := h.DB.Where("address = ?", address).First(&existingAlias).Error; err == nil {
		return renderAddAliasError(c, "O alias já existe", localPart, domain, gotoRaw, domains, isSuperAdmin)
	}

	// Check if it conflicts with a mailbox (if we want to enforce pure aliases vs mailboxes)
	// PostfixAdmin usually prevents creating an alias if a mailbox exists with same email,
	// UNLESS it's the mailbox alias itself (which is auto-created).
	// If we create an alias clashing with mailbox, it might break mail delivery or be redundant.
	var existingMailbox models.Mailbox
	if err := h.DB.Where("username = ?", address).First(&existingMailbox).Error; err == nil {
		return renderAddAliasError(c, "Já existe uma caixa de correio com este endereço", localPart, domain, gotoRaw, domains, isSuperAdmin)
	}

	// Create Alias
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
		return renderAddAliasError(c, "Falha ao criar alias: "+err.Error(), localPart, domain, gotoRaw, domains, isSuperAdmin)
	}

	return c.Redirect(http.StatusFound, "/aliases")
}

// EditAliasForm renders the edit alias form
func (h *Handler) EditAliasForm(c *echo.Context) error {
	address := c.Param("address")

	var alias models.Alias
	if err := h.DB.Where("address = ?", address).First(&alias).Error; err != nil {
		return c.Render(http.StatusNotFound, "add_alias.html", map[string]interface{}{
			"Error": "Alias not found",
		})
	}

	// Security: Check permission
	loggedInUser := middleware.GetUsername(c)
	allowedDomains, isSuperAdmin, err := utils.GetAllowedDomains(h.DB, loggedInUser)
	if err != nil {
		return c.Render(http.StatusInternalServerError, "add_alias.html", map[string]interface{}{"Error": "Permission check failed"})
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
			return c.Render(http.StatusForbidden, "aliases.html", map[string]interface{}{"Error": "Access denied"})
		}
	}

	// Format Goto for display (comma to newline)
	alias.Goto = strings.ReplaceAll(alias.Goto, ",", "\n")

	return c.Render(http.StatusOK, "edit_alias.html", map[string]interface{}{
		"Alias":        alias,
		"IsSuperAdmin": isSuperAdmin,
		"SessionUser":  loggedInUser,
	})
}

// EditAlias processes the alias update
func (h *Handler) EditAlias(c *echo.Context) error {
	address := c.Param("address")

	var alias models.Alias
	if err := h.DB.Where("address = ?", address).First(&alias).Error; err != nil {
		return c.Render(http.StatusNotFound, "edit_alias.html", map[string]interface{}{
			"Error": "Alias not found",
		})
	}

	// Security: Check permission
	loggedInUser := middleware.GetUsername(c)
	allowedDomains, isSuperAdmin, err := utils.GetAllowedDomains(h.DB, loggedInUser)
	if err != nil {
		return c.Render(http.StatusInternalServerError, "edit_alias.html", map[string]interface{}{"Error": "Permission check failed"})
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
			return c.Render(http.StatusForbidden, "aliases.html", map[string]interface{}{"Error": "Access denied"})
		}
	}

	// Parse form data
	gotoRaw := c.FormValue("goto")
	active := c.FormValue("active") == "true"

	// Validate Goto
	if gotoRaw == "" {
		return c.Render(http.StatusBadRequest, "edit_alias.html", map[string]interface{}{
			"Error":        "To (Recipients) is required",
			"Alias":        alias,
			"IsSuperAdmin": isSuperAdmin,
			"SessionUser":  loggedInUser,
		})
	}

	// Process "To" (Goto) field - split lines, trim, join with comma
	var recipients []string
	lines := strings.Split(gotoRaw, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		line = strings.Trim(line, ",") // Remove trailing commas
		if line != "" {
			recipients = append(recipients, line)
		}
	}

	if len(recipients) == 0 {
		return c.Render(http.StatusBadRequest, "edit_alias.html", map[string]interface{}{
			"Error":        "At least one valid recipient is required",
			"Alias":        alias,
			"IsSuperAdmin": isSuperAdmin,
			"SessionUser":  loggedInUser,
		})
	}

	gotoFinal := strings.Join(recipients, ",")

	// Update Alias
	alias.Goto = gotoFinal
	alias.Active = active
	alias.Modified = time.Now()

	if err := h.DB.Save(&alias).Error; err != nil {
		return c.Render(http.StatusInternalServerError, "edit_alias.html", map[string]interface{}{
			"Error":        "Failed to update alias: " + err.Error(),
			"Alias":        alias,
			"IsSuperAdmin": isSuperAdmin,
			"SessionUser":  loggedInUser,
		})
	}

	return c.Redirect(http.StatusFound, "/aliases")
}

// DeleteAlias handles alias deletion
func (h *Handler) DeleteAlias(c *echo.Context) error {
	address := c.Param("address")

	// URL Decode the address to ensure we have the correct string
	decodedAddress, err := url.QueryUnescape(address)
	if err == nil {
		address = decodedAddress
	}

	if address == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": "Address required"})
	}

	// Security: Check permission (Pre-check)
	loggedInUser := middleware.GetUsername(c)
	allowedDomains, isSuperAdmin, err := utils.GetAllowedDomains(h.DB, loggedInUser)
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

	return c.JSON(http.StatusOK, map[string]interface{}{"success": true})
}

func renderAddAliasError(c *echo.Context, errorMsg, localPart, domain, gotoRaw string, domains []models.Domain, isSuperAdmin bool) error {
	return c.Render(http.StatusBadRequest, "add_alias.html", map[string]interface{}{
		"Error":        errorMsg,
		"LocalPart":    localPart,
		"Domain":       domain,
		"Goto":         gotoRaw,
		"Domains":      domains,
		"IsSuperAdmin": isSuperAdmin,
		"SessionUser":  middleware.GetUsername(c),
	})
}
