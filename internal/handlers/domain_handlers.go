package handlers

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go-postfixadmin/internal/middleware"
	"go-postfixadmin/internal/models"
	"go-postfixadmin/internal/utils"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

// DomainDisplay representa um domínio com contadores de aliases e mailboxes
type DomainDisplay struct {
	models.Domain
	AliasCount   int64
	MailboxCount int64
}

// ListDomains lista todos os domínios com contadores de aliases e mailboxes
func (h *Handler) ListDomains(c *echo.Context) error {
	var domains []models.Domain
	var displayDomains []DomainDisplay

	username := middleware.GetUsername(c)
	allowedDomains, isSuperAdmin, err := utils.GetAllowedDomains(h.DB, username, middleware.GetIsSuperAdmin(c))
	if err != nil {
		return c.Render(http.StatusInternalServerError, "domains.html", map[string]interface{}{
			"Error": "Failed to check permissions: " + err.Error(),
		})
	}

	if h.DB != nil {
		query := h.DB.Where("domain != ?", "ALL")
		if !isSuperAdmin {
			if len(allowedDomains) == 0 {
				query = query.Where("1 = 0") // No domains allowed
			} else {
				query = query.Where("domain IN ?", allowedDomains)
			}
		}
		query.Find(&domains)

		for _, d := range domains {
			var aliasCount int64
			var mailboxCount int64

			// Count aliases excluding those that are mailboxes
			h.DB.Model(&models.Alias{}).
				Where("domain = ?", d.Domain).
				Where("address NOT IN (?)", h.DB.Table("mailbox").Select("username")).
				Count(&aliasCount)

			h.DB.Model(&models.Mailbox{}).Where("domain = ?", d.Domain).Count(&mailboxCount)

			displayDomains = append(displayDomains, DomainDisplay{
				Domain:       d,
				AliasCount:   aliasCount,
				MailboxCount: mailboxCount,
			})
		}
	}
	return c.Render(http.StatusOK, "domains.html", map[string]interface{}{
		"Domains":      displayDomains,
		"IsSuperAdmin": isSuperAdmin,
		"SessionUser":  username,
	})
}

// AddDomainForm exibe o formulário de adicionar domínio
func (h *Handler) AddDomainForm(c *echo.Context) error {
	// Security: Only Superadmins can add domains
	username := middleware.GetUsername(c)
	isSuper, err := utils.IsSuperAdmin(h.DB, username)
	if err != nil {
		return c.Render(http.StatusInternalServerError, "domains.html", map[string]interface{}{"Error": "Permission check failed"})
	}
	if !isSuper {
		return c.Render(http.StatusForbidden, "domains.html", map[string]interface{}{"Error": "Access denied: Only Superadmins can create domains"})
	}
	return c.Render(http.StatusOK, "add_domain.html", map[string]interface{}{
		"SessionUser": username,
	})
}

// AddDomain processa a criação de um novo domínio
func (h *Handler) AddDomain(c *echo.Context) error {
	// Security: Only Superadmins can add domains
	username := middleware.GetUsername(c)
	isSuper, err := utils.IsSuperAdmin(h.DB, username)
	if err != nil || !isSuper {
		return c.Render(http.StatusForbidden, "domains.html", map[string]interface{}{"Error": "Access denied"})
	}

	// Parse form data
	domainName := strings.TrimSpace(c.FormValue("domain"))
	description := c.FormValue("description")
	active := c.FormValue("active") == "true"
	backupMX := c.FormValue("backupmx") == "true"

	// Parse numeric fields with defaults
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

	// Parse quota (in MB, store as MB or Bytes? DB schema says int64, usually postfixadmin uses MB)
	// standard postfixadmin uses MB in the UI and stores MB in the DB (quota field).
	// MaxQuota is usually Bytes? Let's check model. Domain struct has Quota int64.
	// Let's assume input is MB.
	quota := int64(2048) // Default
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

	// Validation: domain is required
	if domainName == "" {
		return c.Render(http.StatusBadRequest, "add_domain.html", map[string]interface{}{
			"Error":       "Domain name is required",
			"Domain":      domainName,
			"Description": description,
			"Active":      active,
			"BackupMX":    backupMX,
			"Aliases":     aliases,
			"Mailboxes":   mailboxes,
			"SessionUser": username,
		})
	}

	// Validation: basic DNS format (simplified)
	domainRegex := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{0,61}[a-zA-Z0-9]?(\.[a-zA-Z0-9][a-zA-Z0-9-]{0,61}[a-zA-Z0-9]?)+$`)
	if !domainRegex.MatchString(domainName) {
		return c.Render(http.StatusBadRequest, "add_domain.html", map[string]interface{}{
			"Error":       "Invalid domain format. Please enter a valid domain name (e.g., example.com)",
			"Domain":      domainName,
			"Description": description,
			"Active":      active,
			"BackupMX":    backupMX,
			"Aliases":     aliases,
			"Mailboxes":   mailboxes,
			"SessionUser": username,
		})
	}

	// Check if domain already exists
	var existingDomain models.Domain
	result := h.DB.Where("domain = ?", domainName).First(&existingDomain)
	if result.Error == nil {
		return c.Render(http.StatusBadRequest, "add_domain.html", map[string]interface{}{
			"Error":       "Domain already exists",
			"Domain":      domainName,
			"Description": description,
			"Active":      active,
			"BackupMX":    backupMX,
			"Aliases":     aliases,
			"Mailboxes":   mailboxes,
			"SessionUser": username,
		})
	}

	// Create new domain
	now := time.Now()
	newDomain := models.Domain{
		Domain:         domainName,
		Description:    description,
		Aliases:        aliases,
		Mailboxes:      mailboxes,
		MaxQuota:       0,     // Not used in this form yet? Or is Quota the max quota for the domain?
		Quota:          quota, // Domain quota
		Transport:      "",
		BackupMX:       backupMX,
		Created:        now,
		Modified:       now,
		Active:         active,
		PasswordExpiry: passwordExpiry,
	}

	if err := h.DB.Create(&newDomain).Error; err != nil {
		return c.Render(http.StatusInternalServerError, "add_domain.html", map[string]interface{}{
			"Error":       "Failed to create domain: " + err.Error(),
			"Domain":      domainName,
			"Description": description,
			"Active":      active,
			"BackupMX":    backupMX,
			"Aliases":     aliases,
			"Mailboxes":   mailboxes,
			"SessionUser": username,
		})
	}

	// Log Action
	if err := utils.LogAction(h.DB, username, c.RealIP(), domainName, "create_domain", domainName); err != nil {
		fmt.Printf("Failed to log create_domain: %v\n", err)
	}

	// Redirect to domains list on success
	return c.Redirect(http.StatusFound, "/domains")
}

// EditDomainForm exibe o formulário de edição de domínio
func (h *Handler) EditDomainForm(c *echo.Context) error {
	// Security: Only Superadmins can edit domains
	username := middleware.GetUsername(c)
	isSuper, err := utils.IsSuperAdmin(h.DB, username)
	if err != nil || !isSuper {
		return c.Render(http.StatusForbidden, "domains.html", map[string]interface{}{"Error": "Access denied: Only Superadmins can edit domains"})
	}

	domainName := c.Param("domain")

	var domain models.Domain
	if err := h.DB.Where("domain = ?", domainName).First(&domain).Error; err != nil {
		return c.Render(http.StatusNotFound, "add_domain.html", map[string]interface{}{
			"Error": "Domain not found",
		})
	}

	return c.Render(http.StatusOK, "edit_domain.html", map[string]interface{}{
		"Domain":      domain,
		"SessionUser": username,
	})
}

// EditDomain processa a edição de um domínio existente
func (h *Handler) EditDomain(c *echo.Context) error {
	// Security: Only Superadmins can edit domains
	username := middleware.GetUsername(c)
	isSuper, err := utils.IsSuperAdmin(h.DB, username)
	if err != nil || !isSuper {
		return c.Render(http.StatusForbidden, "domains.html", map[string]interface{}{"Error": "Access denied"})
	}

	domainName := c.Param("domain")

	// Find existing domain
	var domain models.Domain
	if err := h.DB.Where("domain = ?", domainName).First(&domain).Error; err != nil {
		return c.Render(http.StatusNotFound, "edit_domain.html", map[string]interface{}{
			"Error": "Domain not found",
		})
	}

	// Parse form data
	description := c.FormValue("description")
	active := c.FormValue("active") == "true"
	backupMX := c.FormValue("backupmx") == "true"

	// ...
	// Parse numeric fields
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

	var quota int64
	// Default to existing quota if not provided? Or parse from form?
	if val := c.FormValue("quota"); val != "" {
		if parsed, err := strconv.ParseInt(val, 10, 64); err == nil {
			quota = parsed
		}
	} else {
		quota = domain.Quota
	}

	var passwordExpiry *int
	if val := c.FormValue("password_expiry"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			passwordExpiry = &parsed
		}
	}

	// Check if active state is changing
	activeChanged := domain.Active != active

	// Update domain fields
	domain.Description = description
	domain.Aliases = aliases
	domain.Mailboxes = mailboxes
	domain.Quota = quota
	domain.BackupMX = backupMX
	domain.Modified = time.Now()
	domain.Active = active
	domain.PasswordExpiry = passwordExpiry

	// Use transaction to ensure atomicity (especially for cascading updates)
	err = h.DB.Transaction(func(tx *gorm.DB) error {
		if activeChanged {
			// Update all mailboxes for this domain to match the new domain active state
			if err := tx.Model(&models.Mailbox{}).Where("domain = ?", domain.Domain).Update("active", active).Error; err != nil {
				return err
			}
			// Update all aliases for this domain to match the new domain active state
			if err := tx.Model(&models.Alias{}).Where("domain = ?", domain.Domain).Update("active", active).Error; err != nil {
				return err
			}

		}

		if err := tx.Save(&domain).Error; err != nil {
			return err
		}

		// Log Action
		if err := utils.LogAction(tx, username, c.RealIP(), domainName, "edit_domain", domainName); err != nil {
			fmt.Printf("Failed to log edit_domain: %v\n", err)
			return nil
		}
		return nil
	})

	if err != nil {
		return c.Render(http.StatusInternalServerError, "edit_domain.html", map[string]interface{}{
			"Error":       "Failed to update domain: " + err.Error(),
			"Domain":      domain,
			"SessionUser": username,
		})
	}

	// Redirect to domains list on success
	return c.Redirect(http.StatusFound, "/domains")
}

// DeleteDomain remove um domínio e todos os dados associados (aliases e mailboxes)
func (h *Handler) DeleteDomain(c *echo.Context) error {
	// Security: Only Superadmins can delete domains
	username := middleware.GetUsername(c)
	isSuper, err := utils.IsSuperAdmin(h.DB, username)
	if err != nil || !isSuper {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"error": "Access denied: Only Superadmins can delete domains"})
	}

	domainName := c.Param("domain")

	// Check if domain exists
	var domain models.Domain
	if err := h.DB.Where("domain = ?", domainName).First(&domain).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"success": false,
			"error":   "Domain not found",
		})
	}

	// Use utility function to delete domain and all associated data
	if err := utils.DeleteDomain(h.DB, domainName, username, c.RealIP()); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "Failed to delete domain: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Domain deleted successfully",
	})
}
